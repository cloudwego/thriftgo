// Copyright 2023 CloudWeGo Authors
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

package trim

import (
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

// mark the used part of ast
func (t *Trimmer) markAST(ast *parser.Thrift) {
	t.marks[ast.Filename] = make(map[interface{}]bool)
	t.preProcess(ast, ast.Filename)
	for _, service := range ast.Services {
		t.markService(service, ast, ast.Filename)
	}

	t.markKeptPart(ast, ast.Filename)
}

func (t *Trimmer) markService(svc *parser.Service, ast *parser.Thrift, filename string) {
	if t.marks[filename][svc] {
		return
	}

	if len(t.trimMethods) == 0 {
		t.marks[filename][svc] = true
	}

	for _, function := range svc.Functions {
		if len(t.trimMethods) != 0 {
			funcName := svc.Name + "." + function.Name
			for i, method := range t.trimMethods {
				if ok, _ := method.MatchString(funcName); ok {
					t.marks[filename][svc] = true
					t.markFunction(function, ast, filename)
					t.trimMethodValid[i] = true
				}
			}
			continue
		}
		t.markFunction(function, ast, filename)
	}

	if len(t.trimMethods) != 0 && (svc.Extends != "" || svc.Reference != nil) {
		t.traceExtendMethod(svc, svc, ast, filename)
	}

	if svc.Extends != "" && t.marks[filename][svc] {
		// handle extension
		if svc.Reference != nil {
			theInclude := ast.Includes[svc.Reference.Index]
			t.markInclude(ast.Includes[svc.Reference.Index], filename)
			for _, service := range theInclude.Reference.Services {
				if service.Name == svc.Reference.Name {
					t.markService(service, theInclude.Reference, filename)
					break
				}
			}
		}
	}
}

func (t *Trimmer) markFunction(function *parser.Function, ast *parser.Thrift, filename string) {
	t.marks[filename][function] = true
	for _, arg := range function.Arguments {
		t.markType(arg.Type, ast, filename)
	}
	for _, throw := range function.Throws {
		t.markType(throw.Type, ast, filename)
	}
	if !function.Void {
		t.markType(function.FunctionType, ast, filename)
	}
}

func (t *Trimmer) markType(theType *parser.Type, ast *parser.Thrift, filename string) {
	// plain type
	if theType.Category <= 8 && theType.IsTypedef == nil {
		return
	}

	if theType.KeyType != nil {
		t.markType(theType.KeyType, ast, filename)
	}
	if theType.ValueType != nil {
		t.markType(theType.ValueType, ast, filename)
	}
	baseAST := ast
	if theType.Reference != nil {
		// if referenced, redirect to included ast
		baseAST = ast.Includes[theType.Reference.Index].Reference
		t.markInclude(ast.Includes[theType.Reference.Index], filename)
	}

	if theType.IsTypedef != nil {
		t.markTypeDef(theType, baseAST, filename)
		return
	}

	if theType.Category.IsStruct() {
		for _, str := range baseAST.Structs {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsException() {
		for _, str := range baseAST.Exceptions {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsUnion() {
		for _, str := range baseAST.Unions {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsEnum() {
		for _, enum := range baseAST.Enums {
			if enum.Name == theType.Name || (theType.Reference != nil && enum.Name == theType.Reference.Name) {
				t.markEnum(enum, filename)
				break
			}
		}
	}
}

func (t *Trimmer) markStructLike(str *parser.StructLike, ast *parser.Thrift, filename string) {
	if t.marks[filename][str] {
		return
	}
	t.marks[filename][str] = true
	for _, field := range str.Fields {
		t.markType(field.Type, ast, filename)
	}
}

func (t *Trimmer) markEnum(enum *parser.Enum, filename string) {
	t.marks[filename][enum] = true
}

func (t *Trimmer) markTypeDef(theType *parser.Type, ast *parser.Thrift, filename string) {
	if theType.IsTypedef == nil {
		return
	}

	for i, typedef := range ast.Typedefs {
		if typedef.Alias == theType.Name {
			if !t.marks[filename][ast.Typedefs[i]] {
				t.marks[filename][ast.Typedefs[i]] = true
				t.markType(typedef.Type, ast, filename)
			}
			return
		}
	}
}

func (t *Trimmer) markInclude(include *parser.Include, filename string) {
	include.Reference.Name2Category = nil
	if t.marks[filename][include] {
		return
	}
	t.marks[filename][include] = true
	// t.markKeptPart(include.Reference, filename)
}

func (t *Trimmer) markKeptPart(ast *parser.Thrift, filename string) {
	for _, constant := range ast.Constants {
		t.markType(constant.Type, ast, filename)
	}

	for _, typedef := range ast.Typedefs {
		t.markType(typedef.Type, ast, filename)
	}

	if !t.forceTrimming {
		for _, str := range ast.Structs {
			if !t.marks[filename][str] && t.checkPreserve(str) {
				t.markStructLike(str, ast, filename)
			}
		}

		for _, str := range ast.Unions {
			if !t.marks[filename][str] && t.checkPreserve(str) {
				t.markStructLike(str, ast, filename)
			}
		}

		for _, str := range ast.Exceptions {
			if !t.marks[filename][str] && t.checkPreserve(str) {
				t.markStructLike(str, ast, filename)
			}
		}
	}
}

// for -m, trace the extends and find specified method to base on
func (t *Trimmer) traceExtendMethod(father, svc *parser.Service, ast *parser.Thrift, filename string) (ret bool) {
	for _, function := range svc.Functions {
		funcName := father.Name + "." + function.Name
		for i, method := range t.trimMethods {
			if ok, _ := method.MatchString(funcName); ok {
				t.marks[filename][svc] = true
				t.markFunction(function, ast, filename)
				t.trimMethodValid[i] = true
				ret = true
			}
		}
	}
	if svc.Extends != "" {
		var nextSvc *parser.Service
		var nextAst *parser.Thrift
		if svc.Reference == nil {
			for i, extend := range ast.Services {
				if extend.Name == svc.Extends {
					nextSvc = ast.Services[i]
					nextAst = ast
					break
				}
			}
		} else {
			for i, extend := range ast.Includes[svc.Reference.Index].Reference.Services {
				if extend.Name == svc.Reference.Name {
					nextSvc = ast.Includes[svc.Reference.Index].Reference.Services[i]
					nextAst = ast.Includes[svc.Reference.Index].Reference
					break
				}
			}
		}
		back := t.traceExtendMethod(father, nextSvc, nextAst, filename)
		ret = back || ret
	}
	if ret {
		t.marks[filename][svc] = true
		if svc.Reference != nil {
			t.markInclude(ast.Includes[svc.Reference.Index], filename)
		}
	}
	return ret
}

// check for @Preserve comments
func (t *Trimmer) checkPreserve(theStruct *parser.StructLike) bool {
	if t.forceTrimming {
		return false
	}
	for _, name := range t.preservedStructs {
		if name == theStruct.Name {
			return true
		}
	}
	return t.preserveRegex.MatchString(strings.ToLower(theStruct.ReservedComments))
}
