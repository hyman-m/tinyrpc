// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compressor

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
)

// ZlibCompressor implements the Compressor interface
type ZlibCompressor struct {
}

// Zip .
func (_ ZlibCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := zlib.NewWriter(buf)
	defer w.Close()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

// Unzip .
func (_ ZlibCompressor) Unzip(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}
