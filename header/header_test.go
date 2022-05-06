// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package header

import (
	"github.com/zehuamama/tinyrpc/compressor"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequestHeader_Marshal .
func TestRequestHeader_Marshal(t *testing.T) {
	header := &RequestHeader{
		CompressType: 0,
		Method:       "Add",
		ID:           12455,
		RequestLen:   266,
		Checksum:     3845236589,
	}

	assert.Equal(t, []byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
		0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5}, header.Marshal())
}

// TestRequestHeader_Unmarshal .
func TestRequestHeader_Unmarshal(t *testing.T) {
	type expect struct {
		header *RequestHeader
		err    error
	}
	cases := []struct {
		name   string
		data   []byte
		expect expect
	}{
		{
			"test-1",
			[]byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
				0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
			expect{&RequestHeader{
				CompressType: 0,
				Method:       "Add",
				ID:           12455,
				RequestLen:   266,
				Checksum:     3845236589,
			}, nil},
		},
		{
			"test-2",
			nil,
			expect{&RequestHeader{},
				UnmarshalError},
		},
		{
			"test-3",
			[]byte{0x0},
			expect{&RequestHeader{},
				UnmarshalError},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := &RequestHeader{}
			err := h.Unmarshal(c.data)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.header, h))
			assert.Equal(t, err, c.expect.err)
		})
	}
}

// TestRequestHeader_ResetHeader .
func TestRequestHeader_ResetHeader(t *testing.T) {
	header := &RequestHeader{
		CompressType: 0,
		Method:       "Add",
		ID:           12455,
		RequestLen:   266,
		Checksum:     3845236589,
	}
	header.ResetHeader()
	assert.Equal(t, true, reflect.DeepEqual(header, &RequestHeader{}))
}

// TestResponseHeader_Marshal .
func TestResponseHeader_Marshal(t *testing.T) {
	header := &ResponseHeader{
		CompressType: 0,
		Error:        "error",
		ID:           12455,
		ResponseLen:  266,
		Checksum:     3845236589,
	}

	assert.Equal(t, []byte{0x0, 0x0, 0xa7, 0x61, 0x5, 0x65, 0x72,
		0x72, 0x6f, 0x72, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5}, header.Marshal())
}

// TestResponseHeader_Unmarshal .
func TestResponseHeader_Unmarshal(t *testing.T) {
	type expect struct {
		header *ResponseHeader
		err    error
	}
	cases := []struct {
		name   string
		data   []byte
		expect expect
	}{
		{
			"test-1",
			[]byte{0x0, 0x0, 0xa7, 0x61, 0x5, 0x65, 0x72,
				0x72, 0x6f, 0x72, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
			expect{&ResponseHeader{
				CompressType: 0,
				Error:        "error",
				ID:           12455,
				ResponseLen:  266,
				Checksum:     3845236589,
			}, nil},
		},
		{
			"test-2",
			nil,
			expect{&ResponseHeader{},
				UnmarshalError},
		},
		{
			"test-3",
			[]byte{0x0},
			expect{&ResponseHeader{},
				UnmarshalError},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := &ResponseHeader{}
			err := h.Unmarshal(c.data)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.header, h))
			assert.Equal(t, err, c.expect.err)
		})
	}
}

// TestResponseHeader_ResetHeader .
func TestResponseHeader_ResetHeader(t *testing.T) {
	header := &ResponseHeader{
		CompressType: 0,
		Error:        "error",
		ID:           12455,
		ResponseLen:  266,
		Checksum:     3845236589,
	}
	header.ResetHeader()
	assert.Equal(t, true, reflect.DeepEqual(header, &ResponseHeader{}))
}

// TestResponseHeader_GetCompressType .
func TestResponseHeader_GetCompressType(t *testing.T) {
	header := &ResponseHeader{
		CompressType: 0,
		Error:        "error",
		ID:           12455,
		ResponseLen:  266,
		Checksum:     3845236589,
	}

	assert.Equal(t, true, reflect.DeepEqual(compressor.CompressType(0), header.GetCompressType()))
}

// TestRequestHeader_GetCompressType .
func TestRequestHeader_GetCompressType(t *testing.T) {
	header := &RequestHeader{
		CompressType: 0,
		Method:       "Add",
		ID:           12455,
		RequestLen:   266,
		Checksum:     3845236589,
	}

	assert.Equal(t, true, reflect.DeepEqual(compressor.CompressType(0), header.GetCompressType()))
}
