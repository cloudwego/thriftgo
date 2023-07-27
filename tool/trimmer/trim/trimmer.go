package trim

import (
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"os"
)

type Trimmer struct {
	files []string
	// ast of the file
	asts map[string]*parser.Thrift
	// mark the parts of the file's ast that is used
	marks  map[string]map[interface{}]bool
	outDir string
}

func TrimAST(ast *parser.Thrift) error {
	// todo: simplify this trim function. And seems like trimmer struct is unnecessary to handle file input.
	trimmer1, err := newTrimmer(nil, "")
	if err != nil {
		return err
	}
	trimmer1.asts[ast.Filename] = ast
	trimmer1.markAST(ast)
	trimmer1.traversal(ast, ast.Filename)
	return nil
}

// Trim to trim thrift files to remove unused fields
func Trim(files []string, includeDir []string, outDir string) error {
	trimmer, err := newTrimmer(files, outDir)
	if err != nil {
		return err
	}

	for _, filename := range files {
		// go through parse process
		ast, err := parser.ParseFile(filename, includeDir, true)
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
		//TODO: 多文件处理/dump成.thrift
	}

	return nil
}

// make and init a trimmer with related parameters
func newTrimmer(files []string, outDir string) (*Trimmer, error) {
	trimmer := &Trimmer{
		files:  files,
		outDir: outDir,
	}
	trimmer.asts = make(map[string]*parser.Thrift)
	trimmer.marks = make(map[string]map[interface{}]bool)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			println(fmt.Errorf("input file invalid: %s", file))
		}
	}

	info, err := os.Stat(outDir)
	if err != nil || !info.IsDir() {
		println(fmt.Errorf("output dir invalid: %s", outDir))
	}

	return trimmer, nil
}

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}
