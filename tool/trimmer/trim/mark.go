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

	"github.com/cloudwego/thriftgo/utils/dir_utils"

	"github.com/cloudwego/thriftgo/parser"
)

// mark the used part of ast
func (t *Trimmer) markAST(ast *parser.Thrift) {
	t.marks[ast.Filename] = make(map[interface{}]struct{})
	t.preProcess(ast, ast.Filename)
	for _, service := range ast.Services {
		t.markService(service, ast, ast.Filename)
	}
	t.cleanServiceExtends()

	t.markKeptPart(ast, ast.Filename)
}

func toGoName(input string) string {
	words := strings.Split(input, "_")
	var result strings.Builder
	for _, word := range words {
		if word != "" {
			upperWord := strings.ToUpper(string(word[0])) + word[1:]
			result.WriteString(upperWord)
		}
	}
	return result.String()
}

func (t *Trimmer) markService(svc *parser.Service, ast *parser.Thrift, filename string) {
	currentMap := t.marks[filename]
	if _, ok := currentMap[svc]; ok {
		return
	}

	if len(t.trimMethods) == 0 {
		currentMap[svc] = struct{}{}
	}

	for _, function := range svc.Functions {
		if len(t.trimMethods) != 0 {
			funcName := svc.Name + "." + function.Name
			for i, method := range t.trimMethods {
				if t.matchGoName {
					funcName = svc.Name + "." + toGoName(function.Name)
				}
				if ok, _ := method.MatchString(funcName); ok {
					if funcName == method.String() || !strings.HasPrefix(funcName, method.String()) {
						currentMap[svc] = struct{}{}
						t.markFunction(function, ast, filename)
						t.trimMethodValid[i] = true
					}
				}
			}
			continue
		}
		t.markFunction(function, ast, filename)
	}

	if len(t.trimMethods) != 0 && (svc.Extends != "" || svc.Reference != nil) {
		t.traceExtendMethod([]*parser.Service{svc}, svc, ast, filename)
	}

	if svc.Extends != "" {
		if _, ok := currentMap[svc]; ok {
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
}

func (t *Trimmer) markFunction(function *parser.Function, ast *parser.Thrift, filename string) {
	t.marks[filename][function] = struct{}{}
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
	currentMap := t.marks[filename]
	if _, ok := currentMap[str]; ok {
		return
	}
	currentMap[str] = struct{}{}
	for _, field := range str.Fields {
		t.markType(field.Type, ast, filename)
	}
}

func (t *Trimmer) markEnum(enum *parser.Enum, filename string) {
	t.marks[filename][enum] = struct{}{}
}

func (t *Trimmer) markTypeDef(theType *parser.Type, ast *parser.Thrift, filename string) {
	if theType.IsTypedef == nil {
		return
	}

	currentMap := t.marks[filename]
	for i, typedef := range ast.Typedefs {
		if typedef.Alias == theType.Name {
			if _, ok := currentMap[ast.Typedefs[i]]; !ok {
				currentMap[ast.Typedefs[i]] = struct{}{}
				t.markType(typedef.Type, ast, filename)
			}
			return
		}
	}
}

func (t *Trimmer) markInclude(include *parser.Include, filename string) {
	include.Reference.Name2Category = nil
	currentMap := t.marks[filename]
	if _, ok := currentMap[include]; ok {
		return
	}
	currentMap[include] = struct{}{}
	// t.markKeptPart(include.Reference, filename)
}

func (t *Trimmer) markServiceExtends(svc *parser.Service) {
	if t.extServices == nil {
		t.extServices = []*parser.Service{svc}
	} else {
		t.extServices = append(t.extServices, svc)
	}
}

func (t *Trimmer) cleanServiceExtends() {
	for _, svc := range t.extServices {
		svc.Reference = nil
		svc.Extends = ""
	}
}

var keptPartCache = make(map[*parser.Thrift]bool, 200)

func (t *Trimmer) markKeptPart(ast *parser.Thrift, filename string) (ret bool) {
	if kept, ok := keptPartCache[ast]; ok {
		return kept
	}
	defer func() {
		keptPartCache[ast] = ret
	}()

	for _, constant := range ast.Constants {
		t.markType(constant.Type, ast, filename)
		ret = true
	}

	for _, typedef := range ast.Typedefs {
		t.markType(typedef.Type, ast, filename)
		ret = true
	}

	if !t.forceTrimming {
		currentMap := t.marks[filename]
		structs := make([]*parser.StructLike, 0, len(ast.Structs)+len(ast.Unions)+len(ast.Exceptions))
		structs = append(structs, ast.Structs...)
		structs = append(structs, ast.Unions...)
		structs = append(structs, ast.Exceptions...)
		for _, str := range structs {
			_, ok := currentMap[str]
			if !ok && t.checkPreserve(str) {
				t.markStructLike(str, ast, filename)
				ret = true
			}
		}
	}
	return
}

// for -m, trace the extends and find specified method to base on
func (t *Trimmer) traceExtendMethod(fathers []*parser.Service, svc *parser.Service, ast *parser.Thrift, filename string) (ret bool) {
	currentMap := t.marks[filename]
	for _, function := range svc.Functions {
		for _, father := range fathers {
			// 子 method 写了来自 extends 的某个名字的时候，都间接向上查找，遍历所有子节点的名字尝试匹配
			funcName := father.Name + "." + function.Name
			for i, method := range t.trimMethods {
				if ok, _ := method.MatchString(funcName); ok {
					currentMap[svc] = struct{}{}
					t.markFunction(function, ast, filename)
					t.trimMethodValid[i] = true
					ret = true
				}
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
		back := t.traceExtendMethod(append(fathers, nextSvc), nextSvc, nextAst, filename)
		if !back {
			t.markServiceExtends(svc)
		}
		ret = back || ret
	}
	if ret {
		currentMap[svc] = struct{}{}
		if svc.Reference != nil {
			t.markInclude(ast.Includes[svc.Reference.Index], filename)
		}
	}
	return ret
}

// check for @Preserve comments
func (t *Trimmer) checkPreserve(theStruct *parser.StructLike) (preserve bool) {
	if t.forceTrimming {
		return false
	}
	if res, ok := t.preserveCache[theStruct]; ok {
		return res
	}
	defer func() {
		t.preserveCache[theStruct] = preserve
	}()

	currentStructName := theStruct.Name
	if t.matchGoName {
		currentStructName = toGoName(currentStructName)
	}
	if _, exists := t.preservedStructsMap[currentStructName]; exists {
		preserve = true
		return
	}

	if !t.disablePreserveComment {
		// 当 struct 总量相当大的时候，关闭 comment 校验，速度会提高很多
		if t.preserveRegex.MatchString(strings.ToLower(theStruct.ReservedComments)) {
			preserve = true
			return
		}
	}

	// 如果整个文件也是要保留的，那么里面的结构体也不删除
	_, ok := t.preserveFileStructs[theStruct]
	preserve = ok
	return
}

func (t *Trimmer) loadPreserveFiles(ast *parser.Thrift, preserveFiles []string) {
	preserveFilesMap := map[string]bool{}
	for _, fn := range preserveFiles {
		// 这里统一转换为绝对路径
		absFn, err := dir_utils.ToAbsolute(fn)
		if err == nil {
			fn = absFn
		}
		preserveFilesMap[absFn] = true
	}
	t.preserveFileStructs = map[*parser.StructLike]struct{}{}
	for th := range ast.DepthFirstSearch() {
		// 两边都用绝对路径来对比，ast 里的 filename 有时候是相对有时候是绝对
		absFilename, err := dir_utils.ToAbsolute(th.Filename)
		if err == nil && preserveFilesMap[absFilename] {
			for _, st := range th.Structs {
				t.preserveFileStructs[st] = struct{}{}
			}
		}
	}
}
