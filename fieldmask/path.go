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
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"unsafe"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

type pathType int

const (
	pathTypeLitStr pathType = 1 + iota
	pathTypeLitInt pathType = 1 + iota
	pathTypeStr
	pathTypeRoot
	pathTypeField
	pathTypeIndexL
	pathTypeIndexR
	pathTypeMapL
	pathTypeMapR
	pathTypeElem
	pathTypeAny

	pathTypeEOF pathType = -1
	pathTypeERR pathType = -2
)

type pathSep byte

const (
	pathSepRoot       = '$'
	pathSepField      = '.'
	pathSepIndexLeft  = '['
	pathSepIndexRight = ']'
	pathSepMapLeft    = '{'
	pathSepMapRight   = '}'
	pathSepElem       = ','
	pathSepAny        = '*'
	pathSepQuote      = '"'
	pathSepSlash      = '\\'
)

var (
	jsonPathAny  = json.RawMessage(`"*"`)
	jsonPathRoot = json.RawMessage(`"$"`)
)

type pathValue struct {
	pv unsafe.Pointer
	iv int
}

func newPathValueStr(val string) pathValue {
	if val == "" {
		return pathValue{iv: len(val), pv: nil}
	} else {
		return pathValue{iv: len(val), pv: *(*unsafe.Pointer)(unsafe.Pointer(&val))}
	}
}

func newPathValueInt(val int) pathValue {
	return pathValue{iv: val}
}

func (v pathValue) Str() string {
	return *(*string)(unsafe.Pointer(&v))
}

func (v pathValue) Int() int {
	return v.iv
}

type pathToken struct {
	typ pathType
	val pathValue
	loc [2]int
}

func (p pathToken) Type() pathType {
	return p.typ
}

// func (p pathToken) ToInt() (int, bool) {
// 	if p.typ == pathTypeLitStr || p.typ == pathTypeStr {
// 		i, e := strconv.ParseInt(p.val.Str(), 10, 64)
// 		if e != nil {
// 			return 0, false
// 		}
// 		return int(i), true
// 	} else if p.typ == pathTypeLitInt {
// 		return p.val.Int(), true
// 	} else {
// 		return 0, false
// 	}
// }

// func (p pathToken) ToStr() (string, bool) {
// 	if p.typ == pathTypeLitStr || p.typ == pathTypeStr {
// 		return p.val.Str(), true
// 	} else if p.typ == pathTypeLitInt {
// 		str := strconv.Itoa(p.val.Int())
// 		return str, true
// 	} else {
// 		return "", false
// 	}
// }

func (p pathToken) Pos() (int, int) {
	return p.loc[0], p.loc[1]
}

func (p pathToken) Err() error {
	switch p.typ {
	case pathTypeEOF:
		return io.EOF
	default:
		return nil
	}
}

func (p pathToken) String() string {
	switch p.typ {
	case pathTypeEOF:
		return fmt.Sprintf("EOF at %d", p.loc[0])
	case pathTypeAny:
		return fmt.Sprintf("* at %d", p.loc[0])
	case pathTypeElem:
		return fmt.Sprintf(", at %d", p.loc[0])
	case pathTypeField:
		return fmt.Sprintf(". at %d", p.loc[0])
	case pathTypeRoot:
		return fmt.Sprintf("$ at %d", p.loc[0])
	case pathTypeIndexL:
		return fmt.Sprintf("[ at %d", p.loc[0])
	case pathTypeIndexR:
		return fmt.Sprintf("] at %d", p.loc[0])
	case pathTypeMapL:
		return fmt.Sprintf("{ at %d", p.loc[0])
	case pathTypeMapR:
		return fmt.Sprintf("} at %d", p.loc[0])
	// case pathTypeLitInt:
	// 	return fmt.Sprintf("%d(%d:%d)", p.val.Int(), p.loc[0], p.loc[1])
	case pathTypeLitStr:
		return fmt.Sprintf("Lit(%s) at %d-%d", p.val.Str(), p.loc[0], p.loc[1])
	case pathTypeLitInt:
		return fmt.Sprintf("Lit(%d) at %d-%d", p.val.Int(), p.loc[0], p.loc[1])
	case pathTypeStr:
		return fmt.Sprintf("Str(%q) at %d-%d", p.val.Str(), p.loc[0], p.loc[1])
	case pathTypeERR:
		return fmt.Sprintf("Err(%s) at %d-%d", p.val.Str(), p.loc[0], p.loc[1])
	default:
		return fmt.Sprintf("UnknownToken(%d) at %d:%d", p.typ, p.loc[0], p.loc[1])
	}
}

func newPathToken(typ pathType, val string, s, e int) pathToken {
	switch typ {
	case pathTypeEOF:
		return pathToken{typ: typ}
	case pathTypeStr, pathTypeAny, pathTypeElem, pathTypeField, pathTypeIndexL, pathTypeIndexR, pathTypeLitStr, pathTypeMapR, pathTypeMapL, pathTypeRoot:
		return pathToken{typ: typ, val: newPathValueStr(val), loc: [2]int{s, e}}
	case pathTypeLitInt:
		i, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		return pathToken{typ: typ, val: newPathValueInt(i), loc: [2]int{s, e}}
	default:
		panic("unspported pathType " + val)
	}
}

type pathIterator struct {
	pos int
	src string
}

func newPathIter(src string) pathIterator {
	return pathIterator{src: src, pos: 0}
}

func (p *pathIterator) Pos() int {
	return p.pos
}

func (p *pathIterator) LeftPath() string {
	if p.pos >= len(p.src) {
		return ""
	}
	return p.src[p.pos:]
}

func (p *pathIterator) HasNext() bool {
	return p.pos < len(p.src)
}

func (p *pathIterator) Next() pathToken {
	if !p.HasNext() {
		return newPathToken(pathTypeEOF, "", p.pos, p.pos)
	}
	s := p.Pos()
	c := p.char()
	switch c {
	case pathSepRoot:
		return newPathToken(pathTypeRoot, "", s, p.Pos())
	case pathSepField:
		return newPathToken(pathTypeField, "", s, p.Pos())
	case pathSepIndexLeft:
		return newPathToken(pathTypeIndexL, "", s, p.Pos())
	case pathSepIndexRight:
		return newPathToken(pathTypeIndexR, "", s, p.Pos())
	case pathSepMapLeft:
		return newPathToken(pathTypeMapL, "", s, p.Pos())
	case pathSepMapRight:
		return newPathToken(pathTypeMapR, "", s, p.Pos())
	case pathSepElem:
		return newPathToken(pathTypeElem, "", s, p.Pos())
	case pathSepAny:
		return newPathToken(pathTypeAny, "", s, p.Pos())
	case pathSepQuote:
		p.Unwind(s)
		v, e := p.str()
		if e != nil {
			return newPathToken(pathTypeERR, "invalid quote string", s, p.Pos())
		}
		return newPathToken(pathTypeStr, v, s, p.Pos())
	default:
		p.Unwind(s)
		val, isInt := p.lit()
		if isInt {
			return newPathToken(pathTypeLitInt, val, s, p.Pos())
		}
		return newPathToken(pathTypeLitStr, val, s, p.Pos())
	}
}

func (p *pathIterator) char() byte {
	c := p.src[p.pos]
	p.pos += 1
	return c
}

func (p *pathIterator) Unwind(pos int) {
	p.pos = pos
}

func (p *pathIterator) lit() (string, bool) {
	i := p.pos
	var isInt bool
	for ; i < len(p.src); i++ {
		switch cc := p.src[i]; cc {
		case pathSepElem, pathSepAny, pathSepRoot, pathSepField, pathSepIndexLeft, pathSepIndexRight, pathSepMapLeft, pathSepMapRight, pathSepQuote, pathSepSlash:
			goto ret
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if i == p.pos {
				isInt = true
			} else {
				isInt = isInt && true
			}
		default:
			isInt = false
		}
	}
ret:
	val := p.src[p.pos:i]
	p.pos = i
	return val, isInt
}

func (p *pathIterator) str() (string, error) {
	i := p.pos
	open := false
	for ; i < len(p.src); i++ {
		switch cc := p.src[i]; cc {
		case pathSepSlash:
			i += 1
		case pathSepQuote:
			open = !open
			if !open {
				i += 1
				goto ret
			}
		}
	}
ret:
	val := p.src[p.pos:i]
	p.pos = i
	val, err := strconv.Unquote(val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// PathInMask tells if a given path is already in current fieldmask
func (fm *FieldMask) PathInMask(desc *thrift_reflection.TypeDescriptor, path string) bool {
	_, ex := fm.GetPath(desc, path)
	return ex
}

// getPathAncestor tells if a given path is in current fieldmask, and return the nearest settled ancestor (include itself)
func (cur *FieldMask) GetPath(desc *thrift_reflection.TypeDescriptor, path string) (*FieldMask, bool) {
	it := newPathIter(path)
	// println("[PathInMask]")
	last := cur
	for it.HasNext() {
		// NOTICE: desc shoudn't empty here
		// println("desc: ", curDesc.Name)

		// NOTICE: empty fm for path means **IN MASK**
		if cur == nil {
			return last, true
		}
		last = cur

		stok := it.Next()
		if stok.Err() != nil {
			return nil, false
		}
		styp := stok.Type()
		// println("stoken: ", stok.String())
		// j, _ := cur.MarshalJSON()
		// println("cur mask: ", string(j), cur.isAll, cur.all)

		if styp == pathTypeRoot {
			continue
		} else if styp == pathTypeField {
			// get struct descriptor
			st, err := desc.GetStructDescriptor()
			if err != nil {
				return nil, false
			}
			// println("struct: ", st.Name)
			if cur.typ != FtStruct {
				return nil, false
			}

			tok := it.Next()
			if tok.Err() != nil {
				return nil, false
			}
			typ := tok.Type()
			// println("token", tok.String())

			var f *thrift_reflection.FieldDescriptor
			if typ == pathTypeLitInt {
				f = st.GetFieldById(int32(tok.val.Int()))
				if f == nil {
					return nil, false
				}
			} else if typ == pathTypeLitStr {
				name := tok.val.Str()
				f = st.GetFieldByName(name)
				if f == nil {
					return nil, false
				}
			} else if typ == pathTypeAny {
				if !cur.All() {
					return nil, false
				}
			} else {
				return nil, false
			}

			// println("all", all, "FieldInMask:", cur.FieldInMask(int32(f.GetID())))
			// check if name set mask
			// println("field ", f.GetID())
			nextFm, exist := cur.Field(int16(f.GetID()))
			if !exist {
				// println("return nil, false")
				return nil, false
			}

			// deep to next desc
			desc = f.GetType()
			if desc == nil {
				return nil, false
			}

			cur = nextFm

		} else if styp == pathTypeIndexL {

			// get element desc
			if !desc.IsList() {
				return nil, false
			}
			et := desc.GetValueType()
			if et == nil {
				return nil, false
			}

			if cur.typ != FtList {
				return nil, false
			}

			all := cur.All()
			next := cur.all
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				// println("token", tok.String())
				if tok.Err() != nil {
					return nil, false
				}

				if typ == pathTypeIndexR {
					break
				}
				if all || typ == pathTypeElem {
					continue
				}
				if typ == pathTypeAny {
					return nil, false
				}
				if typ != pathTypeLitInt {
					return nil, false
				}

				// check mask
				v := tok.val.Int()
				nextFm, exist := cur.Int(v)
				if !exist {
					return nil, false
				}
				// NOTICE: always use last elem's fieldmask
				next = nextFm
			}

			// next fieldmask
			desc = et
			cur = next
		} else if styp == pathTypeMapL {
			// get element and key desc
			if !desc.IsMap() {
				return nil, false
			}
			et := desc.GetValueType()
			if et == nil {
				return nil, false
			}
			kt := desc.GetKeyType()
			if kt == nil {
				return nil, false
			}

			// println("cur.typ::", cur.typ, "cur::", cur.String(curDesc))
			if cur.typ != FtIntMap && cur.typ != FtStrMap && cur.typ != FtScalar {
				return nil, false
			}
			// spew.Dump("cur ", cur)
			next := cur.all
			// iter indexies...
			for it.HasNext() {
				tok := it.Next()
				typ := tok.Type()
				if tok.Err() != nil {
					return nil, false
				}
				// println("token", tok.String())

				if typ == pathTypeMapR {
					break
				}
				if cur.All() || typ == pathTypeElem {
					continue
				}
				if typ == pathTypeAny {
					return nil, false
				}

				if typ == pathTypeLitInt {
					if cur.typ != FtIntMap {
						return nil, false
					}
					v := tok.val.Int()
					nextFm, exist := cur.Int(v)
					if !exist {
						return nil, false
					}
					// NOTICE: always use last elem's fieldmask
					next = nextFm
				} else if typ == pathTypeStr {
					if cur.typ != FtStrMap {
						return nil, false
					}
					v := tok.val.Str()
					nextFm, exist := cur.Str(v)
					if !exist {
						return nil, false
					}
					// NOTICE: always use last elem's fieldmask
					next = nextFm
				} else {
					return nil, false
				}
			}

			// next fieldmask
			desc = et
			cur = next
			// spew.Dump("next ", cur)
		} else {
			return nil, false
		}
	}

	return cur, !it.HasNext()
}
