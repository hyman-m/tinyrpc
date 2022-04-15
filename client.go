// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tinyrpc

import (
	"io"
	"net/rpc"

	"github.com/zehuamama/tinyrpc/codec"
	"github.com/zehuamama/tinyrpc/compressor"
)

// Client tinyrpc client based on net/rpc implementation
type Client struct {
	*rpc.Client
}

//Option provides options for tinyrpc client
type Option interface {
	apply(*options)
}

type options struct {
	compressType compressor.CompressType
}

type compressOption compressor.CompressType

func (c compressOption) apply(opt *options) {
	opt.compressType = compressor.CompressType(c)
}

// WithCompress set client compression format
func WithCompress(c compressor.CompressType) Option {
	return compressOption(c)
}

// NewClient Create a new tinyrpc client
func NewClient(conn io.ReadWriteCloser, opts ...Option) *Client {
	options := options{
		compressType: compressor.Raw,
	}
	for _, o := range opts {
		o.apply(&options)
	}
	return &Client{rpc.NewClientWithCodec(
		codec.NewClientCodec(conn, options.compressType))}
}

// Call synchronously calls the tinyrpc function
func (c *Client) Call(serviceMethod string, args any, reply any) error {
	return c.Client.Call(serviceMethod, args, reply)
}

// AsyncCall asynchronously calls the tinyrpc function and returns a channel of *rpc.Call
func (c *Client) AsyncCall(serviceMethod string, args any, reply any) chan *rpc.Call {
	return c.Go(serviceMethod, args, reply, nil).Done
}
