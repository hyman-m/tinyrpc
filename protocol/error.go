// Copyright 2021 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import "errors"

var (
	NotImplementProtoMessageError = errors.New("param does not implement proto.Message")
	InvalidSequenceError          = errors.New("invalid sequence number in response")
	MaxLimitLengthHeaderError     = errors.New("header exceeds the maximum limit length")
	UnexpectedChecksumError       = errors.New("unexpected checksum")
)
