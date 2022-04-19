// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

// GzipCompressor implements the Compressor interface
type GzipCompressor struct {
}

// Zip .
func (_ GzipCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
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
func (_ GzipCompressor) Unzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(data))
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
