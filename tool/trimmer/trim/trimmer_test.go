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
	test.Assert(t, len(ast.Includes) == 2)
	test.Assert(t, len(ast.Typedefs) == 5)
	test.Assert(t, len(ast.Namespaces) == 1)
	test.Assert(t, len(ast.Enums) == 1, fmt.Sprintf("Expected 1 enum after trimming, got %d", len(ast.Enums)))
	test.Assert(t, ast.Enums[0].Name == "Gender", "Gender enum should be kept")
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

	_, err = TrimAST(&TrimASTArg{
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

	_, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: nil,
		Preserve:    &preserve,
	})
	check(err)
	test.Assert(t, len(ast.Structs) == 0)
}

// Test enum trimming functionality
func TestEnumTrimming(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_enum", "enum_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming: should have 5 enums (Status, UnusedColor, Gender, UnusedPriority, ResponseCode)
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums before trimming, got %d", len(ast.Enums)))

	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	trimmer.asts[filename] = ast
	trimmer.trimEnums = true
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	// After trimming: should only have 3 enums (Status, Gender, ResponseCode)
	// UnusedColor and UnusedPriority should be removed
	test.Assert(t, len(ast.Enums) == 3, fmt.Sprintf("Expected 3 enums after trimming, got %d", len(ast.Enums)))

	// Verify the correct enums are kept
	enumNames := make(map[string]bool)
	for _, enum := range ast.Enums {
		enumNames[enum.Name] = true
	}

	test.Assert(t, enumNames["Status"], "Status enum should be kept")
	test.Assert(t, enumNames["Gender"], "Gender enum should be kept")
	test.Assert(t, enumNames["ResponseCode"], "ResponseCode enum should be kept")
	test.Assert(t, !enumNames["UnusedColor"], "UnusedColor enum should be trimmed")
	test.Assert(t, !enumNames["UnusedPriority"], "UnusedPriority enum should be trimmed")
}

// Test enum trimming with TrimAST API
func TestEnumTrimmingWithAPI(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_enum", "enum_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums before trimming, got %d", len(ast.Enums)))

	// Trim using TrimAST API
	trimEnums := true
	resultInfo, err := TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: nil,
		Preserve:    nil,
		TrimEnums:   &trimEnums,
	})
	check(err)

	// After trimming: should only have 3 enums
	test.Assert(t, len(ast.Enums) == 3, fmt.Sprintf("Expected 3 enums after trimming, got %d", len(ast.Enums)))

	// Verify result info includes enum counts
	test.Assert(t, resultInfo != nil, "Result info should not be nil")
	test.Assert(t, resultInfo.StructsTrimmed == 2, fmt.Sprintf("Expected 2 enums trimmed, got %d", resultInfo.StructsTrimmed))
}

// Test enum trimming with includes
func TestEnumTrimmingWithIncludes(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_enum", "include_test.thrift")
	ast, err := parser.ParseFile(filename, []string{filepath.Join("..", "test_cases", "test_enum")}, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming: base.thrift has 3 enums
	test.Assert(t, len(ast.Includes) == 1, "Should have 1 include")
	baseAST := ast.Includes[0].Reference
	test.Assert(t, len(baseAST.Enums) == 3, fmt.Sprintf("Expected 3 enums in base.thrift before trimming, got %d", len(baseAST.Enums)))

	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	trimmer.asts[filename] = ast
	trimmer.trimEnums = true
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	// After trimming: base.thrift should only have 2 enums (AccountType and Country)
	// UnusedLevel should be removed
	test.Assert(t, len(baseAST.Enums) == 2, fmt.Sprintf("Expected 2 enums in base.thrift after trimming, got %d", len(baseAST.Enums)))

	// Verify the correct enums are kept
	enumNames := make(map[string]bool)
	for _, enum := range baseAST.Enums {
		enumNames[enum.Name] = true
	}

	test.Assert(t, enumNames["AccountType"], "AccountType enum should be kept")
	test.Assert(t, enumNames["Country"], "Country enum should be kept (used via typedef)")
	test.Assert(t, !enumNames["UnusedLevel"], "UnusedLevel enum should be trimmed")
}

// Test TrimBatchContentWithConfig API with enum trimming
func TestTrimBatchContentWithConfigEnums(t *testing.T) {
	// Read test files
	enumTestContent := `
enum UsedEnum {
    VALUE1 = 1,
    VALUE2 = 2
}

enum UnusedEnum {
    VALUE3 = 1,
    VALUE4 = 2
}

struct TestStruct {
    1: UsedEnum field
}

service TestService {
    TestStruct test()
}
`

	IDLFileContentMap := map[string]string{
		"test.thrift": enumTestContent,
	}

	// Trim the content
	trimEnums := true
	trimmedContent, err := TrimBatchContentWithConfig("test.thrift", IDLFileContentMap, TrimASTArg{
		TrimMethods: nil,
		Preserve:    nil,
		TrimEnums:   &trimEnums,
	})
	test.Assert(t, err == nil, err)
	test.Assert(t, trimmedContent != nil, "Trimmed content should not be nil")

	// Parse the trimmed result to verify
	ast, err := parser.ParseString("test.thrift", trimmedContent["test.thrift"])
	test.Assert(t, err == nil, err)

	// Should only have UsedEnum, UnusedEnum should be removed
	test.Assert(t, len(ast.Enums) == 1, fmt.Sprintf("Expected 1 enum after trimming, got %d", len(ast.Enums)))
	test.Assert(t, ast.Enums[0].Name == "UsedEnum", fmt.Sprintf("Expected UsedEnum, got %s", ast.Enums[0].Name))
}

// Test backward compatibility: enums should NOT be trimmed by default
func TestEnumBackwardCompatibility(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_enum", "enum_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming: should have 5 enums
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums before trimming, got %d", len(ast.Enums)))

	// Trim WITHOUT setting TrimEnums (test default behavior)
	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	trimmer.asts[filename] = ast
	// Note: NOT setting trimmer.trimEnums = true
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	// After trimming: should STILL have 5 enums (backward compatibility)
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums after trimming (backward compatibility), got %d", len(ast.Enums)))
}

// Test backward compatibility with TrimAST API
func TestEnumBackwardCompatibilityWithAPI(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_enum", "enum_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums before trimming, got %d", len(ast.Enums)))

	// Trim using TrimAST API WITHOUT setting TrimEnums
	_, err = TrimAST(&TrimASTArg{
		Ast:         ast,
		TrimMethods: nil,
		Preserve:    nil,
		// TrimEnums not set - should default to false
	})
	check(err)

	// After trimming: should STILL have 5 enums (backward compatibility)
	test.Assert(t, len(ast.Enums) == 5, fmt.Sprintf("Expected 5 enums after trimming (backward compatibility), got %d", len(ast.Enums)))
}

// Test constant trimming functionality
func TestConstTrimming(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_const", "const_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming: should have 3 constants
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants before trimming, got %d", len(ast.Constants)))

	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	trimmer.asts[filename] = ast
	trimmer.trimConsts = true
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	// After trimming: should have 0 constants (all unused)
	test.Assert(t, len(ast.Constants) == 0, fmt.Sprintf("Expected 0 constants after trimming, got %d", len(ast.Constants)))
}

// Test constant trimming with TrimAST API
func TestConstTrimmingWithAPI(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_const", "const_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants before trimming, got %d", len(ast.Constants)))

	// Trim using TrimAST API
	trimConsts := true
	resultInfo, err := TrimAST(&TrimASTArg{
		Ast:        ast,
		TrimConsts: &trimConsts,
	})
	check(err)

	// After trimming: should have 0 constants
	test.Assert(t, len(ast.Constants) == 0, fmt.Sprintf("Expected 0 constants after trimming, got %d", len(ast.Constants)))

	// Verify result info includes constant counts
	test.Assert(t, resultInfo != nil, "Result info should not be nil")
	test.Assert(t, resultInfo.StructsTrimmed == 3, fmt.Sprintf("Expected 3 constants trimmed, got %d", resultInfo.StructsTrimmed))
}

// Test backward compatibility: constants should NOT be trimmed by default
func TestConstBackwardCompatibility(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_const", "const_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants before trimming, got %d", len(ast.Constants)))

	// Trim WITHOUT setting TrimConsts (test default behavior)
	trimmer, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	trimmer.asts[filename] = ast
	// Note: NOT setting trimmer.trimConsts = true
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)

	// After trimming: should STILL have 3 constants (backward compatibility)
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants after trimming (backward compatibility), got %d", len(ast.Constants)))
}

// Test backward compatibility with TrimAST API
func TestConstBackwardCompatibilityWithAPI(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "test_const", "const_test.thrift")
	ast, err := parser.ParseFile(filename, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	// Before trimming
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants before trimming, got %d", len(ast.Constants)))

	// Trim using TrimAST API WITHOUT setting TrimConsts
	_, err = TrimAST(&TrimASTArg{
		Ast: ast,
		// TrimConsts not set - should default to false
	})
	check(err)

	// After trimming: should STILL have 3 constants (backward compatibility)
	test.Assert(t, len(ast.Constants) == 3, fmt.Sprintf("Expected 3 constants after trimming (backward compatibility), got %d", len(ast.Constants)))
}

// Test that empty IDL files are not included in the result
func TestTrimBatchContentExcludeEmptyFiles(t *testing.T) {
	// Create main IDL that doesn't use anything from the include
	mainIDL := `
include "base.thrift"

struct MainStruct {
    1: string name
}

service MainService {
    MainStruct getMain()
}
`

	// Create base IDL with only unused structs (structs are trimmed by default)
	baseIDL := `
struct UnusedStruct {
    1: i32 id
    2: string name
}
`

	IDLFileContentMap := map[string]string{
		"main.thrift": mainIDL,
		"base.thrift": baseIDL,
	}

	// Trim the content with default settings
	trimmedContent, err := TrimBatchContent("main.thrift", IDLFileContentMap)
	test.Assert(t, err == nil, err)
	test.Assert(t, trimmedContent != nil, "Trimmed content should not be nil")

	// base.thrift should NOT be in the result because it's completely unused
	_, hasBase := trimmedContent["base.thrift"]
	test.Assert(t, !hasBase, "base.thrift should not be in the result when it's completely empty after trimming")

	// main.thrift should be in the result
	_, hasMain := trimmedContent["main.thrift"]
	test.Assert(t, hasMain, "main.thrift should be in the result")

	// Verify main.thrift still has correct content
	ast, err := parser.ParseString("main.thrift", trimmedContent["main.thrift"])
	test.Assert(t, err == nil, err)
	test.Assert(t, len(ast.Structs) == 1, fmt.Sprintf("Expected 1 struct in main.thrift, got %d", len(ast.Structs)))
	test.Assert(t, len(ast.Services) == 1, fmt.Sprintf("Expected 1 service in main.thrift, got %d", len(ast.Services)))
}

// Test that empty IDL files are excluded when trimming with enum/const options
func TestTrimBatchContentExcludeEmptyFilesWithConfig(t *testing.T) {
	// Create main IDL
	mainIDL := `
include "unused.thrift"

struct MainStruct {
    1: string name
}

service MainService {
    MainStruct getMain()
}
`

	// Create an include file that only has enums/consts which will be trimmed
	unusedIDL := `
enum UnusedEnum {
    VALUE1 = 1,
    VALUE2 = 2
}

const i32 UnusedConst = 42
`

	IDLFileContentMap := map[string]string{
		"main.thrift":   mainIDL,
		"unused.thrift": unusedIDL,
	}

	// Trim the content with enum and const trimming enabled
	trimEnums := true
	trimConsts := true
	trimmedContent, err := TrimBatchContentWithConfig("main.thrift", IDLFileContentMap, TrimASTArg{
		TrimEnums:  &trimEnums,
		TrimConsts: &trimConsts,
	})
	test.Assert(t, err == nil, err)
	test.Assert(t, trimmedContent != nil, "Trimmed content should not be nil")

	// unused.thrift should NOT be in the result because all its content was trimmed
	_, hasUnused := trimmedContent["unused.thrift"]
	test.Assert(t, !hasUnused, "unused.thrift should not be in the result when all content is trimmed")

	// main.thrift should be in the result
	_, hasMain := trimmedContent["main.thrift"]
	test.Assert(t, hasMain, "main.thrift should be in the result")
}

// Test that partially empty IDL files are still included if they have some content
func TestTrimBatchContentIncludePartiallyUsedFiles(t *testing.T) {
	// Create main IDL
	mainIDL := `
include "base.thrift"

struct MainStruct {
    1: base.UsedStruct field
}

service MainService {
    MainStruct getMain()
}
`

	// Create base IDL with both used and unused content
	baseIDL := `
struct UsedStruct {
    1: i32 id
}

struct UnusedStruct {
    1: string name
}

enum UnusedEnum {
    VALUE1 = 1
}
`

	IDLFileContentMap := map[string]string{
		"main.thrift": mainIDL,
		"base.thrift": baseIDL,
	}

	// Trim the content
	trimmedContent, err := TrimBatchContent("main.thrift", IDLFileContentMap)
	test.Assert(t, err == nil, err)
	test.Assert(t, trimmedContent != nil, "Trimmed content should not be nil")

	// base.thrift SHOULD be in the result because it has used content
	baseContent, hasBase := trimmedContent["base.thrift"]
	test.Assert(t, hasBase, "base.thrift should be in the result when it has used content")

	// Verify base.thrift has the used struct but not the unused ones
	ast, err := parser.ParseString("base.thrift", baseContent)
	test.Assert(t, err == nil, err)
	test.Assert(t, len(ast.Structs) == 1, fmt.Sprintf("Expected 1 struct in base.thrift, got %d", len(ast.Structs)))
	test.Assert(t, ast.Structs[0].Name == "UsedStruct", fmt.Sprintf("Expected UsedStruct, got %s", ast.Structs[0].Name))
}
