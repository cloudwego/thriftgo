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

func (cur *FieldMask) SetPath(path string, curDesc *thrift_reflection.TypeDescriptor) error {
	println("[SetPath]:", path)
	if path == "" {
		return nil
	}

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
				return errDesc(curDesc, "expect STRUCT")
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

			// deep down to the next fieldmask
			curDesc = unwrapDesc(f.GetType())
			if curDesc == nil {
				return errDesc(curDesc, "nil field '"+f.GetName()+"' type descriptor")
			}
			println("field type: ", curDesc.GetName())

			cur = cur.setFieldID(fieldID(f.GetID()), st)
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
				return errDesc(curDesc, "expect LIST or SET")
			}

			et := unwrapDesc(curDesc.GetValueType())
			if et == nil {
				return errDesc(curDesc, "nil element descriptor")
			}
			println("et: ", et.GetName())

			var all = false
			var ids = []int{}
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				println("sub tok: ", tok.String())

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

				ids = append(ids, tok.Val().Int())
			}

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			if all {
				cur.all = &FieldMask{}
				cur.all.typ = nextFt
				cur = cur.all
				continue
			}

			nextPath := it.LeftPath()
			for _, id := range ids {
				next := cur.setInt(id)
				next.typ = nextFt
				if err := next.SetPath(nextPath, et); err != nil {
					return err
				}
			}
			return nil
		} else if styp == pathTypeMapL {
			// get element and key desc
			if !curDesc.IsMap() {
				return errDesc(curDesc, "isn't MAP")
			}
			if cur.typ != ftIntMap && cur.typ != ftStrMap {
				return errDesc(curDesc, "expect MAP")
			}

			et := unwrapDesc(curDesc.GetValueType())
			if et == nil {
				return errDesc(curDesc, "nil element descriptor")
			}
			kt := curDesc.GetKeyType()
			if kt == nil {
				return errDesc(curDesc, "nil key descriptor")
			}
			println("et: ", et.GetName())

			// iter indexies...
			isInt := false
			isStr := false
			all := false
			ids := []int{}
			strs := []string{}
			for it.HasNext() {

				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return errPath(tok, tok.Err().Error())
				}
				println("sub tok: ", tok.String())

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
					ids = append(ids, tok.Val().Int())
				} else if typ == pathTypeLitStr {
					if isInt {
						return errPath(tok, "isn't string")
					}
					isStr = true
					strs = append(strs, tok.Val().Str())
				} else {
					return errPath(tok, "expect integer or string element")
				}
			}

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			if all {
				cur.all = &FieldMask{}
				cur.all.typ = nextFt
				cur = cur.all
				continue
			}

			nextPath := it.LeftPath()
			if isInt {
				if cur.typ != ftIntMap {
					return errDesc(et, "should be integer-key map")
				}

				for _, id := range ids {
					next := cur.setInt(id)
					next.typ = nextFt
					if err := next.SetPath(nextPath, et); err != nil {
						return err
					}
				}

			} else if isStr {
				if cur.typ != ftStrMap {
					return errDesc(et, "should be string-key map")
				}

				for _, id := range strs {
					next := cur.setStr(id)
					next.typ = nextFt
					if err := next.SetPath(nextPath, et); err != nil {
						return err
					}
				}
			}
			return nil
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
		printIndent(buf, indent, "[\n")
		if self.intMask == nil {
			printIndent(buf, indent+2, "*\n")
		} else {
			for k := range self.intMask {
				self.printElem(buf, indent+2, k, desc.GetValueType())
			}
		}
		printIndent(buf, indent, "]\n")
	} else if self.typ == ftIntMap || self.typ == ftStrMap {
		printIndent(buf, indent, "{\n")
		if (self.typ == ftIntMap && self.intMask == nil) || (self.typ == ftStrMap && self.strMask == nil) {
			printIndent(buf, indent+2, "*\n")
		} else {
			for k := range self.intMask {
				self.printElem(buf, indent+2, k, desc.GetValueType())
			}
		}
		printIndent(buf, indent, "}\n")
	}
}

func printIndent(buf *strings.Builder, indent int, v string) {
	for i := 0; i < indent; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(v)
}

func (self FieldMask) printField(buf *strings.Builder, indent int, field *thrift_reflection.FieldDescriptor) {
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
	next := self.Field(int32(field.GetID()))
	if next != nil {
		next.print(buf, indent, nd)
	} else {
		printIndent(buf, indent+2, "...\n")
	}
}

func (self FieldMask) printElem(buf *strings.Builder, indent int, id interface{}, desc *thrift_reflection.TypeDescriptor) {
	printIndent(buf, indent, "+")
	var next *FieldMask
	switch v := id.(type) {
	case int:
		buf.WriteString(strconv.Itoa(v))
		next = self.Int(v)
	case string:
		buf.WriteString(v)
		next = self.Str(v)
	}
	buf.WriteString("\n")
	if next != nil {
		next.print(buf, indent, desc)
	} else {
		printIndent(buf, indent+2, "...\n")
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

			var all = cur.all != nil
			var next = cur.all
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				println("token", tok.String())
				if tok.Err() != nil {
					return false
				}

				if typ == pathTypeIndexR {
					break
				}
				if all || typ == pathTypeElem {
					continue
				}
				if typ == pathTypeAny {
					return all
				}
				if typ != pathTypeLitInt {
					return false
				}

				// check mask
				v := tok.Val().Int()
				if !cur.IntInMask(v) {
					return false
				}

				//NOTICE: always use last elem's fieldmask
				next = cur.Int(v)
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

			// for not specify any sub path
			if (cur.typ == ftIntMap && cur.intMask == nil) || (cur.typ == ftStrMap && cur.strMask == nil) {
				return true
			}

			var all = cur.all != nil
			var next = cur.all
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return false
				}
				println("token", tok.String())

				if typ == pathTypeMapR {
					break
				}
				if all || typ == pathTypeElem {
					continue
				}
				if typ == pathTypeAny {
					return all
				}

				if typ == pathTypeLitInt {
					if cur.typ != ftIntMap {
						return false
					}
					v := tok.Val().Int()
					if !cur.IntInMask(v) {
						return false
					}
					//NOTICE: always use last elem's fieldmask
					next = cur.Int(v)
				} else if typ == pathTypeLitStr {
					if cur.typ != ftStrMap {
						return false
					}
					v := tok.val.Str()
					if !cur.StrInMask(v) {
						return false
					}
					//NOTICE: always use last elem's fieldmask
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
