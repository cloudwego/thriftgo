package trimmer

import (
	"fmt"
	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrimmer(t *testing.T) {
	t.Run("trim AST - case 1", testCase1)
	t.Run("trim AST - test many", testMany)
}

// test single file ast trimming
func testCase1(t *testing.T) {
	trimmer1, err := newTrimmer(nil, "")
	test.Assert(t, err == nil, err)
	filename := filepath.Join("test_cases", "sample1.thrift")
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

// test multiple files for trim_ast
func testMany(t *testing.T) {
	files, err := os.ReadDir("test_cases")
	check(err)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".thrift") {
			os.Mkdir("tmp", os.ModePerm)
			t.Log("testing " + file.Name())
			t.Setenv("verbose", "true")
			trimmer1, err := newTrimmer(nil, "")
			test.Assert(t, err == nil, err)
			filename := filepath.Join("test_cases", file.Name())
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

			req := &plugin.Request{
				Version:    "0.2.12",
				OutputPath: "tmp/gen-go",
				Recursive:  true,
				AST:        ast,
				Language:   "go",
			}

			runCmd("rm -rf gen-go", filename)
			runCmd("rm -f go.mod", filename)
			runCmd("rm -f go.sum", filename)
			runCmd("go mod init tg", filename)
			runCmd("mkdir gen-go", filename)

			var g generator.Generator
			_ = g.RegisterBackend(new(golang.GoBackend))
			logger := backend.DummyLogFunc()
			out := &generator.LangSpec{Language: "go", Options: []plugin.Option{{Name: "package_prefix", Desc: "tg/gen-go"}}}
			arg := &generator.Arguments{Out: out, Req: req, Log: logger}
			res := g.Generate(arg)
			logger.MultiWarn(res.Warnings)
			err = g.Persist(res)
			check(err)

			runCmd("go mod edit -replace github.com/apache/thrift=github.com/apache/thrift@v0.13.0", filename)
			runCmd("go mod tidy", filename)
			runCmd("go build ./...", filename)
			os.RemoveAll("tmp")
		}
	}

}

func runCmd(command string, filename string) {
	splitCommand := strings.Split(command, " ")
	cmd := exec.Command(splitCommand[0], splitCommand[1:]...)
	cmd.Dir = "tmp"
	err := cmd.Run()
	if err != nil {
		stderr := ""
		io.WriteString(cmd.Stderr, stderr)
		fmt.Printf("run %s error for %s : %v\n", command, filename, err.Error())
		fmt.Println(stderr)
		panic(err)
	}
}
