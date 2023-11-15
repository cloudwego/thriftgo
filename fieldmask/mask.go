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
	"sync"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

type fieldMaskType uint8

const (
	ftInvalid fieldMaskType = iota
	ftScalar
	ftArray
	ftStruct
	ftStrMap
	ftIntMap
)

// FieldMask represents a collection of thrift pathes
// See
type FieldMask struct {
	typ fieldMaskType

	isAll bool

	all *FieldMask

	fdMask *fieldMap

	strMask strMap

	intMask intMap
}

var fmsPool = sync.Pool{
	New: func() interface{} {
		return &FieldMask{}
	},
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

// GetFieldMask reuse fieldmask from pool
func GetFieldMask(desc *thrift_reflection.TypeDescriptor, paths ...string) (*FieldMask, error) {
	ret := fmsPool.Get().(*FieldMask)
	err := ret.init(desc, paths...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Recycle puts fieldmask into pool
func (self *FieldMask) Recycle() {
	self.Reset()
	fmsPool.Put(self)
}

// Reset clears fieldmask's all path
func (self *FieldMask) Reset() {
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
// For example:
// pathes `[]string{"$.Extra[0].List", "$.Extra[*].Set", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr"}` will print:
//
//	  (Base)
//		.Extra (list<ExtraInfo>)
//		[
//		  *
//		]
//		.Meta (MetaInfo)
//		  .F2 (map<i8,Base>)
//		  {
//		    *
//		  }
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

// FieldInMask tells if a id in the mask
func (self *FieldMask) FieldInMask(id int16) bool {
	return !self.Exist() || self.isAll || (self.typ == ftStruct && self.fdMask.Get(fieldID(id)) != nil)
}

// IntInMask tells if a index in the mask
func (self *FieldMask) IntInMask(id int) bool {
	return !self.Exist() || self.isAll || ((self.typ == ftArray || self.typ == ftIntMap) && (self.intMask.Get(id) != nil))
}

// StrInMask tells if a string in the mask
func (self *FieldMask) StrInMask(id string) bool {
	return !self.Exist() || self.isAll || (self.typ == ftStrMap && (self.strMask.Get(id) != nil))
}

// Field returns the specific sub mask for given id
func (self *FieldMask) Field(id int16) *FieldMask {
	if self == nil || self.typ == 0 {
		return nil
	}
	if self.isAll {
		return self.all
	}
	return self.fdMask.Get(fieldID(id))
}

// Int returns the specific sub mask for given index
func (self *FieldMask) Int(id int) *FieldMask {
	if self == nil || self.typ == 0 {
		return nil
	}
	if self.isAll {
		return self.all
	}
	return self.intMask.Get(id)
}

// Field returns the specific sub mask for given string
func (self *FieldMask) Str(id string) *FieldMask {
	if self == nil || self.typ == 0 {
		return nil
	}
	if self.isAll {
		return self.all
	}
	return self.strMask.Get(id)
}

// All tells if the mask allows all elements pass (*)
func (self *FieldMask) All() bool {
	if self == nil {
		return true
	}
	switch self.typ {
	case ftStruct, ftArray, ftIntMap, ftStrMap:
		return self.isAll
	default:
		return true
	}
}
