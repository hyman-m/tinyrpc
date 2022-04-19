// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serializer

import (
	"github.com/golang/protobuf/proto"
	errs "github.com/zehuamama/tinyrpc/errors"
)

// ProtoSerializer implements the Serializer interface
type ProtoSerializer struct {
}

// Marshal .
func (_ ProtoSerializer) Marshal(message any) ([]byte, error) {
	var body proto.Message
	if message != nil {
		var ok bool
		if body, ok = message.(proto.Message); !ok {
			return nil, errs.NotImplementProtoMessageError
		}
	}
	return proto.Marshal(body)
}

// Unmarshal .
func (_ ProtoSerializer) Unmarshal(data []byte, message any) error {
	var body proto.Message
	if message != nil {
		var ok bool
		body, ok = message.(proto.Message)
		if !ok {
			return errs.NotImplementProtoMessageError
		}
	}
	return proto.Unmarshal(data, body)
}
