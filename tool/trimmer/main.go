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
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
	"github.com/cloudwego/thriftgo/tool/trimmer/trim"
	"os"

	"github.com/cloudwego/thriftgo/generator"
)

// Version of trimmer tool.
const Version = "0.0.1"

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
	println("IDL TRIMMER.....")
	check(a.Parse(os.Args))
	if a.AskVersion {
		println("thriftgo trimmer tool ", Version)
		os.Exit(0)
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

	// trim ast
	check(trim.TrimAST(ast))

	// dump the trimmed ast to idl
	idl, err := dump.DumpIDL(ast)
	check(err)
	check(writeStringToFile(a.OutputFile, idl))

	os.Exit(0)
}

func writeStringToFile(filename string, content string) error {
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
