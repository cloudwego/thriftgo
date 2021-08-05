// Copyright 2021 CloudWeGo Authors
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

package plugin

import (
	"bytes"
	"context"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// Writer writes out data to an output thrift.TProtocol.
type Writer interface {
	Write(oprot thrift.TProtocol) error
}

// Reader reads data from an input thrift.TProtocol.
type Reader interface {
	Read(iprot thrift.TProtocol) error
}

// CtxWriter writes out data to an output thrift.TProtocol.
type CtxWriter interface {
	Write(ctx context.Context, oprot thrift.TProtocol) error
}

// CtxReader reads data from an input thrift.TProtocol.
type CtxReader interface {
	Read(ctx context.Context, iprot thrift.TProtocol) error
}

// Marshal serializes the data with binary protocol.
func Marshal(data interface{}) ([]byte, error) {
	ctx := context.TODO()
	var buf bytes.Buffer
	trans := thrift.NewStreamTransportRW(&buf)
	proto := thrift.NewTBinaryProtocol(trans, true, true)

	var err error
	switch v := data.(type) {
	case Writer:
		err = v.Write(proto)
	case CtxWriter:
		err = v.Write(ctx, proto)
	default:
		return nil, fmt.Errorf("Marshal accepts only Writer or CtxWriter, got: %T", data)
	}
	if err != nil {
		return nil, err
	}
	err = proto.Flush(ctx) // NOTE: important
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal deserializes the data from bytes with binary protocol.
func Unmarshal(data interface{}, bs []byte) error {
	trans := thrift.NewStreamTransportR(bytes.NewReader(bs))
	proto := thrift.NewTBinaryProtocolTransport(trans)
	switch v := data.(type) {
	case Reader:
		return v.Read(proto)
	case CtxReader:
		return v.Read(context.TODO(), proto)
	default:
		return fmt.Errorf("Unmarshal accepts only Reader or CtxReader, got: %T", data)
	}
}

// MarshalRequest encodes a request with binary protocol.
func MarshalRequest(data *Request) ([]byte, error) {
	return Marshal(data)
}

// UnmarshalRequest decodes a request with binary protocol.
func UnmarshalRequest(bs []byte) (*Request, error) {
	req := NewRequest()
	if err := Unmarshal(req, bs); err != nil {
		return nil, err
	}
	return req, nil
}

// MarshalResponse encodes a response with binary protocol.
func MarshalResponse(data *Response) ([]byte, error) {
	return Marshal(data)
}

// UnmarshalResponse decodes a response with binary protocol.
func UnmarshalResponse(bs []byte) (*Response, error) {
	res := NewResponse()
	if err := Unmarshal(res, bs); err != nil {
		return nil, err
	}
	return res, nil
}
