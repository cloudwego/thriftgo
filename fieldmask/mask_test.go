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
	"testing"
)

func TestFieldMask_ForEachChild(t *testing.T) {
	st := GetDescriptor(baseIDL, "Base")
	fmBase, err := NewFieldMask(st,
		"$.LogID",
		"$.TrafficEnv.Open",
		"$.Extra[0]",
		"$.Extra[1].List",
		"$.Extra[1].Set[1].A",
		"$.Extra[3].IntMap{1}",
		"$.Extra[3].IntMap{3}.A",
		"$.Extra[3].StrMap{\"x\"}",
		"$.Extra[3].StrMap{\"y\"}.A",
	)
	if err != nil {
		t.Fatal(err)
	}
	getter := func(path string) *FieldMask {
		fm, err := fmBase.GetPath(st, path)
		if !err {
			panic(path)
		}
		return fm
	}

	type KV struct {
		SK string
		IK int
		V  *FieldMask
	}
	var kvs []KV
	scanner := func(strKey string, intKey int, child *FieldMask) bool {
		kvs = append(kvs, KV{strKey, intKey, child})
		return true
	}

	tests := []struct {
		name string
		fm   *FieldMask
		exp  interface{}
	}{
		{
			name: "Base",
			fm:   fmBase,
			exp: map[int]*FieldMask{
				1: getter("$.LogID"),
				5: getter("$.TrafficEnv"),
				6: getter("$.Extra"),
			},
		},
		{
			name: "TrafficEnv",
			fm:   getter("$.TrafficEnv.Open"),
			exp:  nil,
		},
		{
			name: "Extra",
			fm:   getter("$.Extra"),
			exp: map[int]*FieldMask{
				0: getter("$.Extra[0]"),
				1: getter("$.Extra[1]"),
				3: getter("$.Extra[3]"),
			},
		},
		{
			name: "Set",
			fm:   getter("$.Extra[1].Set"),
			exp: map[int]*FieldMask{
				1: getter("$.Extra[1].Set[1]"),
			},
		},
		{
			name: "IntMap",
			fm:   getter("$.Extra[3].IntMap"),
			exp: map[int]*FieldMask{
				1: getter("$.Extra[3].IntMap{1}"),
				3: getter("$.Extra[3].IntMap{3}"),
			},
		},
		{
			name: "StrMap",
			fm:   getter("$.Extra[3].StrMap"),
			exp: map[string]*FieldMask{
				"x": getter("$.Extra[3].StrMap{\"x\"}"),
				"y": getter("$.Extra[3].StrMap{\"y\"}"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvs = kvs[:0]
			tt.fm.ForEachChild(scanner)
			switch tt.fm.Type() {
			case FtStrMap:
				sm := tt.exp.(map[string]*FieldMask)
				for _, kv := range kvs {
					if kv.IK != 0 {
						t.Fail()
					}
					if sm[kv.SK] != kv.V {
						t.Fail()
					}
					delete(sm, kv.SK)
				}
				if len(sm) > 0 {
					t.Fail()
				}
			case FtIntMap, FtList, FtStruct:
				sm := tt.exp.(map[int]*FieldMask)
				for _, kv := range kvs {
					if kv.SK != "" {
						t.Fail()
					}
					if sm[kv.IK] != kv.V {
						t.Fail()
					}
					delete(sm, kv.IK)
				}
				if len(sm) > 0 {
					t.Fail()
				}
			default:
				if len(kvs) != 0 {
					t.Fail()
				}
			}
		})
	}
}
