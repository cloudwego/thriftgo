/*
 * Copyright 2024 CloudWeGo Authors
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

package fastgo

import (
	"strconv"

	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
)

const nocopyWriteThreshold = 4096

func (g *FastGoBackend) genFastWrite(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	w.UsePkg("github.com/cloudwego/gopkg/protocol/thrift", "")
	w.f("func (p *%s) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }\n\n", s.GoName())

	w.f("func (p *%s) FastWriteNocopy(b []byte, w thrift.NocopyWriter) (n int) {", s.GoName())
	w.f(`if n = len(p.FastAppend(b[:0])); n > len(b) {`)
	w.f(`panic ("buffer overflow. concurrency issue?")`)
	w.f(`}`)
	w.f(`return`)
	w.f("}\n\n") // end of FastWriteNocopy

	g.genFastAppend(w, scope, s)
}

func (g *FastGoBackend) genFastAppend(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	// var conventions:
	// - p is the var of pointer to the struct going to be generated
	// - b is the buf to write into
	// - w is the var of thrift.NocopyWriter
	// - x is the shortcut of thrift.BinaryProtocol

	w.UsePkg("github.com/cloudwego/gopkg/protocol/thrift", "")
	w.f("func (p *%s) FastAppend(b []byte) []byte {", s.GoName())
	defer w.f("}\n\n")

	// case nil, STOP and return
	w.f(`if p == nil { return append(b, 0) }`)

	// shortcut for encoding
	w.f("x := thrift.BinaryProtocol{}")
	w.f("_ = x")

	// fields
	ff := getSortedFields(s)
	for _, f := range ff {
		rwctx, err := g.utils.MkRWCtx(scope, f)
		if err != nil {
			// never goes here, should fail early in generator/golang pkg
			panic(err)
		}
		genFastAppendField(w, rwctx, f)
	}
	w.f("\nreturn append(b, 0)") // return including the STOP byte
}

func genFastAppendField(w *codewriter, rwctx *golang.ReadWriteContext, f *golang.Field) {
	// the real var name ref to the field
	varname := string("p." + f.GoName())

	// add comment like // ${FieldName} ${FieldID} ${FieldType}
	w.f("\n// %s", rwctx.Target)

	// check skip cases
	// only for optional fields
	if f.Requiredness == parser.FieldType_Optional {
		if f.GoTypeName().IsPointer() || isContainerType(f.Type) {
			// case 1: optional and nil
			w.f("if %s != nil {", varname)
			defer w.f("}")
		} else if !f.GoTypeName().IsPointer() && f.Default != nil {
			// case 2: optional and equals to default value
			w.f("if %s != %v {", varname, f.DefaultValue())
			defer w.f("}")
		}
	}

	// field header
	w.f("b = append(b, %d, %d, %d)", // AppendFieldBegin
		category2ThriftWireType[f.Type.Category], byte(f.ID>>8), byte(f.ID))

	// field value
	genFastAppendAny(w, rwctx, varname, 0)
}

func genFastAppendAny(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	pointer := rwctx.IsPointer
	switch t.Category {
	case parser.Category_Bool:
		genFastAppendBool(w, pointer, varname)
	case parser.Category_Byte:
		genFastAppendByte(w, pointer, varname)
	case parser.Category_I16:
		genFastAppendInt16(w, pointer, varname)
	case parser.Category_I32, parser.Category_Enum:
		genFastAppendInt32(w, pointer, varname)
	case parser.Category_I64:
		genFastAppendInt64(w, pointer, varname)
	case parser.Category_Double:
		genFastAppendDouble(w, pointer, varname)
	case parser.Category_String:
		genFastAppendString(w, pointer, varname)
	case parser.Category_Binary:
		genFastAppendBinary(w, pointer, varname)
	case parser.Category_Map:
		genFastAppendMap(w, rwctx, varname, depth)
	case parser.Category_List, parser.Category_Set:
		genFastAppendList(w, rwctx, varname, depth)
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		// TODO: fix for parser.Category_Union? must only one field set
		genFastAppendStruct(w, rwctx, varname)
	}
}

func genFastAppendBool(w *codewriter, pointer bool, varname string) {
	// for bool, the underlying byte of true is always 1, and 0 for false
	// which is same as thrift binary protocol
	w.UsePkg("unsafe", "")
	w.f("b = append(b, *(*byte)(unsafe.Pointer(%s)))", varnamePtr(pointer, varname))
}

func genFastAppendByte(w *codewriter, pointer bool, varname string) {
	w.f("b = append(b, byte(%s))", varnameVal(pointer, varname))
}

func genFastAppendDouble(w *codewriter, pointer bool, varname string) {
	w.f("b = x.AppendDouble(b, float64(%s))", varnameVal(pointer, varname))
}

func genFastAppendInt16(w *codewriter, pointer bool, varname string) {
	w.f("b = x.AppendI16(b, int16(%s))", varnameVal(pointer, varname))
}

func genFastAppendInt32(w *codewriter, pointer bool, varname string) {
	w.f("b = x.AppendI32(b, int32(%s))", varnameVal(pointer, varname))
}

func genFastAppendInt64(w *codewriter, pointer bool, varname string) {
	w.f("b = x.AppendI64(b, int64(%s))", varnameVal(pointer, varname))
}

func genFastAppendBinary(w *codewriter, pointer bool, varname string) {
	varname = varnameVal(pointer, varname)
	w.f("b = x.AppendI32(b, int32(len(%s)))", varname)
	w.f("b = append(b, %s...)", varname)
}

func genFastAppendString(w *codewriter, pointer bool, varname string) {
	genFastAppendBinary(w, pointer, varname)
}

func genFastAppendStruct(w *codewriter, rwctx *golang.ReadWriteContext, varname string) {
	w.f("b = %s.FastAppend(b)", varname)
}

func genFastAppendList(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	rwctx = rwctx.ValCtx
	t := rwctx.Type

	// list header
	w.f("b = x.AppendListBegin(b, %s, len(%s))", category2GopkgConsts[t.Category], varname)

	// iteration tmp var
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpv = "v" + strconv.Itoa(depth-1)
	}
	w.f("for _, %s := range %s {", tmpv, varname)
	genFastAppendAny(w, rwctx, tmpv, depth+1)
	w.f("}")
}

func genFastAppendMap(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	kt := t.KeyType
	vt := t.ValueType
	// map header
	w.f("b = x.AppendMapBegin(b, %s, %s, len(%s))",
		category2GopkgConsts[kt.Category], category2GopkgConsts[vt.Category], varname)

	// iteration tmp var
	tmpk := "k"
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpk = "k" + strconv.Itoa(depth-1)
		tmpv = "v" + strconv.Itoa(depth-1)
	}
	w.f("for %s, %s := range %s {", tmpk, tmpv, varname)
	genFastAppendAny(w, rwctx.KeyCtx, tmpk, depth+1)
	genFastAppendAny(w, rwctx.ValCtx, tmpv, depth+1)
	w.f("}")
}
