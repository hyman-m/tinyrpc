// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net"
	"time"

	"github.com/zehuamama/tinyrpc"
	"github.com/zehuamama/tinyrpc/compressor"
	pb "github.com/zehuamama/tinyrpc/example/message"
)

// listenAndServe start a tinyrpc service
func listenAndServe() error {
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		return err
	}

	server := tinyrpc.NewServer()
	server.RegisterName("ArithService", new(pb.ArithService))
	go server.Serve(lis)
	return nil
}

func rpcClient() {
	conn, err := net.Dial("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := tinyrpc.NewClient(conn)
	resq := pb.ArithRequest{A: 20, B: 5}
	resp := pb.ArithResponse{}
	err = client.Call("ArithService.Add", &resq, &resp)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Arith.Add(%v, %v): %v", resq.A, resq.B, resp.C)
	err = client.Call("ArithService.Div", &resq, &resp)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Arith.Div(%v, %v): %v", resq.A, resq.B, resp.C)
}

func rpcClientWithCompress() {
	conn, err := net.Dial("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := tinyrpc.NewClientWithCompress(conn, compressor.Snappy)
	resq := pb.ArithRequest{A: 4, B: 15}
	resp := pb.ArithResponse{}
	err = client.Call("ArithService.Mul", &resq, &resp)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Arith.Mul(%v, %v): %v", resq.A, resq.B, resp.C)
}

func rpcClientWithAyncCall() {
	conn, err := net.Dial("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := tinyrpc.NewClient(conn)
	resq := pb.ArithRequest{A: 20, B: 5}
	resp := pb.ArithResponse{}
	result := client.AsyncCall("ArithService.Sub", &resq, &resp)
	select {
	case call := <-result:
		log.Printf("Arith.Sub(%v, %v): %v ,Error: %v", resq.A, resq.B, resp.C, call.Error)
	case <-time.After(100 * time.Microsecond):
		log.Fatal("time out")
	}
}

func main() {
	err := listenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	rpcClient()
	rpcClientWithCompress()
	rpcClientWithAyncCall()
}
