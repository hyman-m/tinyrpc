package tinyrpc

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zehuamama/tinyrpc/compressor"
	js "github.com/zehuamama/tinyrpc/test.data/json"
	pb "github.com/zehuamama/tinyrpc/test.data/message"
)

func init() {
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()
	err = server.Register(new(pb.ArithService))
	if err != nil {
		log.Fatal(err)
	}

	go server.Serve(lis)

	lis, err = net.Listen("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}

	server = NewServer(WithSerializer(&Json{}))
	err = server.Register(new(js.TestService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)
}

// test client synchronously call
func client_call(t *testing.T, comporessType compressor.CompressType) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(comporessType))
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
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 15},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 100},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 4},
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
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	client_call(t, compressor.Raw)
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
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 15},
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 100},
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 4},
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
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err.Error)
		})
	}
}

// TestNewClientWithSnappyCompress test snappy comressor
func TestNewClientWithSnappyCompress(t *testing.T) {
	client_call(t, compressor.Snappy)
}

// TestNewClientWithGzipCompress test gzip comressor
func TestNewClientWithGzipCompress(t *testing.T) {
	client_call(t, compressor.Gzip)
}

// TestNewClientWithZlibCompress test zlib compressor
func TestNewClientWithZlibCompress(t *testing.T) {
	client_call(t, compressor.Zlib)
}

// TestServer_Register .
func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.RegisterName("ArithService", new(pb.ArithService))
	assert.Equal(t, nil, err)
	err = server.Register(new(pb.ArithService))
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
			client:         client,
			name:           "test-1",
			serviceMenthod: "TestService.Add",
			arg:            &js.Request{A: 20, B: 5},
			expect: expect{
				reply: &js.Response{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &js.Response{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}
