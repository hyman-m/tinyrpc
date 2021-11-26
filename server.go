// Copyright 2021 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tinyrpc

import (
	"errors"
	"go/token"
	"io"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	typeOfError  = reflect.TypeOf((*error)(nil)).Elem()
	TimeOutError = errors.New("tinyrpc: call timeout")
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type Server struct {
	serviceMap sync.Map
	reqLock    sync.Mutex
	freeReq    *Request
	respLock   sync.Mutex
	freeResp   *Response
}

type ServerCodec interface {
	ReadRequestHeader(*Request) error
	ReadRequestBody(interface{}) error
	WriteResponse(*Response, interface{}) error
	Close() error
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		return errors.New("tinyrpc.Register: no service name for type " + s.typ.String())
	}
	if !token.IsExported(sname) && !useName {
		return errors.New("tinyrpc.Register: type " + sname + " is not exported")
	}
	s.name = sname
	s.method = suitableMethods(s.typ)
	if len(s.method) == 0 {
		return errors.New("tinyrpc.Register: type " + sname + " has no exported methods of suitable type")
	}
	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("tinyrpc: service already registered: " + sname)
	}
	return nil
}

func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if !method.IsExported() {
			continue
		}
		if mtype.NumIn() != 3 {
			continue
		}
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		if argType.Kind() != reflect.Ptr {
			continue
		}
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		if !isExportedOrBuiltinType(replyType) {
			continue
		}
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

var invalidRequest = struct{}{}

func (server *Server) sendResponse(
	sending *sync.Mutex, req *Request, reply interface{}, codec ServerCodec, errmsg string) {
	resp := server.getResponse()
	// Encode the response header
	resp.ServiceMethod = req.ServiceMethod
	if errmsg != "" {
		resp.Error = errmsg
		reply = invalidRequest
	}
	resp.Seq = req.Seq
	sending.Lock()
	err := codec.WriteResponse(resp, reply)
	if err != nil {
		log.Println("tinyrpc: writing response error:", err)
	}
	sending.Unlock()
	server.freeResponse(resp)
}

func (s *service) call(server *Server, sending *sync.Mutex, wg *sync.WaitGroup,
	mtype *methodType, req *Request, argv, replyv reflect.Value, codec ServerCodec) {
	if wg != nil {
		defer wg.Done()
	}
	timeout := false

	errChannel := make(chan interface{})
	go s.runCall(mtype, argv, replyv, errChannel)
	var errInter interface{}

	if req.TTL == time.Unix(0, 0) {
		errInter = <-errChannel
	} else {
		select {
		case errInter = <-errChannel:
		case <-time.After(req.TTL.Sub(time.Now())):
			timeout = true
		}
	}
	var errMsg string
	if timeout {
		errMsg = TimeOutError.Error()
	}
	if !timeout && errInter != nil {
		errMsg = errInter.(error).Error()
	}
	server.sendResponse(sending, req, replyv.Interface(), codec, errMsg)
	server.freeRequest(req)
}

func (s *service) runCall(mtype *methodType, argv, replyv reflect.Value, ch chan interface{}) {
	function := mtype.method.Func
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	errInter := returnValues[0].Interface()
	ch <- errInter
}

func (server *Server) ServeCodec(codec ServerCodec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
		if err != nil {
			if !keepReading {
				break
			}
			if req != nil {
				server.sendResponse(sending, req, invalidRequest, codec, err.Error())
				server.freeRequest(req)
			}
			continue
		}
		wg.Add(1)
		go service.call(server, sending, wg, mtype, req, argv, replyv, codec)
	}
	wg.Wait()
	codec.Close()
}

// ServeRequest is like ServeCodec but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(codec ServerCodec) error {
	sending := new(sync.Mutex)
	service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
	if err != nil {
		if !keepReading {
			return err
		}
		// send a response if we actually managed to read a header.
		if req != nil {
			server.sendResponse(sending, req, invalidRequest, codec, err.Error())
			server.freeRequest(req)
		}
		return err
	}
	service.call(server, sending, nil, mtype, req, argv, replyv, codec)
	return nil
}

func (server *Server) readRequest(codec ServerCodec) (
	service *service, mtype *methodType, req *Request, argv, replyv reflect.Value, keepReading bool, err error) {
	service, mtype, req, keepReading, err = server.readRequestHeader(codec)
	if err != nil {
		if !keepReading {
			return
		}
		codec.ReadRequestBody(nil)
		return
	}
	argv = reflect.New(mtype.ArgType.Elem())
	// argv guaranteed to be a pointer now.
	if err = codec.ReadRequestBody(argv.Interface()); err != nil {
		return
	}
	replyv = reflect.New(mtype.ReplyType.Elem())
	return
}

func (server *Server) readRequestHeader(codec ServerCodec) (
	svc *service, mtype *methodType, req *Request, keepReading bool, err error) {
	req = server.getRequest()
	err = codec.ReadRequestHeader(req)
	if err != nil {
		req = nil
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		err = errors.New("tinyrpc: server cannot decode request: " + err.Error())
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	dot := strings.LastIndex(req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("tinyrpc: service/method request ill-formed: " + req.ServiceMethod)
		return
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]

	// Look up the request.
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("tinyrpc: can't find service " + req.ServiceMethod)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("tinyrpc: can't find method " + req.ServiceMethod)
	}
	return
}
