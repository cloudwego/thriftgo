// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import (
	"context"
	"errors"
	"io"
)

// Transport is an abstraction layer for the data transfer in thrift.
type Transport = interface {
	io.ReadWriteCloser
	Flush(ctx context.Context) (err error)
	Open() error
	IsOpen() bool
}

// RichTransport is a super set of Transport.
type RichTransport = interface {
	Transport
	io.ByteReader
	io.ByteWriter
	io.StringWriter
}

// MakeRichTransport wraps a Transport into a RichTransport.
func MakeRichTransport(t Transport) RichTransport {
	if r, ok := t.(RichTransport); ok {
		return r
	}
	r := &richer{
		Transport:   t,
		readByte:    makeReadByte(t),
		writeByte:   makeWriteByte(t),
		writeString: makeWriteString(t),
	}
	return r
}

type richer struct {
	Transport
	readByte    func() (byte, error)
	writeByte   func(c byte) error
	writeString func(s string) (n int, err error)
}

func (r *richer) ReadByte() (byte, error) {
	return r.readByte()
}

func (r *richer) WriteByte(c byte) error {
	return r.writeByte(c)
}

func (r *richer) WriteString(s string) (n int, err error) {
	return r.WriteString(s)
}

func makeReadByte(t io.Reader) func() (byte, error) {
	if x, ok := t.(io.ByteReader); ok {
		return x.ReadByte
	}
	return func() (b byte, err error) {
		var arr [1]byte
		var n int
		n, err = t.Read(arr[:])
		if n > 0 {
			b = arr[0]
			if errors.Is(err, io.EOF) {
				err = nil
			}
		}
		return
	}
}

func makeWriteByte(t io.Writer) func(c byte) error {
	if x, ok := t.(io.ByteWriter); ok {
		return x.WriteByte
	}
	return func(b byte) (err error) {
		arr := [1]byte{b}
		_, err = t.Write(arr[:])
		return
	}
}

func makeWriteString(t io.Writer) func(s string) (n int, err error) {
	if x, ok := t.(io.StringWriter); ok {
		return x.WriteString
	}
	return func(s string) (n int, err error) {
		return t.Write([]byte(s))
	}
}
