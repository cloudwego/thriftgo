// Copyright 2024 CloudWeGo Authors
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

package sdk

import (
	"fmt"

	"github.com/cloudwego/thriftgo/utils/dir_utils"

	"github.com/cloudwego/thriftgo/config"

	targs "github.com/cloudwego/thriftgo/args"
	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/version"
)

func init() {
	_ = g.RegisterBackend(new(golang.GoBackend))
}

var (
	g generator.Generator
)

func RunThriftgoAsSDK(wd string, plugins []plugin.SDKPlugin, args ...string) error {

	// this should execute at the first line!
	dir_utils.SetGlobalwd(wd)

	err := config.LoadConfig()
	if err != nil {
		return err
	}

	var a targs.Arguments

	err = a.Parse(append([]string{"thriftgo"}, args...))
	if err != nil {
		if err.Error() == "flag: help requested" {
			return nil
		}
		return err
	}

	if a.AskVersion {
		println("thriftgo", version.ThriftgoVersion)
		return nil
	}

	ast, err := parser.ParseFile(a.IDL, a.Includes, true)
	if err != nil {
		return err
	}

	if path := parser.CircleDetect(ast); len(path) > 0 {
		return fmt.Errorf("found include circle:\n\t%s", path)
	}

	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	if err != nil {
		return err
	}

	err = semantic.ResolveSymbols(ast)
	if err != nil {
		return err
	}

	req := &plugin.Request{
		Version:    version.ThriftgoVersion,
		OutputPath: a.OutputPath,
		Recursive:  a.Recursive,
		AST:        ast,
	}

	langs, err := a.Targets()
	if err != nil {
		return err
	}

	if len(langs) == 0 {
		return fmt.Errorf("No output language(s) specified")
	}

	log := backend.DummyLogFunc()
	for _, out := range langs {
		out.SDKPlugins = plugins
		req.Language = out.Language
		req.OutputPath = a.Output(out.Language)

		arg := &generator.Arguments{Out: out, Req: req, Log: log}
		res := g.Generate(arg)

		err = g.Persist(res)
		if err != nil {
			return err
		}
	}
	return nil
}
