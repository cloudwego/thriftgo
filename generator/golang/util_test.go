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

package golang

import (
	"testing"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/parser"
)

func TestSnakify(t *testing.T) {
	cases := []struct{ original, expected string }{
		{"a", "a"},
		{"A", "a"},
		{"AB", "ab"},
		{"HTTPRequest", "http_request"},
		{"HTTP1Method", "http1_method"},
		{"GetUserIP", "get_user_ip"},
	}
	for _, c := range cases {
		res := snakify(c.original)
		if res != c.expected {
			t.Logf("snakify(%q) => %q. Expected: %q", c.original, res, c.expected)
			t.Fail()
		}
	}
}

func TestLowerCamelCase(t *testing.T) {
	cases := []struct{ original, expected string }{
		{"a", "a"},
		{"A", "a"},
		{"AB", "ab"},
		{"HTTPRequest", "httpRequest"},
		{"HTTP1Method", "http1Method"},
		{"GetUserIP", "getUserIp"},
		{"GetAPI", "getApi"},
		{"Get_API", "getApi"},
	}
	for _, c := range cases {
		res := lowerCamelCase(c.original)
		if res != c.expected {
			t.Logf("lowerCamelCase(%q) => %q. Expected: %q", c.original, res, c.expected)
			t.Fail()
		}
	}
}

func TestGenAnnotations(t *testing.T) {
	cases := []struct {
		desc     string
		getter   func() interface{ GetAnnotations() parser.Annotations }
		expected string
	}{
		{
			desc: "normal case",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{"val"},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`val`},\n",
		},
		{
			desc: "single value seperated by comma",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{"val1,val2"},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`val1,val2`},\n",
		},
		{
			desc: "single empty value",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{""},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{``},\n",
		},
		{
			desc: "multiple keys",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key1",
						Values: []string{"val1,val2"},
					},
					{
						Key:    "key2",
						Values: []string{"val3,val4"},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key1`: []string{`val1,val2`},\n`key2`: []string{`val3,val4`},\n",
		},
		{
			desc: "single key, multiple values",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{"val1", "val2"},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`val1`,`val2`},\n",
		},
		{
			desc: "double quotes are not escaped",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{`\"val\"`},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`\\\"val\\\"`},\n",
		},
		{
			desc: "double quotes are escaped",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{"\"val\""},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`\"val\"`},\n",
		},
		{
			desc: "single quotes are not escaped",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{`\'val\'`},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`\\'val\\'`},\n",
		},
		{
			desc: "single quotes are escaped",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				annos := parser.Annotations{
					{
						Key:    "key",
						Values: []string{"'val'"},
					},
				}
				return &parser.EnumValue{
					Annotations: annos,
				}
			},
			expected: "`key`: []string{`'val'`},\n",
		},
		{
			desc: "nil Annotations",
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				return &parser.EnumValue{}
			},
			expected: ``,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			arg := c.getter()
			res := genAnnotations(arg)
			if res != c.expected {
				t.Logf("genAnnotations(%+v) => %q. Expected: %q", arg, res, c.expected)
				t.Fail()
			}
		})
	}
}

func TestGenFieldTags(t *testing.T) {
	newField := func(annos parser.Annotations) *Field {
		return &Field{Field: &parser.Field{
			ID:           1,
			Name:         "Name",
			Requiredness: parser.FieldType_Default,
			Annotations:  annos,
		}}
	}

	cases := []struct {
		desc     string
		field    *Field
		feats    func(f *Features)
		expected string
	}{
		{
			desc:     "no go.tag, gen_json_tag default on",
			field:    newField(nil),
			expected: "`thrift:\"Name,1\" json:\"Name\"`",
		},
		{
			desc:     "go.tag with json: keeps go.tag json, no duplicate",
			field:    newField(parser.Annotations{{Key: "go.tag", Values: []string{`json:"n" yaml:"n"`}}}),
			expected: "`thrift:\"Name,1\" json:\"n\" yaml:\"n\"`",
		},
		{
			desc:     "go.tag without json: still gets default json tag",
			field:    newField(parser.Annotations{{Key: "go.tag", Values: []string{`yaml:"n"`}}}),
			expected: "`thrift:\"Name,1\" yaml:\"n\" json:\"Name\"`",
		},
		{
			desc:  "go.tag without json: but gen_json_tag off → no json tag",
			field: newField(parser.Annotations{{Key: "go.tag", Values: []string{`yaml:"n"`}}}),
			feats: func(f *Features) {
				f.GenerateJSONTag = false
			},
			expected: "`thrift:\"Name,1\" yaml:\"n\"`",
		},
		{
			desc:  "always_gen_json_tag forces json even when go.tag has json",
			field: newField(parser.Annotations{{Key: "go.tag", Values: []string{`json:"n"`}}}),
			feats: func(f *Features) {
				f.AlwaysGenerateJSONTag = true
			},
			expected: "`thrift:\"Name,1\" json:\"n\" json:\"Name\"`",
		},
		{
			// e.g. single-quoted thrift literal `'json:\"n\"'` survives the parser as raw `json:\"n\"`;
			// the lookup must run after EscapeDoubleInTag or the json key is missed.
			desc:     "go.tag with escaped quotes still detects json key",
			field:    newField(parser.Annotations{{Key: "go.tag", Values: []string{`json:\"n\"`}}}),
			expected: "`thrift:\"Name,1\" json:\"n\"`",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			cu := NewCodeUtils(backend.DummyLogFunc())
			feats := defaultFeatures
			if c.feats != nil {
				c.feats(&feats)
			}
			cu.SetFeatures(feats)
			got, err := cu.GenFieldTags(c.field, "")
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != c.expected {
				t.Fatalf("got %q, want %q", got, c.expected)
			}
		})
	}
}
