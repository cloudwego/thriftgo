// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package golang

import (
	"fmt"

	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	"github.com/cloudwego/thriftgo/parser"
)

func deref(pAST **parser.Thrift, pType **parser.Type) {
	for (*pType).GetIsTypedef() {
		name := (*pType).Name
		if ref := (*pType).Reference; ref != nil {
			*pAST = (*pAST).Includes[ref.Index].Reference
			name = ref.Name
		}
		tmp, _ := (*pAST).GetTypedef(name)
		*pType = tmp.Type
	}
}

func type2typeInfo(ast *parser.Thrift, typ *parser.Type, index int) *meta.TypeMeta {
	switch typ.Category {
	case parser.Category_Bool:
		return &meta.TypeMeta{TypeID: meta.TTypeID_BOOL}
	case parser.Category_Byte:
		return &meta.TypeMeta{TypeID: meta.TTypeID_BYTE}
	case parser.Category_I16:
		return &meta.TypeMeta{TypeID: meta.TTypeID_I16}
	case parser.Category_I32, parser.Category_Enum:
		return &meta.TypeMeta{TypeID: meta.TTypeID_I32}
	case parser.Category_I64:
		return &meta.TypeMeta{TypeID: meta.TTypeID_I64}
	case parser.Category_Double:
		return &meta.TypeMeta{TypeID: meta.TTypeID_DOUBLE}
	case parser.Category_String, parser.Category_Binary:
		return &meta.TypeMeta{TypeID: meta.TTypeID_STRING}
	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		return &meta.TypeMeta{TypeID: meta.TTypeID_STRUCT}
	case parser.Category_Map:
		deref(&ast, &typ)
		return &meta.TypeMeta{
			TypeID:    meta.TTypeID_MAP,
			KeyType:   type2typeInfo(ast, typ.KeyType, index),
			ValueType: type2typeInfo(ast, typ.ValueType, index),
		}
	case parser.Category_List:
		deref(&ast, &typ)
		return &meta.TypeMeta{
			TypeID:    meta.TTypeID_LIST,
			ValueType: type2typeInfo(ast, typ.ValueType, index),
		}
	case parser.Category_Set:
		deref(&ast, &typ)
		return &meta.TypeMeta{
			TypeID:    meta.TTypeID_SET,
			ValueType: type2typeInfo(ast, typ.ValueType, index),
		}
	default:
		panic(fmt.Errorf("unexpected category: %s", typ.Category))
	}
}

func buildMeta(ast *parser.Thrift, sl *parser.StructLike) (sm *meta.StructMeta) {
	sm = &meta.StructMeta{
		Category: sl.Category,
		Name:     sl.Name,
	}
	for i, f := range sl.Fields {
		fm := &meta.FieldMeta{
			FieldID:      int16(f.ID),
			Name:         f.Name,
			Requiredness: meta.TRequiredness(f.Requiredness),
			FieldType:    type2typeInfo(ast, f.Type, i),
		}
		sm.Fields = append(sm.Fields, fm)
	}
	return
}

func prettifyBytesLiteral(s string) string {
	var rs []rune
	cnt := 0
	for _, r := range s {
		if r == '}' {
			rs = append(rs, ',', '\n')
		}
		rs = append(rs, r)
		if r == ',' {
			cnt++
			if cnt == 16 {
				rs = append(rs, '\n')
				cnt = 0
			}
		}
		if r == '{' {
			rs = append(rs, '\n')
		}
	}
	return string(rs)
}
