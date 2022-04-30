// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codec

import (
	"bufio"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/zehuamama/tinyrpc/compressor"
	errs "github.com/zehuamama/tinyrpc/errors"
	"github.com/zehuamama/tinyrpc/header"
	"github.com/zehuamama/tinyrpc/serializer"
)

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request header.RequestHeader

	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]uint64
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		r:       bufio.NewReader(conn),
		w:       bufio.NewWriter(conn),
		c:       conn,
		pending: make(map[uint64]uint64),
	}
}

// ReadRequestHeader read the rpc request header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	s.request.ResetHeader()
	err := readRequestHeader(s.r, &s.request)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	s.seq++
	s.pending[s.seq] = s.request.Id
	r.ServiceMethod = s.request.Method
	r.Seq = s.seq
	s.mutex.Unlock()
	return nil
}

// ReadRequestBody read the rpc request body from the io stream
func (s *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		if s.request.RequestLen != 0 {
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
		}
		return nil
	}

	err := readRequestBody(s.r, &s.request, x)
	if err != nil {
		return nil
	}
	return nil
}

// WriteResponse Write the rpc response header and body to the io stream
func (s *serverCodec) WriteResponse(r *rpc.Response, param interface{}) error {
	s.mutex.Lock()
	id, ok := s.pending[r.Seq]
	if !ok {
		s.mutex.Unlock()
		return errs.InvalidSequenceError
	}
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	err := writeResponse(s.w, id, r.Error, compressor.CompressType(s.request.CompressType), param)
	if err != nil {
		return err
	}

	return nil
}

func readRequestHeader(r io.Reader, h *header.RequestHeader) error {
	pbHeader, err := recvFrame(r)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(pbHeader, h)
	if err != nil {
		return err
	}
	return nil
}

func readRequestBody(r io.Reader, h *header.RequestHeader, param interface{}) error {
	reqBody := make([]byte, h.RequestLen)

	err := read(r, reqBody)
	if err != nil {
		return err
	}

	if h.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != h.Checksum {
			return errs.UnexpectedChecksumError
		}
	}

	if _, ok := compressor.Compressors[compressor.CompressType(h.CompressType)]; !ok {
		return errs.NotFoundCompressorError
	}

	req, err := compressor.Compressors[compressor.CompressType(h.CompressType)].Unzip(reqBody)
	if err != nil {
		return err
	}

	return serializer.Serializers[serializer.Proto].Unmarshal(req, param)
}

func writeResponse(w io.Writer, id uint64, serr string,
	compressType compressor.CompressType, param interface{}) (err error) {
	if serr != "" {
		param = nil
	}
	if _, ok := compressor.Compressors[compressType]; !ok {
		return errs.NotFoundCompressorError
	}

	var respBody []byte
	if param != nil {
		respBody, err = serializer.Serializers[serializer.Proto].Marshal(param)
		if err != nil {
			return err
		}
	}

	compressedRespBody, err := compressor.Compressors[compressType].Zip(respBody)
	if err != nil {
		return err
	}
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.ResetHeader()
		header.ResponsePool.Put(h)
	}()
	h.Id = id
	h.Error = serr
	h.ResponseLen = uint32(len(compressedRespBody))
	h.Checksum = crc32.ChecksumIEEE(compressedRespBody)
	h.CompressType = header.Compress(compressType)

	pbHeader, err := proto.Marshal(h)
	if err != err {
		return
	}

	if err = sendFrame(w, pbHeader); err != nil {
		return
	}

	if err = write(w, compressedRespBody); err != nil {
		return
	}
	w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}
