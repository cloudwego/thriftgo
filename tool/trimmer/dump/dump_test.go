package dump

import (
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestDumpSingle(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "sample1.thrift")
	ast, err := parser.ParseFile(filename, []string{"test_cases"}, true)
	test.Assert(t, err == nil, err)
	out, err := DumpIDL(ast)
	test.Assert(t, err == nil, err)
	println(out)
}

func TestDumpMany(t *testing.T) {
	dir := filepath.Join("..", "test_cases")
	testDir(dir, t)
}

func testDir(dir string, t *testing.T) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return
	}
	var thriftFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".thrift") {
			filePath := filepath.Join(dir, file.Name())
			thriftFiles = append(thriftFiles, filePath)
		}
		if file.IsDir() {
			testDir(filepath.Join(dir, file.Name()), t)
		}
	}

	for _, f := range thriftFiles {
		ast, err := parser.ParseFile(f, []string{"test_cases"}, true)
		test.Assert(t, err == nil, err)
		out, err := DumpIDL(ast)
		test.Assert(t, err == nil, err)
		println("out of ", f, " :")
		println(out)
		println("===================")
	}
}
