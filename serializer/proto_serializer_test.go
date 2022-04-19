// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serializer

import (
	"errors"
	"github.com/stretchr/testify/assert"
	pb "github.com/zehuamama/tinyrpc/example/message"
	"testing"
)

type test struct{}

// TestProtoSerializer_Marshal .
func TestProtoSerializer_Marshal(t *testing.T) {
	type expect struct {
		data []byte
		err  error
	}
	cases := []struct {
		name   string
		arg    any
		expect expect
	}{
		{
			name: "test-1",
			arg:  &pb.ArithRequest{1, 2},
			expect: expect{
				data: []byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0,
					0x3f, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40},
				err: nil,
			},
		},
		{
			name: "test-2",
			arg:  test{},
			expect: expect{
				data: nil,
				err:  errors.New("param does not implement proto.Message"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := ProtoSerializer{}.Marshal(c.arg)
			assert.Equal(t, c.expect.data, data)
			assert.Equal(t, c.expect.err, err)
		})
	}
}

// TestProtoSerializer_Unmarshal .
func TestProtoSerializer_Unmarshal(t *testing.T) {
	type expect struct {
		message any
		err     error
	}
	cases := []struct {
		name   string
		arg    []byte
		expect expect
	}{
		{
			name: "test-1",
			arg: []byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0,
				0x3f, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40},
			expect: expect{
				message: &pb.ArithRequest{1, 2},
				err:     nil,
			},
		},
		{
			name: "test-2",
			arg:  nil,
			expect: expect{
				message: &pb.ArithRequest{0, 0},
				err:     nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			message := &pb.ArithRequest{}
			err := ProtoSerializer{}.Unmarshal(c.arg, message)
			assert.Equal(t, c.expect.message, message)
			assert.Equal(t, c.expect.err, err)
		})
	}
}
