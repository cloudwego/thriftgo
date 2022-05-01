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

func TestLiteralEscape(t *testing.T) {
	ast, err := parser.ParseString("main.thrift", `
const string str1 = "a\'b\"c\td\ve\nf\rg\\h"
const string str2 = 'a\'b\"c\td\ve\nf\rg\\h'
	`)
	test.Assert(t, err == nil, err)
	test.Assert(t, len(ast.Constants) == 2)
	test.Assert(t, ast.Constants[0].Value.TypedValue.GetLiteral() == `a\'b"c\td\ve\nf\rg\\h`)
	test.Assert(t, ast.Constants[1].Value.TypedValue.GetLiteral() == `a'b\"c\td\ve\nf\rg\\h`)
}

const testReservedComments = `
// service definition
service test_service {
	// one-line comment
	// one-line comment
	string method0(1: string req) // non-reserved comment
	# one-line comment
	/* one-line comment */
	string method1(1: string req) # non-reserved comment
	/* cross-line
		comment */
	string method2(1: string req) /* non-reserved comment
	non-reserved comment*/
	string method3(1: string req)
	// no reserved comment before
	string method4(1: string req)
}
`

func TestServiceReservedComment(t *testing.T) {
	ast, err := parser.ParseString("main.thrift", testReservedComments)
	if err != nil {
		t.Fatal(err)
	}
	test.Assert(t, ast.Services[0].ReservedComments == `// service definition`)
	for _, f := range ast.Services[0].Functions {
		switch f.Name {
		case "method0":
			test.Assert(t, f.ReservedComments == `// one-line comment
// one-line comment`)
		case "method1":
			test.Assert(t, f.ReservedComments == `// one-line comment
/* one-line comment */`)
		case "method2":
			test.Assert(t, f.ReservedComments == `/* cross-line
		comment */`)
		case "method3":
			test.Assert(t, f.ReservedComments == ``)
		case "method4":
			test.Assert(t, f.ReservedComments == `// no reserved comment before`)
		}
	}
}

const testSpaceSkip = `
namespace
*
test
enum
Numbers
{
ONE
=
1
,
TWO
,
}
const
Numbers
MyNumber
=
ONE
typedef
i8
MyByte
struct
MyStruct
{
1
:
string
str
,
2
:
list
<
string
>
strList
}
service
MyService
{
list
<
string
>
getStrList
(
1
:
i64
id
,
)
}
`

const testCommentSkip = `
namespace /*c*/ * /*c*/test /*c*/ 
enum /*c*/ Numbers /*c*/ { /*c*/ ONE /*c*/ = /*c*/ 1 /*c*/ , /*c*/ TWO /*c*/ , /*c*/ } /*c*/ 
const /*c*/ Numbers /*c*/ MyNumber /*c*/ = /*c*/ ONE /*c*/ 
typedef /*c*/ i8 /*c*/ MyByte /*c*/ 
struct /*c*/ MyStruct /*c*/ { /*c*/ 1 /*c*/ : /*c*/ string /*c*/ str /*c*/ , /*c*/ 2 /*c*/ : /*c*/ list /*c*/ < /*c*/ string /*c*/ > /*c*/ strList /*c*/ } /*c*/ 
service /*c*/ MyService /*c*/ { /*c*/ list /*c*/ < /*c*/ string /*c*/ > /*c*/ getStrList /*c*/ ( /*c*/ 1 /*c*/ : /*c*/ i64 /*c*/ id /*c*/ , /*c*/ ) /*c*/ } /*c*/
`

func TestSkip(t *testing.T) {
	_, err := parser.ParseString("main.thrift", testSpaceSkip)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parser.ParseString("main.thrift", testCommentSkip)
	if err != nil {
		t.Fatal(err)
	}
}

const testEscape = `
const string str = "hello%s\nworld"

struct s {
	1: string f1 = "\"\'1\a2\\\t3\007本\u12e4456" (a = "vd:\"\'1\a2\\\t3\007本\u12e4456\"")
	2: string f2 = '\"\'1\a2\\\t3\007本\u12e4456' (a = 'vd:\"\'1\a2\\\t3\007本\u12e4456\"')
}
`

func TestEscape(t *testing.T) {
	ast, err := parser.ParseString("main.thrift", testEscape)
	test.Assert(t, err == nil, err)

	test.Assert(t, len(ast.Constants) == 1)
	test.Assert(t, *ast.Constants[0].Value.TypedValue.Literal == `hello%s\nworld`)

	test.Assert(t, len(ast.Structs) == 1)
	test.Assert(t, *ast.Structs[0].Fields[0].Default.TypedValue.Literal == `"\'1\a2\\\t3\007本\u12e4456`)
	test.Assert(t, *ast.Structs[0].Fields[1].Default.TypedValue.Literal == `\"'1\a2\\\t3\007本\u12e4456`)
	test.Assert(t, ast.Structs[0].Fields[0].Annotations[0].Values[0] == `vd:"\'1\a2\\\t3\007本\u12e4456"`)
	test.Assert(t, ast.Structs[0].Fields[1].Annotations[0].Values[0] == `vd:\"'1\a2\\\t3\007本\u12e4456\"`)
}
