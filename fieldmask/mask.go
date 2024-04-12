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

	"github.com/cloudwego/thriftgo/internal/utils"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

// FieldMaskType indicates the corresponding thrift message type for a fieldmask
type FieldMaskType uint8

// MarshalText implements encoding.TextMarshaler
func (ft FieldMaskType) MarshalText() ([]byte, error) {
	switch ft {
	case FtScalar:
		return utils.S2B("Scalar"), nil
	case FtList:
		return utils.S2B("List"), nil
	case FtStruct:
		return utils.S2B("Struct"), nil
	case FtStrMap:
		return utils.S2B("StrMap"), nil
	case FtIntMap:
		return utils.S2B("IntMap"), nil
	default:
		return utils.S2B("Invalid"), nil
	}
}

// UnmarshalText implements encoding.TextUnmarshaler
func (ft *FieldMaskType) UnmarshalText(in []byte) error {
	switch utils.B2S(in) {
	case "Scalar":
		*ft = FtScalar
	case "List":
		*ft = FtList
	case "Struct":
		*ft = FtStruct
	case "StrMap":
		*ft = FtStrMap
	case "IntMap":
		*ft = FtIntMap
	default:
		*ft = FtInvalid
	}
	return nil
}

// FieldMaskType Enums
const (
	// Invalid or unsupported thrift type
	FtInvalid FieldMaskType = iota
	// thrift scalar types, including BOOL/I8/I16/I32/I64/DOUBLE/STRING/BINARY, or neither-string-nor-integer-typed-key MAP
	FtScalar
	// thrift LIST/SET
	FtList
	// thrift STRUCT
	FtStruct
	// thrift MAP with string-typed key
	FtStrMap
	// thrift MAP with integer-typed key
	FtIntMap
)

// FieldMask represents a collection of thrift pathes
// See
type FieldMask struct {
	isAll bool

	isBlack bool // black-list mode

	typ FieldMaskType

	all *FieldMask

	fdMask *fieldMap

	strMask strMap

	intMask intMap
}

// NewFieldMask create a new fieldmask
func NewFieldMask(desc *thrift_reflection.TypeDescriptor, pathes ...string) (*FieldMask, error) {
	return Options{}.NewFieldMask(desc, pathes...)
}

// Options for creating FieldMask
type Options struct {
	// BlackListMode enables black-list mode when create FieldMask,
	// which means `Field()/Str()/Int()` will return false for a **Complete** Path in the FieldMask
	BlackListMode bool
}

// NewFieldMask create a new fieldmask with options
func (opts Options) NewFieldMask(desc *thrift_reflection.TypeDescriptor, pathes ...string) (*FieldMask, error) {
	ret := FieldMask{}
	if opts.BlackListMode {
		ret.isBlack = true
	}
	err := ret.init(desc, pathes...)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// reset clears fieldmask's all path
func (self *FieldMask) reset() {
	if self == nil {
		return
	}
	self.isAll = false
	self.typ = 0
	self.fdMask.Reset()
	self.intMask.Reset()
	self.strMask.Reset()
}

func (self *FieldMask) init(desc *thrift_reflection.TypeDescriptor, paths ...string) error {
	// horizontal traversal...
	for _, path := range paths {
		if err := self.addPath(path, desc); err != nil {
			return fmt.Errorf("Parsing path %q  error: %v", path, err)
		}
	}
	return nil
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
			// cur.path = jsonPathRoot
			continue

		} else if styp == pathTypeField {
			// get struct descriptor
			st, err := curDesc.GetStructDescriptor()
			if err != nil || st == nil {
				return errDesc(curDesc, "isn't STRUCT")
			}
			if cur.typ != FtStruct {
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
				return errPath(tok, "field conflicts with previously settled '*'")
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
				cur = cur.setAll(switchFt(fs[0].GetType()))
				// cur.path = jsonPathAny
			} else {
				// println("field: ", f.Name, f.ID)
				// deep down to the next fieldmask
				curDesc = unwrapDesc(f.GetType())
				if curDesc == nil {
					return errDesc(curDesc, "field '"+f.GetName()+"' has nil type descriptor")
				}

				cur = cur.setFieldID(fieldID(f.GetID()), switchFt(st.GetFieldById(int32(f.GetID())).GetType()))
				// cur.path = strconv.Itoa(int(f.GetID()))
			}
			// continue for deeper path..

		} else if styp == pathTypeIndexL {
			// get element desc
			if !curDesc.IsList() {
				return errDesc(curDesc, "isn't LIST or SET")
			}
			if cur.typ != FtList {
				return errDesc(curDesc, "expect LIST or SET")
			}

			et := unwrapDesc(curDesc.GetValueType())
			if et == nil {
				return errDesc(curDesc, "nil element descriptor")
			}
			// println("et: ", et.GetName())

			nextFt := switchFt(et)
			if nextFt == FtInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			all := cur.All()
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

				if typ == pathTypeElem {
					continue
				}

				if typ == pathTypeAny {
					cur.intMask.Reset()
					cur.isAll = true
					all = true
					continue
				}

				if all {
					return errPath(tok, "id conflicts with previously settled '*'")
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
				cur = cur.setAll(nextFt)
				// cur.path = jsonPathAny
				continue
			}

			nextPath := it.LeftPath()
			for _, id := range ids {
				// println("setInt ", id, nextFt)
				next := cur.setInt(id, nextFt, len(ids))
				// next.path = strconv.Itoa(id)
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
			if cur.typ != FtIntMap && cur.typ != FtStrMap && cur.typ != FtScalar {
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
			if nextFt == FtInvalid {
				return errDesc(et, "unspported type for fieldmask")
			}

			all := cur.All()
			isInt := cur.typ == FtIntMap
			isStr := cur.typ == FtStrMap
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

				if typ == pathTypeElem {
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

				if all {
					return errPath(tok, "key conflicts with previous settled '*'")
				}

				if typ == pathTypeLitInt {
					if !isInt {
						return errPath(tok, "expect string but got integer")
					}
					id := tok.val.Int()
					ids = append(ids, id)
				} else if typ == pathTypeStr {
					if !isStr {
						return errPath(tok, "expect integer but got string")
					}
					id := tok.val.Str()
					strs = append(strs, id)
				} else {
					return errPath(tok, "expect integer or string or '*' as key")
				}
			}

			// println("all:", all, "ids:", ids, "strs:", strs, isInt, isStr)

			if all {
				// println("all for map")
				curDesc = et
				cur = cur.setAll(nextFt)
				// cur.path = jsonPathAny
				continue
			}

			nextPath := it.LeftPath()
			if isInt {
				if cur.typ != FtIntMap {
					return errDesc(et, "should be integer-key map")
				}
				for _, id := range ids {
					next := cur.setInt(id, nextFt, len(ids))
					// next.path = strconv.Itoa(id)
					if err := next.addPath(nextPath, et); err != nil {
						return err
					}
				}

			} else if isStr {
				if cur.typ != FtStrMap {
					return errDesc(et, "should be string-key map")
				}
				for _, id := range strs {
					next := cur.setStr(id, nextFt, len(strs))
					// next.path = strconv.Quote(id)
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

	// for scalar type, isAll is always true
	cur.isAll = true
	return nil
}

// Exist tells if the fieldmask is setted
func (self *FieldMask) Exist() bool {
	return self != nil && self.typ != 0
}

func (self *FieldMask) ret(fm *FieldMask) (*FieldMask, bool) {
	if self.isBlack {
		return fm, !fm.Exist()
	} else {
		return fm, fm.Exist()
	}
}

// Field returns the specific sub mask for a given id, and tells if the id in the mask
func (self *FieldMask) Field(id int16) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.fdMask.Get(fieldID(id))
	return self.ret(fm)
}

// Int returns the specific sub mask for a given index, and tells if the index in the mask
func (self *FieldMask) Int(id int) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.intMask.Get(id)
	return self.ret(fm)
}

// Field returns the specific sub mask for a given string, and tells if the string in the mask
func (self *FieldMask) Str(id string) (*FieldMask, bool) {
	if self == nil || self.typ == 0 {
		return nil, true
	}
	if self.isAll {
		return self.all, true
	}
	fm := self.strMask.Get(id)
	return self.ret(fm)
}

// All tells if the mask covers all elements (* or empty or scalar)
func (self *FieldMask) All() bool {
	if self == nil {
		return true
	}
	switch self.typ {
	case FtStruct, FtList, FtIntMap, FtStrMap:
		return self.isAll
	default:
		return true
	}
}

// SetBlack sets the FieldMask to be black-list or white-list
func (self *FieldMask) SetBlack(black bool) {
	self.isBlack = black
}

// SetBlack tells if the FieldMask is black-list or white-list
func (self *FieldMask) IsBlack() bool {
	return self.isBlack
}
