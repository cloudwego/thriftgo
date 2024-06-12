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
	"errors"
	"github.com/cloudwego/thriftgo/utils/dir_utils"
	"os"
	"regexp"

	"github.com/dlclark/regexp2"

	"github.com/cloudwego/thriftgo/parser"
)

type Trimmer struct {
	files []string
	// ast of the file
	asts map[string]*parser.Thrift
	// mark the parts of the file's ast that is used
	// key is IDL filename
	marks  map[string]map[string]bool
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
}

type TrimASTArg struct {
	Ast         *parser.Thrift
	TrimMethods []string
	Preserve    *bool
	MatchGoName *bool
}

type TrimASTWithComposeArg struct {
	Cfg              *IDLComposeArguments
	TargetAST        *parser.Thrift
	ReadCfgFromLocal bool
}

// TrimAST parse the cfg and trim the single AST
func TrimAST(arg *TrimASTArg) (structureTrimmed int, fieldTrimmed int, err error) {
	var preservedStructs []string
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
			preservedStructs = cfg.PreservedStructs
		}
	}
	forceTrim := false
	if arg.Preserve != nil {
		forceTrim = !*arg.Preserve
	}
	preserve := !forceTrim
	matchGoName := false
	if arg.MatchGoName != nil {
		matchGoName = *arg.MatchGoName
	}
	return TrimASTWithCompose(&TrimASTWithComposeArg{
		Cfg: &IDLComposeArguments{
			IDLs: map[string]*IDLArguments{
				arg.Ast.Filename: {
					Trimmer: &YamlArguments{
						Methods:          arg.TrimMethods,
						Preserve:         &preserve,
						PreservedStructs: preservedStructs,
						MatchGoName:      &matchGoName,
					},
				},
			},
		},
		TargetAST: arg.Ast,
	})
}

func TrimASTWithCompose(arg *TrimASTWithComposeArg) (structureTrimmed int, fieldTrimmed int, err error) {
	if arg == nil {
		return structureTrimmed, fieldTrimmed, errors.New("TrimASTWithCompose is nil")
	}
	if arg.TargetAST == nil {
		return structureTrimmed, fieldTrimmed, errors.New("TrimASTWithCompose.TargetAST is nil")
	}
	cfg := arg.Cfg
	// When ReadCfgFromLocal is set, local cfg has higher priority and the passed cfg would be ignored
	if arg.ReadCfgFromLocal {
		wd, err := dir_utils.Getwd()
		if err == nil {
			cfg = extractIDLComposeConfigFromDir(wd, arg.TargetAST.Filename)
		}
	}
	if cfg == nil {
		cfg = newIDLComposeArgumentsWithTargetAST(arg.TargetAST.Filename)
	}
	cfg.setDefault()

	trimmer, err := newTrimmer(nil, "")
	if err != nil {
		return structureTrimmed, fieldTrimmed, err
	}

	for path, idlArg := range cfg.IDLs {
		var ast *parser.Thrift
		if arg.TargetAST.Filename == path {
			ast = arg.TargetAST
		} else {
			ast = parseAndCheckAST(path, nil, true)
		}
		trimmer.markAST(ast, idlArg.Trimmer)
	}
	// trimmer.marks now have the complete context, we can traverse the target AST
	trimmer.traversal(arg.TargetAST)
	checkAST(arg.TargetAST)

	return trimmer.structsTrimmed, trimmer.fieldsTrimmed, nil
}

// Trim to trim thrift files to remove unused fields
func Trim(files, includeDir []string, outDir string) error {
	trimmer, err := newTrimmer(files, outDir)
	if err != nil {
		return err
	}

	for _, filename := range files {
		// go through parse process
		ast := parseAndCheckAST(filename, includeDir, true)
		trimmer.asts[filename] = ast
		trimmer.markAST(ast, nil)
		// TODO: handle multi files and dump to 'xxx.thrift'
	}

	return nil
}

func (t *Trimmer) countStructs(ast *parser.Thrift) {
	// refresh
	t.fieldsTrimmed = 0
	t.structsTrimmed = 0
	t.doCountStructs(ast)
}

func (t *Trimmer) doCountStructs(ast *parser.Thrift) {
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
		t.doCountStructs(v.Reference)
	}
}

// make and init a trimmer with related parameters
func newTrimmer(files []string, outDir string) (*Trimmer, error) {
	trimmer := &Trimmer{
		files:  files,
		outDir: outDir,
	}
	trimmer.asts = make(map[string]*parser.Thrift)
	trimmer.marks = make(map[string]map[string]bool)
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
