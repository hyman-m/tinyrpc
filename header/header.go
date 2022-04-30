package header

import (
	"encoding/binary"
)

const (
	MaxHeaderSize = 64
	Uint32Size    = 4
	Uint16Size    = 2
)

type Compress uint16

type RequestHeader struct {
	CompressType Compress
	Method       string
	ID           uint64
	RequestLen   uint32
	Checksum     uint32
}

func (r *RequestHeader) Marshal() []byte {
	header := make([]byte, MaxHeaderSize+len(r.Method)) // prevent panic
	idx := 0
	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size
	idx += binary.PutUvarint(header[idx:], uint64(len(r.Method)))
	copy(header[idx:], r.Method)
	idx += len(r.Method)
	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += binary.PutUvarint(header[idx:], uint64(r.RequestLen))
	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

func (r *RequestHeader) Unmarshal(data []byte) {
	idx, size := 0, 0
	r.CompressType = Compress(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size
	length, size := binary.Uvarint(data[idx:])
	idx += size
	r.Method = string(data[idx : idx+int(length)])
	idx += len(r.Method)
	r.ID, size = binary.Uvarint(data[idx:])
	idx += size
	length, size = binary.Uvarint(data[idx:])
	r.RequestLen = uint32(length)
	idx += size
	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	idx += Uint32Size
}

// ResetHeader reset request header
func (r *RequestHeader) ResetHeader() {
	r.ID = 0
	r.Checksum = 0
	r.Method = ""
	r.CompressType = 0
	r.RequestLen = 0
}

type ResponseHeader struct {
	CompressType Compress
	ID           uint64
	Error        string
	ResponseLen  uint32
	Checksum     uint32
}

func (r *ResponseHeader) Marshal() []byte {
	header := make([]byte, MaxHeaderSize+len(r.Error)) // prevent panic
	idx := 0
	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size
	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += binary.PutUvarint(header[idx:], uint64(len(r.Error)))
	copy(header[idx:], r.Error)
	idx += len(r.Error)
	idx += binary.PutUvarint(header[idx:], uint64(r.ResponseLen))
	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

func (r *ResponseHeader) Unmarshal(data []byte) {
	idx, size := 0, 0
	r.CompressType = Compress(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size
	r.ID, size = binary.Uvarint(data[idx:])
	idx += size
	length, size := binary.Uvarint(data[idx:])
	idx += size
	r.Error = string(data[idx : idx+int(length)])
	idx += len(r.Error)
	length, size = binary.Uvarint(data[idx:])
	r.ResponseLen = uint32(length)
	idx += size
	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	idx += Uint32Size
}

// ResetHeader reset response header
func (r *ResponseHeader) ResetHeader() {
	r.Error = ""
	r.ID = 0
	r.CompressType = 0
	r.Checksum = 0
	r.ResponseLen = 0
}
