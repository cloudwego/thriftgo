package trim

import (
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
	"github.com/cloudwego/thriftgo/semantic"
	"path/filepath"
	"testing"
)

func TestTrimmer(t *testing.T) {
	t.Run("trim AST - case 1", testCase1)
	//t.Run("trim AST - test many", testMany)
}

// test single file ast trimming
func testCase1(t *testing.T) {
	trimmer1, err := newTrimmer(nil, "")
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
	trimmer1.asts[filename] = ast
	trimmer1.markAST(ast)
	trimmer1.traversal(ast, ast.Filename)

	test.Assert(t, len(ast.Structs) == 6)
	test.Assert(t, len(ast.Includes) == 1)
	test.Assert(t, len(ast.Typedefs) == 2)
	test.Assert(t, len(ast.Namespaces) == 1)
	test.Assert(t, len(ast.Includes[0].Reference.Structs) == 2)
	test.Assert(t, len(ast.Includes[0].Reference.Constants) == 2)
	test.Assert(t, len(ast.Includes[0].Reference.Services) == 1)
	test.Assert(t, len(ast.Includes[0].Reference.Namespaces) == 1)
}
