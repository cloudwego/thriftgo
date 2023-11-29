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
	desc = unwrapDesc(desc)
	if desc.IsBasic() {
		return ftScalar
	} else if desc.IsList() {
		return ftArray
	} else if desc.IsMap() {
		ft := unwrapDesc(desc.GetKeyType())
		switch ft.GetName() {
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

func (cur *FieldMask) addPath(path string, curDesc *thrift_reflection.TypeDescriptor) error {
	// println("[SetPath]: ", path)

	curDesc = unwrapDesc(curDesc)
	if curDesc == nil {
		return errors.New("nil descriptor")
	}

	it := newPathIter(path)
	for it.HasNext() {
		// println("desc: ", curDesc.Name)

		stok := it.Next()
		if stok.Err() != nil {
			return errPath(stok, "")
		}
		styp := stok.Type()
		// println("stoken: ", stok.String())

		if styp == pathTypeRoot {
			cur.typ = switchFt(curDesc)
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
			// println("struct: ", st.Name)

			// get field name or field id
			tok := it.Next()
			if tok.Err() != nil {
				return errPath(tok, "isn't field-name or field-id")
			}
			typ := tok.Type()
			// println("token: ", tok.String())

			all := cur.All()
			if all {
				return errPath(tok, "conflicts with previously-set all (*) fields")
			}

			var f *thrift_reflection.FieldDescriptor
			if typ == pathTypeLitInt {
				id := tok.val.Int()
				f = st.GetFieldById(int32(id))
				if f == nil {
					return errDesc(curDesc, "field "+strconv.Itoa(id)+" doesn't exist")
				}

			} else if typ == pathTypeLitStr {
				name := tok.val.Str()
				f = st.GetFieldByName(name)
				if f == nil {
					return errDesc(curDesc, "field '"+name+"' doesn't exist")
				}
			} else if typ == pathTypeAny {
				cur.fdMask.Reset()
				cur.isAll = true
				all = true

			} else {
				return errPath(stok, "isn't field-name or field-id")
			}

			if all {
				// println("all for struct")
				// NOTICE: for *, just pick first field desc for next loop
				fs := st.GetFields()
				if len(fs) == 0 || fs[0].GetType() == nil {
					return errDesc(curDesc, "doesn't have children fields")
				}
				cur = cur.getAll(switchFt(fs[0].GetType()))
			} else {
				// println("field: ", f.Name, f.ID)
				// deep down to the next fieldmask
				curDesc = unwrapDesc(f.GetType())
				if curDesc == nil {
					return errDesc(curDesc, "field '"+f.GetName()+"' has nil type descriptor")
				}
				cur = cur.setFieldID(fieldID(f.GetID()), st)
			}
			// continue for deeper path..

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
			// println("et: ", et.GetName())

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			all := cur.All()
			if all {
				return errPath(stok, "conflicts with previously-set all (*) index")
			}

			ids := []int{}
			empty := true
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				// println("sub tok: ", tok.String())

				if tok.Err() != nil {
					return errPath(tok, "isn't integer", tok.Err().Error())
				}
				if typ == pathTypeIndexR {
					if empty {
						return errPath(tok, "empty index set")
					}
					break
				}
				empty = false

				if all || typ == pathTypeElem {
					continue
				}

				if typ == pathTypeAny {
					cur.intMask.Reset()
					cur.isAll = true
					all = true
					continue
				}

				if typ != pathTypeLitInt {
					return errPath(tok, "isn't literal")
				}
				id := tok.val.Int()
				ids = append(ids, id)
			}

			if all {
				// println("all for list")
				curDesc = et
				cur = cur.getAll(nextFt)
				continue
			}

			nextPath := it.LeftPath()
			for _, id := range ids {
				// println("setInt ", id, nextFt)
				next := cur.setInt(id, nextFt)
				if err := next.addPath(nextPath, et); err != nil {
					return err
				}
			}
			// stop since all children has been set
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
			// println("et: ", et.GetName())

			nextFt := switchFt(et)
			if nextFt == ftInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			all := cur.All()
			if all {
				return errPath(stok, "conflicts with previously-set all (*) keys")
			}

			isInt := cur.typ == ftIntMap
			isStr := cur.typ == ftStrMap
			empty := true
			ids := []int{}
			strs := []string{}
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return errPath(tok, tok.Err().Error())
				}
				// println("sub tok: ", tok.String())

				if typ == pathTypeMapR {
					if empty {
						return errPath(tok, "empty key set")
					}
					break
				}
				empty = false

				if all || typ == pathTypeElem {
					continue
				}

				if typ == pathTypeAny {
					// println("* for ", curDesc.KeyType.Name, ", path:", it.LeftPath())
					cur.intMask.Reset()
					cur.strMask.Reset()
					cur.isAll = true
					all = true
					continue
				}

				if typ == pathTypeLitInt {
					if isStr {
						return errPath(tok, "expect string but got integer")
					}
					id := tok.val.Int()
					ids = append(ids, id)
				} else if typ == pathTypeStr {
					if isInt {
						return errPath(tok, "expect integer but got string")
					}
					id := tok.val.Str()
					strs = append(strs, id)
				} else {
					return errPath(tok, "expect integer or string element")
				}
			}

			// println("all:", all, "ids:", ids, "strs:", strs, isInt, isStr)

			if all {
				// println("all for map")
				curDesc = et
				cur = cur.getAll(nextFt)
				continue
			}

			nextPath := it.LeftPath()
			if isInt {
				if cur.typ != ftIntMap {
					return errDesc(et, "should be integer-key map")
				}
				for _, id := range ids {
					next := cur.setInt(id, nextFt)
					if err := next.addPath(nextPath, et); err != nil {
						return err
					}
				}

			} else if isStr {
				if cur.typ != ftStrMap {
					return errDesc(et, "should be string-key map")
				}
				for _, id := range strs {
					next := cur.setStr(id, nextFt)
					if err := next.addPath(nextPath, et); err != nil {
						return err
					}
				}

			} else {
				return errPath(stok, "unexpected path "+nextPath)
			}
			// stop since all children has been set
			return nil

		} else {
			return errPath(stok, "unexpected token")
		}
	}

	cur.isAll = true
	return nil
}

func (self *FieldMask) print(buf *strings.Builder, indent int, desc *thrift_reflection.TypeDescriptor) {
	if self.All() {
		printIndent(buf, indent+2, "*\n")
		return
	}
	if self.typ == ftStruct {
		st, err := desc.GetStructDescriptor()
		if err != nil {
			panic(err)
		}
		for _, f := range st.GetFields() {
			if _, exist := self.Field(int16(f.GetID())); !exist {
				continue
			}
			self.printField(buf, indent+2, f)
		}
	} else if self.typ == ftArray {
		for k, v := range self.intMask {
			if v.typ == 0 {
				continue
			}
			self.printElem(buf, indent+2, k, desc.GetValueType())
		}
		printIndent(buf, indent, "]\n")
	} else if self.typ == ftIntMap || self.typ == ftStrMap {
		for k, v := range self.intMask {
			if v.typ == 0 {
				continue
			}
			self.printElem(buf, indent+2, k, desc.GetValueType())
		}
		printIndent(buf, indent, "}\n")
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
