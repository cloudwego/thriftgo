// Copyright 2021 CloudWeGo Authors
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

package parser_test

import (
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
)

const testAnnotation = `
const string (a = "a") str = "str"

typedef map<i32, string> (cpp.template = "std::map") itoa_map (foo='bar')
typedef list<double (string.presentation = "hex")> float_list

enum Enum {
	E1 (value = "10"),
	E2
	E3 (value = "100")
} (eee = "eee")

struct s {
	1: string f1 ( a = "a" );
	2: string f2 ( a = "a", b = "" );
	3: string (str = "str") f3;
	4: string f4;
} (
	xxx = "",
	yyy = "y",
	zzz = "zzz",
)

exception myerror {
  1: i32 error_code ( range = "<0" )
  2: string error_msg
} (hello = "world")

service test_service {
	i32 (what = "response-annotation") method() (what = "method-annotation")
} (
	what.is.this = "service.annotation",
	empty.annotation = "",
	another.one = "more"
)

`

func TestAnnotation(t *testing.T) {
	ast, err := parser.ParseString("main.thrift", testAnnotation)
	test.Assert(t, err == nil, err)

	has := func(m parser.Annotations, k, v string) bool {
		vs := m.Get(k)
		for _, val := range vs {
			if v == val {
				return true
			}
		}
		return false
	}
	test.Assert(t, len(ast.Constants) == 1)
	test.Assert(t, len(ast.Constants[0].Annotations) == 0)
	test.Assert(t, len(ast.Constants[0].Type.Annotations) == 1)
	test.Assert(t, has(ast.Constants[0].Type.Annotations, "a", "a"))

	test.Assert(t, len(ast.Typedefs) == 2)
	test.Assert(t, len(ast.Typedefs[0].Annotations) == 1)
	test.Assert(t, has(ast.Typedefs[0].Annotations, "foo", "bar"))
	test.Assert(t, len(ast.Typedefs[0].Type.Annotations) == 1)
	test.Assert(t, has(ast.Typedefs[0].Type.Annotations, "cpp.template", "std::map"))
	test.Assert(t, len(ast.Typedefs[0].Type.KeyType.Annotations) == 0)
	test.Assert(t, len(ast.Typedefs[0].Type.ValueType.Annotations) == 0)
	test.Assert(t, len(ast.Typedefs[1].Annotations) == 0)
	test.Assert(t, len(ast.Typedefs[1].Type.Annotations) == 0)
	test.Assert(t, len(ast.Typedefs[1].Type.ValueType.Annotations) == 1)
	test.Assert(t, has(ast.Typedefs[1].Type.ValueType.Annotations, "string.presentation", "hex"))

	test.Assert(t, len(ast.Enums) == 1)
	test.Assert(t, len(ast.Enums[0].Annotations) == 1)
	test.Assert(t, has(ast.Enums[0].Annotations, "eee", "eee"))
	test.Assert(t, len(ast.Enums[0].Values) == 3)
	test.Assert(t, len(ast.Enums[0].Values[0].Annotations) == 1)
	test.Assert(t, len(ast.Enums[0].Values[1].Annotations) == 0)
	test.Assert(t, len(ast.Enums[0].Values[2].Annotations) == 1)
	test.Assert(t, has(ast.Enums[0].Values[0].Annotations, "value", "10"))
	test.Assert(t, has(ast.Enums[0].Values[2].Annotations, "value", "100"))

	test.Assert(t, len(ast.Structs) == 1)
	test.Assert(t, len(ast.Structs[0].Annotations) == 3)
	test.Assert(t, has(ast.Structs[0].Annotations, "xxx", ""))
	test.Assert(t, has(ast.Structs[0].Annotations, "yyy", "y"))
	test.Assert(t, has(ast.Structs[0].Annotations, "zzz", "zzz"))
	test.Assert(t, len(ast.Structs[0].Fields) == 4)
	test.Assert(t, len(ast.Structs[0].Fields[0].Annotations) == 1)
	test.Assert(t, len(ast.Structs[0].Fields[1].Annotations) == 2)
	test.Assert(t, len(ast.Structs[0].Fields[2].Annotations) == 0)
	test.Assert(t, len(ast.Structs[0].Fields[3].Annotations) == 0)
	test.Assert(t, has(ast.Structs[0].Fields[0].Annotations, "a", "a"))
	test.Assert(t, has(ast.Structs[0].Fields[1].Annotations, "a", "a"))
	test.Assert(t, has(ast.Structs[0].Fields[1].Annotations, "b", ""))
	test.Assert(t, has(ast.Structs[0].Fields[2].Type.Annotations, "str", "str"))

	test.Assert(t, len(ast.Exceptions) == 1)
	test.Assert(t, len(ast.Exceptions[0].Annotations) == 1)
	test.Assert(t, has(ast.Exceptions[0].Annotations, "hello", "world"))
	test.Assert(t, len(ast.Exceptions[0].Fields) == 2)
	test.Assert(t, len(ast.Exceptions[0].Fields[0].Annotations) == 1)
	test.Assert(t, len(ast.Exceptions[0].Fields[1].Annotations) == 0)
	test.Assert(t, has(ast.Exceptions[0].Fields[0].Annotations, "range", "<0"))
	test.Assert(t, len(ast.Services) == 1)
	test.Assert(t, len(ast.Services[0].Annotations) == 3)
	test.Assert(t, len(ast.Services[0].Functions) == 1)
	test.Assert(t, len(ast.Services[0].Functions[0].Annotations) == 1)
	test.Assert(t, len(ast.Services[0].Functions[0].FunctionType.Annotations) == 1)
	test.Assert(t, has(ast.Services[0].Annotations, "what.is.this", "service.annotation"))
	test.Assert(t, has(ast.Services[0].Annotations, "empty.annotation", ""))
	test.Assert(t, has(ast.Services[0].Annotations, "another.one", "more"))
	test.Assert(t, has(ast.Services[0].Functions[0].Annotations, "what", "method-annotation"))
	test.Assert(t, has(ast.Services[0].Functions[0].FunctionType.Annotations, "what", "response-annotation"))
}
