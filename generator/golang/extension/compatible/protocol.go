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

package compatible

import (
	"context"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

var ctx = context.Background()

// ProtocolDropContext wraps a thrift.TProtocol into a TProtocolWithoutContext.
func ProtocolDropContext(xprot thrift.TProtocol) TProtocolWithoutContext {
	var x interface{} = xprot // to prevent 'impossible type assertion' compile error
	switch v := x.(type) {
	case TProtocolWithoutContext:
		return v
	case TProtocolWithContext:
		return &protocolWithoutContext{TProtocolWithContext: v}
	default:
		panic(fmt.Errorf("unexpected type %T", v))
	}
}

// ProtocolWithContext wraps a thrift.TProtocol into a TProtocolWithContext.
func ProtocolWithContext(xprot thrift.TProtocol) TProtocolWithContext {
	var x interface{} = xprot // to prevent 'impossible type assertion' compile error
	switch v := x.(type) {
	case TProtocolWithoutContext:
		return &protocolWithContext{TProtocolWithoutContext: v}
	case TProtocolWithContext:
		return v
	default:
		panic(fmt.Errorf("unexpected type %T", v))
	}
}

// TProtocolWithContext is a subset of the thrift.TProtocol interface whose methods accept
// context.Context as the first parameter.
type TProtocolWithContext interface {
	WriteMessageBegin(ctx context.Context, name string, typeID thrift.TMessageType, seqID int32) error
	WriteMessageEnd(ctx context.Context) error
	WriteStructBegin(ctx context.Context, name string) error
	WriteStructEnd(ctx context.Context) error
	WriteFieldBegin(ctx context.Context, name string, typeID thrift.TType, id int16) error
	WriteFieldEnd(ctx context.Context) error
	WriteFieldStop(ctx context.Context) error
	WriteMapBegin(ctx context.Context, keyType thrift.TType, valueType thrift.TType, size int) error
	WriteMapEnd(ctx context.Context) error
	WriteListBegin(ctx context.Context, elemType thrift.TType, size int) error
	WriteListEnd(ctx context.Context) error
	WriteSetBegin(ctx context.Context, elemType thrift.TType, size int) error
	WriteSetEnd(ctx context.Context) error
	WriteBool(ctx context.Context, value bool) error
	WriteByte(ctx context.Context, value int8) error
	WriteI16(ctx context.Context, value int16) error
	WriteI32(ctx context.Context, value int32) error
	WriteI64(ctx context.Context, value int64) error
	WriteDouble(ctx context.Context, value float64) error
	WriteString(ctx context.Context, value string) error
	WriteBinary(ctx context.Context, value []byte) error
	ReadMessageBegin(ctx context.Context) (name string, typeID thrift.TMessageType, seqID int32, err error)
	ReadMessageEnd(ctx context.Context) error
	ReadStructBegin(ctx context.Context) (name string, err error)
	ReadStructEnd(ctx context.Context) error
	ReadFieldBegin(ctx context.Context) (name string, typeID thrift.TType, id int16, err error)
	ReadFieldEnd(ctx context.Context) error
	ReadMapBegin(ctx context.Context) (keyType thrift.TType, valueType thrift.TType, size int, err error)
	ReadMapEnd(ctx context.Context) error
	ReadListBegin(ctx context.Context) (elemType thrift.TType, size int, err error)
	ReadListEnd(ctx context.Context) error
	ReadSetBegin(ctx context.Context) (elemType thrift.TType, size int, err error)
	ReadSetEnd(ctx context.Context) error
	ReadBool(ctx context.Context) (value bool, err error)
	ReadByte(ctx context.Context) (value int8, err error)
	ReadI16(ctx context.Context) (value int16, err error)
	ReadI32(ctx context.Context) (value int32, err error)
	ReadI64(ctx context.Context) (value int64, err error)
	ReadDouble(ctx context.Context) (value float64, err error)
	ReadString(ctx context.Context) (value string, err error)
	ReadBinary(ctx context.Context) (value []byte, err error)
	Skip(ctx context.Context, fieldType thrift.TType) (err error)
	Flush(ctx context.Context) (err error)
	Transport() thrift.TTransport
}

// TProtocolWithoutContext is a subset of the thrift.TProtocol interface whose methods may
// not have a context.Context parameter.
type TProtocolWithoutContext interface {
	WriteMessageBegin(name string, typeID thrift.TMessageType, seqID int32) error
	WriteMessageEnd() error
	WriteStructBegin(name string) error
	WriteStructEnd() error
	WriteFieldBegin(name string, typeID thrift.TType, id int16) error
	WriteFieldEnd() error
	WriteFieldStop() error
	WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error
	WriteMapEnd() error
	WriteListBegin(elemType thrift.TType, size int) error
	WriteListEnd() error
	WriteSetBegin(elemType thrift.TType, size int) error
	WriteSetEnd() error
	WriteBool(value bool) error
	WriteByte(value int8) error
	WriteI16(value int16) error
	WriteI32(value int32) error
	WriteI64(value int64) error
	WriteDouble(value float64) error
	WriteString(value string) error
	WriteBinary(value []byte) error
	ReadMessageBegin() (name string, typeID thrift.TMessageType, seqID int32, err error)
	ReadMessageEnd() error
	ReadStructBegin() (name string, err error)
	ReadStructEnd() error
	ReadFieldBegin() (name string, typeID thrift.TType, id int16, err error)
	ReadFieldEnd() error
	ReadMapBegin() (keyType thrift.TType, valueType thrift.TType, size int, err error)
	ReadMapEnd() error
	ReadListBegin() (elemType thrift.TType, size int, err error)
	ReadListEnd() error
	ReadSetBegin() (elemType thrift.TType, size int, err error)
	ReadSetEnd() error
	ReadBool() (value bool, err error)
	ReadByte() (value int8, err error)
	ReadI16() (value int16, err error)
	ReadI32() (value int32, err error)
	ReadI64() (value int64, err error)
	ReadDouble() (value float64, err error)
	ReadString() (value string, err error)
	ReadBinary() (value []byte, err error)
	Skip(fieldType thrift.TType) (err error)
	Flush(ctx context.Context) (err error)
	Transport() thrift.TTransport
}

type protocolWithContext struct {
	TProtocolWithoutContext
}

func (pwc *protocolWithContext) WriteMessageBegin(ctx context.Context, name string, typeID thrift.TMessageType, seqID int32) error {
	return pwc.TProtocolWithoutContext.WriteMessageBegin(name, typeID, seqID)
}
func (pwc *protocolWithContext) WriteMessageEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteMessageEnd()
}
func (pwc *protocolWithContext) WriteStructBegin(ctx context.Context, name string) error {
	return pwc.TProtocolWithoutContext.WriteStructBegin(name)
}
func (pwc *protocolWithContext) WriteStructEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteStructEnd()
}
func (pwc *protocolWithContext) WriteFieldBegin(ctx context.Context, name string, typeID thrift.TType, id int16) error {
	return pwc.TProtocolWithoutContext.WriteFieldBegin(name, typeID, id)
}
func (pwc *protocolWithContext) WriteFieldEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteFieldEnd()
}
func (pwc *protocolWithContext) WriteFieldStop(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteFieldStop()
}
func (pwc *protocolWithContext) WriteMapBegin(ctx context.Context, keyType thrift.TType, valueType thrift.TType, size int) error {
	return pwc.TProtocolWithoutContext.WriteMapBegin(keyType, valueType, size)
}
func (pwc *protocolWithContext) WriteMapEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteMapEnd()
}
func (pwc *protocolWithContext) WriteListBegin(ctx context.Context, elemType thrift.TType, size int) error {
	return pwc.TProtocolWithoutContext.WriteListBegin(elemType, size)
}
func (pwc *protocolWithContext) WriteListEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteListEnd()
}
func (pwc *protocolWithContext) WriteSetBegin(ctx context.Context, elemType thrift.TType, size int) error {
	return pwc.TProtocolWithoutContext.WriteSetBegin(elemType, size)
}
func (pwc *protocolWithContext) WriteSetEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.WriteSetEnd()
}
func (pwc *protocolWithContext) WriteBool(ctx context.Context, value bool) error {
	return pwc.TProtocolWithoutContext.WriteBool(value)
}
func (pwc *protocolWithContext) WriteByte(ctx context.Context, value int8) error {
	return pwc.TProtocolWithoutContext.WriteByte(value)
}
func (pwc *protocolWithContext) WriteI16(ctx context.Context, value int16) error {
	return pwc.TProtocolWithoutContext.WriteI16(value)
}
func (pwc *protocolWithContext) WriteI32(ctx context.Context, value int32) error {
	return pwc.TProtocolWithoutContext.WriteI32(value)
}
func (pwc *protocolWithContext) WriteI64(ctx context.Context, value int64) error {
	return pwc.TProtocolWithoutContext.WriteI64(value)
}
func (pwc *protocolWithContext) WriteDouble(ctx context.Context, value float64) error {
	return pwc.TProtocolWithoutContext.WriteDouble(value)
}
func (pwc *protocolWithContext) WriteString(ctx context.Context, value string) error {
	return pwc.TProtocolWithoutContext.WriteString(value)
}
func (pwc *protocolWithContext) WriteBinary(ctx context.Context, value []byte) error {
	return pwc.TProtocolWithoutContext.WriteBinary(value)
}
func (pwc *protocolWithContext) ReadMessageBegin(ctx context.Context) (name string, typeID thrift.TMessageType, seqID int32, err error) {
	return pwc.TProtocolWithoutContext.ReadMessageBegin()
}
func (pwc *protocolWithContext) ReadMessageEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadMessageEnd()
}
func (pwc *protocolWithContext) ReadStructBegin(ctx context.Context) (name string, err error) {
	return pwc.TProtocolWithoutContext.ReadStructBegin()
}
func (pwc *protocolWithContext) ReadStructEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadStructEnd()
}
func (pwc *protocolWithContext) ReadFieldBegin(ctx context.Context) (name string, typeID thrift.TType, id int16, err error) {
	return pwc.TProtocolWithoutContext.ReadFieldBegin()
}
func (pwc *protocolWithContext) ReadFieldEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadFieldEnd()
}
func (pwc *protocolWithContext) ReadMapBegin(ctx context.Context) (keyType thrift.TType, valueType thrift.TType, size int, err error) {
	return pwc.TProtocolWithoutContext.ReadMapBegin()
}
func (pwc *protocolWithContext) ReadMapEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadMapEnd()
}
func (pwc *protocolWithContext) ReadListBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	return pwc.TProtocolWithoutContext.ReadListBegin()
}
func (pwc *protocolWithContext) ReadListEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadListEnd()
}
func (pwc *protocolWithContext) ReadSetBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	return pwc.TProtocolWithoutContext.ReadSetBegin()
}
func (pwc *protocolWithContext) ReadSetEnd(ctx context.Context) error {
	return pwc.TProtocolWithoutContext.ReadSetEnd()
}
func (pwc *protocolWithContext) ReadBool(ctx context.Context) (value bool, err error) {
	return pwc.TProtocolWithoutContext.ReadBool()
}
func (pwc *protocolWithContext) ReadByte(ctx context.Context) (value int8, err error) {
	return pwc.TProtocolWithoutContext.ReadByte()
}
func (pwc *protocolWithContext) ReadI16(ctx context.Context) (value int16, err error) {
	return pwc.TProtocolWithoutContext.ReadI16()
}
func (pwc *protocolWithContext) ReadI32(ctx context.Context) (value int32, err error) {
	return pwc.TProtocolWithoutContext.ReadI32()
}
func (pwc *protocolWithContext) ReadI64(ctx context.Context) (value int64, err error) {
	return pwc.TProtocolWithoutContext.ReadI64()
}
func (pwc *protocolWithContext) ReadDouble(ctx context.Context) (value float64, err error) {
	return pwc.TProtocolWithoutContext.ReadDouble()
}
func (pwc *protocolWithContext) ReadString(ctx context.Context) (value string, err error) {
	return pwc.TProtocolWithoutContext.ReadString()
}
func (pwc *protocolWithContext) ReadBinary(ctx context.Context) (value []byte, err error) {
	return pwc.TProtocolWithoutContext.ReadBinary()
}
func (pwc *protocolWithContext) Skip(ctx context.Context, fieldType thrift.TType) (err error) {
	return pwc.TProtocolWithoutContext.Skip(fieldType)
}

type protocolWithoutContext struct {
	TProtocolWithContext
}

func (pwc *protocolWithoutContext) WriteMessageBegin(name string, typeID thrift.TMessageType, seqID int32) error {
	return pwc.TProtocolWithContext.WriteMessageBegin(ctx, name, typeID, seqID)
}
func (pwc *protocolWithoutContext) WriteMessageEnd() error {
	return pwc.TProtocolWithContext.WriteMessageEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteStructBegin(name string) error {
	return pwc.TProtocolWithContext.WriteStructBegin(ctx, name)
}
func (pwc *protocolWithoutContext) WriteStructEnd() error {
	return pwc.TProtocolWithContext.WriteStructEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteFieldBegin(name string, typeID thrift.TType, id int16) error {
	return pwc.TProtocolWithContext.WriteFieldBegin(ctx, name, typeID, id)
}
func (pwc *protocolWithoutContext) WriteFieldEnd() error {
	return pwc.TProtocolWithContext.WriteFieldEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteFieldStop() error {
	return pwc.TProtocolWithContext.WriteFieldStop(ctx)
}
func (pwc *protocolWithoutContext) WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error {
	return pwc.TProtocolWithContext.WriteMapBegin(ctx, keyType, valueType, size)
}
func (pwc *protocolWithoutContext) WriteMapEnd() error {
	return pwc.TProtocolWithContext.WriteMapEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteListBegin(elemType thrift.TType, size int) error {
	return pwc.TProtocolWithContext.WriteListBegin(ctx, elemType, size)
}
func (pwc *protocolWithoutContext) WriteListEnd() error {
	return pwc.TProtocolWithContext.WriteListEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteSetBegin(elemType thrift.TType, size int) error {
	return pwc.TProtocolWithContext.WriteSetBegin(ctx, elemType, size)
}
func (pwc *protocolWithoutContext) WriteSetEnd() error {
	return pwc.TProtocolWithContext.WriteSetEnd(ctx)
}
func (pwc *protocolWithoutContext) WriteBool(value bool) error {
	return pwc.TProtocolWithContext.WriteBool(ctx, value)
}
func (pwc *protocolWithoutContext) WriteByte(value int8) error {
	return pwc.TProtocolWithContext.WriteByte(ctx, value)
}
func (pwc *protocolWithoutContext) WriteI16(value int16) error {
	return pwc.TProtocolWithContext.WriteI16(ctx, value)
}
func (pwc *protocolWithoutContext) WriteI32(value int32) error {
	return pwc.TProtocolWithContext.WriteI32(ctx, value)
}
func (pwc *protocolWithoutContext) WriteI64(value int64) error {
	return pwc.TProtocolWithContext.WriteI64(ctx, value)
}
func (pwc *protocolWithoutContext) WriteDouble(value float64) error {
	return pwc.TProtocolWithContext.WriteDouble(ctx, value)
}
func (pwc *protocolWithoutContext) WriteString(value string) error {
	return pwc.TProtocolWithContext.WriteString(ctx, value)
}
func (pwc *protocolWithoutContext) WriteBinary(value []byte) error {
	return pwc.TProtocolWithContext.WriteBinary(ctx, value)
}
func (pwc *protocolWithoutContext) ReadMessageBegin() (name string, typeID thrift.TMessageType, seqID int32, err error) {
	return pwc.TProtocolWithContext.ReadMessageBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadMessageEnd() error {
	return pwc.TProtocolWithContext.ReadMessageEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadStructBegin() (name string, err error) {
	return pwc.TProtocolWithContext.ReadStructBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadStructEnd() error {
	return pwc.TProtocolWithContext.ReadStructEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadFieldBegin() (name string, typeID thrift.TType, id int16, err error) {
	return pwc.TProtocolWithContext.ReadFieldBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadFieldEnd() error {
	return pwc.TProtocolWithContext.ReadFieldEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadMapBegin() (keyType thrift.TType, valueType thrift.TType, size int, err error) {
	return pwc.TProtocolWithContext.ReadMapBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadMapEnd() error {
	return pwc.TProtocolWithContext.ReadMapEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadListBegin() (elemType thrift.TType, size int, err error) {
	return pwc.TProtocolWithContext.ReadListBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadListEnd() error {
	return pwc.TProtocolWithContext.ReadListEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadSetBegin() (elemType thrift.TType, size int, err error) {
	return pwc.TProtocolWithContext.ReadSetBegin(ctx)
}
func (pwc *protocolWithoutContext) ReadSetEnd() error {
	return pwc.TProtocolWithContext.ReadSetEnd(ctx)
}
func (pwc *protocolWithoutContext) ReadBool() (value bool, err error) {
	return pwc.TProtocolWithContext.ReadBool(ctx)
}
func (pwc *protocolWithoutContext) ReadByte() (value int8, err error) {
	return pwc.TProtocolWithContext.ReadByte(ctx)
}
func (pwc *protocolWithoutContext) ReadI16() (value int16, err error) {
	return pwc.TProtocolWithContext.ReadI16(ctx)
}
func (pwc *protocolWithoutContext) ReadI32() (value int32, err error) {
	return pwc.TProtocolWithContext.ReadI32(ctx)
}
func (pwc *protocolWithoutContext) ReadI64() (value int64, err error) {
	return pwc.TProtocolWithContext.ReadI64(ctx)
}
func (pwc *protocolWithoutContext) ReadDouble() (value float64, err error) {
	return pwc.TProtocolWithContext.ReadDouble(ctx)
}
func (pwc *protocolWithoutContext) ReadString() (value string, err error) {
	return pwc.TProtocolWithContext.ReadString(ctx)
}
func (pwc *protocolWithoutContext) ReadBinary() (value []byte, err error) {
	return pwc.TProtocolWithContext.ReadBinary(ctx)
}
func (pwc *protocolWithoutContext) Skip(fieldType thrift.TType) (err error) {
	return pwc.TProtocolWithContext.Skip(ctx, fieldType)
}
