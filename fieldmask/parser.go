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
	"errors"
	"io"
	"strconv"
	"unsafe"
)

type pathType int

const (
	pathTypeFieldName pathType = 1 + iota
	pathTypeFieldID
	pathTypeMapStrKey
	pathTypeMapIntKey
	pathTypeArrIndex
	pathTypeAny

	pathTypeEOF       pathType = -1
	pathTypeSyntaxErr pathType = -2
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

func (p pathToken) StrVal() string {
	switch p.typ {
	case pathTypeFieldName, pathTypeMapStrKey:
		return p.val.Str()
	default:
		panic("the token shouldn't has string val")
	}
}

func (p pathToken) IntVal() int {
	switch p.typ {
	case pathTypeFieldID, pathTypeArrIndex, pathTypeMapIntKey:
		return p.val.Int()
	default:
		panic("the token shouldn't has int val")
	}
}

func (p pathToken) Pos() (int, int) {
	return p.loc[0], p.loc[1]
}

func (p pathToken) Err() error {
	switch p.typ {
	case pathTypeEOF:
		return io.EOF
	case pathTypeSyntaxErr:
		return errors.New("syntax error")
	default:
		return nil
	}
}

func newPathToken(typ pathType, val string, s, e int) pathToken {
	switch typ {
	case pathTypeEOF:
		return pathToken{typ: typ}
	case pathTypeAny:
		return pathToken{typ: typ, loc: [2]int{s, e}}
	case pathTypeSyntaxErr, pathTypeFieldName, pathTypeMapStrKey:
		return pathToken{typ: typ, val: newPathValueStr(val), loc: [2]int{s, e}}
	case pathTypeArrIndex, pathTypeFieldID, pathTypeMapIntKey:
		i, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		return pathToken{typ: typ, val: newPathValueInt(i), loc: [2]int{s, e}}
	default:
		panic("unspported pathType " + val)
	}
}

type pathParser struct {
	pos     int
	src     string
	cur     pathToken
	parents pathSep
}

func (p *pathParser) Pos() int {
	return p.pos
}

func (p *pathParser) Cur() pathToken {
	return p.cur
}

func (p *pathParser) errNoVal(c byte) pathToken {
	return newPathToken(pathTypeSyntaxErr, "miss literal value after "+string(c), p.pos, p.pos)
}

func (p *pathParser) Next() pathToken {
	if p.pos >= len(p.src) {
		return newPathToken(pathTypeEOF, "", p.pos, p.pos)
	}

	c, typ := p.tok()
	if typ < 0 {
		return newPathToken(typ, "", p.pos, p.pos)
	}

	var tok pathToken
	switch c {
	case pathSepField:
		pre := p.pos
		val, isInt := p.lit()
		if val == "" {
			return p.errNoVal(c)
		}
		p.parents = pathSepField
		if isInt {
			tok = newPathToken(pathTypeFieldName, val, pre, p.pos)
		} else {
			tok = newPathToken(pathTypeFieldName, val, pre, p.pos)
		}
	case pathSepIndexLeft:
		pre := p.pos
		val, isInt := p.lit()
		if val == "" {
			return p.errNoVal(c)
		}
		p.parents = pathSepIndexLeft
		if isInt {
			tok = newPathToken(pathTypeArrIndex, val, pre, p.pos)
		} else {
			tok = newPathToken(pathTypeSyntaxErr, "invalid int val after [", pre, p.pos)
		}
	case pathSepIndexRight:
		p.parents = 0
		tok = p.Next()
	case pathSepMapLeft:
		p.parents = pathSepIndexLeft
		pre := p.pos
		val, isInt := p.lit()
		if val == "" {
			return p.errNoVal(c)
		}
		if isInt {
			tok = newPathToken(pathTypeMapIntKey, val, pre, p.pos)
		} else {
			tok = newPathToken(pathTypeMapStrKey, val, pre, p.pos)
		}
	case pathSepMapRight:
		p.parents = 0
		tok = p.Next()
	case pathSepElem:
		pat := p.parents
		if pat != pathSepIndexLeft || pat != pathSepMapLeft {
			tok = newPathToken(pathTypeSyntaxErr, "element sep should has a parent [] or {}", p.pos, p.pos)
		}
		pre := p.pos
		val, isInt := p.lit()
		if val == "" {
			return p.errNoVal(c)
		}
		if isInt {
			if p.cur.typ != pathTypeMapIntKey {
				tok = newPathToken(pathTypeSyntaxErr, "inconsistent int element type", pre, p.pos)
			} else {
				tok = newPathToken(pathTypeMapIntKey, val, pre, p.pos)
			}
		} else {
			if p.cur.typ != pathTypeMapStrKey {
				tok = newPathToken(pathTypeSyntaxErr, "inconsistent string element type", pre, p.pos)
			} else {
				tok = newPathToken(pathTypeMapStrKey, val, pre, p.pos)
			}
		}
	}

	return tok
}

func (p *pathParser) tok() (byte, pathType) {
	if p.pos >= len(p.src) {
		return 0, pathTypeEOF
	}
	c := p.char()
	if c == pathSepRoot {
		p.parents = pathSepRoot
		if p.pos < len(p.src) {
			cc := p.char()
			if cc != pathSepField || cc != pathSepIndexLeft || cc != pathSepMapLeft {
				return 0, pathTypeSyntaxErr
			}
		} else {
			return 0, pathTypeEOF
		}
	}
	return c, 0
}

func (p *pathParser) char() byte {
	c := p.src[p.pos]
	p.pos += 1
	return c
}

func (p *pathParser) lit() (string, bool) {
	var i = p.pos
	var isInt bool
	for ; i < len(p.src); i++ {
		switch cc := p.src[i]; cc {
		case pathSepField, pathSepIndexLeft, pathSepIndexRight, pathSepMapLeft, pathSepMapRight:
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
