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

func (g *FastGoBackend) genFastWrite(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	// var conventions:
	// - p is the var of pointer to the struct going to be generated
	// - b is the buf to write into
	// - w is the var of thrift.NocopyWriter
	// - off is the offset of b

	// func definition
	w.UsePkg("github.com/cloudwego/gopkg/protocol/thrift", "")
	w.f("func (p *%s) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }\n\n", s.GoName())
	w.f("func (p *%s) FastWriteNocopy(b []byte, w thrift.NocopyWriter) int {", s.GoName())

	// case nil, STOP and return
	w.f("if p == nil { b[0] = 0; return 1; }")

	// `off` definition for buf cursor
	w.f("off := 0")

	// fields
	ff := getSortedFields(s)
	for _, f := range ff {
		rwctx, err := g.utils.MkRWCtx(scope, f)
		if err != nil {
			// never goes here, should fail early in generator/golang pkg
			panic(err)
		}
		genFastWriteField(w, rwctx, f)
	}

	// end of field encoding
	w.f("")               // empty line
	w.f("b[off] = 0")     // STOP
	w.f("return off + 1") // return including the STOP byte

	// end of func definition
	w.f("}\n\n")
}

func genFastWriteField(w *codewriter, rwctx *golang.ReadWriteContext, f *golang.Field) {
	// the real var name ref to the field
	varname := string("p." + f.GoName())

	// add comment like // ${FieldName} ${FieldID} ${FieldType}
	w.f("\n// %s ID:%d %s", rwctx.Target, f.ID, category2GopkgConsts[f.Type.Category])

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
	w.UsePkg("encoding/binary", "")
	w.f("b[off] = %d", category2ThriftWireType[f.Type.Category])
	w.f("binary.BigEndian.PutUint16(b[off+1:], %d) ", f.ID)
	w.f("off += 3")

	// field value
	genFastWriteAny(w, rwctx, varname, 0)

}

func genFastWriteAny(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	pointer := rwctx.IsPointer
	switch t.Category {
	case parser.Category_Bool:
		genFastWriteBool(w, pointer, varname)
	case parser.Category_Byte:
		genFastWriteByte(w, pointer, varname)
	case parser.Category_I16:
		genFastWriteInt16(w, pointer, varname)
	case parser.Category_I32, parser.Category_Enum:
		genFastWriteInt32(w, pointer, varname)
	case parser.Category_I64:
		genFastWriteInt64(w, pointer, varname)
	case parser.Category_Double:
		genFastWriteDouble(w, pointer, varname)
	case parser.Category_String:
		genFastWriteString(w, pointer, varname)
	case parser.Category_Binary:
		genFastWriteBinary(w, pointer, varname)
	case parser.Category_Map:
		genFastWriteMap(w, rwctx, varname, depth)
	case parser.Category_List, parser.Category_Set:
		genFastWriteList(w, rwctx, varname, depth)
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		// TODO: fix for parser.Category_Union? must only one field set
		genFastWriteStruct(w, rwctx, varname)
	}
}

func genFastWriteBool(w *codewriter, pointer bool, varname string) {
	// for bool, the underlying byte of true is always 1, and 0 for false
	// which is same as thrift binary protocol
	w.UsePkg("unsafe", "")
	w.f("b[off] = *((*byte)(unsafe.Pointer(%s)))", varnamePtr(pointer, varname))
	w.f("off++")
}

func genFastWriteByte(w *codewriter, pointer bool, varname string) {
	w.f("b[off] = byte(%s)", varnameVal(pointer, varname))
	w.f("off++")
}

func genFastWriteDouble(w *codewriter, pointer bool, varname string) {
	w.UsePkg("unsafe", "")
	w.f("binary.BigEndian.PutUint64(b[off:], *(*uint64)(unsafe.Pointer(%s)))", varnamePtr(pointer, varname))
	w.f("off += 8")
}

func genFastWriteInt16(w *codewriter, pointer bool, varname string) {
	w.UsePkg("encoding/binary", "")
	w.f("binary.BigEndian.PutUint16(b[off:], uint16(%s))", varnameVal(pointer, varname))
	w.f("off += 2")
}

func genFastWriteInt32(w *codewriter, pointer bool, varname string) {
	w.UsePkg("encoding/binary", "")
	w.f("binary.BigEndian.PutUint32(b[off:], uint32(%s))", varnameVal(pointer, varname))
	w.f("off += 4")
}

func genFastWriteInt64(w *codewriter, pointer bool, varname string) {
	w.UsePkg("encoding/binary", "")
	w.f("binary.BigEndian.PutUint64(b[off:], uint64(%s))", varnameVal(pointer, varname))
	w.f("off += 8")
}

func genFastWriteBinary(w *codewriter, pointer bool, varname string) {
	varname = varnameVal(pointer, varname)
	w.f("off += thrift.Binary.WriteBinaryNocopy(b[off:], w, %s)", varname)
}

func genFastWriteString(w *codewriter, pointer bool, varname string) {
	varname = varnameVal(pointer, varname)
	w.f("off += thrift.Binary.WriteStringNocopy(b[off:], w, %s)", varname)
}

func genFastWriteStruct(w *codewriter, rwctx *golang.ReadWriteContext, varname string) {
	w.f("off += %s.FastWriteNocopy(b[off:], w)", varname)
}

func genFastWriteList(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	rwctx = rwctx.ValCtx
	t := rwctx.Type
	w.UsePkg("encoding/binary", "")
	// list header
	w.f("b[off] = %d", category2ThriftWireType[t.Category])
	w.f("binary.BigEndian.PutUint32(b[off+1:], uint32(len(%s)))", varname)
	w.f("off += 5")

	// iteration tmp var
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpv = "v" + strconv.Itoa(depth-1)
	}
	w.f("for _, %s := range %s {", tmpv, varname)
	genFastWriteAny(w, rwctx, tmpv, depth+1)
	w.f("}")
}

func genFastWriteMap(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	kt := t.KeyType
	vt := t.ValueType
	// map header
	w.UsePkg("encoding/binary", "")
	w.f("b[off] = %d", category2ThriftWireType[kt.Category])
	w.f("b[off+1] = %d", category2ThriftWireType[vt.Category])
	w.f("binary.BigEndian.PutUint32(b[off+2:], uint32(len(%s)))", varname)
	w.f("off += 6")

	// iteration tmp var
	tmpk := "k"
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpk = "k" + strconv.Itoa(depth-1)
		tmpv = "v" + strconv.Itoa(depth-1)
	}
	w.f("for %s, %s := range %s {", tmpk, tmpv, varname)
	genFastWriteAny(w, rwctx.KeyCtx, tmpk, depth+1)
	genFastWriteAny(w, rwctx.ValCtx, tmpv, depth+1)
	w.f("}")
}
