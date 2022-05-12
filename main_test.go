// Copyright 2022 CloudWeGo Authors
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
	"testing"

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
)

// Tests in this file all depend on an environment variable `IDL`
// which will be written by gen_test.go for parsing.
// Export the environment variable before testing.
var idl string

func TestParse(t *testing.T) {
	if idl == "" {
		t.SkipNow()
	}
	_, err := parser.ParseFile(idl, nil, true)
	if err != nil {
		t.Fatalf("parse falied: %s", err.Error())
	}
}

func TestGenerate(t *testing.T) {
	if idl == "" {
		t.SkipNow()
	}
	g, a := pair(t)
	res := g.Generate(a)
	if res.Error != nil {
		t.Fatalf("generate falied: %s", *res.Error)
	}
}

func pair(tb testing.TB) (generator.Generator, *generator.Arguments) {
	ast, err := parser.ParseFile(idl, nil, true)
	if err != nil {
		tb.Fatalf("parse falied: %s", err.Error())
	}

	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	resolver, ok := checker.(interface {
		ResolveSymbols(t *parser.Thrift) error
	})
	if ok {
		if err = resolver.ResolveSymbols(ast); err != nil {
			tb.Fatalf("resolve: %s", err.Error())
		}
	}

	var gen generator.Generator
	gen.RegisterBackend(new(golang.GoBackend))
	log := backend.LogFunc{
		Info:      func(v ...interface{}) {},
		Warn:      func(v ...interface{}) {},
		MultiWarn: func(warns []string) {},
	}
	out := &generator.LangSpec{
		Language: "go",
		Options:  nil,
	}
	req := &plugin.Request{
		Language:   out.Language,
		Version:    "?",
		OutputPath: "./gen-go",
		Recursive:  true,
		AST:        ast,
	}
	arg := &generator.Arguments{Out: out, Req: req, Log: log}
	return gen, arg
}

func full(tb testing.TB) {
	if idl == "" {
		tb.SkipNow()
	}
	g, a := pair(tb)
	res := g.Generate(a)
	if res.Error != nil {
		tb.Fatalf("generate: %s", *res.Error)
	}
}

func BenchmarkParse(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(idl, nil, true)
		if err != nil {
			b.Fatalf("parse falied: %s", err.Error())
		}
	}
}

func BenchmarkParseParallel(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.ParseFile(idl, nil, true)
			if err != nil {
				b.Fatalf("parse falied: %s", err.Error())
			}
		}
	})
}

func BenchmarkGenerate(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	g, a := pair(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := g.Generate(a)
		if res.Error != nil {
			b.Fatalf("generate falied: %s", *res.Error)
		}
	}
}

func BenchmarkAll(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	for i := 0; i < b.N; i++ {
		full(b)
	}
}

func BenchmarkSemanticCheck(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	ast, err := parser.ParseFile(idl, nil, true)
	if err != nil {
		b.Fatalf("parse falied: %s", err.Error())
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckAll(ast)
		if err != nil {
			b.Fatalf("generate falied: %s", err.Error())
		}
	}
}

func BenchmarkSemanticResolve(b *testing.B) {
	if idl == "" {
		b.SkipNow()
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	resolver, ok := checker.(interface {
		ResolveSymbols(t *parser.Thrift) error
	})
	if !ok {
		b.Skip()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ast, err := parser.ParseFile(idl, nil, true)
		if err != nil {
			b.Fatalf("parse falied: %s", err.Error())
		}
		if err = resolver.ResolveSymbols(ast); err != nil {
			b.Fatalf("resolve: %s", err.Error())
		}
	}
}
