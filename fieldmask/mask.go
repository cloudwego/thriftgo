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

// FieldMask represents a collection of field paths
type FieldMask struct {
	typ fieldMaskType

	fieldMask *fieldMaskBitmap
	fields    *fieldMap

	strMask strMap

	intMask intMap

	elem *FieldMask
}

var fmsPool = sync.Pool{
	New: func() interface{} {
		return &FieldMask{}
	},
}

func NewFieldMask(desc *thrift_reflection.TypeDescriptor, pathes ...string) (*FieldMask, error) {
	ret := FieldMask{}
	err := ret.init(desc, pathes...)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// GetFieldMask make a new FieldMask from paths and root descriptor,
// each path is the combination of field names from root struct to any layer of its children, separated with PathSep
func GetFieldMask(desc *thrift_reflection.TypeDescriptor, paths ...string) (*FieldMask, error) {
	ret := fmsPool.Get().(*FieldMask)
	err := ret.init(desc, paths...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (self *FieldMask) Recycle() {
	self.Reset()
	fmsPool.Put(self)
}

func (self *FieldMask) Reset() {
	if self == nil {
		return
	}
	*self.fieldMask = (*self.fieldMask)[:0]
	self.fields.Reset()
}

func (self *FieldMask) init(desc *thrift_reflection.TypeDescriptor, paths ...string) error {
	// horizontal traversal...
	for _, path := range paths {
		if err := self.setPath(path, desc); err != nil {
			return err
		}
	}
	return nil
}

// String pretty prints the structure a FieldMask represents
func (self FieldMask) String(desc *thrift_reflection.TypeDescriptor) string {
	buf := strings.Builder{}
	buf.WriteString("(")
	buf.WriteString(desc.GetName())
	buf.WriteString(")\n")
	self.print(&buf, 0, desc)
	return buf.String()
}

func (self *FieldMask) FieldInMask(id int32) bool {
	return self == nil || self.fieldMask == nil || self.fieldMask.Get(fieldID(id))
}

func (self *FieldMask) IntInMask(id int) bool {
	return self == nil || self.intMask == nil || self.intMask.Get(id)
}

func (self *FieldMask) StrInMask(id string) bool {
	return self == nil || self.strMask == nil || self.strMask.Get(id)
}

func (self *FieldMask) Field(id int32) *FieldMask {
	if self == nil || self.fields == nil {
		return nil
	}
	return self.fields.Get(fieldID(id))
}

func (self *FieldMask) Elem() *FieldMask {
	if self == nil {
		return nil
	}
	return self.elem
}

func (self *FieldMask) setFieldID(f fieldID) {
	if self.fieldMask == nil {
		m := make(fieldMaskBitmap, 0)
		self.fieldMask = &m
	}
	self.fieldMask.Set(f)
}

func (self *FieldMask) setInt(v int) {
	if self.intMask == nil {
		self.intMask = make(intMap)
	}
	self.intMask.Set(v)
}

func (self *FieldMask) setStr(v string) {
	if self.strMask == nil {
		self.strMask = make(strMap)
	}
	self.strMask.Set(v)
}

func (self *FieldMask) getField(f fieldID, st *thrift_reflection.StructDescriptor) *FieldMask {
	if self.fields == nil {
		m := makeFieldMaskMap(st)
		self.fields = &m
	}
	return self.fields.GetOrAlloc(f)
}

func (self *FieldMask) getElem() *FieldMask {
	if self.elem == nil {
		self.elem = &FieldMask{}
	}
	return self.elem
}
