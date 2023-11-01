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
	"errors"
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

var (
	baseIDL = `
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
	255: optional list<ExtraInfo> Extra,
	256: MetaInfo Meta,
}

struct ExtraInfo {
	1: map<i32,Val> Map
	2: list<Val> List
	3: set<Val> Set
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
)

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
		err        error
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
			want: &FieldMask{
				fieldMask: (*fieldMaskBitmap)(&[]byte{0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
			},
		},
		{
			name: "List/Set",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.Extra[0]", "$.Extra[1].List", "$.Extra[2].Set[0]"},

				inMasks:    []string{"$.Extra[0].Map", "$.Extra[1].List[0]", "$.Extra[2].Set[0].A"},
				notInMasks: []string{"$.Extra[3]", "$.Extra[1].Map", "$.Extra[1].Set", "$.Extra[2].List", "$.Extra[2].Map", "$.Extra[2].Set[1]"},
			},
		},
		{
			name: "not struct err",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"$.LogID.X"},
				err:        errors.New(`Descriptor "string" isn't STRUCT`),
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
			got, err := GetFieldMask(st, tt.args.paths...)
			if tt.args.err != nil {
				if err == nil {
					t.Fatal(err)
				}
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got.fieldMask, tt.want.fieldMask) {
				t.Fatal("not expected flat, ", tt.want.fieldMask, got.fieldMask)
			}

			println("fieldmask:")
			println(got.String(st))
			// spew.Dump(got)

			// for _, path := range tt.args.paths {
			// 	if !got.PathInMask(st, path) {
			// 		t.Fatal(path)
			// 	}
			// }
			for _, path := range tt.args.inMasks {
				println("path: ", path)
				if !got.PathInMask(st, path) {
					t.Fatal(path)
				}
			}
			for _, path := range tt.args.notInMasks {
				if got.PathInMask(st, path) {
					t.Fatal(path)
				}
			}
		})
	}
}

func BenchmarkNewFieldMask(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	if st == nil {
		b.Fail()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm, _ := GetFieldMask(st, []string{"LogID", "TrafficEnv.Open", "TrafficEnv.Env", "Meta"}...)
		fm.Recycle()
	}
}

func BenchmarkFieldMask_InMask(b *testing.B) {
	st := GetDescriptor(baseIDL, "Base")
	if st == nil {
		b.Fail()
	}
	fm, _ := GetFieldMask(st, []string{"LogID", "TrafficEnv.Open", "TrafficEnv.Env", "Meta"}...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !fm.FieldInMask(5) {
			b.Fail()
		}
		next := fm.Field(5)
		if !next.FieldInMask(1) {
			b.Fail()
		}
		if next.FieldInMask(256) {
			b.Fail()
		}
		if next.FieldInMask(1024) {
			b.Fail()
		}
	}
}
