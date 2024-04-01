// Copyright 2024 CloudWeGo Authors
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

package thrift_option

import (
	"fmt"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

func TestStructOptionWithStructBasic(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "entity.person_basic_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err := opt.GetFieldValue("valuei100")
	assert(t, err != nil)

	v, err = opt.GetFieldValue("valuei8")
	assert(t, err == nil)
	val0, ok := v.(int8)
	assert(t, ok && val0 == 8)

	v, err = opt.GetFieldValue("valuei16")
	assert(t, err == nil)
	val1, ok := v.(int16)
	assert(t, ok && val1 == 16)

	v, err = opt.GetFieldValue("valuei32")
	assert(t, err == nil)
	val2, ok := v.(int32)
	assert(t, ok && val2 == 32)

	v, err = opt.GetFieldValue("valuei64")
	assert(t, err == nil)
	val3, ok := v.(int64)
	assert(t, ok && val3 == 64)

	v, err = opt.GetFieldValue("valuestring")
	assert(t, err == nil)
	val4, ok := v.(string)
	assert(t, ok && val4 == "example@email.com")

	v, err = opt.GetFieldValue("valuebyte")
	assert(t, err == nil)
	val5, ok := v.(int8)
	assert(t, ok && val5 == 1)

	v, err = opt.GetFieldValue("valuebinary")
	assert(t, err == nil)
	val6, ok := v.([]uint8)
	assert(t, ok && len(val6) == 1 && val6[0] == 18)

	v, err = opt.GetFieldValue("valuedouble")
	assert(t, err == nil)
	val7, ok := v.(float64)
	assert(t, ok && val7 == 3.14159)

	v, err = opt.GetFieldValue("valuebool")
	assert(t, err == nil)
	val8, ok := v.(bool)
	assert(t, ok && val8 == true)
}

func TestStructOptionWithStructStruct(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test struct option
	opt, err := ParseStructOption(p, "entity.person_struct_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err := opt.GetFieldValue("valuestruct")
	assert(t, err == nil)
	val0, ok := v.(map[string]interface{})
	assert(t, ok && val0["email"] == "empty email")

	v, err = opt.GetFieldValue("valueteststruct")
	assert(t, err == nil)
	val1, ok := v.(map[string]interface{})
	assert(t, ok && val1["name"] == "lee")
	val2, ok := val1["innerStruct"].(map[string]interface{})
	assert(t, ok && val2["email"] == "no email")

	v, err = opt.GetFieldValue("valueenum")
	assert(t, err == nil)
	val3, ok := v.(int64)
	assert(t, ok && val3 == 1)

	v, err = opt.GetFieldValue("valuestructtypedef")
	assert(t, err == nil)
	val4, ok := v.(map[string]interface{})
	assert(t, ok && val4["email"] == "empty email")

	v, err = opt.GetFieldValue("valuebasictypedef")
	assert(t, err == nil)
	val5, ok := v.(string)
	assert(t, ok && val5 == "hello there")
}

func TestStructOptionWithStructContainer(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test container option
	opt, err := ParseStructOption(p, "entity.person_container_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err := opt.GetFieldValue("valuemap")
	assert(t, err == nil)
	valuemap, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")

	v, err = opt.GetFieldValue("valuelist")
	assert(t, err == nil)
	valuelist, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)
	assert(t, valuelist[0] == "list1")
	assert(t, valuelist[1] == "list2")

	v, err = opt.GetFieldValue("valueset")
	assert(t, err == nil)
	valueset, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)
	assert(t, valueset[0] == "list3")
	assert(t, valueset[1] == "list4")

	v, err = opt.GetFieldValue("valuelistsetstruct")
	assert(t, err == nil)
	valuelistsetstruct, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)

	valuelistsetstruct0, ok := valuelistsetstruct[0].([]interface{})
	assert(t, ok)
	assert(t, len(valuelistsetstruct0) == 2)

	valuelistsetstruct1, ok := valuelistsetstruct[1].([]interface{})
	assert(t, ok)
	assert(t, len(valuelistsetstruct1) == 2)

	v, err = opt.GetFieldValue("valuemapstruct")
	assert(t, err == nil)
	valuemapstruct, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemapstruct) == 2)
	valuemapstructk1, ok := valuemapstruct["k1"].(map[string]interface{})
	assert(t, ok && valuemapstructk1["email"] == "e1")
	valuemapstructk2, ok := valuemapstruct["k2"].(map[string]interface{})
	assert(t, ok && valuemapstructk2["email"] == "e2")
}

func TestStructOptionWithBasic(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "validation.person_string_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "hello")
}

func TestStructOptionWithContainer(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test container option
	opt, err := ParseStructOption(p, "validation.person_map_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valuemap, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")
}

func TestStructOptionWithEnum(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test enum option
	opt, err := ParseStructOption(p, "validation.person_enum_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valueenum, ok := v.(int64)
	assert(t, ok)
	assert(t, valueenum == 2)
}

func TestStructOptionWithTypedef(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test basic typedef option
	opt1, err := ParseStructOption(p, "validation.person_basic_typedef_info")
	assert(t, err == nil)
	assert(t, opt1 != nil)

	v := opt1.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "hello there")

	// test struct typedef option
	opt2, err := ParseStructOption(p, "validation.person_struct_typedef_info")
	assert(t, err == nil)
	assert(t, opt2 != nil)

	v = opt2.GetValue()
	assert(t, err == nil)
	valuestruct, ok := v.(map[string]interface{})
	assert(t, ok)
	assert(t, valuestruct["name"] == "empty name")
}

func TestStructOptionWithDefaultValue(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "validation.person_struct_default_value_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v1, err := opt.GetFieldValue("v1")
	assert(t, err == nil && v1.(string) == "v1 string")

	v2, err := opt.GetFieldValue("v2")
	assert(t, err == nil && v2.(string) == "v2")

	v3, err := opt.GetFieldValue("v3")
	assert(t, err == nil && v3.(int8) == 8)

	v4, err := opt.GetFieldValue("v4")
	assert(t, err == nil && v4.(int16) == 16)

	v5, err := opt.GetFieldValue("v5")
	assert(t, err == nil && v5.(int32) == 32)

	v6, err := opt.GetFieldValue("v6")
	assert(t, err == nil && v6.(int64) == 64)

	v7, err := opt.GetFieldValue("v7")
	assert(t, err == nil && v7.(bool) == true)

	v8, err := opt.GetFieldValue("v8")
	assert(t, err == nil && v8.(float64) == 3.1415926123456)

	v9, err := opt.GetFieldValue("v9")
	assert(t, err == nil && v9.(map[interface{}]interface{})["k1"].(string) == "v1")

	v10, err := opt.GetFieldValue("v10")
	assert(t, err == nil && v10.([]interface{})[0].(string) == "k1" && v10.([]interface{})[1].(string) == "k2")

	v11, err := opt.GetFieldValue("v11")
	assert(t, err == nil && v11.(string) == "hello there")
}

func TestFieldOption(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("Person")
	assert(t, p != nil)
	f := p.GetFieldByName("name")
	assert(t, f != nil)

	opt, err := ParseFieldOption(f, "entity.person_field_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "the name of this person")
}

func TestServiceAndMethodOption(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	// service option
	svc := fd.GetServiceDescriptor("MyService")
	assert(t, svc != nil)

	opt, err := ParseServiceOption(svc, "validation.svc_info")
	assert(t, err == nil, err)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valueInfo, ok := v.(map[string]interface{})
	assert(t, ok)
	assert(t, valueInfo["name"] == "ServiceInfoName")
	number, ok := valueInfo["number"].(int16)
	assert(t, ok && number == 666)

	// method option
	method := svc.GetMethodByName("M1")
	assert(t, method != nil)

	methodOption, err := ParseMethodOption(method, "validation.method_info")
	assert(t, err == nil, err)
	assert(t, methodOption != nil)

	mv := methodOption.GetValue()
	assert(t, err == nil)
	methodValueInfo, ok := mv.(map[string]interface{})
	assert(t, ok)
	assert(t, methodValueInfo["name"] == "MethodInfoName")
	methodNumber, ok := methodValueInfo["number"].(int16)
	assert(t, ok && methodNumber == 555)
}

func TestEnumAndEnumValueOption(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	// enum option
	e := fd.GetEnumDescriptor("MyEnum")
	assert(t, e != nil)

	opt, err := ParseEnumOption(e, "validation.enum_info")
	assert(t, err == nil, err)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valueInfo, ok := v.(map[string]interface{})
	assert(t, ok)
	assert(t, valueInfo["name"] == "EnumInfoName")
	number, ok := valueInfo["number"].(int16)
	assert(t, ok && number == 333)

	// enum value option
	ev := e.GetValues()[0]

	enumValueOption, err := ParseEnumValueOption(ev, "validation.enum_value_info")
	assert(t, err == nil, err)
	assert(t, enumValueOption != nil)

	evv := enumValueOption.GetValue()
	assert(t, err == nil)
	enumValueInfo, ok := evv.(map[string]interface{})
	assert(t, ok)
	assert(t, enumValueInfo["name"] == "EnumValueInfoName")
	enumValueNumber, ok := enumValueInfo["number"].(int16)
	assert(t, ok && enumValueNumber == 222)
}

func TestCommaGrammar(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("PersonB")
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "entity.person_basic_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err := opt.GetFieldValue("valuei100")
	assert(t, err != nil)

	v, err = opt.GetFieldValue("valuei8")
	assert(t, err == nil)
	val0, ok := v.(int8)
	assert(t, ok && val0 == 8)

	v, err = opt.GetFieldValue("valuei16")
	assert(t, err == nil)
	val1, ok := v.(int16)
	assert(t, ok && val1 == 16)

	v, err = opt.GetFieldValue("valuei32")
	assert(t, err == nil)
	val2, ok := v.(int32)
	assert(t, ok && val2 == 32)

	v, err = opt.GetFieldValue("valuei64")
	assert(t, err == nil)
	val3, ok := v.(int64)
	assert(t, ok && val3 == 64)

	v, err = opt.GetFieldValue("valuestring")
	assert(t, err == nil)
	val4, ok := v.(string)
	assert(t, ok && val4 == "example@email.com")

	v, err = opt.GetFieldValue("valuebyte")
	assert(t, err == nil)
	val5, ok := v.(int8)
	assert(t, ok && val5 == 1)

	v, err = opt.GetFieldValue("valuebinary")
	assert(t, err == nil)
	val6, ok := v.([]uint8)
	assert(t, ok && len(val6) == 1 && val6[0] == 18)

	v, err = opt.GetFieldValue("valuedouble")
	assert(t, err == nil)
	val7, ok := v.(float64)
	assert(t, ok && val7 == 3.14159)

	v, err = opt.GetFieldValue("valuebool")
	assert(t, err == nil)
	val8, ok := v.(bool)
	assert(t, ok && val8 == true)

	// test struct option
	opt, err = ParseStructOption(p, "entity.person_struct_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err = opt.GetFieldValue("valuestruct")
	assert(t, err == nil)
	val00, ok := v.(map[string]interface{})
	assert(t, ok && val00["email"] == "empty email")

	vs, err := opt.GetFieldValue("valueteststruct")
	assert(t, err == nil)
	val11, ok := vs.(map[string]interface{})
	assert(t, ok && val11["name"] == "lee")
	val22, ok := val11["innerStruct"].(map[string]interface{})
	assert(t, ok && val22["email"] == "no email")

	v, err = opt.GetFieldValue("valueenum")
	assert(t, err == nil)
	val33, ok := v.(int64)
	assert(t, ok && val33 == 1)

	v, err = opt.GetFieldValue("valuestructtypedef")
	assert(t, err == nil)
	val44, ok := v.(map[string]interface{})
	assert(t, ok && val44["email"] == "empty email")

	v, err = opt.GetFieldValue("valuebasictypedef")
	assert(t, err == nil)
	val55, ok := v.(string)
	assert(t, ok && val55 == "hello there")

	// test container option
	opt, err = ParseStructOption(p, "entity.person_container_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err = opt.GetFieldValue("valuemap")
	assert(t, err == nil)
	valuemap, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")

	v, err = opt.GetFieldValue("valuelist")
	assert(t, err == nil)
	valuelist, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)
	assert(t, valuelist[0] == "list1")
	assert(t, valuelist[1] == "list2")

	v, err = opt.GetFieldValue("valueset")
	assert(t, err == nil)
	valueset, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)
	assert(t, valueset[0] == "list3")
	assert(t, valueset[1] == "list4")

	v, err = opt.GetFieldValue("valuelistsetstruct")
	assert(t, err == nil)
	valuelistsetstruct, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 2)

	valuelistsetstruct0, ok := valuelistsetstruct[0].([]interface{})
	assert(t, ok)
	assert(t, len(valuelistsetstruct0) == 2)

	valuelistsetstruct1, ok := valuelistsetstruct[1].([]interface{})
	assert(t, ok)
	assert(t, len(valuelistsetstruct1) == 2)

	v, err = opt.GetFieldValue("valuemapstruct")
	assert(t, err == nil)
	valuemapstruct, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemapstruct) == 2)
	valuemapstructk1, ok := valuemapstruct["k1"].(map[string]interface{})
	assert(t, ok && valuemapstructk1["email"] == "e1")
	valuemapstructk2, ok := valuemapstruct["k2"].(map[string]interface{})
	assert(t, ok && valuemapstructk2["email"] == "e2")
}

func TestSimpleGrammar(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	_, fd := thrift_reflection.RegisterAST(ast)
	assert(t, fd != nil)

	p := fd.GetStructDescriptor("PersonC")
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "entity.person_basic_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err := opt.GetFieldValue("valuei100")
	assert(t, err != nil)

	v, err = opt.GetFieldValue("valuei8")
	assert(t, err == nil)
	val0, ok := v.(int8)
	assert(t, ok && val0 == 8)

	v, err = opt.GetFieldValue("valuei16")
	assert(t, err == nil)
	val1, ok := v.(int16)
	assert(t, ok && val1 == 16)

	// test struct option
	opt, err = ParseStructOption(p, "entity.person_struct_info")
	assert(t, err == nil)
	assert(t, opt != nil)

	v, err = opt.GetFieldValue("valuestruct")
	assert(t, err == nil)
	val00, ok := v.(map[string]interface{})
	assert(t, ok && val00["email"] == "empty email")

	vs, err := opt.GetFieldValue("valueteststruct")
	assert(t, err == nil)
	val11, ok := vs.(map[string]interface{})
	assert(t, ok && val11["name"] == "lee")
	val22, ok := val11["innerStruct"].(map[string]interface{})
	assert(t, ok && val22["email"] == "no email")

}

func TestBuildTree(t *testing.T) {
	opts := []*subValue{
		{path: "aa.bb.cc", value: "v1"},
		//{path: "aa.bb.dd", value: "v2"},
		//{path: "cc", value: "v3"},
		{path: "aa.dd", value: "v4"},
		{path: "aa.dd.ee.ff.gg", value: "v5"},
	}

	result := buildTree(opts)
	fmt.Println(result)
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
