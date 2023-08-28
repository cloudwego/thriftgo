/**
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
	255: optional ExtraInfo Extra,
	256: MetaInfo Meta,
}

struct ExtraInfo {
	1: map<string, string> KVS
}

struct MetaInfo {
	1: map<string, string> PersistentKVS,
	2: map<string, string> TransientKVS,
	3: Base base,
}

struct BaseResp {
	1: string StatusMessage = "",
	2: i32 StatusCode = 0,
	3: optional map<string, string> Extra,
}`
)

func TestNewFieldMaskFromAST(t *testing.T) {
	type args struct {
		IDL        string
		rootStruct string
		paths      []string
		inMasks    []string
		notInMasks []string
	}
	tests := []struct {
		name string
		args args
		want *FieldMask
	}{
		{
			name: "base",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths:      []string{"LogID", "TrafficEnv.Open", "TrafficEnv.Env", "Meta"},
				inMasks:    []string{"Meta.PersistentKVS", "Meta.TransientKVS"},
				notInMasks: []string{"TrafficEnv.Name", "TrafficEnv.Code", "Caller", "Addr", "Extra", "Extra.KVS"},
			},
			want: &FieldMask{
				flat: fieldMaskBitmap([]byte{0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
				next: fieldMaskMap{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			IDL, err := parser.ParseString("a.thrift", tt.args.IDL)
			if err != nil {
				t.Fatal(err)
			}
			fd := thrift_reflection.RegisterAST(IDL)
			st := fd.GetStructDescriptor(tt.args.rootStruct)
			got := NewFieldMaskFromNames(st, tt.args.paths...)

			if !reflect.DeepEqual(got.flat, tt.want.flat) {
				t.Fatal("not expected flat, ", tt.want.flat, got.flat)
			}
			println(got.String(st))
			for _, path := range tt.args.paths {
				if !got.PathInMask(st, path) {
					t.Fatal(path)
				}
			}
			for _, path := range tt.args.inMasks {
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
