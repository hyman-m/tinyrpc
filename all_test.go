package tinyrpc

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/rpc"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zehuamama/tinyrpc/compressor"
	js "github.com/zehuamama/tinyrpc/test.data/json"
	pb "github.com/zehuamama/tinyrpc/test.data/message"
)

var server Server

func init() {
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()
	server.Register(new(pb.ArithService))
	go server.Serve(lis)

	lis, err = net.Listen("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}

	server = NewServer(WithSerializer(&Json{}))
	server.Register(new(js.TestService))
	go server.Serve(lis)
}

// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn)
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
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

// TestClient_AsyncCall test client asynchronously call
func TestClient_AsyncCall(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn)
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
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
			call := c.client.AsyncCall(c.serviceMenthod, c.arg, reply)
			err := <-call
			assert.Equal(t, c.expect.reply, reply)
			assert.Equal(t, c.expect.err, err.Error)
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
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
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

// TestNewClientWithGzipCompress test gzip comressor
func TestNewClientWithGzipCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
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

// TestNewClientWithZlibCompress test zlib compressor
func TestNewClientWithZlibCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
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

// TestServer_Register .
func TestServer_Register(t *testing.T) {
	server := NewServer()
	server.RegisterName("ArithService", new(pb.ArithService))
	err := server.Register(new(pb.ArithService))
	assert.Equal(t, errors.New("rpc: service already defined: ArithService"), err)
}

// Json .
type Json struct{}

// Marshal .
func (_ *Json) Marshal(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// Unmarshal .
func (_ *Json) Unmarshal(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

// TestNewClientWithSerializer .
func TestNewClientWithSerializer(t *testing.T) {

	conn, err := net.Dial("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithSerializer(&Json{}))
	defer client.Close()

	type expect struct {
		reply *js.Response
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *js.Request
		expect         expect
	}{
		{
			client,
			"test-1",
			"TestService.Add",
			&js.Request{A: 20, B: 5},
			expect{
				&js.Response{25},
				nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &js.Response{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, c.expect.reply, reply)
			assert.Equal(t, c.expect.err, err)
		})
	}
}
