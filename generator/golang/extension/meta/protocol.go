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
)

// Protocol is an abstraction for input and output protocols in thrift.
type Protocol interface {
	ReadMessageBegin(ctx context.Context) (name string, typeID TMessageType, seqID int32, err error)
	ReadMessageEnd(ctx context.Context) error
	ReadBool(ctx context.Context) (value bool, err error)
	ReadByte(ctx context.Context) (value int8, err error)
	ReadI16(ctx context.Context) (value int16, err error)
	ReadI32(ctx context.Context) (value int32, err error)
	ReadI64(ctx context.Context) (value int64, err error)
	ReadDouble(ctx context.Context) (value float64, err error)
	ReadString(ctx context.Context) (value string, err error)
	ReadBinary(ctx context.Context) (value []byte, err error)
	ReadMapBegin(ctx context.Context) (keyType, valueType TTypeID, size int, err error)
	ReadMapEnd(ctx context.Context) error
	ReadListBegin(ctx context.Context) (elemType TTypeID, size int, err error)
	ReadListEnd(ctx context.Context) error
	ReadSetBegin(ctx context.Context) (elemType TTypeID, size int, err error)
	ReadSetEnd(ctx context.Context) error
	ReadStructBegin(ctx context.Context) (name string, err error)
	ReadStructEnd(ctx context.Context) error
	ReadFieldBegin(ctx context.Context) (name string, typeID TTypeID, id int16, err error)
	ReadFieldEnd(ctx context.Context) error
	Skip(ctx context.Context, fieldType TTypeID) (err error)
	WriteMessageBegin(ctx context.Context, name string, typeID TMessageType, seqID int32) error
	WriteMessageEnd(ctx context.Context) error
	WriteBool(ctx context.Context, value bool) error
	WriteByte(ctx context.Context, value int8) error
	WriteI16(ctx context.Context, value int16) error
	WriteI32(ctx context.Context, value int32) error
	WriteI64(ctx context.Context, value int64) error
	WriteDouble(ctx context.Context, value float64) error
	WriteString(ctx context.Context, value string) error
	WriteBinary(ctx context.Context, value []byte) error
	WriteMapBegin(ctx context.Context, keyType, valueType TTypeID, size int) error
	WriteMapEnd(ctx context.Context) error
	WriteListBegin(ctx context.Context, elemType TTypeID, size int) error
	WriteListEnd(ctx context.Context) error
	WriteSetBegin(ctx context.Context, elemType TTypeID, size int) error
	WriteSetEnd(ctx context.Context) error
	WriteStructBegin(ctx context.Context, name string) error
	WriteStructEnd(ctx context.Context) error
	WriteFieldBegin(ctx context.Context, name string, typeID TTypeID, id int16) error
	WriteFieldEnd(ctx context.Context) error
	WriteFieldStop(ctx context.Context) error
	Flush(ctx context.Context) (err error)
}
