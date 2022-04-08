// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compressor

type RawCompressor struct {
}

func (c RawCompressor) Zip(data []byte) ([]byte, error) {
	return data, nil
}

func (c RawCompressor) Unzip(data []byte) ([]byte, error) {
	return data, nil
}
