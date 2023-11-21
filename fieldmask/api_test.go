/*
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fieldmask

import (
	"strings"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

var baseIDL = `
namespace go base

struct TrafficEnv {
	0: string Name = "",
	1: bool Open = false,
	2: string Env = "",
	256: i64 Code,
}

struct Base {
	0: string Addr = "",
	1: string LogID = "",
	2: string Caller = "",
	5: optional TrafficEnv TrafficEnv,
	6: optional list<ExtraInfo> Extra,
	256: MetaInfo Meta,
}

struct ExtraInfo {
	1: map<i32,Val> IntMap
	2: map<string,Val> StrMap
	3: list<Val> List
	4: set<Val> Set
}

struct Val {
	1: string A,
	2: string B,
}

struct MetaInfo {
	1: map<string, Base> F1,
	2: map<i8, Base> F2,
	3: list<Base> F3,
	3: Base Base,
}

struct BaseResp {
	1: string StatusMessage = "",
	2: i32 StatusCode = 0,
	3: optional map<string, string> Extra,
}`

func GetDescriptor(IDL string, root string) *thrift_reflection.TypeDescriptor {
	ast, err := parser.ParseString("a.thrift", IDL)
	if err != nil {
		panic(err.Error())
	}
	fd := thrift_reflection.RegisterAST(ast)
	st := fd.GetStructDescriptor(root)
	return &thrift_reflection.TypeDescriptor{
		Filepath: st.Filepath,
		Name:     st.Name,
	}
}

func TestNewFieldMask(t *testing.T) {
	type args struct {
		IDL        string
		rootStruct string
		paths      []string
		inMasks    []string
		notInMasks []string
		err        []error
	}
	tests := []struct {
		name string
		args args
		want *FieldMask
	}{
		{
			name: "Struct",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.LogID", "$.TrafficEnv.Open", "$.TrafficEnv.Env", "$.Meta"},

				inMasks:    []string{"$.Meta.F1", "$.Meta.F2", "$.Meta.Base.Caller"},
				notInMasks: []string{"$.TrafficEnv.Name", "$.TrafficEnv.Code", "$.Caller", "$.Addr", "$.Extra"},
			},
		},
		{
			name: "List/Set",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[0]", "$.Extra[1].List", "$.Extra[2].Set[0,1]", "$.Extra[4,5].List[*]"},

				inMasks:    []string{"$.Extra[0].List", "$.Extra[2].Set[0].A", "$.Extra[2].Set[1].A", "$.Extra[4].List[0]", "$.Extra[4,5].List[0]", "$.Extra[1,4,5].List"},
				notInMasks: []string{"$.Extra[1].Set", "$.Extra[1].IntMap", "$.Extra[3]", "$.Extra[3,4].Set"},
			},
		},
		{
			name: "Int Map",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[0].IntMap{0}", "$.Extra[0].IntMap{1}.A", "$.Extra[0].IntMap{1}.B", "$.Extra[0].IntMap{2}.A", "$.Extra[0].IntMap{4,5}.A", "$.Meta.F2{*}.TrafficEnv"},
				inMasks:    []string{"$.Extra[0].IntMap{0}.A", "$.Extra[0].IntMap{0}.B", "$.Extra[0].IntMap{4}.A", "$.Extra[0].IntMap{5}.A", "$.Meta.F2{0}.TrafficEnv.Env", "$.Meta.F2{*}.TrafficEnv.Env"},
				notInMasks: []string{"$.Extra[0].IntMap{2}.B", "$.Extra[0].IntMap{3}", "$.Extra[0].IntMap{4}.B", "$.Extra[0].IntMap{5}.B", "$.Meta.F2{0}.Addr", "$.Meta.F2{*}.Addr"},
			},
		},
		{
			name: "Union",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[0].List", "$.Extra[*].Set", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr"},
				inMasks:    []string{"$.Extra[*].Set[0]", "$.Meta.F2{1}.Addr"},
				notInMasks: []string{"$.Extra[0].List", "$.Meta.F2[0].LogID"},
			},
		},
		{
			name: "String Map",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[0].StrMap{\"x\"}", "$.Extra[0].StrMap{\"a\"}.A", "$.Extra[0].StrMap{\"b\"}.B", "$.Extra[0].StrMap{\"c\",\"d\"}", "$.Extra[0].StrMap{\"e\",\"f\"}.A"},
				inMasks:    []string{"$.Extra[0].StrMap{\"x\"}.A", "$.Extra[0].StrMap{\"x\"}.B", "$.Extra[0].StrMap{\"c\"}.A", "$.Extra[0].StrMap{\"c\",\"d\",\"e\",\"f\"}.A"},
				notInMasks: []string{"$.Extra[0].StrMap{\"a\"}.B", "$.Extra[0].StrMap{\"b\"}.A", "$.Extra[0].StrMap{\"s\"}", "$.Extra[0].StrMap{\"s\",\"c\"}", "$.Extra[0].StrMap{\"d\",\"e\"}.B"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// defer func() {
			// 	if v := recover(); v != nil {
			// 		if tt.args.err == nil || v != tt.args.err {
			// 			t.Fatal("panic: ", v)
			// 		}
			// 	}
			// }()

			st := GetDescriptor(tt.args.IDL, tt.args.rootStruct)
			got, err := NewFieldMask(st, tt.args.paths...)
			if tt.args.err != nil {
				if err == nil {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			retry := true
		begin:

			println("fieldmask:")
			println(got.String(st))
			// spew.Dump(got)

			if tt.name != "Union" {
				for _, path := range tt.args.paths {
					println("[paths] ", path)
					if !got.PathInMask(st, path) {
						t.Fatal(path)
					}
				}
			}

			for _, path := range tt.args.inMasks {
				println("[inMasks] ", path)
				if !got.PathInMask(st, path) {
					t.Fatal(path)
				}
			}
			for _, path := range tt.args.notInMasks {
				println("[notInMasks] ", path)
				if got.PathInMask(st, path) {
					t.Fatal(path)
				}
			}

			if retry {
				got.reset()
				if err := got.init(st, tt.args.paths...); err != nil {
					t.Fatal(err)
				}
				retry = false
				goto begin
			}
		})
	}
}

func TestErrors(t *testing.T) {
	type args struct {
		IDL        string
		rootStruct string
		path       []string
		err        string
	}
	tests := []struct {
		name string
		args args
		want *FieldMask
	}{
		{
			name: "desc struct",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID.X"},
				err:        `Descriptor "string" isn't STRUCT`,
			},
		},
		{
			name: "desc list",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID[1]"},
				err:        `Descriptor "string" isn't LIST or SET`,
			},
		},
		{
			name: "desc map",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID{1}"},
				err:        `Descriptor "string" isn't MAP`,
			},
		},
		{
			name: "desc map key",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.IntMap{\"a\"}"},
				err:        `expect integer but got string`,
			},
		},
		{
			name: "desc map key",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.StrMap{1}"},
				err:        `expect string but got integer`,
			},
		},
		{
			name: "syntax index",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.List[\"1\"]"},
				err:        `isn't literal`,
			},
		},
		{
			name: "fields conflict",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.TrafficEnv", "$.TrafficEnv.Env"},
				err:        `onflicts with previously-set all (*) fields`,
			},
		},
		{
			name: "index conflict",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.Extra[*]", "$.Extra[1]"},
				err:        `onflicts with previously-set all (*) index`,
			},
		},
		{
			name: "key conflict",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.IntMap{*}", "$.IntMap{1}"},
				err:        `onflicts with previously-set all (*) keys`,
			},
		},
		{
			name: "empty map set",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.IntMap{}"},
				err:        `empty key set`,
			},
		},
		{
			name: "empty list set",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.List[]"},
				err:        `empty index set`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := GetDescriptor(tt.args.IDL, tt.args.rootStruct)
			_, err := GetFieldMask(st, tt.args.path...)
			if err == nil || !strings.Contains(err.Error(), tt.args.err) {
				t.Fatal(err)
			}
		})
	}
}

func BenchmarkNewFieldMask(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	if st == nil {
		b.Fail()
	}
	b.Run("new", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fm, err := NewFieldMask(st, []string{"$.LogID", "$.TrafficEnv.Open", "$.TrafficEnv.Env", "$.Extra[0]", "$.Extra[1].IntMap{0}", "$.Extra[2].StrMap{\"abcd\"}"}...)
			if err != nil {
				b.Fatal(err)
			}
			_ = fm
		}
	})
	b.Run("reuse", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fm, err := GetFieldMask(st, []string{"$.LogID", "$.TrafficEnv.Open", "$.TrafficEnv.Env", "$.Extra[0]", "$.Extra[1].IntMap{0}", "$.Extra[2].StrMap{\"abcd\"}"}...)
			if err != nil {
				b.Fatal(err)
			}
			fm.Recycle()
		}
	})
}

func BenchmarkFieldMask_InMask(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	if st == nil {
		b.Fail()
	}
	fm, err := NewFieldMask(st, []string{"$.Extra[0]", "$.Extra[1].IntMap{0}", "$.Extra[2].StrMap{\"abcdefghi\"}"}...)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("Field", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if next, exist := fm.Field(6); !exist {
				b.Fail()
			} else {
				_ = next
			}
		}
	})

	b.Run("Index", func(b *testing.B) {
		var v *FieldMask
		if next, ex := fm.Field(6); !ex {
			b.Fail()
		} else {
			v = next
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if next, ex := v.Int(0); !ex {
				b.Fail()
			} else {
				_ = next
			}
		}
	})

	b.Run("Int Map", func(b *testing.B) {
		var v *FieldMask
		if next, ex := fm.Field(6); !ex {
			b.Fail()
		} else if l, ex := next.Int(1); !ex {
			b.Fail()
		} else if f, ex := l.Field(1); !ex {
			b.Fail()
		} else {
			v = f
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if next, ex := v.Int(0); !ex {
				b.Fail()
			} else {
				_ = next
			}
		}
	})

	b.Run("Str Map", func(b *testing.B) {
		var v *FieldMask
		if next, ex := fm.Field(6); !ex {
			b.Fail()
		} else if l, ex := next.Int(2); !ex {
			b.Fail()
		} else if f, ex := l.Field(2); !ex {
			b.Fail()
		} else {
			v = f
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if next, ex := v.Str("abcdefghi"); !ex {
				b.Fail()
			} else {
				_ = next
			}
		}
	})
}
