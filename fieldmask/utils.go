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
			if typ == pathTypeLit {
				id, ok := tok.ToInt()
				if ok {
					f = st.GetFieldById(int32(id))
					if f == nil {
						return errDesc(curDesc, "field "+strconv.Itoa(id)+" doesn't exist")
					}
				} else {
					name, ok := tok.ToStr()
					if !ok {
						return errPath(tok, "isn't string")
					}
					f = st.GetFieldByName(name)
					if f == nil {
						return errDesc(curDesc, "field '"+name+"' doesn't exist")
					}
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
				return errDesc(curDesc, "field '"+f.GetName()+"' has nil type descriptor")
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

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			var all = false
			if cur.all != nil {
				return errPath(stok, "is overlapped with * setted before")
			}
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
					cur.all = &FieldMask{}
					cur.all.typ = nextFt
					continue
				}
				if typ != pathTypeLit {
					return errPath(tok, "isn't literal")
				}
				id, ok := tok.ToInt()
				if !ok {
					return errPath(tok, "isn't integer")
				}

				ids = append(ids, id)
			}

			if all {
				curDesc = et
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

			var all = false
			if cur.all != nil {
				return errPath(stok, "is overlapped with * setted before")
			}

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			isInt := cur.typ == ftIntMap
			isStr := cur.typ == ftStrMap
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
					println("* for ", curDesc.KeyType.Name, ", path:", it.LeftPath())
					all = true
					cur.intMask = nil
					cur.strMask = nil
					cur.all = &FieldMask{}
					cur.all.typ = nextFt
					continue
				}

				if typ == pathTypeLit {
					id, ok := tok.ToInt()
					if !ok {
						return errPath(tok, "isn't integer")
					}
					if isStr {
						return errPath(tok, "expect string but got integer")
					}
					ids = append(ids, id)
				} else if typ == pathTypeStr {
					id, ok := tok.ToStr()
					if !ok {
						return errPath(tok, "isn't string")
					}
					if isInt {
						return errPath(tok, "expect integer but got string")
					}
					strs = append(strs, id)
				} else {
					return errPath(tok, "expect integer or string element")
				}
			}

			println("all:", all, "ids:", ids, "strs:", strs, isInt, isStr)

			if all {
				curDesc = et
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
				return nil
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
				return nil
			} else {
				return errPath(stok, "unexpected path "+nextPath)
			}
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
