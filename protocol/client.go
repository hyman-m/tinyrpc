package protocol

import (
	"bufio"
	"errors"
	"hash/crc32"
	"io"
	"sync"

	"github.com/cloudmzh/tinyrpc"
	"github.com/cloudmzh/tinyrpc/header"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compress bool
	response header.ResponseHeader
	mutex    sync.Mutex // protects pending
	pending  map[uint64]string
}

func NewClientCodec(conn io.ReadWriteCloser, compress bool) tinyrpc.ClientCodec {
	return &clientCodec{
		r:        bufio.NewReader(conn),
		w:        bufio.NewWriter(conn),
		c:        conn,
		compress: compress,
		pending:  make(map[uint64]string),
	}
}

func (c *clientCodec) WriteRequest(r *tinyrpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return errors.New("param does not implement proto.Message")
		}
	}
	err := writeRequest(c.w, r, c.compress, request)
	if err != nil {
		return err
	}
	return nil
}

func writeRequest(w io.Writer, r *tinyrpc.Request, isCompressed bool, request proto.Message) error {
	var pbRequest []byte
	if request != nil {
		var err error
		pbRequest, err = proto.Marshal(request)
		if err != nil {
			return err
		}
	}

	var compressedPbRequest []byte
	if isCompressed {
		compressedPbRequest = snappy.Encode(nil, pbRequest)
	} else {
		compressedPbRequest = pbRequest
	}

	h := &header.RequestHeader{
		Id:           r.Seq,
		Method:       r.ServiceMethod,
		RequestLen:   uint32(len(compressedPbRequest)),
		IsCompressed: isCompressed,
		Checksum:     crc32.ChecksumIEEE(compressedPbRequest),
		Ttl:          uint64(r.TTL.UnixMilli()),
	}

	pbHeader, err := proto.Marshal(h)
	if err != err {
		return err
	}
	if len(pbHeader) > int(header.Const_MAX_HEADER_LEN) {
		return errors.New("header exceeds the maximum limit length")
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

func (c *clientCodec) ReadResponseHeader(r *tinyrpc.Response) error {
	h := header.ResponseHeader{}
	err := readResponseHeader(c.r, &h)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	r.Seq = h.Id
	r.Error = h.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	c.response = h
	return nil
}

// readResponseHeader ...
func readResponseHeader(r io.Reader, h *header.ResponseHeader) error {
	pbHeader, err := recvFrame(r, 0)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(pbHeader, h)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		if c.response.ResponseLen != 0 {
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
		}
		return nil
	}

	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return errors.New("header exceeds the maximum limit length")
		}
	}

	err := readResponseBody(c.r, &c.response, response)
	if err != nil {
		return nil
	}

	c.response = header.ResponseHeader{}
	return nil
}

func readResponseBody(r io.Reader, h *header.ResponseHeader, response proto.Message) error {
	pbResponse := make([]byte, h.ResponseLen)

	err := read(r, pbResponse)
	if err != nil {
		return err
	}

	// checksum
	if h.Checksum != 0 {
		if crc32.ChecksumIEEE(pbResponse) != h.Checksum {
			return errors.New("unexpected checksum")
		}
	}

	var resp []byte
	if h.IsCompressed {
		resp, err = snappy.Decode(nil, pbResponse)
		if err != nil {
			return err
		}
	} else {
		resp = pbResponse
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
