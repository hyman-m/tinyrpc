// Copyright 2021 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	"bufio"
	"hash/crc32"
	"io"
	"sync"
	"time"

	"github.com/cloudmzh/tinyrpc"
	"github.com/cloudmzh/tinyrpc/header"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
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

// NewServerCodec ...
func NewServerCodec(conn io.ReadWriteCloser) tinyrpc.ServerCodec {
	return &serverCodec{
		r:       bufio.NewReader(conn),
		w:       bufio.NewWriter(conn),
		c:       conn,
		pending: make(map[uint64]uint64),
	}
}

// ReadRequestHeader ...
func (s *serverCodec) ReadRequestHeader(r *tinyrpc.Request) error {
	h := header.RequestHeader{}
	err := readRequestHeader(s.r, &h)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	s.seq++
	s.pending[s.seq] = h.Id
	r.ServiceMethod = h.Method
	r.Seq = s.seq
	r.TTL = time.UnixMilli(int64(h.Ttl))
	s.mutex.Unlock()
	s.request = h
	return nil
}

// readRequestHeader ...
func readRequestHeader(r io.Reader, h *header.RequestHeader) (err error) {
	pbHeader, err := recvFrame(r, int(header.Const_MAX_HEADER_LEN))
	if err != nil {
		return err
	}
	err = proto.Unmarshal(pbHeader, h)
	if err != nil {
		return err
	}
	return nil
}

func (s *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		if s.request.RequestLen != 0 {
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
		}
		return nil
	}

	request, ok := x.(proto.Message)
	if !ok {
		return NotImplementProtoMessageError
	}

	err := readRequestBody(s.r, &s.request, request)
	if err != nil {
		return nil
	}
	s.request = header.RequestHeader{}
	return nil
}

// readRequestBody ...
func readRequestBody(r io.Reader, h *header.RequestHeader, request proto.Message) error {
	requestLen := make([]byte, h.RequestLen)

	err := read(r, requestLen)
	if err != nil {
		return err
	}

	if h.Checksum != 0 {
		if crc32.ChecksumIEEE(requestLen) != h.Checksum {
			return UnexpectedChecksumError
		}
	}

	var pbRequest []byte

	if h.IsCompressed {
		pbRequest, err = snappy.Decode(nil, requestLen)
		if err != nil {
			return err
		}
	} else {
		pbRequest = requestLen
	}

	if request != nil {
		err = proto.Unmarshal(pbRequest, request)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *serverCodec) WriteResponse(r *tinyrpc.Response, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		if response, ok = x.(proto.Message); !ok {
			if _, ok = x.(struct{}); !ok {
				s.mutex.Lock()
				delete(s.pending, r.Seq)
				s.mutex.Unlock()
				return NotImplementProtoMessageError
			}
		}
	}

	s.mutex.Lock()
	id, ok := s.pending[r.Seq]
	if !ok {
		s.mutex.Unlock()
		return InvalidSequenceError
	}
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	err := writeResponse(s.w, id, r.Error, s.request.IsCompressed, response)
	if err != nil {
		return err
	}

	return nil
}

// writeResponse ...
func writeResponse(w io.Writer, id uint64, serr string, isCompressed bool, response proto.Message) (err error) {
	if serr != "" {
		response = nil
	}
	var pbResponse []byte
	if response != nil {
		pbResponse, err = proto.Marshal(response)
		if err != nil {
			return err
		}
	}

	var compressedPbResponse []byte
	if isCompressed {
		compressedPbResponse = snappy.Encode(nil, pbResponse)
	} else {
		compressedPbResponse = pbResponse
	}

	h := &header.ResponseHeader{
		Id:           id,
		Error:        serr,
		ResponseLen:  uint32(len(compressedPbResponse)),
		Checksum:     crc32.ChecksumIEEE(compressedPbResponse),
		IsCompressed: isCompressed,
	}

	pbHeader, err := proto.Marshal(h)
	if err != err {
		return
	}

	if err = sendFrame(w, pbHeader); err != nil {
		return
	}

	if err = write(w, compressedPbResponse); err != nil {
		return
	}
	w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}
