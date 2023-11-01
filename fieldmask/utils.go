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

func switchFt(desc *thrift_reflection.TypeDescriptor) fieldMaskType {
	if desc.IsBasic() {
		return ftScalar
	} else if desc.IsList() {
		return ftArray
	} else if desc.IsMap() {
		ft := desc.GetKeyType().GetName()
		switch ft {
		case "i8", "i16", "i32", "i64", "byte":
			return ftIntMap
		case "string", "binary":
			return ftStrMap
		default:
			return ftInvalid
		}
	} else if desc.IsStruct() {
		return ftStruct
	} else {
		return ftInvalid
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

func (cur *FieldMask) setPath(path string, curDesc *thrift_reflection.TypeDescriptor) error {
	it := newPathIter(path)
	curDesc = unwrapDesc(curDesc)
	if curDesc == nil {
		return errors.New("nil descriptor")
	}

	for it.HasNext() {
		println("desc: ", curDesc.Name)

		stok := it.Next()
		if stok.Err() != nil {
			return errPath(stok, "")
		}
		styp := stok.Type()
		println("stoken: ", stok.String())

		if styp == pathTypeRoot {
			cur.typ = switchFt(curDesc)
			if cur.typ == ftInvalid {
				return errDesc(curDesc, "unsupported")
			}
			continue

		} else if styp == pathTypeField {
			// get struct descriptor
			st, err := curDesc.GetStructDescriptor()
			if err != nil || st == nil {
				return errDesc(curDesc, "isn't STRUCT")
			}
			if cur.typ != ftStruct {
				return errDesc(curDesc, "isn't STRUCT")
			}
			println("struct: ", st.Name)

			// get field name or field id
			tok := it.Next()
			if tok.Err() != nil {
				return errPath(tok, "isn't field-name or field-id")
			}
			typ := tok.Type()
			println("token: ", tok.String())

			var f *thrift_reflection.FieldDescriptor
			if typ == pathTypeLitStr {
				// find the field desc
				name := tok.Val().Str()
				f = st.GetFieldByName(name)
				if f == nil {
					return errDesc(curDesc, "field '"+name+"' doesn't exist")
				}
			} else if typ == pathTypeLitInt {
				id := tok.Val().Int()
				f = st.GetFieldById(int32(id))
				if f == nil {
					return errDesc(curDesc, "field "+strconv.Itoa(id)+" doesn't exist")
				}
			} else if typ == pathTypeAny {
				if it.HasNext() {
					return errPath(tok, "* for STRUCT should end here")
				}
				// NOTICE: .fields == nil means all
				cur.fields = nil
				cur.fieldMask = nil
				return nil
			} else {
				return errPath(stok, "isn't field-name or field-id")
			}
			println("field: ", f.Name, f.ID)

			// set mask
			next := cur.setFieldID(fieldID(f.GetID()), st)

			// deep down to the next fieldmask
			curDesc = unwrapDesc(f.GetType())
			if curDesc == nil {
				return errDesc(curDesc, "nil field '"+f.GetName()+"' type descriptor")
			}
			cur = next
			cur.typ = switchFt(curDesc)
			if cur.typ == ftInvalid {
				return errDesc(curDesc, "unspported type for fieldmask")
			}

		} else if styp == pathTypeIndexL {
			// get element desc
			if !curDesc.IsList() {
				return errDesc(curDesc, "isn't LIST or SET")
			}
			if cur.typ != ftArray {
				return errDesc(curDesc, "isn't LIST or SET")
			}

			et := curDesc.GetValueType()
			if et == nil {
				return errDesc(curDesc, "nil element descriptor")
			}
			et = unwrapDesc(et)

			var all = false
			var next = &FieldMask{}
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()

				if tok.Err() != nil {
					return errPath(tok, "isn't integer", tok.Err().Error())
				}
				if typ == pathTypeIndexR {
					break
				}
				if typ == pathTypeElem {
					continue
				}
				if all {
					continue
				}
				if typ == pathTypeAny {
					all = true
					cur.intMask = nil
					continue
				}
				if typ != pathTypeLitInt {
					return errPath(tok, "isn't integer")
				}

				// set mask
				cur.setInt(tok.Val().Int(), next)
			}

			// next fieldmask
			curDesc = et
			cur = next
			cur.typ = switchFt(curDesc)
			if cur.typ == ftInvalid {
				return errDesc(curDesc, "unspported type for fieldmask")
			}

		} else if styp == pathTypeMapL {
			// get element and key desc
			if !curDesc.IsMap() {
				return errDesc(curDesc, "isn't MAP")
			}
			if cur.typ != ftIntMap && cur.typ != ftStrMap {
				return errDesc(curDesc, "isn't MAP")
			}

			et := curDesc.GetValueType()
			if et == nil {
				return errDesc(curDesc, "nil element descriptor")
			}
			et = unwrapDesc(et)
			kt := curDesc.GetKeyType()
			if kt == nil {
				return errDesc(curDesc, "nil key descriptor")
			}

			// iter indexies...
			var next = &FieldMask{}
			isInt := false
			isStr := true
			all := false
			for it.HasNext() {

				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return errPath(tok, tok.Err().Error())
				}

				if typ == pathTypeMapR {
					break
				}
				if typ == pathTypeElem {
					continue
				}
				if all {
					continue
				}
				if typ == pathTypeAny {
					all = true
					cur.intMask = nil
					cur.strMask = nil
					continue
				}

				if typ == pathTypeLitInt {
					if isStr {
						return errPath(tok, "isn't integer")
					}
					isInt = true
					cur.setInt(tok.Val().Int(), next)
				} else if typ == pathTypeLitStr {
					if isInt {
						return errPath(tok, "isn't string")
					}
					isStr = true
					cur.setStr(tok.val.Str(), next)
				} else {
					return errPath(tok, "expect integer or string element")
				}
			}

			if isInt {
				if cur.typ != ftIntMap {
					return errDesc(et, "should be integer-key map")
				}
			} else if isStr {
				if cur.typ != ftStrMap {
					return errDesc(et, "should be string-key map")
				}
			}

			// next fieldmask
			curDesc = et
			cur = next
			cur.typ = switchFt(et)
		} else {
			return errPath(stok, "unexpected token")
		}
	}

	return nil
}

func (self *FieldMask) print(buf *strings.Builder, indent int, desc *thrift_reflection.TypeDescriptor) {
	if self == nil || self.typ == ftScalar {
		return
	}
	if self.typ == ftStruct {
		st, err := desc.GetStructDescriptor()
		if err != nil {
			panic(err)
		}
		for _, f := range st.GetFields() {
			if !self.FieldInMask(int32(f.GetID())) {
				continue
			}
			self.printField(buf, indent+2, f)
		}
	} else if self.typ == ftArray {
		for i := 0; i < indent; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("[")
		if self.intMask == nil {
			buf.WriteString("*")
		} else {
			for k := range self.intMask {
				self.printElem(buf, indent+2, k, desc.GetValueType())
				buf.WriteString("\n")
			}
		}
		for i := 0; i < indent; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("]\n")
	} else if self.typ == ftIntMap || self.typ == ftStrMap {
		for i := 0; i < indent; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("{")
		if (self.typ == ftIntMap && self.intMask == nil) || (self.typ == ftStrMap && self.strMask == nil) {
			buf.WriteString("*")
		} else {
			for k := range self.intMask {
				self.printElem(buf, indent+2, k, desc.GetValueType())
				buf.WriteString("\n")
			}
		}
		for i := 0; i < indent; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("[")
		buf.WriteString("}\n")
	}
}

func (self FieldMask) printField(buf *strings.Builder, indent int, field *thrift_reflection.FieldDescriptor) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(".")
	buf.WriteString(field.GetName())
	buf.WriteString(" (")
	buf.WriteString(field.GetType().GetName())
	buf.WriteString(")\n")
	nd := field.GetType()
	next := self.Field(int32(field.GetID()))
	if next != nil {
		next.print(buf, indent, nd)
	} else {
		for i := 0; i < indent+2; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("...\n")
	}
}

func (self FieldMask) printElem(buf *strings.Builder, indent int, id interface{}, desc *thrift_reflection.TypeDescriptor) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString("+")

	var next *FieldMask

	switch id.(type) {
	case int:
		buf.WriteString(strconv.Itoa(id.(int)))
		next = self.Int(id.(int))
	case string:
		buf.WriteString(id.(string))
		next = self.Str(id.(string))
	}
	buf.WriteString("\n")
	if next != nil {
		next.print(buf, indent, desc)
	} else {
		for i := 0; i < indent+2; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("...\n")
	}
}

func (cur *FieldMask) PathInMask(curDesc *thrift_reflection.TypeDescriptor, path string) bool {
	it := newPathIter(path)
	println("[PathInMask]")
	for it.HasNext() {
		//NOTICE: desc shoudn't empty here
		println("desc: ", curDesc.Name)

		//NOTICE: empty fm for path means **IN MASK**
		if cur == nil {
			return true
		}

		stok := it.Next()
		if stok.Err() != nil {
			return false
		}
		styp := stok.Type()
		println("stoken: ", stok.String())

		if styp == pathTypeRoot {
			continue

		} else if styp == pathTypeField {
			// get struct descriptor
			st, err := curDesc.GetStructDescriptor()
			if err != nil {
				return false
			}
			println("struct: ", st.Name)

			if cur.typ != ftStruct {
				return false
			}

			// for any * directive
			if cur.fieldMask == nil {
				println("nil fields")
				return true
			}

			tok := it.Next()
			if tok.Err() != nil {
				return false
			}
			typ := tok.Type()
			println("token", tok.String())

			var f *thrift_reflection.FieldDescriptor
			if typ == pathTypeLitStr {
				// find the field desc
				name := tok.Val().Str()
				f = st.GetFieldByName(name)
				if f == nil {
					return false
				}
			} else if typ == pathTypeLitInt {
				id := tok.Val().Int()
				f = st.GetFieldById(int32(id))
				if f == nil {
					return false
				}
			} else if typ == pathTypeAny {
				if cur.fields != nil {
					return false
				}
				if it.HasNext() {
					return false
				}
				return true
			} else {
				return false
			}

			// check if name set mask
			if !cur.FieldInMask(int32(f.GetID())) {
				return false
			}

			if !it.HasNext() {
				return true
			}

			// deep to next desc
			curDesc = f.GetType()
			if curDesc == nil {
				return false
			}
			cur = cur.Field(int32(f.GetID()))

		} else if styp == pathTypeIndexL {

			// get element desc
			if !curDesc.IsList() {
				return false
			}
			et := curDesc.GetValueType()
			if et == nil {
				return false
			}

			if cur.typ != ftArray {
				return false
			}

			// for any * directive or not specify any sub path
			if cur.intMask == nil {
				return true
			}

			// iter indexies...
			var next *FieldMask
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()

				if tok.Err() != nil {
					return false
				}

				if typ == pathTypeIndexR {
					break
				}
				if typ == pathTypeElem {
					continue
				}

				if typ == pathTypeAny {
					if cur.intMask != nil {
						return false
					}
					continue
				}
				if typ != pathTypeLitInt {
					return false
				}

				// check mask
				if !cur.IntInMask(tok.Val().Int()) {
					return false
				}
				next = cur.Int(tok.Val().Int())
			}

			// next fieldmask
			curDesc = et
			cur = next

		} else if styp == pathTypeMapL {
			// get element and key desc
			if !curDesc.IsMap() {
				return false
			}
			et := curDesc.GetValueType()
			if et == nil {
				return false
			}
			kt := curDesc.GetKeyType()
			if kt == nil {
				return false
			}

			if cur.typ != ftIntMap && cur.typ != ftStrMap {
				return false
			}

			// for any * directive or not specify any sub path
			if cur.intMask == nil {
				return true
			}

			// iter indexies...
			var next *FieldMask
			for it.HasNext() {

				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return false
				}

				if typ == pathTypeMapR {
					break
				}
				if typ == pathTypeElem {
					continue
				}
				if typ == pathTypeAny {
					if cur.typ == ftIntMap && cur.intMask != nil {
						return false
					}
					if cur.typ == ftStrMap && cur.strMask != nil {
						return false
					}
					continue
				}

				if typ == pathTypeLitInt {
					if cur.typ != ftIntMap {
						return false
					}
					v := tok.Val().Int()
					if !cur.IntInMask(v) {
						return false
					}
					next = cur.Int(v)
				} else if typ == pathTypeLitStr {
					if cur.typ != ftStrMap {
						return false
					}
					v := tok.val.Str()
					if !cur.StrInMask(v) {
						return false
					}
					next = cur.Str(v)
				} else {
					return false
				}
			}

			// next fieldmask
			curDesc = et
			cur = next
		} else {
			return false
		}
	}

	return !it.HasNext()
}
