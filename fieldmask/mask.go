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

// PathSep is the separator of field names in a path
const PathSep = '.'

// FieldMask represents a collection of field paths
type FieldMask struct {
	// current layer of mask
	flat fieldMaskBitmap
	// for lookup next layer of fieldmasks
	next *fieldMaskMap
	// desc *parser.StructLike
}

var fmsPool = sync.Pool{
	New: func() interface{} {
		return &FieldMask{}
	},
}

// NewFieldMaskFromNames make a new FieldMask from paths and root descriptor,
// each path is the combination of field names from root struct to any layer of its children, separated with PathSep
func NewFieldMaskFromNames(desc *thrift_reflection.StructDescriptor, paths ...string) *FieldMask {
	ret := fmsPool.Get().(*FieldMask)
	ret.Init(desc, paths...)
	return ret
}

func (self *FieldMask) Recycle() {
	self.Reset()
	fmsPool.Put(self)
}

func (self *FieldMask) Reset() {
	if self == nil {
		return
	}
	self.flat = self.flat[:0]
	self.next.Reset()
}

func (self *FieldMask) Init(desc *thrift_reflection.StructDescriptor, paths ...string) {
	// horizontal traversal...
	for _, path := range paths {
		self.setPath(path, desc)
	}
}

// String pretty prints the structure a FieldMask represents
func (self FieldMask) String(desc *thrift_reflection.StructDescriptor) string {
	buf := strings.Builder{}
	buf.WriteString("(")
	buf.WriteString(desc.GetName())
	buf.WriteString(")\n")
	self.print(&buf, 0, desc)
	return buf.String()
}

func (self *FieldMask) InMask(id int16) bool {
	return self == nil || self.flat == nil || self.flat.Get(fieldID(id))
}

func (self *FieldMask) Next(id int16) *FieldMask {
	if self == nil || self.next == nil {
		return nil
	}
	return self.next.Get(fieldID(id))
}

func (self *FieldMask) PathInMask(desc *thrift_reflection.StructDescriptor, path string) bool {
	in := true
	iterPath(path, func(name, path string) bool {
		// empty fm or path means **IN MASK**
		if self == nil || name == "" {
			return false
		}

		// check if name exist
		f := desc.GetFieldByName(name)
		if f == nil {
			in = false
			return false
		}

		// check if name set mask
		if !self.InMask(int16(f.GetID())) {
			in = false
			return false
		}

		// deep to next desc
		var err error
		desc, err = f.GetType().GetStructDescriptor()
		if err != nil {
			if path != "" {
				in = false
			}
			return false
		}

		self = self.Next(int16(f.GetID()))
		return true
	})

	return in
}
