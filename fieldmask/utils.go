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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

func errDesc(desc *thrift_reflection.TypeDescriptor, msg ...string) error {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("Descriptor %q ", desc.GetName()))
	for _, m := range msg {
		buf.WriteString(m)
		buf.WriteByte('\n')
	}
	return errors.New(buf.String())
}

func errPath(tok pathToken, msg ...string) error {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("Token %s ", tok.String()))
	for _, m := range msg {
		buf.WriteString(m)
		buf.WriteByte('\n')
	}
	return errors.New(buf.String())
}

func switchFt(desc *thrift_reflection.TypeDescriptor) FieldMaskType {
	desc = unwrapDesc(desc)
	if desc.IsBasic() {
		return FtScalar
	} else if desc.IsList() {
		return FtList
	} else if desc.IsMap() {
		ft := unwrapDesc(desc.GetKeyType())
		if ft.IsEnum() {
			return FtIntMap
		}
		switch ft.GetName() {
		case "i8", "i16", "i32", "i64", "byte":
			return FtIntMap
		case "string", "binary":
			return FtStrMap
		default:
			return FtScalar // NOTICE: mean fieldmask exist and is all
		}
	} else if desc.IsStruct() {
		return FtStruct
	} else if desc.IsEnum() {
		return FtScalar
	} else {
		return FtInvalid // NOTICE: mean fieldmask not exist
	}
}

func unwrapDesc(desc *thrift_reflection.TypeDescriptor) *thrift_reflection.TypeDescriptor {
	if desc == nil {
		return nil
	}
	for desc.IsTypedef() {
		td, _ := desc.GetTypedefDescriptor()
		desc = td.GetType()
	}
	return desc
}

func (self *FieldMask) print(buf *strings.Builder, indent int, desc *thrift_reflection.TypeDescriptor) {
	if !self.Exist() {
		return
	}
	if self.typ == FtStruct {
		st, err := desc.GetStructDescriptor()
		if err != nil {
			panic(err)
		}
		if self.All() {
			printIndent(buf, indent+2, "*\n")
			if fs := st.GetFields(); len(fs) > 0 {
				self.all.print(buf, indent+2, fs[0].GetType())
			}
			return
		}
		for _, f := range st.GetFields() {
			if _, exist := self.Field(int16(f.GetID())); !exist {
				continue
			}
			self.printField(buf, indent+2, f)
		}
	} else if self.typ == FtList || self.typ == FtIntMap {
		if self.All() {
			printIndent(buf, indent+2, "*\n")
			self.all.printElem(buf, indent+2, 0, desc.GetValueType())
			return
		}
		for k, v := range self.intMask {
			if v.typ == 0 {
				continue
			}
			self.printElem(buf, indent+2, k, desc.GetValueType())
		}
	} else if self.typ == FtStrMap {
		if self.All() {
			printIndent(buf, indent+2, "*")
			self.printElem(buf, indent+2, "", desc.GetValueType())
			return
		}
		for k, v := range self.strMask {
			if v.typ == 0 {
				continue
			}
			self.printElem(buf, indent+2, k, desc.GetValueType())
		}
	} else if self.typ == FtScalar {
		buf.WriteString(" (")
		buf.WriteString(desc.GetName())
		buf.WriteString(")\n")
	} else {
		printIndent(buf, indent, "Unknown Fieldmask")
	}
}

func printIndent(buf *strings.Builder, indent int, v string) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(v)
}

func (self *FieldMask) printField(buf *strings.Builder, indent int, field *thrift_reflection.FieldDescriptor) {
	printIndent(buf, indent, ".")
	buf.WriteString(field.GetName())
	buf.WriteString(" (")
	buf.WriteString(field.GetType().GetName())
	if field.GetType().IsList() {
		buf.WriteString("<")
		buf.WriteString(field.GetType().GetValueType().GetName())
		buf.WriteString(">")
	}
	if field.GetType().IsMap() {
		buf.WriteString("<")
		buf.WriteString(field.GetType().GetKeyType().GetName())
		buf.WriteString(",")
		buf.WriteString(field.GetType().GetValueType().GetName())
		buf.WriteString(">")
	}
	buf.WriteString(")\n")
	nd := field.GetType()
	next, exist := self.Field(int16(field.GetID()))
	if exist {
		next.print(buf, indent, nd)
	}
}

func (self *FieldMask) printElem(buf *strings.Builder, indent int, id interface{}, desc *thrift_reflection.TypeDescriptor) {
	printIndent(buf, indent, "+")
	var next *FieldMask
	var e bool
	switch v := id.(type) {
	case int:
		buf.WriteString(strconv.Itoa(v))
		next, e = self.Int(v)
	case string:
		buf.WriteString(v)
		next, e = self.Str(v)
	}
	buf.WriteString("\n")
	if e {
		next.print(buf, indent, desc)
	}
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
