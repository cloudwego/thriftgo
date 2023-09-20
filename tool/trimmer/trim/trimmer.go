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
	"os"
	"regexp"
	"strings"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
)

type Trimmer struct {
	files []string
	// ast of the file
	asts map[string]*parser.Thrift
	// mark the parts of the file's ast that is used
	marks  map[string]map[interface{}]bool
	outDir string
	// use -m
	trimMethods     []string
	trimMethodValid []bool
	preserveRegex   *regexp.Regexp
	forceTrimming   bool
}

// TrimAST trim the single AST, pass method names if -m specified
func TrimAST(ast *parser.Thrift, trimMethods []string, forceTrimming bool) error {
	trimmer, err := newTrimmer(nil, "")
	if err != nil {
		return err
	}
	trimmer.asts[ast.Filename] = ast
	trimmer.trimMethods = trimMethods
	trimmer.trimMethodValid = make([]bool, len(trimMethods))
	trimmer.forceTrimming = forceTrimming
	for i, method := range trimMethods {
		parts := strings.Split(method, ".")
		if len(parts) < 2 {
			if len(ast.Services) == 1 {
				trimMethods[i] = ast.Services[0].Name + "." + method
			} else {
				println("please specify service name!\n  -m usage: -m [service_name.method_name]")
				os.Exit(2)
			}
		}
	}
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)
	ast.Name2Category = nil
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))

	for i, method := range trimMethods {
		if !trimmer.trimMethodValid[i] {
			println("err: method", method, "not found!")
			os.Exit(2)
		}
	}

	return nil
}

// Trim to trim thrift files to remove unused fields
func Trim(files, includeDir []string, outDir string) error {
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
		// TODO: handle multi files and dump to 'xxx.thrift'
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
	pattern := `^[\s]*(\/\/|#)[\s]*@preserve[\s]*$`
	trimmer.preserveRegex = regexp.MustCompile(pattern)
	return trimmer, nil
}

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}
