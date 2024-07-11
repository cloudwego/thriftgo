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

func (g *FastGoBackend) genFastRead(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	// var conventions:
	// - p is the var of pointer to the struct going to be generated
	// - b is the buf to read from
	// - off is the offset of b
	// - err is the return err
	// - ftyp, fid only used in this method
	// - l must be increased after read
	// - enum is the tmp var for enum, it's updated by ReadInt32, and then set to the enum field
	// - x is the decoder of thrift.BinaryProtocol
	//
	// Please update the list if you'r going to add more vars
	// Instead of using consts for vars above, would like to use the names directly making code clear

	// func definition
	w.UsePkg("github.com/cloudwego/gopkg/protocol/thrift", "")
	w.f("func (p *%s) FastRead(b []byte) (off int, err error) {", s.GoName())
	w.f("var ftyp thrift.TType")
	w.f("var fid int16")
	w.f("var l int")

	isset := newBitsetCodeGen("isset", "uint8")
	hasEnum := false
	ff := getSortedFields(s)
	for _, f := range ff {
		if f.Type.Category == parser.Category_Enum {
			hasEnum = true
		}
		if f.Requiredness == parser.FieldType_Required {
			isset.Add(f)
		}
	}
	if hasEnum {
		w.f("var enum int32") // tmp var for enum
	}
	isset.GenVar(w)

	w.f("x := thrift.BinaryProtocol{}") // empty struct, no stack needed, for shorten varname

	w.f("for {")

	w.f("ftyp, fid, l, err = x.ReadFieldBegin(b[off:])")
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldBeginError }")
	w.f("if ftyp == thrift.STOP { break }")

	// fields
	w.f("switch uint32(fid)<<8| uint32(ftyp) {")
	for _, f := range ff {
		rwctx, err := g.utils.MkRWCtx(scope, f)
		if err != nil {
			// never goes here, should fail early in generator/golang pkg
			panic(err)
		}
		w.f("case 0x%x: // %s ID:%d %s",
			uint32(f.ID)<<8|uint32(category2ThriftWireType[f.Type.Category]),
			rwctx.Target, f.ID, category2GopkgConsts[f.Type.Category])
		genFastReadAny(w, rwctx, rwctx.Target, 0)
		if f.Requiredness == parser.FieldType_Required {
			isset.GenSetbit(w, f)
		}
	}
	w.f("default:") // default case, skip
	w.f("	l, err = x.Skip(b[off:], ftyp)")
	w.f("	off += l")
	w.f("	if err != nil { goto SkipFieldError }")
	w.f("}") // switch fid ends
	w.f("}") // for ends

	isset.GenIfNotSet(w, func(w *codewriter, v interface{}) {
		f := v.(*golang.Field)
		w.f("fid = %d // %s", f.ID, f.GoName())
		w.f("goto RequiredFieldNotSetError")
	})

	w.f("return") // no error

	w.UsePkg("fmt", "")
	w.f("ReadFieldBeginError:")
	w.f(`return off, thrift.PrependError(fmt.Sprintf("%%T read field begin error: ", p), err)`)

	if len(ff) > 0 { // fix `label ReadFieldError defined and not used`
		w.f("ReadFieldError:")
		w.f(`return off, thrift.PrependError(fmt.Sprintf("%%T read field %%d '%%s' error: ", p, fid, fieldIDToName_%s[fid]), err)`, s.GoName())
	}

	w.f("SkipFieldError:")
	w.f(`return off, thrift.PrependError(fmt.Sprintf("%%T skip field %%d type %%d error: ", p, fid, ftyp), err)`)

	if isset.Len() > 0 {
		w.f("RequiredFieldNotSetError:")
		w.f(`return off, thrift.NewProtocolException(thrift.INVALID_DATA, fmt.Sprintf("required field %%s is not set", fieldIDToName_%s[fid]))`, s.GoName())
	}

	// end of func definition
	w.f("}\n\n")
}

func genFastReadAny(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	t := rwctx.Type
	pointer := rwctx.IsPointer
	switch t.Category {
	case parser.Category_Bool:
		genFastReadBool(w, pointer, varname)
	case parser.Category_Byte:
		genFastReadByte(w, pointer, varname)
	case parser.Category_I16:
		genFastReadInt16(w, pointer, varname)
	case parser.Category_I32:
		genFastReadInt32(w, pointer, varname)
	case parser.Category_Enum:
		genFastReadEnum(w, rwctx, varname)
	case parser.Category_I64:
		genFastReadInt64(w, pointer, varname)
	case parser.Category_Double:
		genFastReadDouble(w, pointer, varname)
	case parser.Category_String:
		genFastReadString(w, pointer, varname)
	case parser.Category_Binary:
		genFastReadBinary(w, pointer, varname)
	case parser.Category_Map:
		genFastReadMap(w, rwctx, varname, depth)
	case parser.Category_List:
		genFastReadList(w, rwctx, varname, depth)
	case parser.Category_Set:
		genFastReadList(w, rwctx, varname, depth)
	case parser.Category_Struct:
		genFastReadStruct(w, rwctx, varname)
	case parser.Category_Union:
		genFastReadStruct(w, rwctx, varname)
	case parser.Category_Exception:
		genFastReadStruct(w, rwctx, varname)
	}
}

func genFastReadBool(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(bool)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadBool(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadByte(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(int8)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadByte(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadDouble(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(float64)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadDouble(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadInt16(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(int16)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadI16(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadInt32(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(int32)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadI32(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadEnum(w *codewriter, rwctx *golang.ReadWriteContext, varname string) {
	pointer := rwctx.IsPointer
	if pointer {
		w.f("if %s == nil { %s = new(%s)  }", varname, varname, rwctx.TypeName.Deref())
	}

	w.f("enum, l, err = x.ReadI32(b[off:])")
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
	w.f("%s = %s(enum)", varnameVal(pointer, varname), rwctx.TypeName.Deref())
}

func genFastReadInt64(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(int64)  }", varname, varname)
	}
	w.f("%s, l, err = x.ReadI64(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadBinary(w *codewriter, pointer bool, varname string) {
	if pointer { // always false?
		w.f("if %s == nil { %s = new([]byte) } ", varname, varname)
	}
	w.f("%s, l, err = x.ReadBinary(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadString(w *codewriter, pointer bool, varname string) {
	if pointer {
		w.f("if %s == nil { %s = new(string) } ", varname, varname)
	}
	w.f("%s, l, err = x.ReadString(b[off:])", varnameVal(pointer, varname))
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadStruct(w *codewriter, rwctx *golang.ReadWriteContext, varname string) {
	w.f("%s = %s()", varname, rwctx.TypeName.Deref().NewFunc())
	w.f("l, err = %s.FastRead(b[off:])", varname)
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")
}

func genFastReadList(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	// var conventions:
	// - sz is the size of a list
	// - i is unsed to interate for loop
	//
	// you must use the vars below instead of using literal above,
	// coz we may have embedded structs like list<list<i32>>
	if depth != 0 {
		w.f("{") // new block to protect tmp vars
		defer w.f("}")
	}
	tmpsize := "sz" //  for ReadListBegin, size int
	tmpi := "i"     // loop var
	if depth > 0 {  // avoid redeclared vars
		sub := strconv.Itoa(depth - 1)
		tmpsize = tmpsize + sub
		tmpi = tmpi + sub
	}

	w.f("var %s int", tmpsize)

	// ??? thriftgo & kitex always ignore element type of a list/set?
	w.f("_, %s, l, err = x.ReadListBegin(b[off:])", tmpsize)
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")

	w.f("%s = make(%s, %s)", varname, rwctx.TypeName.Deref(), tmpsize)
	w.f("for %s := 0; %s < %s; %s++ {", tmpi, tmpi, tmpsize, tmpi)
	genFastReadAny(w, rwctx.ValCtx, varname+"["+tmpi+"]", depth+1)
	w.f("}")
}

func genFastReadMap(w *codewriter, rwctx *golang.ReadWriteContext, varname string, depth int) {
	// var conventions:
	// - sz is the size of a map
	// - i is the counter for decoding a map
	//
	// you must use the vars below instead of using literal above,
	// coz we may have embedded structs like list<list<i32>>
	if depth != 0 {
		w.f("{") // new block to protect tmp vars
		defer w.f("}")
	}
	tmpsize := "sz" //  for ReadMapBegin, size int
	tmpk := "k"     // for reading keys
	tmpv := "v"     // for reading values
	tmpi := "i"     // loop var
	if depth > 0 {  // avoid redeclared vars
		sub := strconv.Itoa(depth - 1)
		tmpsize = tmpsize + sub
		tmpk = tmpk + sub
		tmpv = tmpv + sub
		tmpi = tmpi + sub
	}

	w.f("var %s int", tmpsize)

	// ??? thriftgo & kitex always ignore kv types of a map?
	w.f("_, _, %s, l, err = x.ReadMapBegin(b[off:])", tmpsize)
	w.f("off += l")
	w.f("if err != nil { goto ReadFieldError }")

	w.f("%s = make(%s, %s)", varname, rwctx.TypeName, tmpsize)
	w.f("for %s := 0; %s < %s; %s++ {", tmpi, tmpi, tmpsize, tmpi)
	if rwctx.KeyCtx.TypeID == "Struct" && !rwctx.KeyCtx.IsPointer {
		// hotfix for struct, it's always pointer for keys
		// remove this check after generator/gopkg fix it
		w.f("var %s *%s", tmpk, rwctx.KeyCtx.TypeName)
	} else {
		w.f("var %s %s", tmpk, rwctx.KeyCtx.TypeName)
	}
	w.f("var %s %s", tmpv, rwctx.ValCtx.TypeName)
	genFastReadAny(w, rwctx.KeyCtx, tmpk, depth+1)
	genFastReadAny(w, rwctx.ValCtx, tmpv, depth+1)
	w.f("%s[%s] = %s", varname, tmpk, tmpv)
	w.f("}")
}
