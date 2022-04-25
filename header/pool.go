// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package header

import "sync"

var (
	RequestPool  sync.Pool
	ResponsePool sync.Pool
)

func init() {
	RequestPool = sync.Pool{New: func() interface{} {
		return &RequestHeader{}
	}}
	ResponsePool = sync.Pool{New: func() interface{} {
		return &ResponseHeader{}
	}}
}

// ResetHeader reset request header
func (h *RequestHeader) ResetHeader() {
	h.Id = 0
	h.Checksum = 0
	h.Method = ""
	h.CompressType = 0
	h.RequestLen = 0
}

// ResetHeader reset response header
func (h *ResponseHeader) ResetHeader() {
	h.Error = ""
	h.Id = 0
	h.CompressType = 0
	h.Checksum = 0
	h.ResponseLen = 0
}
