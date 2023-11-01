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
	"fmt"
	"io"
	"strconv"
	"unsafe"
)

type pathType int

const (
	pathTypeLitInt pathType = 1 + iota
	pathTypeLitStr
	pathTypeRoot
	pathTypeField
	pathTypeIndexL
	pathTypeIndexR
	pathTypeMapL
	pathTypeMapR
	pathTypeElem
	pathTypeAny

	pathTypeEOF pathType = -1
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

func (p pathToken) Val() pathValue {
	return p.val
}

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
		return fmt.Sprintf("EOF(%d)", p.loc[0])
	case pathTypeAny:
		return fmt.Sprintf("*(%d)", p.loc[0])
	case pathTypeElem:
		return fmt.Sprintf(",(%d)", p.loc[0])
	case pathTypeField:
		return fmt.Sprintf(".(%d)", p.loc[0])
	case pathTypeRoot:
		return fmt.Sprintf("$(%d)", p.loc[0])
	case pathTypeIndexL:
		return fmt.Sprintf("[(%d)", p.loc[0])
	case pathTypeIndexR:
		return fmt.Sprintf("](%d)", p.loc[0])
	case pathTypeMapL:
		return fmt.Sprintf("{(%d)", p.loc[0])
	case pathTypeMapR:
		return fmt.Sprintf("}(%d)", p.loc[0])
	case pathTypeLitInt:
		return fmt.Sprintf("%d(%d:%d)", p.val.Int(), p.loc[0], p.loc[1])
	case pathTypeLitStr:
		return fmt.Sprintf("%s(%d:%d)", p.val.Str(), p.loc[0], p.loc[1])
	default:
		return fmt.Sprintf("unknown token %d(%d:%d) ", p.typ, p.loc[0], p.loc[1])
	}
}

func newPathToken(typ pathType, val string, s, e int) pathToken {
	switch typ {
	case pathTypeEOF:
		return pathToken{typ: typ}
	case pathTypeAny, pathTypeElem, pathTypeField, pathTypeIndexL, pathTypeIndexR, pathTypeLitStr, pathTypeMapR, pathTypeMapL, pathTypeRoot:
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
		return newPathToken(pathTypeRoot, string(c), s, p.Pos())
	case pathSepField:
		return newPathToken(pathTypeField, string(c), s, p.Pos())
	case pathSepIndexLeft:
		return newPathToken(pathTypeIndexL, string(c), s, p.Pos())
	case pathSepIndexRight:
		return newPathToken(pathTypeIndexR, string(c), s, p.Pos())
	case pathSepMapLeft:
		return newPathToken(pathTypeMapL, string(c), s, p.Pos())
	case pathSepMapRight:
		return newPathToken(pathTypeMapR, string(c), s, p.Pos())
	case pathSepElem:
		return newPathToken(pathTypeElem, string(c), s, p.Pos())
	case pathSepAny:
		return newPathToken(pathTypeAny, string(c), s, p.Pos())
	default:
		p.Unwind(s)
		val, isInt := p.lit()
		if isInt {
			return newPathToken(pathTypeLitInt, val, s, p.Pos())
		} else {
			return newPathToken(pathTypeLitStr, val, s, p.Pos())
		}
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
	var i = p.pos
	var isInt bool
	for ; i < len(p.src); i++ {
		switch cc := p.src[i]; cc {
		case pathSepElem, pathSepAny, pathSepRoot, pathSepField, pathSepIndexLeft, pathSepIndexRight, pathSepMapLeft, pathSepMapRight:
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
