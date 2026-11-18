// Copyright 2026 CloudWeGo Authors
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
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
)

func oracleDump(a *parser.Thrift, out map[string]string, calls *int) error {
	(*calls)++
	s, e := dump.DumpIDL(a)
	if e != nil {
		return e
	}
	out[a.Filename] = s
	for _, i := range a.Includes {
		if e = oracleDump(i.Reference, out, calls); e != nil {
			return e
		}
	}
	return nil
}
func fixtureDiamond(depth int) *parser.Thrift {
	lower := []*parser.Thrift{{Filename: "leaf.thrift"}}
	for level := 0; level < depth; level++ {
		ns := []*parser.Thrift{}
		for _, s := range []string{"a", "b"} {
			n := &parser.Thrift{Filename: string(rune('A'+level)) + s + ".thrift"}
			for _, l := range lower {
				n.Includes = append(n.Includes, &parser.Include{Path: l.Filename, Reference: l})
			}
			ns = append(ns, n)
		}
		lower = ns
	}
	root := &parser.Thrift{Filename: "root.thrift"}
	for _, l := range lower {
		root.Includes = append(root.Includes, &parser.Include{Path: l.Filename, Reference: l})
	}
	return root
}
func TestRecursiveDumpDedupPreservesOracle(t *testing.T) {
	for _, d := range []int{0, 1, 4, 14} {
		a := fixtureDiamond(d)
		want, got := map[string]string{}, map[string]string{}
		calls := 0
		if e := oracleDump(a, want, &calls); e != nil {
			t.Fatal(e)
		}
		if e := recursiveDump(a, got); e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("depth %d changed output", d)
		}
		t.Logf("depth=%d repeated_dumps=%d unique_files=%d", d, calls, len(got))
	}
}
func TestRecursiveDumpDedupEmptyFileAndInvocationIsolation(t *testing.T) {
	empty := &parser.Thrift{Filename: "empty.thrift"}
	a := &parser.Thrift{Filename: "root.thrift", Includes: []*parser.Include{{Path: "empty.thrift", Reference: empty}, {Path: "empty.thrift", Reference: empty}}}
	one, two := map[string]string{}, map[string]string{}
	if e := recursiveDump(a, one); e != nil {
		t.Fatal(e)
	}
	if e := recursiveDump(a, two); e != nil {
		t.Fatal(e)
	}
	if len(one) != 2 || one["empty.thrift"] != "" || !reflect.DeepEqual(one, two) {
		t.Fatal("empty result or independent invocation changed")
	}
}

func TestRecursiveDumpDedupParallelCalls(t *testing.T) {
	for i := 0; i < 8; i++ {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			t.Parallel()
			a := fixtureDiamond(8)
			want, got := map[string]string{}, map[string]string{}
			calls := 0
			if e := oracleDump(a, want, &calls); e != nil {
				t.Fatal(e)
			}
			if e := recursiveDump(a, got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatal("parallel call changed output")
			}
		})
	}
}

func BenchmarkRecursiveDumpSharedDAG(b *testing.B) {
	ast := fixtureDiamond(14)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := recursiveDump(ast, map[string]string{}); err != nil {
			b.Fatal(err)
		}
	}
}

// With a real completed result, revisiting a shared node should do no more
// serialization allocations. This avoids a wall-time threshold or a dump hook.
func TestRecursiveDumpCompletedASTDoesNoSerializationWork(t *testing.T) {
	a, err := parser.ParseBatchString("root.thrift", map[string]string{
		"root.thrift": "include \"leaf.thrift\"\nnamespace go root\n",
		"leaf.thrift": "namespace go leaf\n",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	if err := recursiveDump(a, out); err != nil {
		t.Fatal(err)
	}
	allocs := testing.AllocsPerRun(10, func() {
		if err := recursiveDump(a, out); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("completed file was serialized again: %.0f allocs", allocs)
	}
}
