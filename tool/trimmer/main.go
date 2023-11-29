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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
	"github.com/cloudwego/thriftgo/tool/trimmer/trim"
	"github.com/cloudwego/thriftgo/version"
)

var (
	a Arguments
	g generator.Generator
)

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}

func main() {
	// you can execute "go install" to install this tool and execute "trimmer" or "trimmer -version"
	err := a.Parse(os.Args)
	if err != nil {
		help()
		check(err)
	}
	if a.AskVersion {
		println("thriftgo", version.ThriftgoVersion)
		os.Exit(0)
	}

	var preserveInput *bool
	if a.Preserve != "" {
		preserve, err := strconv.ParseBool(a.Preserve)
		if err != nil {
			help()
			os.Exit(2)
		}
		preserveInput = &preserve
	}

	// parse file to ast
	ast, err := parser.ParseFile(a.IDL, nil, true)
	check(err)
	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	check(err)
	check(semantic.ResolveSymbols(ast))
	structs, fields := countStructs(ast)

	// trim ast
	_, _, err = trim.TrimAST(&trim.TrimASTArg{
		Ast: ast, TrimMethods: a.Methods, Preserve: preserveInput,
	})
	check(err)

	// dump the trimmed ast to idl
	idl, err := dump.DumpIDL(ast)
	check(err)

	file, err := os.Stat(a.OutputFile)
	if a.OutputFile == "" || err != nil {
		if a.OutputFile == "" && a.Recurse == "" {
			parts := strings.Split(ast.Filename, ".")
			parts = parts[:len(parts)-1]
			a.OutputFile = strings.Join(parts, ".")
			a.OutputFile = a.OutputFile + "_trimmed.thrift"
		}
	} else if file.IsDir() && a.Recurse == "" {
		parts := strings.Split(a.IDL, string(filepath.Separator))
		realSourceFilename := parts[len(parts)-1]
		a.OutputFile = a.OutputFile + string(filepath.Separator) + realSourceFilename
	}

	if a.Recurse != "" {
		if err != nil {
			err = os.MkdirAll(a.OutputFile, os.ModePerm)
		}
		file, err := os.Stat(a.OutputFile)
		if err != nil || !file.IsDir() {
			if err != nil {
				a.OutputFile = "trimmed_idl"
				err = os.MkdirAll(a.OutputFile, os.ModePerm)
				if err != nil {
					println("-o should be set as a valid dir to enable -r", err.Error())
					os.Exit(2)
				}
			} else {
				println("-o should be set as a valid dir to enable -r")
				os.Exit(2)
			}
		}
		relPath, err := filepath.Rel(a.Recurse, a.OutputFile)
		if err == nil && (len(relPath) < 2 || relPath[:2] != "..") {
			println("output-dir should be set outside of -r base-dir to avoid overlay")
			os.Exit(2)
		}
		recurseDump(ast, a.Recurse, a.OutputFile)
	} else {
		check(writeStringToFile(a.OutputFile, idl))
		structsNew, fieldsNew := countStructs(ast)
		fmt.Printf("removed %d unused structures with %d fields\n", structs-structsNew, fields-fieldsNew)
	}
	println("success, dump to", a.OutputFile)

	os.Exit(0)
}

func recurseDump(ast *parser.Thrift, sourceDir, outDir string) {
	if ast == nil {
		return
	}
	out, err := dump.DumpIDL(ast)
	check(err)
	relativeUrl, err := filepath.Rel(sourceDir, ast.Filename)
	if err != nil {
		println("-r input err, range should cover all the target IDLs;", err.Error())
		os.Exit(2)
	}
	outputFileUrl := filepath.Join(outDir, relativeUrl)
	err = os.MkdirAll(filepath.Dir(outputFileUrl), os.ModePerm)
	if err != nil {
		println("mkdir", filepath.Dir(outputFileUrl), "error:", err.Error())
		os.Exit(2)
	}
	check(writeStringToFile(outputFileUrl, out))
	for _, includes := range ast.Includes {
		recurseDump(includes.Reference, sourceDir, outDir)
	}
}

func writeStringToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func countStructs(ast *parser.Thrift) (structs, fields int) {
	structs += len(ast.Structs) + len(ast.Includes) + len(ast.Services) + len(ast.Unions) + len(ast.Exceptions)
	for _, v := range ast.Structs {
		fields += len(v.Fields)
	}
	for _, v := range ast.Services {
		fields += len(v.Functions)
	}
	for _, v := range ast.Unions {
		fields += len(v.Fields)
	}
	for _, v := range ast.Exceptions {
		fields += len(v.Fields)
	}
	return structs, fields
}
