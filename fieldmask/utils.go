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

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

func (self *FieldMask) print(buf *strings.Builder, indent int, desc *thrift_reflection.StructDescriptor) {
	for _, f := range desc.GetFields() {
		if !self.InMask(int16(f.GetID())) {
			continue
		}
		self.printField(buf, indent+2, f)
	}
}

func (self FieldMask) printField(buf *strings.Builder, indent int, field *thrift_reflection.FieldDescriptor) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString("+ ")
	buf.WriteString(field.GetName())
	buf.WriteString(" (")
	buf.WriteString(field.GetType().GetName())
	buf.WriteString(")\n")
	nd, err := field.GetType().GetStructDescriptor()
	if err == nil {
		next := self.Next(int16(field.GetID()))
		if next != nil {
			next.print(buf, indent, nd)
		} else {
			for i := 0; i < indent+2; i++ {
				buf.WriteByte(' ')
			}
			buf.WriteString("...\n")
		}
	}
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
