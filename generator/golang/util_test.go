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
	"github.com/cloudwego/thriftgo/parser"
	"testing"
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
		getter   func() interface{ GetAnnotations() parser.Annotations }
		expected string
	}{
		{
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
			expected: `"key": []string{"val"},
`,
		},
		{
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
			expected: `"key": []string{"val1,val2"},
`,
		},
		{
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
			expected: `"key": []string{""},
`,
		},
		{
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
			expected: `"key1": []string{"val1,val2"},
"key2": []string{"val3,val4"},
`,
		},
		{
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
			expected: `"key": []string{"val1","val2"},
`,
		},
		{
			getter: func() interface{ GetAnnotations() parser.Annotations } {
				return &parser.EnumValue{}
			},
			expected: ``,
		},
	}

	for _, c := range cases {
		arg := c.getter()
		res := genAnnotations(arg)
		if res != c.expected {
			t.Logf("genAnnotations(%+v) => %q. Expected: %q", arg, res, c.expected)
			t.Fail()
		}
	}
}
