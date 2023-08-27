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
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/utils"
)

type fieldID int32

const _MaxFieldIDHead = 255

type fieldMaskMap struct {
	head [_MaxFieldIDHead]*FieldMask
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

func (self *fieldMaskMap) GetOrAlloc(f fieldID) *FieldMask {
	if f < _MaxFieldIDHead {
		s := self.head[f]
		if s == nil {
			s = &FieldMask{}
			self.head[f] = s
		}
		return s
	} else {
		if s := self.tail[f]; s != nil {
			return s
		} else {
			s := &FieldMask{}
			self.tail[f] = s
			return s
		}
	}
}

type fieldMaskBitmap []byte

const _BucketBit = 8

func (self *fieldMaskBitmap) Set(f fieldID) {
	b := int(f / _BucketBit)
	i := int(f % _BucketBit)
	if cap(*self) <= b {
		tmp := make([]byte, len(*self), 2*cap(*self))
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
	i := int(f % _BucketBit)
	if cap(*self) <= b {
		return false
	}
	if len(*self) <= b {
		*self = (*self)[:b+1]
	}
	return ((*self)[b] & byte(1<<i)) != 0
}

type FieldMask struct {
	flat fieldMaskBitmap
	next fieldMaskMap
	desc *parser.StructLike
}

func NewFieldMaskFromAST(IDL *parser.Thrift, rootStruct string, paths ...[]string) *FieldMask {
	if IDL == nil {
		panic("FieldMask must have a IDL!")
	}
	desc := utils.GetStructLike(rootStruct, IDL)
	if desc == nil {
		panic("struct '" + rootStruct + "' doesn't exist for the IDL")
	}
	ret := &FieldMask{}
	ret.desc = desc
	ret.next = makeFieldMaskMap(desc.Fields)
	// horizontal traversal...
	for _, path := range paths {
		setPath(IDL, path, ret, ret.desc)
	}

	return ret
}

func setPath(IDL *parser.Thrift, path []string, cur *FieldMask, curDesc *parser.StructLike) {
	// vertical traversal...
	for j, field := range path {
		// find the field desc
		f, ok := curDesc.GetField(field)
		if !ok {
			panic("field '" + field + "' doesn't exist in current struct " + curDesc.GetName())
		}

		// set the field's mask
		cur.flat.Set(fieldID(f.GetID()))

		// check current FieldMaskMap if it is allocated
		if cur.next.tail == nil {
			cur.next = makeFieldMaskMap(curDesc.GetFields())
		}

		// deep down to the next desc
		ft := f.GetType()
		next := utils.GetStructLike(ft.GetName(), IDL)
		if next == nil {
			if j < len(path)-1 {
				panic("too deep field '" + path[j+1] + "' for current struct " + curDesc.GetName())
			}
			break
		}
		curDesc = next

		// deep down to the next fieldmask
		cur = cur.next.GetOrAlloc(fieldID(f.GetID()))
		cur.desc = curDesc
	}
}
