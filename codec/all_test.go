package codec

import (
	"github.com/stretchr/testify/assert"
	"github.com/zehuamama/tinyrpc/compressor"
	pb "github.com/zehuamama/tinyrpc/example/message"
	"log"
	"net"
	"net/rpc"
	"testing"
)

var server *rpc.Server

func init() {
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}

	server := rpc.NewServer()
	server.RegisterName("ArithService", new(pb.ArithService))
	go serve(lis, server)
}

func serve(lis net.Listener, server *rpc.Server) {
	for {
		conn, _ := lis.Accept()
		go server.ServeCodec(NewServerCodec(conn))
	}
}

// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := rpc.NewClientWithCodec(NewClientCodec(conn, compressor.Raw))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *rpc.Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client,
			"test-1",
			"ArithService.Add",
			&pb.ArithRequest{A: 20, B: 5},
			expect{
				&pb.ArithResponse{25},
				nil,
			},
		},
		{
			client,
			"test-2",
			"ArithService.Sub",
			&pb.ArithRequest{A: 20, B: 5},
			expect{
				&pb.ArithResponse{15},
				nil,
			},
		},
		{
			client,
			"test-3",
			"ArithService.Mul",
			&pb.ArithRequest{A: 20, B: 5},
			expect{
				&pb.ArithResponse{100},
				nil,
			},
		},
		{
			client,
			"test-4",
			"ArithService.Div",
			&pb.ArithRequest{A: 20, B: 5},
			expect{
				&pb.ArithResponse{4},
				nil,
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&pb.ArithRequest{A: 20, B: 0},
			expect{
				&pb.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, c.expect.reply, reply)
			assert.Equal(t, c.expect.err, err)
		})
	}
}

// TestNewClientWithSnappyCompress test snappy comressor
func TestNewClientWithSnappyCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := rpc.NewClientWithCodec(NewClientCodec(conn, compressor.Snappy))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *rpc.Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client,
			"test-1",
			"ArithService.Add",
			&pb.ArithRequest{A: 20, B: 5},
			expect{
				&pb.ArithResponse{25},
				nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, c.expect.reply, reply)
			assert.Equal(t, c.expect.err, err)
		})
	}
}
