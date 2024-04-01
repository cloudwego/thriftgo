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
	"encoding/json"
	"runtime"
	"strconv"
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

typedef Val Key 

struct BaseResp {
	1: required string StatusMessage = "",
	2: required i32 StatusCode = 0,
	3: required bool R3,
	4: required byte R4,
	5: required i16 R5,
	6: required i64 R6,
	7: required double R7,
	8: required string R8,
	9: required Ex R9,
	10: required list<Val> R10,
	11: required set<Val> R11,
	12: required TrafficEnv R12,
	13: required map<string, Key> R13,
	0: required Key R0,

	14: map<Str, Str> F1
	15: map<Int, string> F2,
	16: list<string> F3
	17: set<string> F4,
	18: map<Float, Val> F5
	19: map<double, string> F6
	110: map<Ex, string> F7
	111: map<double, list<Str>> F8
	112: list<map<Float, list<Str>>> F9
	113: map<Key, Val> F10
}
`

func GetDescriptor(IDL string, root string) *thrift_reflection.TypeDescriptor {
	ast, err := parser.ParseString("a.thrift", IDL)
	if err != nil {
		panic(err.Error())
	}
	_, fd := thrift_reflection.RegisterAST(ast)
	st := fd.GetStructDescriptor(root)
	return &thrift_reflection.TypeDescriptor{
		Filepath: st.Filepath,
		Name:     st.Name,
		Extra:    map[string]string{thrift_reflection.GLOBAL_UUID_EXTRA_KEY: st.Extra[thrift_reflection.GLOBAL_UUID_EXTRA_KEY]},
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
			name: "Neither-string-nor-integer-key Map",
			args: args{
				IDL:        baseIDL,
				rootStruct: "BaseResp",
				paths:      []string{"$.F10{*}.A", "$.F5{*}.A"},
				inMasks:    []string{"$.F10{\"a\"}.A", "$.F5{0}.A"},
				notInMasks: []string{`$.F10{"a"}.B`, "$.F10{*}.B", "$.F5{0}.B", "$.F5{*}.B"},
			},
		},
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
			name: "Repeated *",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[*].List", "$.Extra[*].Set", "$.Meta.F2{*}.Caller", "$.Meta.F2{*}.Addr"},
				inMasks:    []string{"$.Extra[*].Set[0]", "$.Meta.F2{1}.Addr"},
				notInMasks: []string{"$.Meta.F2[0].LogID"},
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

			// println("fieldmask:")
			// println(got.String(st))
			// spew.Dump(got)

			// test marshal json
			// println("marshal:")
			out, err := got.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			// println(string(out))
			if !json.Valid(out) {
				t.Fatal("not invalid json")
			}

			// test unmarshal json
			nn := &FieldMask{}
			if err := nn.UnmarshalJSON(out); err != nil {
				t.Fatal(err)
			}

			if tt.name != "Union" {
				for _, path := range tt.args.paths {
					println("[paths] ", path)
					if !got.PathInMask(st, path) {
						t.Fatal(path)
					}
					if !nn.PathInMask(st, path) {
						t.Fatal(path)
					}
				}
			}

			for _, path := range tt.args.inMasks {
				println("[inMasks] ", path)
				if !got.PathInMask(st, path) {
					t.Fatal(path)
				}
				if !nn.PathInMask(st, path) {
					t.Fatal(path)
				}
			}
			for _, path := range tt.args.notInMasks {
				println("[notInMasks] ", path)
				if got.PathInMask(st, path) {
					t.Fatal(path)
				}
				if nn.PathInMask(st, path) {
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

func TestMarshalJSONStable(t *testing.T) {
	st := GetDescriptor(baseIDL, "MetaInfo")
	fm, err := NewFieldMask(st, "$.F2{4,1,3}", "$.F2{0,2}", `$.F1{"c","d","b"}`, `$.F1{"a"}`)
	if err != nil {
		t.Fatal(err)
	}
	jo, err := fm.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	println(string(jo))
	if string(jo) != (`{"path":"$","type":"Struct","children":[{"path":1,"type":"StrMap","children":[{"path":"a","type":"Struct"},{"path":"b","type":"Struct"},{"path":"c","type":"Struct"},{"path":"d","type":"Struct"}]},{"path":2,"type":"IntMap","children":[{"path":0,"type":"Struct"},{"path":1,"type":"Struct"},{"path":2,"type":"Struct"},{"path":3,"type":"Struct"},{"path":4,"type":"Struct"}]}]}`) {
		t.Fatal(string(jo))
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
			name: "desc expect struct",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID.X"},
				err:        `Descriptor "string" isn't STRUCT`,
			},
		},
		{
			name: "desc expect list",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID[1]"},
				err:        `Descriptor "string" isn't LIST or SET`,
			},
		},
		{
			name: "desc expect map",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.LogID{1}"},
				err:        `Descriptor "string" isn't MAP`,
			},
		},
		{
			name: "desc expect map int key",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.IntMap{\"a\"}"},
				err:        `expect integer but got string`,
			},
		},
		{
			name: "desc expect map string key",
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
				err:        `field conflicts with previously settled '*'`,
			},
		},
		{
			name: "index conflict",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				path:       []string{"$.Extra[*]", "$.Extra[1]"},
				err:        `id conflicts with previously settled '*'`,
			},
		},
		{
			name: "key conflict",
			args: args{
				IDL:        baseIDL,
				rootStruct: "ExtraInfo",
				path:       []string{"$.IntMap{*}", "$.IntMap{1}"},
				err:        `key conflicts with previous settled '*'`,
			},
		},
		{
			name: "key conflict2",
			args: args{
				IDL:        baseIDL,
				rootStruct: "BaseResp",
				path:       []string{"$.F5{*}", "$.F5{1}"},
				err:        `key conflicts with previous settled '*'`,
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
			_, err := NewFieldMask(st, tt.args.path...)
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
	// b.Run("reuse", func(b *testing.B) {
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		fm, err := GetFieldMask(st, []string{"$.LogID", "$.TrafficEnv.Open", "$.TrafficEnv.Env", "$.Extra[0]", "$.Extra[1].IntMap{0}", "$.Extra[2].StrMap{\"abcd\"}"}...)
	// 		if err != nil {
	// 			b.Fatal(err)
	// 		}
	// 		fm.Recycle()
	// 	}
	// })
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

func BenchmarkMarshal(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	got, err := NewFieldMask(st, "$.Extra[0].List", "$.Extra[*].Set", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr")
	if err != nil {
		b.Fatal(err)
	}
	j, err := got.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	if !json.Valid(j) {
		b.Fatal("invalid json:", string(j))
	}
	j2, e2 := Marshal(got)
	if e2 != nil {
		b.Fatal(e2)
	}
	if !json.Valid(j2) {
		b.Fatal("invalid json2", string(j2))
	}

	b.Run("MarshalJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = got.MarshalJSON()
		}
	})

	b.Run("Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Marshal(got)
		}
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	got, err := NewFieldMask(st, "$.Extra[0].List", "$.Extra[*].Set", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr")
	if err != nil {
		b.Fatal(err)
	}
	j, err := got.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	if !json.Valid(j) {
		b.Fatal("invalid json:", string(j))
	}
	act := new(FieldMask)
	if err := act.UnmarshalJSON(j); err != nil {
		b.Fatal(err)
	}
	// if !reflect.DeepEqual(got, act) {
	// 	b.Fatal()
	// }

	_, err = Unmarshal(j)
	if err != nil {
		b.Fatal(err)
	}
	// if !reflect.DeepEqual(got, act2) {
	// 	b.Fatal()
	// }

	b.Run("UnmarshalJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			act := new(FieldMask)
			_ = act.UnmarshalJSON(j)
		}
	})

	b.Run("Umarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Unmarshal(j)
		}
	})
}

func BenchmarkMemory(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	_, err := NewFieldMask(st, []string{"$.Extra[0].List", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr"}...)
	if err != nil {
		b.Fatal(err)
	}

	go func() {
		for {
			runtime.GC()
		}
	}()

	tester := func(X int, b *testing.B) {
		for i := 0; i < b.N; i++ {
			for x := 0; x < X; x++ {
				tt, err := NewFieldMask(st, "$.Extra["+strconv.Itoa(x)+"].List", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr")
				if err != nil {
					b.Fatal(err)
				}
				j, err := Marshal(tt)
				if err != nil {
					b.Fatal(err)
				}
				_, err = Unmarshal(j)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}

	b.Run("10", func(b *testing.B) {
		tester(10, b)
	})

	b.Run("100", func(b *testing.B) {
		tester(100, b)
	})

	b.Run("1000", func(b *testing.B) {
		tester(1000, b)
	})

	b.Run("10000", func(b *testing.B) {
		tester(10000, b)
	})
}
