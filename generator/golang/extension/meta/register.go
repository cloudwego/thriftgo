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
	"reflect"
)

var (
	structs map[reflect.Type]*structType
	nul     = reflect.ValueOf(nil)
)

// RegisterStruct associates a constructor of a thrift struct type
// with some meta data to describes its meta data.
func RegisterStruct(newFunc interface{}, data []byte) {
	f := reflect.TypeOf(newFunc)
	if f.Kind() != reflect.Func || f.NumIn() != 0 || f.NumOut() != 1 {
		panic(fmt.Errorf("non creator: %T", newFunc))
	}

	rt := f.Out(0)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid type: %T", newFunc))
	}

	rt = rt.Elem()
	if structs[rt] != nil {
		panic(fmt.Errorf("already registered: %T", newFunc))
	}

	st := &structType{newFunc: reflect.ValueOf(newFunc)}
	if err := Unmarshal(data, &st.StructMeta); err != nil {
		panic(err)
	}
	structs[rt] = st
}

// Struct .
type Struct interface {
	Read(ctx context.Context, iprot Protocol) (err error)
	Write(ctx context.Context, oprot Protocol) (err error)
}

// AsStruct tries to wrap the given object into a Struct.
// The object is expected to be a pointer returned by a
// New function in the generated code.
func AsStruct(x interface{}) (Struct, error) {
	ptr := reflect.ValueOf(x)
	rt := ptr.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	st := structs[rt]
	if st == nil {
		return nil, fmt.Errorf("unregistered type: %s", rt)
	}
	return &instance{typ: st, ptr: ptr}, nil
}

type structType struct {
	StructMeta
	newFunc reflect.Value
}

type instance struct {
	typ *structType
	ptr reflect.Value
}

func (i *instance) Read(ctx context.Context, iprot Protocol) (err error) {
	absent := i.typ.StructMeta.requiredFields()
	_, err = iprot.ReadStructBegin(ctx)
	if err != nil {
		return fmt.Errorf("%s read struct begin: %w", i.ptr.Type(), err)
	}

	for {
		_, fieldTypeID, fieldID, err := iprot.ReadFieldBegin(ctx)
		if err != nil {
			return fmt.Errorf("%s read field begin: %w", i.ptr.Type(), err)
		}
		if fieldTypeID == TTypeID_STOP {
			break
		}

		idx := i.findField(fieldID, fieldTypeID)
		if idx != -1 {
			err = i.readField(ctx, iprot, idx)
			if err != nil {
				return fmt.Errorf("%s read field %d: %w", i.ptr.Type(), idx, err)
			}
			err = iprot.ReadFieldEnd(ctx)
			if err != nil {
				return fmt.Errorf("%s read field %d end: %w", i.ptr.Type(), idx, err)
			}
			delete(absent, fieldID)
			continue
		}

		if false /* TODO: support unknown fields */ {
			/* ... */
		} else {
			err = iprot.Skip(ctx, fieldTypeID)
			if err != nil {
				return fmt.Errorf("%s skip %s: %w", i.ptr.Type(), TTypeID(fieldTypeID), err)
			}
		}
	}

	err = iprot.ReadStructEnd(ctx)
	if err != nil {
		return fmt.Errorf("%s read struct end: %w", i.ptr.Type(), err)
	}

	if len(absent) > 0 {
		for _, idx := range absent {
			return fmt.Errorf("%s required field %s is not set", i.ptr.Type(), i.typ.StructMeta.Fields[idx].Name)
		}
	}
	return nil
}

func (i *instance) readField(ctx context.Context, iprot Protocol, index int) error {
	f := i.typ.Fields[index]
	p := i.ptr.Elem().Field(index)
	v, err := read(ctx, iprot, f.FieldType, p.Type())
	if err == nil {
		p.Set(v)
	}
	return err
}

func (i *instance) findField(fid int16, tid TTypeID) int {
	for idx, f := range i.typ.Fields {
		if f.FieldID == fid {
			if f.FieldType.TypeID == tid {
				return idx
			}
		}
	}
	return -1
}

func force(t reflect.Type, u interface{}) (p reflect.Value) {
	v, ok := u.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(u)
	}
	if t.Kind() == reflect.Ptr {
		et := t.Elem()
		p = reflect.New(et)
		p.Elem().Set(v.Convert(et))
	} else {
		p = v.Convert(t)
	}
	return
}

func read(ctx context.Context, iprot Protocol, tt *TypeMeta, gt reflect.Type) (v reflect.Value, err error) {
	switch tt.TypeID {
	case TTypeID_BOOL:
		v, err := iprot.ReadBool(ctx)
		return force(gt, v), err
	case TTypeID_BYTE:
		v, err := iprot.ReadByte(ctx)
		return force(gt, v), err
	case TTypeID_DOUBLE:
		v, err := iprot.ReadDouble(ctx)
		return force(gt, v), err
	case TTypeID_I16:
		v, err := iprot.ReadI16(ctx)
		return force(gt, v), err
	case TTypeID_I32:
		v, err := iprot.ReadI32(ctx)
		return force(gt, v), err
	case TTypeID_I64:
		v, err := iprot.ReadI64(ctx)
		return force(gt, v), err
	case TTypeID_STRING:
		if gt.Kind() == reflect.Slice { // binary
			v, err := iprot.ReadBinary(ctx)
			return force(gt, v), err
		}
		v, err := iprot.ReadString(ctx)
		return force(gt, v), err
	case TTypeID_MAP:
		_, _, size, err := iprot.ReadMapBegin(ctx)
		if err != nil {
			return nul, err
		}
		m := reflect.MakeMapWithSize(gt, size)
		for i := 0; i < size; i++ {
			k, err := read(ctx, iprot, tt.KeyType, gt.Key())
			if err != nil {
				return nul, err
			}
			v, err := read(ctx, iprot, tt.ValueType, gt.Elem())
			if err != nil {
				return nul, err
			}
			m.SetMapIndex(k, v)
		}
		return m, iprot.ReadMapEnd(ctx)
	case TTypeID_SET:
		_, size, err := iprot.ReadSetBegin(ctx)
		if err != nil {
			return nul, err
		}
		s := reflect.MakeSlice(gt, size, size)
		for i := 0; i < size; i++ {
			v, err := read(ctx, iprot, tt.ValueType, gt.Elem())
			if err != nil {
				return nul, err
			}
			s.Index(i).Set(v)
		}
		return s, iprot.ReadSetEnd(ctx)
	case TTypeID_LIST:
		_, size, err := iprot.ReadListBegin(ctx)
		if err != nil {
			return nul, err
		}
		s := reflect.MakeSlice(gt, size, size)
		for i := 0; i < size; i++ {
			v, err := read(ctx, iprot, tt.ValueType, gt.Elem())
			if err != nil {
				return nul, err
			}
			s.Index(i).Set(v)
		}
		return s, iprot.ReadListEnd(ctx)
	case TTypeID_STRUCT:
		if gt.Kind() != reflect.Ptr {
			panic(fmt.Errorf("expect pointer type to a struct: %s", gt))
		}
		st := structs[gt.Elem()]
		if st == nil {
			panic(fmt.Errorf("type not registered: %s", gt.Elem()))
		}
		v := st.newFunc.Call(nil)[0]
		i := &instance{typ: st, ptr: v}
		return v, i.Read(ctx, iprot)
	default:
		panic(fmt.Errorf("invalid typeID: %d", tt.TypeID))
	}
}

func (i *instance) countFields() error {
	if i.typ.Category != "union" || i.ptr.IsNil() {
		return nil
	}
	count := 0
	for idx := range i.typ.Fields {
		f := i.ptr.Elem().Field(idx)
		if !f.IsZero() {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%s write union: exactly one field must be set (%d set)", i.ptr.Type(), count)
	}
	return nil
}

func (i *instance) Write(ctx context.Context, oprot Protocol) (err error) {
	if err = i.countFields(); err != nil {
		return
	}
	if err = oprot.WriteStructBegin(ctx, i.typ.Name); err != nil {
		return fmt.Errorf("%s write struct begin: %w", i.ptr.Type(), err)
	}
	if !i.ptr.IsNil() {
		n := i.ptr.Elem().NumField()
		for idx := 0; idx < n; idx++ {
			err = i.writeField(ctx, oprot, idx)
			if err != nil {
				return fmt.Errorf("%s write field %d: %w", i.ptr.Type(), idx, err)
			}
		}
	}
	if err = oprot.WriteFieldStop(ctx); err != nil {
		return fmt.Errorf("%s write field stop: %w", i.ptr.Type(), err)
	}
	if err = oprot.WriteStructEnd(ctx); err != nil {
		return fmt.Errorf("%s write struct end: %w", i.ptr.Type(), err)
	}
	return nil
}

func (i *instance) writeField(ctx context.Context, oprot Protocol, index int) (err error) {
	f := i.typ.Fields[index]
	p := i.ptr.Elem().Field(index)

	if f.Requiredness == TRequiredness_OPTIONAL && p.IsZero() {
		return nil
	}

	if err = oprot.WriteFieldBegin(ctx, f.Name, f.FieldType.TypeID, f.FieldID); err != nil {
		return err
	}
	if err = write(ctx, oprot, f.FieldType, p); err != nil {
		return err
	}
	if err = oprot.WriteFieldEnd(ctx); err != nil {
		return err
	}
	return nil
}

func write(ctx context.Context, oprot Protocol, tt *TypeMeta, gv reflect.Value) error {
	if tt.TypeID == TTypeID_STRUCT {
		rt := gv.Type().Elem()
		st := structs[rt]
		if st == nil {
			panic(fmt.Errorf("type not registered: %s", rt))
		}
		i := &instance{typ: st, ptr: gv}
		return i.Write(ctx, oprot)
	}

	if gv.Kind() == reflect.Ptr {
		gv = gv.Elem()
	}
	switch tt.TypeID {
	case TTypeID_BOOL:
		return oprot.WriteBool(ctx, gv.Bool())
	case TTypeID_BYTE:
		return oprot.WriteByte(ctx, int8(gv.Int()))
	case TTypeID_DOUBLE:
		return oprot.WriteDouble(ctx, gv.Float())
	case TTypeID_I16:
		return oprot.WriteI16(ctx, int16(gv.Int()))
	case TTypeID_I32:
		return oprot.WriteI32(ctx, int32(gv.Int()))
	case TTypeID_I64:
		return oprot.WriteI64(ctx, int64(gv.Int()))
	case TTypeID_STRING:
		if gv.Kind() == reflect.Slice { // binary
			return oprot.WriteBinary(ctx, gv.Bytes())
		}
		return oprot.WriteString(ctx, gv.String())
	case TTypeID_MAP:
		if err := oprot.WriteMapBegin(ctx, tt.KeyType.TypeID, tt.ValueType.TypeID, gv.Len()); err != nil {
			return err
		}
		iter := gv.MapRange()
		for iter.Next() {
			if err := write(ctx, oprot, tt.KeyType, iter.Key()); err != nil {
				return err
			}
			if err := write(ctx, oprot, tt.ValueType, iter.Value()); err != nil {
				return err
			}
		}
		if err := oprot.WriteMapEnd(ctx); err != nil {
			return err
		}
	case TTypeID_SET:
		size := gv.Len()
		if err := oprot.WriteSetBegin(ctx, tt.ValueType.TypeID, size); err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			if err := write(ctx, oprot, tt.ValueType, gv.Index(i)); err != nil {
				return err
			}
		}
		if err := oprot.WriteSetEnd(ctx); err != nil {
			return err
		}
	case TTypeID_LIST:
		size := gv.Len()
		if err := oprot.WriteListBegin(ctx, tt.ValueType.TypeID, size); err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			if err := write(ctx, oprot, tt.ValueType, gv.Index(i)); err != nil {
				return err
			}
		}
		if err := oprot.WriteListEnd(ctx); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("invalid typeID: %d", tt.TypeID))
	}
	return nil
}
