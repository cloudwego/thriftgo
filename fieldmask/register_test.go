// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fieldmask

import (
	"testing"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

func TestRegisterFieldMask(t *testing.T) {
	type args struct {
		id        uint64
		desc      *thrift_reflection.StructDescriptor
		fm        *FieldMask
		inMask    []string
		notInMask []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{
				id:        1,
				desc:      GetDescriptor(baseIDL, "Base"),
				fm:        NewFieldMaskFromNames(GetDescriptor(baseIDL, "Base"), "Caller"),
				inMask:    []string{"Caller"},
				notInMask: []string{"Addr"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterFieldMask(tt.args.id, tt.args.desc, tt.args.fm)
			fm := GetFieldMask(tt.args.id, tt.args.desc)
			if fm == nil {
				t.Fail()
			}
			for _, path := range tt.args.inMask {
				if !fm.PathInMask(tt.args.desc, path) {
					t.Fatal(path)
				}
			}
			for _, path := range tt.args.notInMask {
				if fm.PathInMask(tt.args.desc, path) {
					t.Fatal(path)
				}
			}
		})
	}
}

func BenchmarkGetFieldMask(b *testing.B) {
	desc := GetDescriptor(baseIDL, "Base")
	fm := NewFieldMaskFromNames(desc, "Caller")
	RegisterFieldMask(1, desc, fm)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetFieldMask(1, desc)
	}
}
