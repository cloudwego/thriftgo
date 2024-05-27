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
	trimmer.markAST(ast, nil)
	trimmer.traversal(ast)

	test.Assert(t, len(ast.Structs) == 7)
	test.Assert(t, len(ast.Includes) == 2)
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
	trimmer.markAST(ast, nil)
	trimmer.traversal(ast)
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

	_, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: methods,
		Preserve:    nil,
	})
	check(err)
	test.Assert(t, len(ast.Services[0].Functions) == 1)
}

func TestTrimMethodWithExtend(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_extend", "common1.thrift")
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
	methods[0] = "Echo"

	_, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: methods,
		Preserve:    nil,
	})
	check(err)
	// for common1.thrift, Common1Struct1 and ProcessCommon1 are trimmed
	test.Assert(t, len(ast.Structs) <= 0)
	test.Assert(t, len(ast.Services) == 1)
	test.Assert(t, len(ast.Services[0].Functions) <= 0)

	// for common2.thrift, Common2Struct1 and ProcessCommon2 are trimmed
	common2AST := ast.Includes[0].Reference
	test.Assert(t, len(common2AST.Structs) <= 0)
	test.Assert(t, len(common2AST.Services) == 1)
	test.Assert(t, len(common2AST.Services[0].Functions) <= 0)

	// for common3.thrift, Common3Struct1 and ProcessCommon3 are trimmed, Echo and Common3Struct2 are preserved
	common3AST := common2AST.Includes[0].Reference
	test.Assert(t, len(common3AST.Structs) == 1)
	test.Assert(t, len(common3AST.Services) == 1)
	test.Assert(t, len(common3AST.Services[0].Functions) == 1)
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

	_, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: nil,
		Preserve:    &preserve,
	})
	check(err)
	test.Assert(t, len(ast.Structs) == 0)
}

func TestTrimASTWithCompose(t *testing.T) {
	testcases := []struct {
		desc       string
		composeArg func() *TrimASTWithComposeArg
		expect     func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error)
	}{
		{
			desc: "two unrelated idls refer to a common idl",
			composeArg: func() *TrimASTWithComposeArg {
				test1 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test1.thrift")
				test2 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test2.thrift")
				test1AST := parseAndCheckAST(test1, nil, true)
				return &TrimASTWithComposeArg{
					Cfg: &IDLComposeArguments{
						IDLs: map[string]*IDLArguments{
							test1: nil,
							test2: nil,
						},
					},
					TargetAST: test1AST,
				}
			},
			expect: func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error) {
				// all the elements have been preserved
				test.Assert(t, res.StructsTrimmed == 0)
				test.Assert(t, res.FieldsTrimmed == 0)
				test.Assert(t, res.StructsTotal == 5)
				test.Assert(t, res.FieldsTotal == 5)
				test.Assert(t, err == nil)
				// all the content in common.thrift has been marked to be preserved
				commonAst1 := ast.Includes[0].Reference
				test.Assert(t, len(commonAst1.Enums) == 2)
				test.Assert(t, len(commonAst1.Structs) == 2)
			},
		},
		{
			desc: "two unrelated idls refer to two common idls",
			composeArg: func() *TrimASTWithComposeArg {
				test3 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test3.thrift")
				test4 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test4.thrift")
				test3AST := parseAndCheckAST(test3, nil, true)
				return &TrimASTWithComposeArg{
					Cfg: &IDLComposeArguments{
						IDLs: map[string]*IDLArguments{
							test3: nil,
							test4: nil,
						},
					},
					TargetAST: test3AST,
				}
			},
			expect: func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error) {
				// all the elements have been preserved
				test.Assert(t, res.StructsTrimmed == 0)
				test.Assert(t, res.FieldsTrimmed == 0)
				test.Assert(t, res.StructsTotal == 8)
				test.Assert(t, res.FieldsTotal == 9)
				test.Assert(t, err == nil)
				commonAst1 := ast.Includes[0].Reference
				test.Assert(t, len(commonAst1.Enums) == 2)
				test.Assert(t, len(commonAst1.Structs) == 2)
				commonAst2 := ast.Includes[1].Reference
				test.Assert(t, len(commonAst2.Enums) == 2)
				test.Assert(t, len(commonAst2.Structs) == 2)

			},
		},
		{
			desc: "use local idl_compose.yaml",
			composeArg: func() *TrimASTWithComposeArg {
				test1 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test1.thrift")
				test1Ast := parseAndCheckAST(test1, nil, true)
				return &TrimASTWithComposeArg{
					TargetAST:        test1Ast,
					ReadCfgFromLocal: true,
				}
			},
			expect: func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error) {
				// all the elements have been preserved
				test.Assert(t, res.StructsTrimmed == 0)
				test.Assert(t, res.FieldsTrimmed == 0)
				test.Assert(t, err == nil)
				// all the content in common.thrift has been marked to be preserved
				commonAst1 := ast.Includes[0].Reference
				test.Assert(t, len(commonAst1.Enums) == 2)
				test.Assert(t, len(commonAst1.Structs) == 2)
			},
		},
		{
			desc: "test5.thrift refer to test6.thrift, they both refer to two common idls",
			composeArg: func() *TrimASTWithComposeArg {
				test5 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test5.thrift")
				test6 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test6.thrift")
				test5AST := parseAndCheckAST(test5, nil, true)
				return &TrimASTWithComposeArg{
					Cfg: &IDLComposeArguments{
						IDLs: map[string]*IDLArguments{
							test5: nil,
							test6: nil,
						},
					},
					TargetAST: test5AST,
				}
			},
			expect: func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error) {
				// all the elements have been preserved
				test.Assert(t, res.StructsTrimmed == 0)
				test.Assert(t, res.FieldsTrimmed == 0)
				test.Assert(t, err == nil)
				// common parts
				commonAST1 := ast.Includes[0].Reference
				test.Assert(t, len(commonAST1.Enums) == 2)
				test.Assert(t, len(commonAST1.Structs) == 2)
				commonAST2 := ast.Includes[1].Reference
				test.Assert(t, len(commonAST2.Enums) == 2)
				test.Assert(t, len(commonAST2.Structs) == 2)
				// self
				test.Assert(t, len(ast.Structs) == 2)
			},
		},
		{
			desc: "two unrelated idls refer to two common idls with preserved methods and structs",
			composeArg: func() *TrimASTWithComposeArg {
				test7 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test7.thrift")
				test8 := filepath.Join("..", "test_cases", "multiple_idls", "multiple_idls_include_common", "test8.thrift")
				test7AST := parseAndCheckAST(test7, nil, true)
				return &TrimASTWithComposeArg{
					Cfg: &IDLComposeArguments{
						IDLs: map[string]*IDLArguments{
							test7: {
								Trimmer: &YamlArguments{
									Methods:          []string{"Process"},
									PreservedStructs: []string{"Test7Struct3"},
								},
							},
							test8: nil,
						},
					},
					TargetAST: test7AST,
				}
			},
			expect: func(t *testing.T, ast *parser.Thrift, res *TrimResultInfo, err error) {
				test.Assert(t, res.StructsTrimmed == 0, res.StructsTrimmed)
				// Echo Method has been trimmed
				test.Assert(t, res.FieldsTrimmed == 1)
				test.Assert(t, res.StructsTotal == 10)
				test.Assert(t, res.FieldsTotal == 12)
				test.Assert(t, err == nil)
				commonAst1 := ast.Includes[0].Reference
				test.Assert(t, len(commonAst1.Enums) == 2)
				test.Assert(t, len(commonAst1.Structs) == 2)
				commonAst2 := ast.Includes[1].Reference
				test.Assert(t, len(commonAst2.Enums) == 2)
				test.Assert(t, len(commonAst2.Structs) == 2)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			arg := tc.composeArg()
			res, err := TrimASTWithCompose(arg)
			tc.expect(t, arg.TargetAST, res, err)
		})
	}
}
