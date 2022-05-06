// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serializer

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// NotImplementProtoMessageError refers to param not implemented by proto.Message
var NotImplementProtoMessageError = errors.New("param does not implement proto.Message")

var Proto = ProtoSerializer{}

// ProtoSerializer implements the Serializer interface
type ProtoSerializer struct {
}

// Marshal .
func (_ ProtoSerializer) Marshal(message interface{}) ([]byte, error) {
	var body proto.Message
	if message == nil {
		return []byte{}, nil
	}
	var ok bool
	if body, ok = message.(proto.Message); !ok {
		return nil, NotImplementProtoMessageError
	}
	return proto.Marshal(body)
}

// Unmarshal .
func (_ ProtoSerializer) Unmarshal(data []byte, message interface{}) error {
	var body proto.Message
	if message == nil {
		return nil
	}

	var ok bool
	body, ok = message.(proto.Message)
	if !ok {
		return NotImplementProtoMessageError
	}

	return proto.Unmarshal(data, body)
}
