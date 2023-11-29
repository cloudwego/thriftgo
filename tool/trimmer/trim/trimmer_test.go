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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
	"github.com/cloudwego/thriftgo/semantic"
)

// test single file ast trimming
func TestSingleFile(t *testing.T) {
	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	filename := filepath.Join("..", "test_cases", "sample1.thrift")
	ast, err := parser.ParseFile(filename, []string{"test_cases"}, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))
	trimmer.asts[filename] = ast
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	test.Assert(t, len(ast.Structs) == 7)
	test.Assert(t, len(ast.Includes) == 1)
	test.Assert(t, len(ast.Typedefs) == 5)
	test.Assert(t, len(ast.Namespaces) == 1)
	test.Assert(t, len(ast.Includes[0].Reference.Structs) == 2)
	test.Assert(t, len(ast.Includes[0].Reference.Constants) == 2)
	test.Assert(t, len(ast.Includes[0].Reference.Services) == 1)
	test.Assert(t, len(ast.Includes[0].Reference.Namespaces) == 1)
}

func TestInclude(t *testing.T) {
	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	filename := filepath.Join("..", "test_cases/test_include", "example.thrift")
	ast, err := parser.ParseFile(filename, []string{"test_cases/test_include"}, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))
	trimmer.asts[filename] = ast
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker = semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	test.Assert(t, len(ast.Structs) == 0)
	test.Assert(t, len(ast.Includes) == 1)
	test.Assert(t, ast.Includes[0].Used == nil)
}

func TestTrimMethod(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "tests", "dir", "dir2", "test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	methods := make([]string, 1)
	methods[0] = "func1"

	_, _, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: methods,
		Preserve:    nil,
	})
	check(err)
	test.Assert(t, len(ast.Services[0].Functions) == 1)
}

func TestPreserve(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "tests", "dir", "dir2", "test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	preserve := false

	_, _, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: nil,
		Preserve:    &preserve,
	})
	check(err)
	test.Assert(t, len(ast.Structs) == 0)
}
