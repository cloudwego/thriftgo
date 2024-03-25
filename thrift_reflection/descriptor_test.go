// Copyright 2023 CloudWeGo Authors
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

package thrift_reflection

import (
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
)

func TestDescriptor(t *testing.T) {
	// file descriptor
	ast, err := parser.ParseFile("reflection_test_idl.thrift", []string{"reflection_test_idl"}, true)
	assert(t, err == nil)
	_, fd := RegisterAST(ast)
	assert(t, fd != nil)
	assert(t, fd.Namespaces["go"] == "thrift_reflection_test")

	// struct descriptor
	pd := fd.GetStructDescriptor("Person")

	assert(t, pd != nil)
	assert(t, pd.GetName() == "Person")
	assert(t, pd.GetComments() == "// Person Comment")
	assert(t, pd.GetAnnotations()["k1"][0] == "hello")
	assert(t, pd.GetAnnotations()["k2"][0] == "hey")
	assert(t, pd.GetFilepath() == "reflection_test_idl.thrift")
	assert(t, len(pd.GetFields()) == 8)

	// file descriptor getter test
	personDesc := fd.GetStructDescriptor("Person")
	assert(t, personDesc == pd)
	personUnknownDesc := fd.GetStructDescriptor("person_unknown")
	assert(t, personUnknownDesc == nil)

	// struct descriptor reflection api, currently no go type registered
	assert(t, pd.GetGoType() == nil)

	// field descriptor
	basicField := pd.GetFieldByName("name")
	assert(t, basicField != nil)
	assert(t, basicField.GetName() == "name")
	assert(t, basicField.GetRequiredness() == "Required")
	assert(t, basicField.GetID() == 1)

	// type descriptor
	basicType := basicField.GetType()
	assert(t, basicType != nil)
	assert(t, basicType.IsBasic())
	assert(t, basicType.Name == "string")

	basicGoType, err := basicType.GetGoType()
	assert(t, err == nil)
	assert(t, basicGoType == reflect.TypeOf(string("")))

	// get struct descriptor from type descriptor
	structField := pd.GetFieldByName("id")
	structTypeDesc := structField.GetType()
	assert(t, structTypeDesc.IsStruct())
	stDesc, err := structTypeDesc.GetStructDescriptor()
	assert(t, err == nil)
	assert(t, stDesc.GetName() == "IDCard")
	assert(t, stDesc.GetGoType() == nil)
}

func TestLookup(t *testing.T) {
	ast, err := parser.ParseFile("reflection_test_idl.thrift", []string{"reflection_test_idl"}, true)
	assert(t, err == nil)
	gd, fd := RegisterAST(ast)
	// lookup struct
	pd := fd.GetStructDescriptor("Person")
	assert(t, pd == gd.LookupStruct(pd.GetName(), ""))
	assert(t, pd == gd.LookupStruct(pd.GetName(), pd.GetFilepath()))
	// lookup const
	cs := fd.GetConstDescriptor("MY_CONST")
	assert(t, cs == gd.LookupConst(cs.GetName(), ""))
	assert(t, cs == gd.LookupConst(cs.GetName(), cs.GetFilepath()))
	// lookup enum
	en := fd.GetEnumDescriptor("Gender")
	assert(t, en == gd.LookupEnum(en.GetName(), ""))
	assert(t, en == gd.LookupEnum(en.GetName(), en.GetFilepath()))
	// lookup typedef
	td := fd.GetTypedefDescriptor("SpecialPerson")
	assert(t, td == gd.LookupTypedef(td.GetAlias(), ""))
	assert(t, td == gd.LookupTypedef(td.GetAlias(), td.GetFilepath()))
	// lookup union
	un := fd.GetUnionDescriptor("MyUnion")
	assert(t, un == gd.LookupUnion(un.GetName(), ""))
	assert(t, un == gd.LookupUnion(un.GetName(), td.GetFilepath()))
	// lookup exception
	ex := fd.GetExceptionDescriptor("MyException")
	assert(t, ex == gd.LookupException(ex.GetName(), ""))
	assert(t, ex == gd.LookupException(ex.GetName(), td.GetFilepath()))
	// lookup service
	svc := fd.GetServiceDescriptor("MyService")
	assert(t, svc == gd.LookupService(svc.GetName(), ""))
	assert(t, svc == gd.LookupService(svc.GetName(), svc.GetFilepath()))
	// lookup method
	md := fd.GetMethodDescriptor("MyService", "M1")
	assert(t, md == gd.LookupMethod(md.GetName(), "", ""))
	assert(t, md == gd.LookupMethod(md.GetName(), "", md.GetFilepath()))
	assert(t, md == gd.LookupMethod(md.GetName(), svc.GetName(), ""))
	assert(t, md == gd.LookupMethod(md.GetName(), svc.GetName(), md.GetFilepath()))
}

func assert(t *testing.T, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			val = append([]interface{}{"assertion failed:"}, val...)
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}
