// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serializer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/zehuamama/tinyrpc/test.data/message"
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
		arg    interface{}
		expect expect
	}{
		{
			name: "test-1",
			arg:  &pb.ArithRequest{A: 1, B: 2},
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
		{
			name: "test-3",
			arg:  nil,
			expect: expect{
				data: []byte{},
				err:  nil,
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
		message interface{}
		err     error
	}
	cases := []struct {
		name    string
		arg     []byte
		message interface{}
		expect  expect
	}{
		{
			name: "test-1",
			arg: []byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0,
				0x3f, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40},
			message: &pb.ArithRequest{},
			expect: expect{
				message: &pb.ArithRequest{A: 1, B: 2},
				err:     nil,
			},
		},
		{
			name:    "test-2",
			arg:     nil,
			message: nil,
			expect: expect{
				message: nil,
				err:     nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ProtoSerializer{}.Unmarshal(c.arg, c.message)
			if err != nil {
				assert.Equal(t, c.expect.message.(*pb.ArithRequest).A,
					c.message.(*pb.ArithRequest).A)
				assert.Equal(t, c.expect.message.(*pb.ArithRequest).B,
					c.message.(*pb.ArithRequest).B)
			}

			assert.Equal(t, c.expect.err, err)
		})
	}
}
