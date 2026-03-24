// Copyright 2025 CloudWeGo Authors
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

package options_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
)

var testIDL string

func init() {
	_, f, _, _ := runtime.Caller(0)
	testIDL = filepath.Join(filepath.Dir(f), "test.thrift")
}

func parseIDL(tb testing.TB) *parser.Thrift {
	tb.Helper()
	ast, err := parser.ParseFile(testIDL, nil, true)
	if err != nil {
		tb.Fatal(err)
	}
	return ast
}

func resolveAST(tb testing.TB, ast *parser.Thrift) {
	tb.Helper()
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	if _, err := checker.CheckAll(ast); err != nil {
		tb.Fatal(err)
	}
	if err := semantic.ResolveSymbols(ast); err != nil {
		tb.Fatal(err)
	}
}

func makeGenerator(tb testing.TB, ast *parser.Thrift) (generator.Generator, *generator.Arguments) {
	tb.Helper()
	var gen generator.Generator
	gen.RegisterBackend(new(golang.GoBackend))
	log := backend.LogFunc{
		Info:      func(v ...any) {},
		Warn:      func(v ...any) {},
		MultiWarn: func(warns []string) {},
	}
	out := &generator.LangSpec{Language: "go"}
	req := &plugin.Request{
		Language:   "go",
		Version:    "test",
		OutputPath: tb.TempDir(),
		Recursive:  false,
		AST:        ast,
	}
	return gen, &generator.Arguments{Out: out, Req: req, Log: log}
}

func BenchmarkDefaultParse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parseIDL(b)
	}
}

func BenchmarkDefaultSemanticCheck(b *testing.B) {
	ast := parseIDL(b)
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckAll(ast)
	}
}

func BenchmarkDefaultGenerate(b *testing.B) {
	ast := parseIDL(b)
	resolveAST(b, ast)
	gen, arg := makeGenerator(b, ast)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := gen.Generate(arg)
		if res.Error != nil {
			b.Fatal(*res.Error)
		}
	}
}

func BenchmarkDefaultAll(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ast := parseIDL(b)
		resolveAST(b, ast)
		gen, arg := makeGenerator(b, ast)
		res := gen.Generate(arg)
		if res.Error != nil {
			b.Fatal(*res.Error)
		}
	}
}
