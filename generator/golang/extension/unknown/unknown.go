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
)

// WithUnknownFields is the interface of all structures that supports keeping unknown fields.
type WithUnknownFields interface {
	// CarryingUnknownFields tells whether the structure contains data from fields not recognized by the current IDL.
	CarryingUnknownFields() bool
}

// errors .
var (
	ErrExceedDepthLimit = errors.New("depth limit exceeded")

	ErrUnknownType = func(t int) error {
		return fmt.Errorf("unknown data type %d", t)
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
	Type    int
	KeyType int
	ValType int
	Value   interface{}
}

// Fields is a list of Field.
type Fields []*Field

// Append reads an unrecognized field and append it to the current slice.
func (fs *Fields) Append(xprot TProtocol, name string, fieldType TType, id int16) error {
	iprot, err := convert(xprot)
	if err != nil {
		return err
	}
	f, err := read(iprot, name, asInt(fieldType), id, maxNestingDepth)
	if err != nil {
		return err
	}
	*fs = append(*fs, f)
	return nil
}

// Write writes out the unknown fields.
func (fs *Fields) Write(xprot TProtocol) (err error) {
	oprot, err := convert(xprot)
	if err != nil {
		return err
	}
	var i int
	var f *Field
	for i, f = range *fs {
		if err = oprot.WriteFieldBegin(ctx, f.Name, f.Type, f.ID); err != nil {
			break
		}
		if err = write(oprot, f); err != nil {
			break
		}
		if err = oprot.WriteFieldEnd(ctx); err != nil {
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("write field error unknown.%d(name:%s type:%d id:%d): %w",
			i, f.Name, f.Type, f.ID, err)
	}
	return err
}

// write writes out the unknown field.
func write(oprot *protocol, f *Field) (err error) {
	switch f.Type {
	case TBool:
		return oprot.WriteBool(ctx, f.Value.(bool))
	case TByte:
		return oprot.WriteByte(ctx, f.Value.(int8))
	case TDouble:
		return oprot.WriteDouble(ctx, f.Value.(float64))
	case TI16:
		return oprot.WriteI16(ctx, f.Value.(int16))
	case TI32:
		return oprot.WriteI32(ctx, f.Value.(int32))
	case TI64:
		return oprot.WriteI64(ctx, f.Value.(int64))
	case TString:
		return oprot.WriteString(ctx, f.Value.(string))
	case TSet:
		vs := f.Value.([]*Field)
		if err = oprot.WriteSetBegin(ctx, f.ValType, len(vs)); err != nil {
			return fmt.Errorf("write set begin error: %w", err)
		}
		for _, v := range vs {
			if err = write(oprot, v); err != nil {
				return fmt.Errorf("write set elem error: %w", err)
			}
		}
		if err = oprot.WriteSetEnd(ctx); err != nil {
			return fmt.Errorf("write set end error: %w", err)
		}
	case TList:
		vs := f.Value.([]*Field)
		if err = oprot.WriteListBegin(ctx, f.ValType, len(vs)); err != nil {
			return fmt.Errorf("write list begin error: %w", err)
		}
		for _, v := range vs {
			if err = write(oprot, v); err != nil {
				return fmt.Errorf("write list elem error: %w", err)
			}
		}
		if err = oprot.WriteListEnd(ctx); err != nil {
			return fmt.Errorf("write list end error: %w", err)
		}
	case TMap:
		kvs := f.Value.([]*Field)
		if err = oprot.WriteMapBegin(ctx, f.KeyType, f.ValType, len(kvs)/2); err != nil {
			return fmt.Errorf("write map begin error: %w", err)
		}
		for i := 0; i < len(kvs); i += 2 {
			if err = write(oprot, kvs[i]); err != nil {
				return fmt.Errorf("write map key error: %w", err)
			}
			if err = write(oprot, kvs[i+1]); err != nil {
				return fmt.Errorf("write map value error: %w", err)
			}
		}
		if err = oprot.WriteMapEnd(ctx); err != nil {
			return fmt.Errorf("write map end error: %w", err)
		}
	case TStruct:
		fs := Fields(f.Value.([]*Field))
		if err = oprot.WriteStructBegin(ctx, f.Name); err != nil {
			return fmt.Errorf("write struct begin error: %w", err)
		}
		if err = fs.Write(oprot); err != nil {
			return fmt.Errorf("write struct field error: %w", err)
		}
		if err = oprot.WriteFieldStop(ctx); err != nil {
			return fmt.Errorf("write struct stop error: %w", err)
		}
		if err = oprot.WriteStructEnd(ctx); err != nil {
			return fmt.Errorf("write struct end error: %w", err)
		}
	default:
		return ErrUnknownType(f.Type)
	}
	return
}

// read reads an unknown field from the given TProtocol.
func read(iprot *protocol, name string, fieldType int, id int16, maxDepth int) (f *Field, err error) {
	if maxDepth <= 0 {
		return nil, ErrExceedDepthLimit
	}

	var size int
	f = &Field{Name: name, ID: id, Type: asInt(fieldType)}
	switch fieldType {
	case TBool:
		f.Value, err = iprot.ReadBool(ctx)
	case TByte:
		f.Value, err = iprot.ReadByte(ctx)
	case TI16:
		f.Value, err = iprot.ReadI16(ctx)
	case TI32:
		f.Value, err = iprot.ReadI32(ctx)
	case TI64:
		f.Value, err = iprot.ReadI64(ctx)
	case TDouble:
		f.Value, err = iprot.ReadDouble(ctx)
	case TString:
		f.Value, err = iprot.ReadString(ctx)
	case TSet:
		f.ValType, size, err = iprot.ReadSetBegin(ctx)
		if err != nil {
			return nil, fmt.Errorf("read set begin error: %w", err)
		}
		set := make([]*Field, 0, size)
		for i := 0; i < size; i++ {
			v, err2 := read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, fmt.Errorf("read set elem error: %w", err)
			}
			set = append(set, v)
		}
		if err = iprot.ReadSetEnd(ctx); err != nil {
			return nil, fmt.Errorf("read set end error: %w", err)
		}
		f.Value = set
	case TList:
		f.ValType, size, err = iprot.ReadListBegin(ctx)
		if err != nil {
			return nil, fmt.Errorf("read list begin error: %w", err)
		}
		list := make([]*Field, 0, size)
		for i := 0; i < size; i++ {
			v, err2 := read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, fmt.Errorf("read list elem error: %w", err)
			}
			list = append(list, v)
		}
		if err = iprot.ReadListEnd(ctx); err != nil {
			return nil, fmt.Errorf("read list end error: %w", err)
		}
		f.Value = list
	case TMap:
		f.KeyType, f.ValType, size, err = iprot.ReadMapBegin(ctx)
		if err != nil {
			return nil, fmt.Errorf("read map begin error: %w", err)
		}
		flatMap := make([]*Field, 0, size*2)
		for i := 0; i < size; i++ {
			k, err2 := read(iprot, "", f.KeyType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, fmt.Errorf("read map key error: %w", err)
			}
			v, err2 := read(iprot, "", f.ValType, int16(i), maxDepth-1)
			if err2 != nil {
				return nil, fmt.Errorf("read map value error: %w", err)
			}
			flatMap = append(flatMap, k, v)
		}
		if err = iprot.ReadMapEnd(ctx); err != nil {
			return nil, fmt.Errorf("read map end error: %w", err)
		}
		f.Value = flatMap
	case TStruct:
		_, err = iprot.ReadStructBegin(ctx)
		if err != nil {
			return nil, fmt.Errorf("read struct begin error: %w", err)
		}
		var fields []*Field
		for {
			name, fieldTypeID, fieldID, err := iprot.ReadFieldBegin(ctx)
			if err != nil {
				return nil, fmt.Errorf("read field begin error: %w", err)
			}
			if fieldTypeID == TStop {
				break
			}
			v, err := read(iprot, name, fieldTypeID, fieldID, maxDepth-1)
			if err != nil {
				return nil, fmt.Errorf("read struct field error: %w", err)
			}
			if err := iprot.ReadFieldEnd(ctx); err != nil {
				return nil, fmt.Errorf("read field end error: %w", err)
			}
			fields = append(fields, v)
		}
		if err = iprot.ReadStructEnd(ctx); err != nil {
			return nil, fmt.Errorf("read struct end error: %w", err)
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
