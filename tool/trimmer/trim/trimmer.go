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

	"github.com/cloudwego/thriftgo/utils/dir_utils"

	"github.com/dlclark/regexp2"

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
	trimMethods      []*regexp2.Regexp
	matchGoName      bool
	trimMethodValid  []bool
	preserveRegex    *regexp.Regexp
	forceTrimming    bool
	preservedStructs []string
	structsTrimmed   int
	fieldsTrimmed    int
	extServices      []*parser.Service
	PreservedFiles   []string
}

type TrimASTArg struct {
	Ast             *parser.Thrift
	TrimMethods     []string
	Preserve        *bool
	MatchGoName     *bool
	PreserveStructs []string
	PreservedFiles  []string
}

type TrimResultInfo struct {
	StructsTrimmed int
	FieldsTrimmed  int
	StructsTotal   int
	FieldsTotal    int
}

func (t *TrimResultInfo) StructsLeft() int {
	return t.StructsTotal - t.StructsTrimmed
}

func (t *TrimResultInfo) FieldsLeft() int {
	return t.FieldsTotal - t.FieldsTrimmed
}

func (t *TrimResultInfo) StructTrimmedPercentage() float64 {
	return float64(t.StructsTrimmed) / float64(t.StructsTotal) * 100
}

func (t *TrimResultInfo) FieldTrimmedPercentage() float64 {
	return float64(t.FieldsTrimmed) / float64(t.FieldsTotal) * 100
}

// TrimAST parse the cfg and trim the single AST
func TrimAST(arg *TrimASTArg) (trimResultInfo *TrimResultInfo, err error) {
	var preservedStructs, preservedFiles []string
	preservedStructs = arg.PreserveStructs
	preservedFiles = arg.PreservedFiles
	if wd, err := dir_utils.Getwd(); err == nil {
		cfg := ParseYamlConfig(wd)
		if cfg != nil {
			if len(arg.TrimMethods) == 0 && len(cfg.Methods) > 0 {
				arg.TrimMethods = cfg.Methods
			}
			if arg.Preserve == nil && !(*cfg.Preserve) {
				preserve := false
				arg.Preserve = &preserve
			}
			if arg.MatchGoName == nil && cfg.MatchGoName != nil {
				arg.MatchGoName = cfg.MatchGoName
			}
			if len(preservedStructs) == 0 {
				preservedStructs = cfg.PreservedStructs
			}
			if len(preservedFiles) == 0 {
				preservedFiles = cfg.PreservedFiles
			}
		}
	}
	forceTrim := false
	if arg.Preserve != nil {
		forceTrim = !*arg.Preserve
	}
	matchGoName := false
	if arg.MatchGoName != nil {
		matchGoName = *arg.MatchGoName
	}
	return doTrimAST(arg.Ast, arg.TrimMethods, forceTrim, matchGoName, preservedStructs, preservedFiles)
}

// doTrimAST trim the single AST, pass method names if -m specified
func doTrimAST(ast *parser.Thrift, trimMethods []string, forceTrimming, matchGoName bool, preservedStructs, preserveFiles []string) (
	trimResultInfo *TrimResultInfo, err error) {
	trimmer, err := newTrimmer(nil, "")
	if err != nil {
		return nil, err
	}
	trimmer.asts[ast.Filename] = ast
	trimmer.trimMethods = make([]*regexp2.Regexp, len(trimMethods))
	trimmer.trimMethodValid = make([]bool, len(trimMethods))
	trimmer.forceTrimming = forceTrimming
	trimmer.matchGoName = matchGoName
	for i, method := range trimMethods {
		parts := strings.Split(method, ".")
		if len(parts) < 2 {
			if len(ast.Services) == 1 {
				trimMethods[i] = ast.Services[0].Name + "." + method
			} else {
				trimMethods[i] = ast.Services[len(ast.Services)-1].Name + "." + method
				// println("please specify service name!\n  -m usage: -m [service_name.method_name]")
				// os.Exit(2)
			}
		}
		trimmer.trimMethods[i], err = regexp2.Compile(trimMethods[i], 0)
		if err != nil {
			return nil, err
		}
	}
	trimmer.preservedStructs = preservedStructs
	trimmer.countStructs(ast)
	originStructsNum := trimmer.structsTrimmed
	originFieldNum := trimmer.fieldsTrimmed
	trimmer.loadPreserveFiles(ast, preserveFiles)
	trimmer.markAST(ast)
	trimmer.traversal(ast, ast.Filename)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		return nil, fmt.Errorf("found include circle:\n\t%s", path)
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	if err != nil {
		return nil, err
	}
	err = semantic.ResolveSymbols(ast)
	if err != nil {
		return nil, err
	}

	for i, method := range trimMethods {
		if !trimmer.trimMethodValid[i] {
			return nil, fmt.Errorf("err: method %s not found!\n", method)
		}
	}

	return &TrimResultInfo{
		StructsTrimmed: trimmer.structsTrimmed,
		FieldsTrimmed:  trimmer.fieldsTrimmed,
		StructsTotal:   originStructsNum,
		FieldsTotal:    originFieldNum,
	}, nil
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

func (t *Trimmer) countStructs(ast *parser.Thrift) {
	t.structsTrimmed += len(ast.Structs) + len(ast.Includes) + len(ast.Services) + len(ast.Unions) + len(ast.Exceptions)
	for _, v := range ast.Structs {
		t.fieldsTrimmed += len(v.Fields)
	}
	for _, v := range ast.Services {
		t.fieldsTrimmed += len(v.Functions)
	}
	for _, v := range ast.Unions {
		t.fieldsTrimmed += len(v.Fields)
	}
	for _, v := range ast.Exceptions {
		t.fieldsTrimmed += len(v.Fields)
	}
	for _, v := range ast.Includes {
		t.countStructs(v.Reference)
	}
}

// make and init a trimmer with related parameters
func newTrimmer(files []string, outDir string) (*Trimmer, error) {
	trimmer := &Trimmer{
		files:  files,
		outDir: outDir,
	}
	trimmer.asts = make(map[string]*parser.Thrift)
	trimmer.marks = make(map[string]map[interface{}]bool)
	pattern := `(?m)^[\s]*(\/\/|#)[\s]*@preserve[\s]*$`
	trimmer.preserveRegex = regexp.MustCompile(pattern)
	return trimmer, nil
}

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}
