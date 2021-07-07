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

// Package unknown provides definitions that work with the thriftgo `keep_unknown_fields` option.
// When the option is turned on, thriftgo generates an extra field `_unknownFields` for each structure
// defined in the IDL to store fields that are not recognized by the current IDL when deserializing data.
// Those unknown fields will be written out at the end of the data stream when serializing the structure
// that carries them.
package unknown

import (
	"errors"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// WithUnknownFields is the interface of all structures that supports keeping unknown fields.
type WithUnknownFields interface {
	// CarryingUnknownFields tells whether the structure contains data from fields not recognized by the current IDL.
	CarryingUnknownFields() bool
}

// errors .
var (
	ErrExceedDepthLimit = thrift.NewTProtocolExceptionWithType(thrift.DEPTH_LIMIT, errors.New("depth limit exceeded"))

	ErrUnknownType = func(t thrift.TType) error {
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("unknown data type %d", t))
	}

	maxNestingDepth = 64
)

// SetNestingDepthLimit sets the max number of nesting level.
func SetNestingDepthLimit(d int) {
	maxNestingDepth = d
}

// Field is used to store unrecognized field when deserializing data.
type Field struct {
	Name    string
	ID      int16
	Type    thrift.TType
	KeyType thrift.TType
	ValType thrift.TType
	Value   interface{}
}

// Fields is a list of Field.
type Fields []*Field

// Append reads an unrecognized field and append it to the current slice.
func (fs *Fields) Append(iprot thrift.TProtocol, name string, fieldType thrift.TType, id int16) error {
	f, err := Read(iprot, name, fieldType, id, maxNestingDepth)
	if err != nil {
		return err
	}
	*fs = append(*fs, f)
	return nil
}

// Write writes out the unknown fields.
func (fs *Fields) Write(oprot thrift.TProtocol) (err error) {
	var i int
	var f *Field
	for i, f = range *fs {
		if err = oprot.WriteFieldBegin(f.Name, f.Type, f.ID); err != nil {
			break
		}
		if err = Write(oprot, f); err != nil {
			break
		}
		if err = oprot.WriteFieldEnd(); err != nil {
			break
		}
	}
	if err != nil {
		msg := fmt.Sprintf("write field error unknown.%d(name:%s type:%d id:%d): ", i, f.Name, f.Type, f.ID)
		err = thrift.PrependError(msg, err)
	}
	return err
}

// Write writes out the unknown field.
func Write(oprot thrift.TProtocol, f *Field) (err error) {
	switch f.Type {
	case thrift.BOOL:
		return oprot.WriteBool(f.Value.(bool))
	case thrift.BYTE:
		return oprot.WriteByte(f.Value.(int8))
	case thrift.DOUBLE:
		return oprot.WriteDouble(f.Value.(float64))
	case thrift.I16:
		return oprot.WriteI16(f.Value.(int16))
	case thrift.I32:
		return oprot.WriteI32(f.Value.(int32))
	case thrift.I64:
		return oprot.WriteI64(f.Value.(int64))
	case thrift.STRING:
		return oprot.WriteString(f.Value.(string))
	case thrift.SET:
		vs := f.Value.([]*Field)
		if err = oprot.WriteSetBegin(f.ValType, len(vs)); err != nil {
			return thrift.PrependError("write set begin error: ", err)
		}
		for _, v := range vs {
			if err = Write(oprot, v); err != nil {
				return thrift.PrependError("write set elem error: ", err)
			}
		}
		if err = oprot.WriteSetEnd(); err != nil {
			return thrift.PrependError("write set end error: ", err)
		}
	case thrift.LIST:
		vs := f.Value.([]*Field)
		if err = oprot.WriteListBegin(f.ValType, len(vs)); err != nil {
			return thrift.PrependError("write list begin error: ", err)
		}
		for _, v := range vs {
			if err = Write(oprot, v); err != nil {
				return thrift.PrependError("write list elem error: ", err)
			}
		}
		if err = oprot.WriteListEnd(); err != nil {
			return thrift.PrependError("write list end error: ", err)
		}
	case thrift.MAP:
		kvs := f.Value.([]*Field)
		if err = oprot.WriteMapBegin(f.KeyType, f.ValType, len(kvs)/2); err != nil {
			return thrift.PrependError("write map begin error: ", err)
		}
		for i := 0; i < len(kvs); i += 2 {
			if err = Write(oprot, kvs[i]); err != nil {
				return thrift.PrependError("write map key error: ", err)
			}
			if err = Write(oprot, kvs[i+1]); err != nil {
				return thrift.PrependError("write map value error: ", err)
			}
		}
		if err = oprot.WriteMapEnd(); err != nil {
			return thrift.PrependError("write map end error: ", err)
		}
	case thrift.STRUCT:
		fs := Fields(f.Value.([]*Field))
		if err = oprot.WriteStructBegin(f.Name); err != nil {
			return thrift.PrependError("write struct begin error: ", err)
		}
		if err = fs.Write(oprot); err != nil {
			return thrift.PrependError("write struct field error: ", err)
		}
		if err = oprot.WriteFieldStop(); err != nil {
			return thrift.PrependError("write struct stop error: ", err)
		}
		if err = oprot.WriteStructEnd(); err != nil {
			return thrift.PrependError("write struct end error: ", err)
		}
	default:
		return ErrUnknownType(f.Type)
	}
	return
}

// Read reads an unknown field from the given TProtocol.
func Read(iprot thrift.TProtocol, name string, fieldType thrift.TType, id int16, maxDepth int) (f *Field, err error) {
	if maxDepth <= 0 {
		return nil, ErrExceedDepthLimit
	}

	var size int
	f = &Field{Name: name, ID: id, Type: fieldType}
	switch fieldType {
	case thrift.BOOL:
		f.Value, err = iprot.ReadBool()
	case thrift.BYTE:
		f.Value, err = iprot.ReadByte()
	case thrift.I16:
		f.Value, err = iprot.ReadI16()
	case thrift.I32:
		f.Value, err = iprot.ReadI32()
	case thrift.I64:
		f.Value, err = iprot.ReadI64()
	case thrift.DOUBLE:
		f.Value, err = iprot.ReadDouble()
	case thrift.STRING:
		f.Value, err = iprot.ReadString()
	case thrift.SET:
		f.ValType, size, err = iprot.ReadSetBegin()
		if err != nil {
			return nil, thrift.PrependError("read set begin error: ", err)
		}
		set := make([]*Field, 0, size)
		for i := 0; i < size; i++ {
			v, err2 := Read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, thrift.PrependError("read set elem error: ", err)
			}
			set = append(set, v)
		}
		if err = iprot.ReadSetEnd(); err != nil {
			return nil, thrift.PrependError("read set end error: ", err)
		}
		f.Value = set
	case thrift.LIST:
		f.ValType, size, err = iprot.ReadListBegin()
		if err != nil {
			return nil, thrift.PrependError("read list begin error: ", err)
		}
		list := make([]*Field, 0, size)
		for i := 0; i < size; i++ {
			v, err2 := Read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, thrift.PrependError("read list elem error: ", err)
			}
			list = append(list, v)
		}
		if err = iprot.ReadListEnd(); err != nil {
			return nil, thrift.PrependError("read list end error: ", err)
		}
		f.Value = list
	case thrift.MAP:
		f.KeyType, f.ValType, size, err = iprot.ReadMapBegin()
		if err != nil {
			return nil, thrift.PrependError("read map begin error: ", err)
		}
		flatMap := make([]*Field, 0, size*2)
		for i := 0; i < size; i++ {
			k, err2 := Read(iprot, "", f.KeyType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, thrift.PrependError("read map key error: ", err)
			}
			v, err2 := Read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, thrift.PrependError("read map value error: ", err)
			}
			flatMap = append(flatMap, k, v)
		}
		if err = iprot.ReadMapEnd(); err != nil {
			return nil, thrift.PrependError("read map end error: ", err)
		}
		f.Value = flatMap
	case thrift.STRUCT:
		_, err = iprot.ReadStructBegin()
		if err != nil {
			return nil, thrift.PrependError("read struct begin error: ", err)
		}
		var fields []*Field
		for {
			name, fieldTypeID, fieldID, err := iprot.ReadFieldBegin()
			if err != nil {
				return nil, thrift.PrependError("read field begin error: ", err)
			}
			if fieldTypeID == thrift.STOP {
				break
			}
			v, err := Read(iprot, name, fieldTypeID, fieldID, maxDepth-1)
			if err != nil {
				return nil, thrift.PrependError("read struct field error: ", err)
			}
			if err := iprot.ReadFieldEnd(); err != nil {
				return nil, thrift.PrependError("read field end error: ", err)
			}
			fields = append(fields, v)
		}
		if err = iprot.ReadStructEnd(); err != nil {
			return nil, thrift.PrependError("read struct end error: ", err)
		}
		f.Value = fields
	default:
		return nil, ErrUnknownType(fieldType)
	}
	if err != nil {
		return nil, err
	}
	return
}
