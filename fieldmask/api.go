/**
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

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/utils"
)

type fieldID int32

const PathSep = '.'

const _MaxFieldIDHead = 255

type fieldMaskMap struct {
	head [_MaxFieldIDHead + 1]*FieldMask
	tail map[fieldID]*FieldMask
}

func makeFieldMaskMap(fields []*parser.Field) fieldMaskMap {
	max := 0
	count := 0
	for _, f := range fields {
		if max < int(f.GetID()) {
			max = int(f.GetID())
			count = 0
		} else {
			count += 1
		}
	}
	return fieldMaskMap{
		tail: make(map[fieldID]*FieldMask, count),
	}
}

func (self fieldMaskMap) IsInitialized() bool {
	return self.tail != nil
}

func (self *fieldMaskMap) GetOrAlloc(f fieldID) *FieldMask {
	if f <= _MaxFieldIDHead {
		s := self.head[f]
		if s == nil {
			s = &FieldMask{}
			self.head[f] = s
		}
		return s
	} else {
		s := self.tail[f]
		if s == nil {
			s = &FieldMask{}
			self.tail[f] = s
		}
		return s
	}
}

func (self *fieldMaskMap) Get(f fieldID) *FieldMask {
	if f <= _MaxFieldIDHead {
		return self.head[f]
	} else {
		return self.tail[f]
	}
}

type fieldMaskBitmap []byte

const _BucketBit = 8

func (self *fieldMaskBitmap) Set(f fieldID) {
	b := int(f / _BucketBit)
	i := int(f % _BucketBit)
	c := cap(*self)
	if c <= b+1 {
		tmp := make([]byte, len(*self), (c + b + 1))
		copy(tmp, *self)
		*self = tmp
	}
	if len(*self) <= b {
		*self = (*self)[:b+1]
	}
	(*self)[b] |= byte(1 << i)
}

func (self *fieldMaskBitmap) Get(f fieldID) bool {
	b := int(f / _BucketBit)
	if len(*self) <= b {
		return false
	}
	i := int(f % _BucketBit)
	return ((*self)[b] & byte(1<<i)) != 0
}

type FieldMask struct {
	// current layer of mask
	flat fieldMaskBitmap
	// for lookup next layer of fieldmasks
	next fieldMaskMap
	desc *parser.StructLike
}

func NewFieldMaskFromAST(IDL *parser.Thrift, rootStruct string, paths ...string) *FieldMask {
	if IDL == nil {
		panic("FieldMask must have a IDL!")
	}
	desc := utils.GetStructLike(rootStruct, IDL)
	if desc == nil {
		panic("struct '" + rootStruct + "' doesn't exist for the IDL")
	}

	ret := &FieldMask{}
	ret.desc = desc
	// horizontal traversal...
	for _, path := range paths {
		setPath(IDL, path, ret, ret.desc)
	}

	return ret
}

func (self FieldMask) String() string {
	buf := strings.Builder{}
	buf.WriteString("(")
	buf.WriteString(self.desc.GetName())
	buf.WriteString(")\n")
	self.print(&buf, 0)
	return buf.String()
}

func (self FieldMask) print(buf *strings.Builder, indent int) {
	for _, f := range self.desc.GetFields() {
		if !self.InMask(f.GetID()) {
			continue
		}
		self.printField(buf, indent+2, f)
	}
}

func (self FieldMask) printField(buf *strings.Builder, indent int, field *parser.Field) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString("+ ")
	buf.WriteString(field.GetName())
	buf.WriteString(" (")
	buf.WriteString(field.GetType().GetName())
	buf.WriteString(")\n")
	next := self.Next(field.GetID())
	if next != nil {
		next.print(buf, indent)
	}
}

func setPath(IDL *parser.Thrift, path string, cur *FieldMask, curDesc *parser.StructLike) {
	// vertical traversal...
	iterPath(path, func(name string, path string) bool {
		// find the field desc
		f, ok := curDesc.GetField(name)
		if !ok {
			panic("path '" + name + "' doesn't exist in current struct " + curDesc.GetName())
		}

		// set the field's mask
		cur.flat.Set(fieldID(f.GetID()))

		ft := f.GetType()
		nextDesc := utils.GetStructLike(ft.GetName(), IDL)
		if nextDesc == nil {
			if path != "" {
				panic("too deep path '" + path + "' for current struct " + curDesc.GetName())
			}
		} else {
			// check current FieldMaskMap if it is allocated
			if !cur.next.IsInitialized() {
				cur.next = makeFieldMaskMap(curDesc.GetFields())
			}
			curDesc = nextDesc
			// deep down to the next fieldmask
			cur = cur.next.GetOrAlloc(fieldID(f.GetID()))
			cur.desc = curDesc
		}

		// continue next layer
		return true
	})
}

func (self *FieldMask) InMask(id int32) bool {
	return self == nil || self.flat == nil || self.flat.Get(fieldID(id))
}

func (self *FieldMask) Next(id int32) *FieldMask {
	if self == nil {
		return nil
	}
	return self.next.Get(fieldID(id))
}

func iterPath(path string, f func(name, path string) bool) {
	for path != "" {
		name := path
		idx := strings.IndexByte(path, PathSep)
		if idx != -1 {
			name = path[:idx]
			path = path[idx+1:]
		} else {
			path = ""
		}
		if !f(name, path) {
			return
		}
	}
}

func (self *FieldMask) PathInMask(path string) bool {
	in := true
	iterPath(path, func(name, path string) bool {
		// empty fm or path means **IN MASK**
		if self == nil || name == "" {
			return false
		}

		// check if name exist
		f, ok := self.desc.GetField(name)
		if !ok {
			in = false
			return false
		}

		// check if name set mask
		if !self.InMask(f.GetID()) {
			in = false
			return false
		}

		self = self.next.Get(fieldID(f.GetID()))
		return true
	})

	return in
}
