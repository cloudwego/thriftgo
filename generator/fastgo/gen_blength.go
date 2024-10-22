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

// genBLength must be aligned with genBLength
// XXX: the code looks a bit redundant ...
func (g *FastGoBackend) genBLength(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	// var conventions:
	// - p is the var of pointer to the struct going to be generated
	// - off is the counter of BLength

	// func definition
	w.UsePkg("github.com/cloudwego/gopkg/protocol/thrift", "")
	w.f("func (p *%s) BLength() int {", s.GoName())

	// case nil, STOP
	w.f("if p == nil { return 1; }")

	w.f("off := 0")

	// fields
	ff := getSortedFields(s)
	for _, f := range ff {
		rwctx, err := g.utils.MkRWCtx(scope, f)
		if err != nil {
			// never goes here, should fail early in generator/golang pkg
			panic(err)
		}
		genBLengthField(w, rwctx, f)
	}

	// end of field encoding
	w.f("return off + 1") // return including the STOP byte

	// end of func definition
	w.f("}\n\n")
}

func genBLengthField(w *codewriter, rwctx *golang.ReadWriteContext, f *golang.Field) {
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
	w.f("off += 3") // type + fid

	// field value
	genBLengthAny(w, rwctx, varname, 0)

}

func genBLengthAny(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	if sz := category2WireSize[t.Category]; sz > 0 {
		w.f("off += %d", sz)
		return
	}
	pointer := rwctx.IsPointer
	switch t.Category {
	case parser.Category_String, parser.Category_Binary:
		genBLengthString(w, pointer, varname)
	case parser.Category_Map:
		genBLengthMap(w, rwctx, varname, depth)
	case parser.Category_List, parser.Category_Set:
		genBLengthList(w, rwctx, varname, depth)
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		genBLengthStruct(w, rwctx, varname)
	}
}

func genBLengthBinary(w *codewriter, pointer bool, varname string) {
	varname = varnameVal(pointer, varname)
	w.f("off += 4 + len(%s)", varname)
}

func genBLengthString(w *codewriter, pointer bool, varname string) {
	varname = varnameVal(pointer, varname)
	w.f("off += 4 + len(%s)", varname)
}

func genBLengthStruct(w *codewriter, _ *golang.ReadWriteContext, varname string) {
	w.f("off += %s.BLength()", varname)
}

func genBLengthList(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	// list header
	w.f("off += 5")

	// if element is basic type like int32, we can speed up the calc by sizeof(int32) * len(l)
	if sz := category2WireSize[t.ValueType.Category]; sz > 0 { // fast path for less code
		w.f("off += len(%s) * %d", varnameVal(rwctx.IsPointer, varname), sz)
		return
	}

	// iteration tmp var
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpv = "v" + strconv.Itoa(depth-1)
	}
	w.f("for _, %s := range %s {", tmpv, varname)
	genBLengthAny(w, rwctx.ValCtx, tmpv, depth+1)
	w.f("}")
}

func genBLengthMap(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	kt := t.KeyType
	vt := t.ValueType

	// map header
	w.f("off += 6")

	// iteration tmp var
	tmpk := "k"
	tmpv := "v"
	if depth > 0 { // avoid redeclared vars
		tmpk = "k" + strconv.Itoa(depth-1)
		tmpv = "v" + strconv.Itoa(depth-1)
	}

	// if key or value is basic type like int32, we can speed up the calc by sizeof(int32) * len(m)
	varname = varnameVal(rwctx.IsPointer, varname)
	ksz := category2WireSize[kt.Category]
	vsz := category2WireSize[vt.Category]
	if ksz > 0 && vsz > 0 {
		w.f("off += len(%s) * (%d+%d)", varname, ksz, vsz)
	} else if ksz > 0 {
		w.f("off += len(%s) * %d", varname, ksz)
		w.f("for _, %s := range %s {", tmpv, varname)
		genBLengthAny(w, rwctx.ValCtx, tmpv, depth+1)
		w.f("}")
	} else if vsz > 0 {
		w.f("off += len(%s) * %d", varname, vsz)
		w.f("for %s, _ := range %s {", tmpk, varname)
		genBLengthAny(w, rwctx.KeyCtx, tmpk, depth+1)
		w.f("}")
	} else {
		w.f("for %s, %s := range %s {", tmpk, tmpv, varname)
		genBLengthAny(w, rwctx.KeyCtx, tmpk, depth+1)
		genBLengthAny(w, rwctx.ValCtx, tmpv, depth+1)
		w.f("}")
	}
}
