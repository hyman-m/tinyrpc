package tinyrpc

import (
	"log"
	"net"
	"net/rpc"

	"github.com/zehuamama/tinyrpc/codec"
)

type Server struct {
	*rpc.Server
}

func NewServer() *Server {
	return &Server{&rpc.Server{}}
}

func (s *Server) Register(rcvr ...any) error {
	for r := range rcvr {
		err := s.Server.Register(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) RegisterName(name string, rcvr any) error {
	return s.Server.RegisterName(name, rcvr)
}

func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Print("tinyrpc.Serve: accept:", err.Error())
			return
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn))
	}
}
