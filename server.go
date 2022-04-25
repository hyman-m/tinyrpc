// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tinyrpc

import (
	"net"
	"net/rpc"

	"github.com/zehuamama/tinyrpc/codec"
)

// Server tinyrpc server based on net/rpc implementation
type Server struct {
	*rpc.Server
}

// NewServer Create a new tinyrpc server
func NewServer() *Server {
	return &Server{&rpc.Server{}}
}

// Register register rpc function
func (s *Server) Register(rcvr ...interface{}) error {
	for r := range rcvr {
		err := s.Server.Register(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterName register the rpc function with the specified name
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.Server.RegisterName(name, rcvr)
}

// Serve start service
func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn))
	}
}
