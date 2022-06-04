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
	"reflect"
)

// Marshal serializes the object with binary protocol.
func Marshal(obj interface{}) ([]byte, error) {
	x, err := AsStruct(obj)
	if err != nil {
		return nil, err
	}

	mem := new(MemoryTransport)
	oprot := NewBinaryProtocol(mem)

	err = x.Write(context.Background(), oprot)
	if err != nil {
		return nil, err
	}
	return mem.Bytes(), nil
}

// Unmarshal deserializes the data from bytes with binary protocol.
func Unmarshal(data []byte, obj interface{}) error {
	x, err := AsStruct(obj)
	if err != nil {
		return err
	}

	mem := new(MemoryTransport)
	mem.Write(data)
	iprot := NewBinaryProtocol(mem)

	return x.Read(context.Background(), iprot)
}

func (sm *StructMeta) requiredFields() map[int16]int {
	m := make(map[int16]int)
	for i, f := range sm.Fields {
		if f.Requiredness == TRequiredness_REQUIRED {
			m[f.FieldID] = i
		}
	}
	return m
}

func init() {
	structs = map[reflect.Type]*structType{
		reflect.TypeOf(TypeMeta{}): {
			newFunc: reflect.ValueOf(NewTypeMeta),
			StructMeta: StructMeta{
				Name:     string("TypeMeta"),
				Category: string("struct"),
				Fields: []*FieldMeta{
					{
						FieldID:      int16(1),
						Name:         string("type_id"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(8),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(2),
						Name:         string("key_type"),
						Requiredness: TRequiredness(2),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(12),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(3),
						Name:         string("value_type"),
						Requiredness: TRequiredness(2),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(12),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
				},
			},
		},
		reflect.TypeOf(FieldMeta{}): {
			newFunc: reflect.ValueOf(NewFieldMeta),
			StructMeta: StructMeta{
				Name:     string("FieldMeta"),
				Category: string("struct"),
				Fields: []*FieldMeta{
					{
						FieldID:      int16(1),
						Name:         string("field_id"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(6),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(2),
						Name:         string("name"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(11),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(3),
						Name:         string("requiredness"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(8),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(4),
						Name:         string("field_type"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(12),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
				},
			},
		},
		reflect.TypeOf(StructMeta{}): {
			newFunc: reflect.ValueOf(NewStructMeta),
			StructMeta: StructMeta{
				Name:     string("StructMeta"),
				Category: string("struct"),
				Fields: []*FieldMeta{
					{
						FieldID:      int16(1),
						Name:         string("name"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(11),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(2),
						Name:         string("category"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:    TTypeID(11),
							KeyType:   (*TypeMeta)(nil),
							ValueType: (*TypeMeta)(nil),
						},
					},
					{
						FieldID:      int16(3),
						Name:         string("fields"),
						Requiredness: TRequiredness(1),
						FieldType: &TypeMeta{
							TypeID:  TTypeID(15),
							KeyType: (*TypeMeta)(nil),
							ValueType: &TypeMeta{
								TypeID:    TTypeID(12),
								KeyType:   (*TypeMeta)(nil),
								ValueType: (*TypeMeta)(nil),
							},
						},
					},
				},
			},
		},
	}
}
