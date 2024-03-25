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

package runtime_test

import (
	"testing"

	"github.com/cloudwego/thriftgo/extension/thrift_option"
	"github.com/cloudwego/thriftgo/extension/thrift_option/runtime_test/option_gen"
	"github.com/cloudwego/thriftgo/extension/thrift_option/runtime_test/option_gen/annotation/entity"
	"github.com/cloudwego/thriftgo/extension/thrift_option/runtime_test/option_gen/annotation/validation"
)

func TestRuntimeSimpleGrammarOption(t *testing.T) {
	// test basic option
	option, err := thrift_option.GetStructOption(option_gen.NewPersonC().GetDescriptor(), entity.STRUCT_OPTION_PERSON_BASIC_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.GetInstance().(*entity.PersonBasicInfo)
	assert(t, ok)

	assert(t, opt.GetValuei8() == 8)
	assert(t, opt.GetValuei16() == 16)

	// test struct option
	option2, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_STRUCT_INFO)
	assert(t, err == nil && option != nil)
	opt2, ok := option2.GetInstance().(*entity.PersonStructInfo)
	assert(t, ok)

	innerStruct := opt2.GetValuestruct()
	assert(t, innerStruct != nil && innerStruct.GetEmail() == "empty email")

	testStruct := opt2.GetValueteststruct()
	assert(t, testStruct != nil)
	assert(t, testStruct.GetName() == "lee")
	innerStruct = testStruct.GetInnerStruct()
	assert(t, innerStruct != nil && innerStruct.GetEmail() == "no email")

}

func TestRuntimeBasicStructOption(t *testing.T) {
	// test basic option
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_BASIC_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.GetInstance().(*entity.PersonBasicInfo)
	assert(t, ok)

	assert(t, opt.GetValuei8() == 8)
	assert(t, opt.GetValuei16() == 16)
	assert(t, opt.GetValuei32() == 32)
	assert(t, opt.GetValuei64() == 64)
	assert(t, opt.GetValuestring() == "example@email.com")
	assert(t, opt.GetValuebyte() == 1)
	assert(t, len(opt.GetValuebinary()) == 1 && opt.GetValuebinary()[0] == 18)
	assert(t, opt.GetValuedouble() == 3.14159)
	assert(t, opt.GetValuebool() == true)
}

func TestRuntimeStructStructOption(t *testing.T) {
	// test struct option
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_STRUCT_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.GetInstance().(*entity.PersonStructInfo)
	assert(t, ok)

	innerStruct := opt.GetValuestruct()
	assert(t, innerStruct != nil && innerStruct.GetEmail() == "empty email")

	testStruct := opt.GetValueteststruct()
	assert(t, testStruct != nil)
	assert(t, testStruct.GetName() == "lee")
	innerStruct = testStruct.GetInnerStruct()
	assert(t, innerStruct != nil && innerStruct.GetEmail() == "no email")

	// test enum option
	valueenum := opt.GetValueenum()
	assert(t, valueenum == entity.TestEnum_B)

	// test basic typedef option
	valuebasicTypedef := opt.GetValuebasictypedef()
	assert(t, ok && valuebasicTypedef == "hello there")

	// test struct typedef option
	valuestructTypedef := opt.GetValuestructtypedef()
	assert(t, valuestructTypedef != nil)
	assert(t, ok && valuestructTypedef.GetEmail() == "empty email")
}

func TestRuntimeContainerStructOption(t *testing.T) {
	// test container option
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_CONTAINER_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.GetInstance().(*entity.PersonContainerInfo)
	assert(t, ok)

	valuemap := opt.GetValuemap()
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")

	valuelist := opt.GetValuelist()
	assert(t, len(valuelist) == 2)
	assert(t, valuelist[0] == "list1")
	assert(t, valuelist[1] == "list2")

	valueset := opt.GetValueset()
	assert(t, len(valuelist) == 2)
	assert(t, valueset[0] == "list3")
	assert(t, valueset[1] == "list4")

	valuelistsetstruct := opt.GetValuelistsetstruct()
	valuelistsetstruct0 := valuelistsetstruct[0]
	assert(t, len(valuelistsetstruct0) == 2)
	valuelistsetstruct1 := valuelistsetstruct[1]
	assert(t, len(valuelistsetstruct1) == 2)

	valuemapstruct := opt.GetValuemapstruct()
	assert(t, len(valuemapstruct) == 2)
	valuemapstructk1 := valuemapstruct["k1"]
	assert(t, ok && valuemapstructk1.GetEmail() == "e1")
	valuemapstructk2 := valuemapstruct["k2"]
	assert(t, ok && valuemapstructk2.GetEmail() == "e2")
}

func TestRuntimeBasicOption(t *testing.T) {
	// test basic string option
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRING_INFO)
	assert(t, err == nil && option != nil)
	valuestring, ok := option.GetInstance().(string)
	assert(t, ok)
	assert(t, valuestring == "hello")
}

func TestRuntimeContainerOption(t *testing.T) {
	// test container
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_MAP_INFO)
	assert(t, err == nil && option != nil)
	valuemap, ok := option.GetInstance().(map[string]string)
	assert(t, ok)
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")
}

func TestRuntimeEnumOption(t *testing.T) {
	// test enum option
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_ENUM_INFO)
	assert(t, err == nil && option.GetInstance() == validation.MyEnum_XXL)
}

func TestRuntimeTypedefOption(t *testing.T) {
	// test basic typedef option
	option1, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_BASIC_TYPEDEF_INFO)
	assert(t, err == nil && option1 != nil)
	valuebasicTypedef, ok := option1.GetInstance().(validation.MyBasicTypedef)
	assert(t, ok && valuebasicTypedef == "hello there")

	// test struct typedef option
	option2, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRUCT_TYPEDEF_INFO)
	assert(t, err == nil && option2 != nil)
	valuestructTypedef, ok := option2.GetInstance().(*validation.MyStructTypedef)
	assert(t, ok && valuestructTypedef.GetName() == "empty name")
}

func TestRuntimeStructOptionWithDefaultValue(t *testing.T) {
	option, err := thrift_option.GetStructOption(option_gen.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRUCT_DEFAULT_VALUE_INFO)
	assert(t, err == nil)

	opt := option.GetInstance().(*validation.MyStructWithDefaultVal)
	assert(t, opt.GetV1() == "v1 string")
	assert(t, opt.GetV2() == "v2")
	assert(t, opt.GetV3() == 8)
	assert(t, opt.GetV4() == 16)
	assert(t, opt.GetV5() == 32)
	assert(t, opt.GetV6() == 64)
	assert(t, opt.GetV7() == true)
	assert(t, opt.GetV8() == 3.1415926123456)
	assert(t, opt.GetV9()["k1"] == "v1")
	assert(t, opt.GetV10()[0] == "k1" && opt.GetV10()[1] == "k2")
	assert(t, opt.GetV11() == "hello there")
}

func TestRuntimeFieldOption(t *testing.T) {
	pd := option_gen.NewPerson().GetDescriptor()
	fd := pd.GetFieldByName("name")
	// test basic string option
	opt, err := thrift_option.GetFieldOption(fd, entity.FIELD_OPTION_PERSON_FIELD_INFO)
	assert(t, err == nil && opt != nil)
	valuestring, ok := opt.GetInstance().(string)
	assert(t, ok)
	assert(t, valuestring == "the name of this person")
}

func TestRuntimeServiceAndMethodOption(t *testing.T) {
	// service option
	svc := option_gen.GetFileDescriptorForTest().GetServiceDescriptor("MyService")
	opt, err := thrift_option.GetServiceOption(svc, validation.SERVICE_OPTION_SVC_INFO)
	assert(t, err == nil, err)

	valueInfo, ok := opt.GetInstance().(*validation.TestInfo)
	assert(t, ok)
	assert(t, valueInfo.GetName() == "ServiceInfoName")
	assert(t, valueInfo.GetNumber() == 666)

	// method option

	// method := option_gen.GetMethodDescriptorForMyServiceM1()
	method := svc.GetMethodByName("M1")
	assert(t, method != nil)
	methodOption, err := thrift_option.GetMethodOption(method, validation.METHOD_OPTION_METHOD_INFO)
	assert(t, err == nil, err)

	methodValueInfo, ok := methodOption.GetInstance().(*validation.TestInfo)
	assert(t, ok)
	assert(t, methodValueInfo.GetName() == "MethodInfoName")
	assert(t, methodValueInfo.GetNumber() == 555)
}

func TestRuntimeEnumAndEnumValueOption(t *testing.T) {
	// enum option
	e := option_gen.MyEnum(0).GetDescriptor()
	assert(t, e != nil)
	opt, err := thrift_option.GetEnumOption(e, validation.ENUM_OPTION_ENUM_INFO)
	assert(t, err == nil, err)

	valueInfo, ok := opt.GetInstance().(*validation.TestInfo)
	assert(t, ok)
	assert(t, valueInfo.GetName() == "EnumInfoName")
	assert(t, valueInfo.GetNumber() == 333)

	// enum value option
	ev := e.GetValues()[0]
	enumValueOption, err := thrift_option.GetEnumValueOption(ev, validation.ENUM_VALUE_OPTION_ENUM_VALUE_INFO)
	assert(t, err == nil, err)

	enumValueInfo, ok := enumValueOption.GetInstance().(*validation.TestInfo)
	assert(t, ok)
	assert(t, enumValueInfo.GetName() == "EnumValueInfoName")
	assert(t, enumValueInfo.GetNumber() == 222)
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
