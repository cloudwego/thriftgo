// Copyright 2021 CloudWeGo Authors
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

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
)

// Version of thriftgo.
const Version = "0.1.7"

var (
	a Arguments
	g generator.Generator
)

func init() {
	_ = g.RegisterBackend(new(golang.GoBackend))
}

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}

func main() {
	check(a.Parse(os.Args))
	if a.AskVersion {
		println("thriftgo", Version)
		os.Exit(0)
	}

	log := a.MakeLogFunc()

	ast, err := parser.ParseFile(a.IDL, a.Includes, true)
	check(err)

	if path := parser.CircleDetect(ast); len(path) > 0 {
		check(fmt.Errorf("found include circle:\n\t%s", path))
	}

	if a.CheckKeyword {
		if warns := parser.DetectKeyword(ast); len(warns) > 0 {
			log.MultiWarn(warns)
		}
	}

	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	warns, err := checker.CheckAll(ast)
	log.MultiWarn(warns)
	check(err)

	check(semantic.ResolveSymbols(ast))

	req := &plugin.Request{
		Version:    Version,
		OutputPath: a.OutputPath,
		Recursive:  a.Recursive,
		AST:        ast,
	}

	plugins, err := a.UsedPlugins()
	check(err)

	langs, err := a.Targets()
	check(err)

	if len(langs) == 0 {
		println("No output language(s) specified")
		os.Exit(2)
	}

	for _, out := range langs {
		out.UsedPlugins = plugins
		req.Language = out.Language
		req.OutputPath = a.Output(out.Language)

		arg := &generator.Arguments{Out: out, Req: req, Log: log}
		res := g.Generate(arg)
		log.MultiWarn(res.Warnings)

		err = g.Persist(res)
		check(err)
	}
}
