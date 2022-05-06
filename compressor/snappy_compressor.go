// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compressor

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
)

// SnappyCompressor implements the Compressor interface
type SnappyCompressor struct {
}

// Zip .
func (_ SnappyCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := snappy.NewBufferedWriter(buf)
	defer func() {
		w.Close()
	}()
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
func (_ SnappyCompressor) Unzip(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewBuffer(data))
	data, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}
