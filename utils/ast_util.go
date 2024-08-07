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

package utils

import (
	"github.com/cloudwego/thriftgo/parser"
)

// reuse builtin types
var builtinTypes = map[string]bool{
	"void":   true,
	"bool":   true,
	"byte":   true,
	"i8":     true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"double": true,
	"string": true,
	"binary": true,
	"list":   true,
	"map":    true,
	"set":    true,
}

// TypeToStructLike try to find the defined parser.StructLike of a parser.Type in ast
func GetStructLike(name string, ast *parser.Thrift) *parser.StructLike {
	tname := name
	if builtinTypes[tname] {
		return nil
	}
	typePkg, typeName := SplitSubfix(name)
	if typePkg != "" {
		ref, ok := ast.GetReference(typePkg)
		if !ok {
			return nil
		}
		ast = ref
	}
	if _, ok := ast.GetEnum(typeName); ok {
		return nil
	}
	if typDef, ok := ast.GetTypedef(typeName); ok {
		return GetStructLike(typDef.Type.Name, ast)
	}
	st, ok := ast.GetStruct(typeName)
	if !ok {
		st, ok = ast.GetUnion(typeName)
		if !ok {
			st, _ = ast.GetException(typeName)
		}
	}
	return st
}
