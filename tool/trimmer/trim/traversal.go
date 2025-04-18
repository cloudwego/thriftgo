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
func (t *Trimmer) traversal(ast *parser.Thrift, filename string) {
	currentMap := t.marks[filename]

	listInclude := make([]*parser.Include, 0, len(ast.Includes))
	for i := range ast.Includes {
		_, ok := currentMap[ast.Includes[i]]
		if ok || len(ast.Includes[i].Reference.Constants)+
			len(ast.Includes[i].Reference.Enums)+len(ast.Includes[i].Reference.Typedefs) > 0 {
			t.traversal(ast.Includes[i].Reference, filename)
			listInclude = append(listInclude, ast.Includes[i])
		}
	}
	ast.Includes = listInclude

	listStruct := make([]*parser.StructLike, 0, len(ast.Structs))
	for i := range ast.Structs {
		_, ok := currentMap[ast.Structs[i]]
		if ok || t.checkPreserve(ast.Structs[i]) {
			listStruct = append(listStruct, ast.Structs[i])
			t.fieldsTrimmed -= len(ast.Structs[i].Fields)
		}
	}
	ast.Structs = listStruct

	listUnion := make([]*parser.StructLike, 0, len(ast.Unions))
	for i := range ast.Unions {
		_, ok := currentMap[ast.Unions[i]]
		if ok || t.checkPreserve(ast.Unions[i]) {
			listUnion = append(listUnion, ast.Unions[i])
			t.fieldsTrimmed -= len(ast.Unions[i].Fields)
		}
	}
	ast.Unions = listUnion

	listException := make([]*parser.StructLike, 0, len(ast.Exceptions))
	for i := range ast.Exceptions {
		_, ok := currentMap[ast.Exceptions[i]]
		if ok || t.checkPreserve(ast.Exceptions[i]) {
			listException = append(listException, ast.Exceptions[i])
			t.fieldsTrimmed -= len(ast.Exceptions[i].Fields)
		}
	}
	ast.Exceptions = listException

	listService := make([]*parser.Service, 0, len(ast.Services))
	for i := range ast.Services {
		_, ok := currentMap[ast.Services[i]]
		if ok {
			if len(t.trimMethods) != 0 {
				var trimmedMethods []*parser.Function
				for j := range ast.Services[i].Functions {
					_, okFunc := currentMap[ast.Services[i].Functions[j]]
					if okFunc {
						trimmedMethods = append(trimmedMethods, ast.Services[i].Functions[j])
					}
				}
				ast.Services[i].Functions = trimmedMethods
			}
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
