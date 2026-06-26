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
	"testing"
	"time"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
)

func TestRecurseDump_DedupSharedDAG(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()

	const k = 10 // 22 files; 2^10 = 1024 distinct paths reach the leaf.

	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name+".thrift"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("leaf", "struct Leaf { 1: string s }\n")
	for i := 1; i <= k-1; i++ {
		body := fmt.Sprintf("include \"L%dA.thrift\"\ninclude \"L%dB.thrift\"\n", i+1, i+1)
		write("L"+strconv.Itoa(i)+"A", body)
		write("L"+strconv.Itoa(i)+"B", body)
	}
	write("L"+strconv.Itoa(k)+"A", "include \"leaf.thrift\"\n")
	write("L"+strconv.Itoa(k)+"B", "include \"leaf.thrift\"\n")
	write("root", `include "L1A.thrift"
include "L1B.thrift"
struct Root { 1: string s }
`)

	ast, err := parser.ParseFile(filepath.Join(src, "root.thrift"), nil, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	visited := make(map[*parser.Thrift]struct{})
	start := time.Now()
	recurseDump(ast, src, out, visited)
	elapsed := time.Since(start)

	wantFiles := 2*k + 2 // root + leaf + 2 per mid layer
	if got := len(visited); got != wantFiles {
		t.Errorf("visited entries = %d, want %d (root + leaf + 2*k mid files)", got, wantFiles)
	}

	for _, name := range []string{"root", "leaf", "L1A", "L" + strconv.Itoa(k) + "B"} {
		if _, err := os.Stat(filepath.Join(out, name+".thrift")); err != nil {
			t.Errorf("missing output %s.thrift: %v", name, err)
		}
	}

	if elapsed > 1000*time.Millisecond {
		t.Errorf("recurseDump took %v on k=%d case", elapsed, k)
	}
}

func recurseDumpNoDedup(t *testing.T, ast *parser.Thrift, sourceDir, outDir string) {
	t.Helper()
	if ast == nil {
		return
	}
	out, err := dump.DumpIDL(ast)
	if err != nil {
		t.Fatal(err)
	}
	absFileName, err := filepath.Abs(ast.Filename)
	if err != nil {
		t.Fatal(err)
	}
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(absSourceDir, absFileName)
	if err != nil {
		t.Fatal(err)
	}
	outputFileUrl := filepath.Join(outDir, rel)
	if err := os.MkdirAll(filepath.Dir(outputFileUrl), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := writeStringToFile(outputFileUrl, out); err != nil {
		t.Fatal(err)
	}
	for _, inc := range ast.Includes {
		recurseDumpNoDedup(t, inc.Reference, sourceDir, outDir)
	}
}

func readDirFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestRecurseDump_OutputUnchangedByDedup(t *testing.T) {
	src := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name+".thrift"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("leaf", `struct Deep {
  1: i64 v
}
`)
	write("base", `include "leaf.thrift"

const i64 MAGIC = 42
typedef string UserId

enum Color {
  RED = 0
  GREEN = 1
}

struct Inner {
  1: UserId id
  2: Color c
  3: leaf.Deep d
}

union Choice {
  1: i64 a
  2: string b
}

exception Err {
  1: string msg
}
`)
	write("a", `include "base.thrift"

struct A {
  1: base.Inner inner
}
`)
	write("b", `include "base.thrift"

struct B {
  1: base.Choice choice
}
`)
	write("root", `include "base.thrift"
include "a.thrift"
include "b.thrift"

service Svc {
  base.Inner ping(1: base.UserId id) throws (1: base.Err e)
}
`)

	rootPath := filepath.Join(src, "root.thrift")

	// dedup version (production recurseDump)
	ast1, err := parser.ParseFile(rootPath, nil, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	outDedup := t.TempDir()
	recurseDump(ast1, src, outDedup, make(map[*parser.Thrift]struct{}))

	// no-dedup version
	ast2, err := parser.ParseFile(rootPath, nil, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	outNoDedup := t.TempDir()
	recurseDumpNoDedup(t, ast2, src, outNoDedup)

	got := readDirFiles(t, outDedup)
	want := readDirFiles(t, outNoDedup)

	if len(got) != len(want) {
		t.Fatalf("file count diff: dedup=%d no-dedup=%d", len(got), len(want))
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 output files (root/base/a/b/leaf), got %d", len(got))
	}
	for name, wantContent := range want {
		gotContent, ok := got[name]
		if !ok {
			t.Errorf("file %q produced by no-dedup but missing in dedup", name)
			continue
		}
		if gotContent != wantContent {
			t.Errorf("file %q differs between dedup and no-dedup:\n--- dedup ---\n%s\n--- no-dedup ---\n%s",
				name, gotContent, wantContent)
		}
	}
}
