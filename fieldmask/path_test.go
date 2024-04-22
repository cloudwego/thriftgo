/**
 * Copyright 2024 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
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
	"reflect"
	"testing"
)

func TestGetPath(t *testing.T) {
	v1 := &FieldMaskTransfer{jsonPathRoot, FtScalar, false, nil}
	fm1, err := v1.TransferTo()
	if err != nil {
		t.Fatal(err)
	}
	fm2, err := NewFieldMask(GetDescriptor(baseIDL, "TrafficEnv"), "$.Open")
	if err != nil {
		t.Fatal(err)
	}
	v3 := &FieldMaskTransfer{jsonPathRoot, FtList, false, []FieldMaskTransfer{
		{json.RawMessage("1"), FtStruct, false, []FieldMaskTransfer{
			{json.RawMessage("1"), FtScalar, false, nil},
		}},
	}}
	fm3, err := v3.TransferTo()
	if err != nil {
		t.Fatal(err)
	}
	fm4, err := NewFieldMask(GetDescriptor(baseIDL, "ExtraInfo"), "$.List", "$.Set[1].A")
	if err != nil {
		t.Fatal(err)
	}
	fm5, err := NewFieldMask(GetDescriptor(baseIDL, "ExtraInfo"), "$.IntMap{1}", "$.IntMap{3}.A", "$.StrMap{\"x\"}", "$.StrMap{\"y\"}.A")
	if err != nil {
		t.Fatal(err)
	}
	v6 := &FieldMaskTransfer{jsonPathRoot, FtIntMap, false, []FieldMaskTransfer{
		{json.RawMessage("1"), FtStruct, false, nil},
		{json.RawMessage("3"), FtStruct, false, []FieldMaskTransfer{
			{json.RawMessage("1"), FtScalar, false, nil},
		}},
	}}
	fm6, err := v6.TransferTo()
	if err != nil {
		t.Fatal(err)
	}
	v7 := &FieldMaskTransfer{jsonPathRoot, FtStrMap, false, []FieldMaskTransfer{
		{json.RawMessage(`"x"`), FtStruct, false, nil},
		{json.RawMessage(`"y"`), FtStruct, false, []FieldMaskTransfer{
			{json.RawMessage("1"), FtScalar, false, nil},
		}},
	}}
	fm7, err := v7.TransferTo()
	if err != nil {
		t.Fatal(err)
	}
	v8 := &FieldMaskTransfer{jsonPathRoot, FtList, false, nil}
	fm8, err := v8.TransferTo()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		opts       Options
		IDL        string
		rootStruct string
		paths      []string
		err        []error
	}
	type res struct {
		path string
		fm   *FieldMask
		ex   bool
	}
	tests := []struct {
		name string
		args args
		res  []res
	}{
		{
			name: "Base",
			args: args{
				IDL:        baseIDL,
				rootStruct: "Base",
				paths: []string{
					"$.LogID",
					"$.TrafficEnv.Open",
					"$.Extra[0]",
					"$.Extra[1].List",
					"$.Extra[1].Set[1].A",
					"$.Extra[3].IntMap{1}",
					"$.Extra[3].IntMap{3}.A",
					"$.Extra[3].StrMap{\"x\"}",
					"$.Extra[3].StrMap{\"y\"}.A",
				},
			},
			res: []res{
				{
					path: "$.LogID",
					fm:   fm1,
					ex:   true,
				},
				{
					path: "$.Addr",
					fm:   (*FieldMask)(nil),
					ex:   false,
				},
				{
					path: "$.TrafficEnv",
					fm:   fm2,
					ex:   true,
				},
				{
					path: "$.Extra[1].List",
					fm:   fm8,
					ex:   true,
				},
				{
					path: "$.Extra[1].List.A",
					fm:   (*FieldMask)(nil),
					ex:   false,
				},
				{
					path: "$.Extra[1].Set",
					fm:   fm3,
					ex:   true,
				},
				{
					path: "$.Extra[1]",
					fm:   fm4,
					ex:   true,
				},
				{
					path: "$.Extra[3]",
					fm:   fm5,
					ex:   true,
				},
				{
					path: "$.Extra[3].IntMap",
					fm:   fm6,
					ex:   true,
				},
				{
					path: "$.Extra[3].StrMap",
					fm:   fm7,
					ex:   true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := GetDescriptor(tt.args.IDL, tt.args.rootStruct)
			root, err := tt.args.opts.NewFieldMask(st, tt.args.paths...)
			if tt.args.err != nil {
				if err == nil {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, re := range tt.res {
				println("re.Path:", re.path)
				got, ex := root.GetPath(st, re.path)
				if ex != re.ex {
					t.Fatal(ex)
				}
				gj, err := got.MarshalJSON()
				if err != nil {
					t.Fatal(err)
				}
				println(string(gj))
				var act *FieldMask
				if err := json.Unmarshal(gj, &act); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(re.fm, act) {
					t.Fatalf("exp:%#v,\ngot:%#v", re.fm, *act)
				}
			}
		})
	}
}
