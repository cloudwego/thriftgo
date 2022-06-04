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
	"fmt"
	"strings"
)

// DebugProtocol wraps another Protocol and writes a log every time
// every time its method is being invoked.
// The default logging function writes to the standard output.
type DebugProtocol struct {
	impl   Protocol
	indent int
	logf   func(format string, a ...interface{})
}

func defaultDebugLogFunc(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Println()
}

// NewDebugProtocol creates a DebugProtocol wrapping the given Protocol.
func NewDebugProtocol(impl Protocol) *DebugProtocol {
	return &DebugProtocol{impl: impl, logf: defaultDebugLogFunc}
}

// WithLogFunc sets the logging function of the DebugProtocol.
func (p *DebugProtocol) WithLogFunc(f func(format string, a ...interface{})) *DebugProtocol {
	p.logf = f
	return p
}

// ReadMessageBegin .
func (p *DebugProtocol) ReadMessageBegin(ctx context.Context) (name string, typeID TMessageType, seqID int32, err error) {
	name, typeID, seqID, err = p.impl.ReadMessageBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadMessageBegin() (name=%#v, typeID=%#v, seqID=%#v, err=%#v)", indent, name, typeID, seqID, err)
	p.indent++
	return
}

// ReadMessageEnd .
func (p *DebugProtocol) ReadMessageEnd(ctx context.Context) (err error) {
	err = p.impl.ReadMessageEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadMessageEnd() err=%#v", indent, err)
	p.indent--
	return
}

// ReadBool .
func (p *DebugProtocol) ReadBool(ctx context.Context) (value bool, err error) {
	value, err = p.impl.ReadBool(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadBool() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadByte .
func (p *DebugProtocol) ReadByte(ctx context.Context) (value int8, err error) {
	value, err = p.impl.ReadByte(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadByte() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadI16 .
func (p *DebugProtocol) ReadI16(ctx context.Context) (value int16, err error) {
	value, err = p.impl.ReadI16(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadI16() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadI32 .
func (p *DebugProtocol) ReadI32(ctx context.Context) (value int32, err error) {
	value, err = p.impl.ReadI32(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadI32() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadI64 .
func (p *DebugProtocol) ReadI64(ctx context.Context) (value int64, err error) {
	value, err = p.impl.ReadI64(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadI64() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadDouble .
func (p *DebugProtocol) ReadDouble(ctx context.Context) (value float64, err error) {
	value, err = p.impl.ReadDouble(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadDouble() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadString .
func (p *DebugProtocol) ReadString(ctx context.Context) (value string, err error) {
	value, err = p.impl.ReadString(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadString() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadBinary .
func (p *DebugProtocol) ReadBinary(ctx context.Context) (value []byte, err error) {
	value, err = p.impl.ReadBinary(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadBinary() (value=%#v, err=%#v)", indent, value, err)
	return
}

// ReadMapBegin .
func (p *DebugProtocol) ReadMapBegin(ctx context.Context) (keyType, valueType TTypeID, size int, err error) {
	keyType, valueType, size, err = p.impl.ReadMapBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadMapBegin() (keyType=%#v, valueType=%#v, size=%#v, err=%#v)", indent, keyType, valueType, size, err)
	p.indent++
	return
}

// ReadMapEnd .
func (p *DebugProtocol) ReadMapEnd(ctx context.Context) (err error) {
	err = p.impl.ReadMapEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadMapEnd() err=%#v", indent, err)
	p.indent--
	return
}

// ReadListBegin .
func (p *DebugProtocol) ReadListBegin(ctx context.Context) (elemType TTypeID, size int, err error) {
	elemType, size, err = p.impl.ReadListBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadListBegin() (elemType=%#v, size=%#v, err=%#v)", indent, elemType, size, err)
	p.indent++
	return
}

// ReadListEnd .
func (p *DebugProtocol) ReadListEnd(ctx context.Context) (err error) {
	err = p.impl.ReadListEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadListEnd() err=%#v", indent, err)
	p.indent--
	return
}

// ReadSetBegin .
func (p *DebugProtocol) ReadSetBegin(ctx context.Context) (elemType TTypeID, size int, err error) {
	elemType, size, err = p.impl.ReadSetBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadSetBegin() (elemType=%#v, size=%#v, err=%#v)", indent, elemType, size, err)
	p.indent++
	return
}

// ReadSetEnd .
func (p *DebugProtocol) ReadSetEnd(ctx context.Context) (err error) {
	err = p.impl.ReadSetEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadSetEnd() err=%#v", indent, err)
	p.indent--
	return
}

// ReadStructBegin .
func (p *DebugProtocol) ReadStructBegin(ctx context.Context) (name string, err error) {
	name, err = p.impl.ReadStructBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadStructBegin() (name%#v, err=%#v)", indent, name, err)
	p.indent++
	return
}

// ReadStructEnd .
func (p *DebugProtocol) ReadStructEnd(ctx context.Context) (err error) {
	err = p.impl.ReadStructEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadStructEnd() err=%#v", indent, err)
	p.indent--
	return
}

// ReadFieldBegin .
func (p *DebugProtocol) ReadFieldBegin(ctx context.Context) (name string, typeID TTypeID, id int16, err error) {
	name, typeID, id, err = p.impl.ReadFieldBegin(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadFieldBegin() (name=%#v, typeID=%#v, id=%#v, err=%#v)", indent, name, typeID, id, err)
	p.indent++
	return
}

// ReadFieldEnd .
func (p *DebugProtocol) ReadFieldEnd(ctx context.Context) (err error) {
	err = p.impl.ReadFieldEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sReadFieldEnd() err=%#v", indent, err)
	p.indent--
	return
}

// Skip .
func (p *DebugProtocol) Skip(ctx context.Context, fieldType TTypeID) (err error) {
	err = p.impl.Skip(ctx, fieldType)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sSkip(fieldType=%#v) (err=%#v)", indent, fieldType, err)
	return
}

// WriteMessageBegin .
func (p *DebugProtocol) WriteMessageBegin(ctx context.Context, name string, typeID TMessageType, seqID int32) (err error) {
	err = p.impl.WriteMessageBegin(ctx, name, typeID, seqID)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteMessageBegin(name=%#v, typeID=%#v, seqID=%#v) => %#v", indent, name, typeID, seqID, err)
	p.indent++
	return
}

// WriteMessageEnd .
func (p *DebugProtocol) WriteMessageEnd(ctx context.Context) (err error) {
	err = p.impl.WriteMessageEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteMessageEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteBool .
func (p *DebugProtocol) WriteBool(ctx context.Context, value bool) (err error) {
	err = p.impl.WriteBool(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteBool(value=%#v) => %#v", indent, value, err)
	return
}

// WriteByte .
func (p *DebugProtocol) WriteByte(ctx context.Context, value int8) (err error) {
	err = p.impl.WriteByte(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteByte(value=%#v) => %#v", indent, value, err)
	return
}

// WriteI16 .
func (p *DebugProtocol) WriteI16(ctx context.Context, value int16) (err error) {
	err = p.impl.WriteI16(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteI16(value=%#v) => %#v", indent, value, err)
	return
}

// WriteI32 .
func (p *DebugProtocol) WriteI32(ctx context.Context, value int32) (err error) {
	err = p.impl.WriteI32(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteI32(value=%#v) => %#v", indent, value, err)
	return
}

// WriteI64 .
func (p *DebugProtocol) WriteI64(ctx context.Context, value int64) (err error) {
	err = p.impl.WriteI64(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteI64(value=%#v) => %#v", indent, value, err)
	return
}

// WriteDouble .
func (p *DebugProtocol) WriteDouble(ctx context.Context, value float64) (err error) {
	err = p.impl.WriteDouble(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteDouble(value=%#v) => %#v", indent, value, err)
	return
}

// WriteString .
func (p *DebugProtocol) WriteString(ctx context.Context, value string) (err error) {
	err = p.impl.WriteString(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteString(value=%#v) => %#v", indent, value, err)
	return
}

// WriteBinary .
func (p *DebugProtocol) WriteBinary(ctx context.Context, value []byte) (err error) {
	err = p.impl.WriteBinary(ctx, value)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteBinary(value=%#v) => %#v", indent, value, err)
	return
}

// WriteMapBegin .
func (p *DebugProtocol) WriteMapBegin(ctx context.Context, keyType, valueType TTypeID, size int) (err error) {
	err = p.impl.WriteMapBegin(ctx, keyType, valueType, size)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteMapBegin(keyType=%#v, valueType=%#v, size=%#v) => %#v", indent, keyType, valueType, size, err)
	p.indent++
	return
}

// WriteMapEnd .
func (p *DebugProtocol) WriteMapEnd(ctx context.Context) (err error) {
	err = p.impl.WriteMapEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteMapEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteListBegin .
func (p *DebugProtocol) WriteListBegin(ctx context.Context, elemType TTypeID, size int) (err error) {
	err = p.impl.WriteListBegin(ctx, elemType, size)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteListBegin(elemType=%#v, size=%#v) => %#v", indent, elemType, size, err)
	p.indent++
	return
}

// WriteListEnd .
func (p *DebugProtocol) WriteListEnd(ctx context.Context) (err error) {
	err = p.impl.WriteListEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteListEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteSetBegin .
func (p *DebugProtocol) WriteSetBegin(ctx context.Context, elemType TTypeID, size int) (err error) {
	err = p.impl.WriteSetBegin(ctx, elemType, size)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteSetBegin(elemType=%#v, size=%#v) => %#v", indent, elemType, size, err)
	p.indent++
	return
}

// WriteSetEnd .
func (p *DebugProtocol) WriteSetEnd(ctx context.Context) (err error) {
	err = p.impl.WriteSetEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteSetEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteStructBegin .
func (p *DebugProtocol) WriteStructBegin(ctx context.Context, name string) (err error) {
	err = p.impl.WriteStructBegin(ctx, name)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteStructBegin(name=%#v) => %#v", indent, name, err)
	p.indent++
	return
}

// WriteStructEnd .
func (p *DebugProtocol) WriteStructEnd(ctx context.Context) (err error) {
	err = p.impl.WriteStructEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteStructEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteFieldBegin .
func (p *DebugProtocol) WriteFieldBegin(ctx context.Context, name string, typeID TTypeID, id int16) (err error) {
	err = p.impl.WriteFieldBegin(ctx, name, typeID, id)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteFieldBegin(name=%#v, typeID=%#v, id%#v) => %#v", indent, name, typeID, id, err)
	p.indent++
	return
}

// WriteFieldEnd .
func (p *DebugProtocol) WriteFieldEnd(ctx context.Context) (err error) {
	err = p.impl.WriteFieldEnd(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteFieldEnd() => %#v", indent, err)
	p.indent--
	return
}

// WriteFieldStop .
func (p *DebugProtocol) WriteFieldStop(ctx context.Context) (err error) {
	err = p.impl.WriteFieldStop(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sWriteFieldStop() => %#v", indent, err)
	return
}

// Flush .
func (p *DebugProtocol) Flush(ctx context.Context) (err error) {
	err = p.impl.Flush(ctx)
	indent := strings.Repeat("  ", p.indent)
	p.logf("%sFlush() (err=%#v)", indent, err)
	return
}
