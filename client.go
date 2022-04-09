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

// NewClient Create a new tinyrpc client
func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{rpc.NewClientWithCodec(codec.NewClientCodec(conn, compressor.Raw))}
}

// NewClientWithCompress Create a new tinyrpc client (based on compression encoding)
func NewClientWithCompress(conn io.ReadWriteCloser, compressType compressor.CompressType) *Client {
	return &Client{rpc.NewClientWithCodec(codec.NewClientCodec(conn, compressType))}
}

// Call synchronously call tinyrpc function
func (c *Client) Call(serviceMethod string, args any, reply any) error {
	return c.Client.Call(serviceMethod, args, reply)
}

// AsyncCall asynchronously calls the tinyrpc method and returns a channel of *rpc.Call
func (c *Client) AsyncCall(serviceMethod string, args any, reply any) chan *rpc.Call {
	return c.Go(serviceMethod, args, reply, nil).Done
}
