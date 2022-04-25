// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serializer

// SerializeType serialized type supported by tinyrpc
type SerializeType int32

const (
	Proto SerializeType = iota
)

// Serializers which supported by tinyrpc
var Serializers = map[SerializeType]Serializer{
	Proto: ProtoSerializer{},
}

// Serializer is interface, each serializer has Marshal and Unmarshal functions
type Serializer interface {
	Marshal(message interface{}) ([]byte, error)
	Unmarshal(data []byte, message interface{}) error
}
