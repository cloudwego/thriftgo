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

package unknown

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Type IDs.
const (
	TStop   = 0
	TVoid   = 1
	TBool   = 2
	TByte   = 3
	TDouble = 4
	TI16    = 6
	TI32    = 8
	TI64    = 10
	TString = 11
	TStruct = 12
	TMap    = 13
	TSet    = 14
	TList   = 15
	TUtf8   = 16
	TUtf16  = 17
)

// TType is supposed to be an integer.
type TType interface{}

// TProtocol is supposed to have methods that a thrift.TProtocol requires.
type TProtocol interface{}

type protocol struct {
	impl reflect.Value
}

func (p *protocol) call(method string, args ...interface{}) (res []reflect.Value) {
	m := p.impl.MethodByName(method)
	t := m.Type()
	if t.NumIn() == 0 || !t.In(0).Implements(contextInterface) {
		args = args[1:]
	}
	var vs []reflect.Value
	for i, a := range args {
		v := reflect.ValueOf(a).Convert(t.In(i))
		vs = append(vs, v)
	}
	return m.Call(vs)
}

func (p *protocol) convert(res []reflect.Value, ps ...interface{}) {
	for i, p := range ps {
		v := reflect.ValueOf(p).Elem()
		v.Set(res[i].Convert(v.Type()))
	}
}

func (p *protocol) ReadMessageBegin(ctx context.Context) (name string, typeID int, seqID int32, err error) {
	p.convert(p.call("ReadMessageBegin", ctx), &name, &typeID, &seqID, &err)
	return
}

func (p *protocol) ReadMessageEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadMessageEnd", ctx), &err)
	return
}

func (p *protocol) ReadBool(ctx context.Context) (value bool, err error) {
	p.convert(p.call("ReadBool", ctx), &value, &err)
	return
}

func (p *protocol) ReadByte(ctx context.Context) (value int8, err error) {
	p.convert(p.call("ReadByte", ctx), &value, &err)
	return
}

func (p *protocol) ReadI16(ctx context.Context) (value int16, err error) {
	p.convert(p.call("ReadI16", ctx), &value, &err)
	return
}

func (p *protocol) ReadI32(ctx context.Context) (value int32, err error) {
	p.convert(p.call("ReadI32", ctx), &value, &err)
	return
}

func (p *protocol) ReadI64(ctx context.Context) (value int64, err error) {
	p.convert(p.call("ReadI64", ctx), &value, &err)
	return
}

func (p *protocol) ReadDouble(ctx context.Context) (value float64, err error) {
	p.convert(p.call("ReadDouble", ctx), &value, &err)
	return
}

func (p *protocol) ReadString(ctx context.Context) (value string, err error) {
	p.convert(p.call("ReadString", ctx), &value, &err)
	return
}

func (p *protocol) ReadBinary(ctx context.Context) (value []byte, err error) {
	p.convert(p.call("ReadBinary", ctx), &value, &err)
	return
}

func (p *protocol) ReadMapBegin(ctx context.Context) (keyType, valueType, size int, err error) {
	p.convert(p.call("ReadMapBegin", ctx), &keyType, &valueType, &size, &err)
	return
}

func (p *protocol) ReadMapEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadMapEnd", ctx), &err)
	return
}

func (p *protocol) ReadListBegin(ctx context.Context) (elemType, size int, err error) {
	p.convert(p.call("ReadListBegin", ctx), &elemType, &size, &err)
	return
}

func (p *protocol) ReadListEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadListEnd", ctx), &err)
	return
}

func (p *protocol) ReadSetBegin(ctx context.Context) (elemType, size int, err error) {
	p.convert(p.call("ReadSetBegin", ctx), &elemType, &size, &err)
	return
}

func (p *protocol) ReadSetEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadSetEnd", ctx), &err)
	return
}

func (p *protocol) ReadStructBegin(ctx context.Context) (name string, err error) {
	p.convert(p.call("ReadStructBegin", ctx), &name, &err)
	return
}

func (p *protocol) ReadStructEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadStructEnd", ctx), &err)
	return
}

func (p *protocol) ReadFieldBegin(ctx context.Context) (name string, typeID int, id int16, err error) {
	p.convert(p.call("ReadFieldBegin", ctx), &name, &typeID, &id, &err)
	return
}

func (p *protocol) ReadFieldEnd(ctx context.Context) (err error) {
	p.convert(p.call("ReadFieldEnd", ctx), &err)
	return
}

func (p *protocol) Skip(ctx context.Context, fieldType int) (err error) {
	p.convert(p.call("Skip", ctx, fieldType), &err)
	return
}

func (p *protocol) WriteMessageBegin(ctx context.Context, name string, typeID int, seqID int32) (err error) {
	p.convert(p.call("WriteMessageBegin", ctx, name, typeID, seqID), &err)
	return
}

func (p *protocol) WriteMessageEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteMessageEnd", ctx), &err)
	return
}

func (p *protocol) WriteBool(ctx context.Context, value bool) (err error) {
	p.convert(p.call("WriteBool", ctx, value), &err)
	return
}

func (p *protocol) WriteByte(ctx context.Context, value int8) (err error) {
	p.convert(p.call("WriteByte", ctx, value), &err)
	return
}

func (p *protocol) WriteI16(ctx context.Context, value int16) (err error) {
	p.convert(p.call("WriteI16", ctx, value), &err)
	return
}

func (p *protocol) WriteI32(ctx context.Context, value int32) (err error) {
	p.convert(p.call("WriteI32", ctx, value), &err)
	return
}

func (p *protocol) WriteI64(ctx context.Context, value int64) (err error) {
	p.convert(p.call("WriteI64", ctx, value), &err)
	return
}

func (p *protocol) WriteDouble(ctx context.Context, value float64) (err error) {
	p.convert(p.call("WriteDouble", ctx, value), &err)
	return
}

func (p *protocol) WriteString(ctx context.Context, value string) (err error) {
	p.convert(p.call("WriteString", ctx, value), &err)
	return
}

func (p *protocol) WriteBinary(ctx context.Context, value []byte) (err error) {
	p.convert(p.call("WriteBinary", ctx, value), &err)
	return
}

func (p *protocol) WriteMapBegin(ctx context.Context, keyType, valueType, size int) (err error) {
	p.convert(p.call("WriteMapBegin", ctx, keyType, valueType, size), &err)
	return
}

func (p *protocol) WriteMapEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteMapEnd", ctx), &err)
	return
}

func (p *protocol) WriteListBegin(ctx context.Context, elemType, size int) (err error) {
	p.convert(p.call("WriteListBegin", ctx, elemType, size), &err)
	return
}

func (p *protocol) WriteListEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteListEnd", ctx), &err)
	return
}

func (p *protocol) WriteSetBegin(ctx context.Context, elemType, size int) (err error) {
	p.convert(p.call("WriteSetBegin", ctx, elemType, size), &err)
	return
}

func (p *protocol) WriteSetEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteSetEnd", ctx), &err)
	return
}

func (p *protocol) WriteStructBegin(ctx context.Context, name string) (err error) {
	p.convert(p.call("WriteStructBegin", ctx, name), &err)
	return
}

func (p *protocol) WriteStructEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteStructEnd", ctx), &err)
	return
}

func (p *protocol) WriteFieldBegin(ctx context.Context, name string, typeID int, id int16) (err error) {
	p.convert(p.call("WriteFieldBegin", ctx, name, typeID, id), &err)
	return
}

func (p *protocol) WriteFieldEnd(ctx context.Context) (err error) {
	p.convert(p.call("WriteFieldEnd", ctx), &err)
	return
}

func (p *protocol) WriteFieldStop(ctx context.Context) (err error) {
	p.convert(p.call("WriteFieldStop", ctx), &err)
	return
}

func (p *protocol) Flush(ctx context.Context) (err error) {
	p.convert(p.call("Flush", ctx), &err)
	return
}

var (
	contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
	ctx              = context.TODO()
	protocolType     = reflect.TypeOf((*protocol)(nil))
	protocols        sync.Map // reflect.Type => errro
	intType          = reflect.TypeOf((*int)(nil)).Elem()
)

func convert(x interface{}) (*protocol, error) {
	v := reflect.ValueOf(x)
	t := v.Type()
	if tmp, ok := protocols.Load(t); ok {
		if tmp != nil {
			return nil, tmp.(error)
		}
		return &protocol{impl: v}, nil
	}
	for i := 0; i < protocolType.NumMethod(); i++ {
		n := protocolType.Method(i).Name
		if !v.MethodByName(n).IsValid() {
			err := fmt.Errorf("type %s does not implement TProtocol (missing %s method)", t, n)
			protocols.LoadOrStore(t, err)
			return nil, err
		}
	}
	protocols.LoadOrStore(t, nil)
	return &protocol{impl: v}, nil
}

func asInt(x interface{}) int {
	v, ok := x.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(x)
	}
	i := v.Convert(intType)
	if i.IsValid() {
		return int(i.Int())
	}
	panic(fmt.Errorf("expected int or uint type, got %s", v.Type()))
}
