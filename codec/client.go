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
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.CompressType // rpc compress type(raw,gzip,snappy,zlib)
	response   header.ResponseHeader   // rpc response header
	mutex      sync.Mutex              // protect pending map
	pending    map[uint64]string
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser,
	compressType compressor.CompressType) rpc.ClientCodec {

	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressType,
		pending:    make(map[uint64]string),
	}
}

// WriteRequest Write the rpc request header and body to the io stream
func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()
	err := writeRequest(c.w, r, c.compressor, param)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	c.response.ResetHeader()
	err := readResponseHeader(c.r, &c.response)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	r.Seq = c.response.Id
	r.Error = c.response.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()
	return nil
}

func (c *clientCodec) ReadResponseBody(x any) error {
	if x == nil {
		if c.response.ResponseLen != 0 {
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
		}
		return nil
	}

	err := readResponseBody(c.r, &c.response, x)
	if err != nil {
		return nil
	}
	return nil
}

// readResponseHeader ...
func readResponseHeader(r io.Reader, h *header.ResponseHeader) error {
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

// writeRequest Marshal the request body into a pb byte slice, compress it with a compressor,
// marshal the request header into a pb byte slice, and then send the request header and body to the io stream
func writeRequest(w io.Writer, r *rpc.Request,
	compressType compressor.CompressType, param any) error {
	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return errs.NotImplementProtoMessageError
		}
	}

	var pbRequest []byte
	if request != nil {
		var err error
		pbRequest, err = proto.Marshal(request)
		if err != nil {
			return err
		}
	}

	var compressedPbRequest []byte
	compressedPbRequest, err := compressor.Compressors[compressType].Zip(pbRequest)
	if err != nil {
		return err
	}

	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()
	h.Id = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(compressedPbRequest))
	h.CompressType = header.Compress(compressType)
	h.Checksum = crc32.ChecksumIEEE(compressedPbRequest)

	pbHeader, err := proto.Marshal(h)
	if err != err {
		return err
	}

	if err := sendFrame(w, pbHeader); err != nil {
		return err
	}

	if err := write(w, compressedPbRequest); err != nil {
		return err
	}

	w.(*bufio.Writer).Flush()
	return nil
}

func readResponseBody(r io.Reader, h *header.ResponseHeader, x any) error {
	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return errs.NotImplementProtoMessageError
		}
	}

	pbResponse := make([]byte, h.ResponseLen)
	err := read(r, pbResponse)
	if err != nil {
		return err
	}

	// checksum
	if h.Checksum != 0 {
		if crc32.ChecksumIEEE(pbResponse) != h.Checksum {
			return errs.UnexpectedChecksumError
		}
	}

	var resp []byte
	resp, err = compressor.Compressors[compressor.CompressType(h.CompressType)].Unzip(pbResponse)
	if err != nil {
		return err
	}

	if response != nil {
		err = proto.Unmarshal(resp, response)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
