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

package parser

import (
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestDetectKeyword(t *testing.T) {
	tree := &Thrift{}
	test.Assert(t, len(DetectKeyword(tree)) == 0)

	cnt := 0
	count := func(s string) string {
		cnt++
		return s
	}
	tree = &Thrift{
		Typedefs: []*Typedef{
			{Alias: "ok"}, {Alias: count("if")},
		},
		Constants: []*Constant{
			{Name: "ok"}, {Name: count("else")},
		},
		Enums: []*Enum{
			{Name: "ok"}, {Name: count("for")},
		},
		Structs: []*StructLike{
			{
				Category: "struct",
				Name:     "ok",
				Fields: []*Field{
					{Name: "fok"},
				},
			},
			{
				Category: "struct",
				Name:     count("do"),
				Fields: []*Field{
					{Name: count("while")},
				},
			},
		},
		Unions: []*StructLike{
			{
				Category: "union",
				Name:     "ok",
				Fields: []*Field{
					{Name: "fok"},
				},
			},
			{
				Category: "union",
				Name:     count("continue"),
				Fields: []*Field{
					{Name: count("break")},
				},
			},
		},
		Exceptions: []*StructLike{
			{
				Category: "exception",
				Name:     "ok",
				Fields: []*Field{
					{Name: "fok"},
				},
			},
			{
				Category: "exception",
				Name:     count("True"),
				Fields: []*Field{
					{Name: count("False")},
				},
			},
		},
		Services: []*Service{
			{
				Name: "ok",
				Functions: []*Function{
					{
						Name: "fok",
						Arguments: []*Field{
							{Name: "aok"},
						},
						Throws: []*Field{
							{Name: "tok"},
						},
					},
				},
			},
			{
				Name: count("goto"),
				Functions: []*Function{
					{
						Name: count("lambda"),
						Arguments: []*Field{
							{Name: count("try")},
						},
						Throws: []*Field{
							{Name: count("except")},
						},
					},
				},
			},
		},
	}
	warns := DetectKeyword(tree)
	test.Assert(t, len(warns) == cnt)
}
