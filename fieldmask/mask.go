/*
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fieldmask

import (
	"fmt"
	"strings"

	"github.com/cloudwego/thriftgo/internal/utils"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

// FieldMaskType indicates the corresponding thrift message type for a fieldmask
type FieldMaskType uint8

// MarshalText implements encoding.TextMarshaler
func (ft FieldMaskType) MarshalText() ([]byte, error) {
	switch ft {
	case FtScalar:
		return utils.S2B("Scalar"), nil
	case FtList:
		return utils.S2B("List"), nil
	case FtStruct:
		return utils.S2B("Struct"), nil
	case FtStrMap:
		return utils.S2B("StrMap"), nil
	case FtIntMap:
		return utils.S2B("IntMap"), nil
	default:
		return utils.S2B("Invalid"), nil
	}
}

// UnmarshalText implements encoding.TextUnmarshaler
func (ft *FieldMaskType) UnmarshalText(in []byte) error {
	switch utils.B2S(in) {
	case "Scalar":
		*ft = FtScalar
	case "List":
		*ft = FtList
	case "Struct":
		*ft = FtStruct
	case "StrMap":
		*ft = FtStrMap
	case "IntMap":
		*ft = FtIntMap
	default:
		*ft = FtInvalid
	}
	return nil
}

// FieldMaskType Enums
const (
	// Invalid or unsupported thrift type
	FtInvalid FieldMaskType = iota
	// thrift scalar types, including BOOL/I8/I16/I32/I64/DOUBLE/STRING/BINARY, or neither-string-nor-integer-typed-key MAP
	FtScalar
	// thrift LIST/SET
	FtList
	// thrift STRUCT
	FtStruct
	// thrift MAP with string-typed key
	FtStrMap
	// thrift MAP with integer-typed key
	FtIntMap
)

// FieldMask represents a collection of thrift pathes
// See
type FieldMask struct {
	isAll bool

	typ FieldMaskType

	all *FieldMask

	fdMask *fieldMap

	strMask strMap

	intMask intMap
}

// NewFieldMask create a new fieldmask
func NewFieldMask(desc *thrift_reflection.TypeDescriptor, pathes ...string) (*FieldMask, error) {
	ret := FieldMask{}
	err := ret.init(desc, pathes...)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// reset clears fieldmask's all path
func (self *FieldMask) reset() {
	if self == nil {
		return
	}
	self.isAll = false
	self.typ = 0
	self.fdMask.Reset()
	self.intMask.Reset()
	self.strMask.Reset()
}

func (self *FieldMask) init(desc *thrift_reflection.TypeDescriptor, paths ...string) error {
	// horizontal traversal...
	for _, path := range paths {
		if err := self.addPath(path, desc); err != nil {
			return fmt.Errorf("Parsing path %q  error: %v", path, err)
		}
	}
	return nil
}

// String pretty prints the structure a FieldMask represents
//
// WARING: This is unstable API, the printer format is not guaranteed
func (self FieldMask) String(desc *thrift_reflection.TypeDescriptor) string {
	buf := strings.Builder{}
	buf.WriteString("(")
	buf.WriteString(desc.GetName())
	buf.WriteString(")\n")
	self.print(&buf, 0, desc)
	return buf.String()
}

// Exist tells if the fieldmask is setted
func (self *FieldMask) Exist() bool {
	return self != nil && self.typ != 0
}

// Field returns the specific sub mask for a given id, and tells if the id in the mask
func (self *FieldMask) Field(id int16) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.fdMask.Get(fieldID(id))
	return fm, fm != nil
}

// Int returns the specific sub mask for a given index, and tells if the index in the mask
func (self *FieldMask) Int(id int) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.intMask.Get(id)
	return fm, fm != nil
}

// Field returns the specific sub mask for a given string, and tells if the string in the mask
func (self *FieldMask) Str(id string) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.strMask.Get(id)
	return fm, fm != nil
}

// All tells if the mask allows all elements pass (*)
func (self *FieldMask) All() bool {
	if self == nil {
		return true
	}
	switch self.typ {
	case FtStruct, FtList, FtIntMap, FtStrMap:
		return self.isAll
	default:
		return true
	}
}
