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
	"github.com/cloudwego/thriftgo/parser"
)

// traverse and remove the unmarked part of ast
func (t *Trimmer) traversal(ast *parser.Thrift) {
	t.countStructs(ast)
	t.doTraversal(ast)
}

func (t *Trimmer) doTraversal(ast *parser.Thrift) {
	// deal with trimmed statistical data
	filename := ast.Filename
	var listInclude []*parser.Include
	for i := range ast.Includes {
		if t.marks[filename][includePrefix+ast.Includes[i].Path] || len(ast.Includes[i].Reference.Constants)+
			len(ast.Includes[i].Reference.Enums)+len(ast.Includes[i].Reference.Typedefs) > 0 {
			t.doTraversal(ast.Includes[i].Reference)
			listInclude = append(listInclude, ast.Includes[i])
		}
	}
	ast.Includes = listInclude

	var listStruct []*parser.StructLike
	for i := range ast.Structs {
		if t.marks[filename][ast.Structs[i].Name] || t.checkPreserve(ast.Structs[i]) {
			listStruct = append(listStruct, ast.Structs[i])
			t.fieldsTrimmed -= len(ast.Structs[i].Fields)
		}
	}
	ast.Structs = listStruct

	var listUnion []*parser.StructLike
	for i := range ast.Unions {
		if t.marks[filename][ast.Unions[i].Name] || t.checkPreserve(ast.Unions[i]) {
			listUnion = append(listUnion, ast.Unions[i])
			t.fieldsTrimmed -= len(ast.Unions[i].Fields)
		}
	}
	ast.Unions = listUnion

	var listException []*parser.StructLike
	for i := range ast.Exceptions {
		if t.marks[filename][ast.Exceptions[i].Name] || t.checkPreserve(ast.Exceptions[i]) {
			listException = append(listException, ast.Exceptions[i])
			t.fieldsTrimmed -= len(ast.Exceptions[i].Fields)
		}
	}
	ast.Exceptions = listException

	var listService []*parser.Service
	for i := range ast.Services {
		if t.marks[filename][ast.Services[i].Name] {
			var trimmedMethods []*parser.Function
			for j := range ast.Services[i].Functions {
				if t.marks[filename][functionIdentifier(ast.Services[i].Name, ast.Services[i].Functions[j].Name)] {
					trimmedMethods = append(trimmedMethods, ast.Services[i].Functions[j])
				}
			}
			ast.Services[i].Functions = trimmedMethods
			listService = append(listService, ast.Services[i])
			t.fieldsTrimmed -= len(ast.Services[i].Functions)
		}
	}
	ast.Services = listService
	ast.Name2Category = nil
	for _, inc := range ast.Includes {
		inc.Used = nil
	}

	t.structsTrimmed -= len(ast.Structs) + len(ast.Includes) + len(ast.Services) + len(ast.Unions) + len(ast.Exceptions)
}
