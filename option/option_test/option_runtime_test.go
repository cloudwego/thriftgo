package option_test_test

import (
	"github.com/cloudwego/thriftgo/option"
	"github.com/cloudwego/thriftgo/option/option_test"
	"github.com/cloudwego/thriftgo/option/option_test/annotation/entity"
	"github.com/cloudwego/thriftgo/option/option_test/annotation/validation"
	"strings"
	"testing"
)

func TestRuntimeBasicStructOption(t *testing.T) {
	// test basic option
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_BASIC_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.(*entity.PersonBasicInfo)
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
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_STRUCT_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.(*entity.PersonStructInfo)
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
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), entity.STRUCT_OPTION_PERSON_CONTAINER_INFO)
	assert(t, err == nil && option != nil)
	opt, ok := option.(*entity.PersonContainerInfo)
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
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRING_INFO)
	assert(t, err == nil && option != nil)
	valuestring, ok := option.(string)
	assert(t, ok)
	assert(t, valuestring == "hello")

}

func TestRuntimeContainerOption(t *testing.T) {

	// test container
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_MAP_INFO)
	assert(t, err == nil && option != nil)
	valuemap, ok := option.(map[string]string)
	assert(t, ok)
	assert(t, len(valuemap) == 1)
	assert(t, valuemap["hey1"] == "value1")

}

func TestRuntimeEnumOption(t *testing.T) {
	// test enum option
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_ENUM_INFO)
	assert(t, err == nil && option == validation.MyEnum_XXL)
}

func TestRuntimeTypedefOption(t *testing.T) {
	// test basic typedef option
	option1, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_BASIC_TYPEDEF_INFO)
	assert(t, err == nil && option1 != nil)
	valuebasicTypedef, ok := option1.(validation.MyBasicTypedef)
	assert(t, ok && valuebasicTypedef == "hello there")

	// test struct typedef option
	option2, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRUCT_TYPEDEF_INFO)
	assert(t, err == nil && option2 != nil)
	valuestructTypedef, ok := option2.(*validation.MyStructTypedef)
	assert(t, ok && valuestructTypedef.GetName() == "empty name")
}

func TestRuntimeStructOptionWithDefaultValue(t *testing.T) {
	option, err := option.GetStructOption(option_test.NewPerson().GetDescriptor(), validation.STRUCT_OPTION_PERSON_STRUCT_DEFAULT_VALUE_INFO)
	assert(t, err == nil)

	opt := option.(*validation.MyStructWithDefaultVal)
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

// 检测各种报错提示场景
func TestRuntimeOptionError(t *testing.T) {

	// 获取不存在的 option
	_, err := option.GetStructOption(option_test.NewPersonA().GetDescriptor(), entity.STRUCT_OPTION_PERSON_BASIC_INFO)
	assert(t, err != nil && err.Error() == "option not exist on current descriptor")

	// 错误的 field value
	_, err = option.GetStructOption(option_test.NewPersonB().GetDescriptor(), entity.STRUCT_OPTION_PERSON_BASIC_INFO)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_basic_info:strconv.ParseInt: parsing \"hellostring\": invalid syntax", err)

	// 错误的 field name
	_, err = option.GetStructOption(option_test.NewPersonC().GetDescriptor(), entity.STRUCT_OPTION_PERSON_STRUCT_INFO)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_struct_info:field not exist:value_xxx", err)

	// 错误的 option 语法
	_, err = option.GetStructOption(option_test.NewPersonD().GetDescriptor(), entity.STRUCT_OPTION_PERSON_STRUCT_INFO)
	assert(t, err != nil && strings.HasPrefix(err.Error(), "grammar error"), err)

	// 错误的 kv 语法
	_, err = option.GetStructOption(option_test.NewPersonE().GetDescriptor(), entity.STRUCT_OPTION_PERSON_CONTAINER_INFO)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_container_info:{} not match", err)

}

func TestRuntimeFieldOption(t *testing.T) {

	pd := option_test.NewPerson().GetDescriptor()
	fd := pd.GetFieldByName("name")
	// test basic string option
	option, err := option.GetFieldOption(fd, entity.FIELD_OPTION_PERSON_FIELD_INFO)
	assert(t, err == nil && option != nil)
	valuestring, ok := option.(string)
	assert(t, ok)
	assert(t, valuestring == "the name of this person")

}

func TestServiceAndMethodOption(t *testing.T) {

	// service option
	svc := option_test.GetServiceDescriptorForMyService()
	opt, err := option.GetServiceOption(svc, validation.SERVICE_OPTION_SVC_INFO)
	assert(t, err == nil, err)

	valueInfo, ok := opt.(*validation.TestInfo)
	assert(t, ok)
	assert(t, valueInfo.GetName() == "ServiceInfoName")
	assert(t, valueInfo.GetNumber() == 666)

	// method option

	//method := option_test.GetMethodDescriptorForMyServiceM1()
	method := svc.GetMethodByName("M1")
	assert(t, method != nil)
	methodOption, err := option.GetMethodOption(method, validation.METHOD_OPTION_METHOD_INFO)
	assert(t, err == nil, err)

	methodValueInfo, ok := methodOption.(*validation.TestInfo)
	assert(t, ok)
	assert(t, methodValueInfo.GetName() == "MethodInfoName")
	assert(t, methodValueInfo.GetNumber() == 555)

}

func TestEnumAndEnumValueOption(t *testing.T) {

	// enum option
	e := option_test.MyEnum(0).GetDescriptor()
	assert(t, e != nil)
	opt, err := option.GetEnumOption(e, validation.ENUM_OPTION_ENUM_INFO)
	assert(t, err == nil, err)

	valueInfo, ok := opt.(*validation.TestInfo)
	assert(t, ok)
	assert(t, valueInfo.GetName() == "EnumInfoName")
	assert(t, valueInfo.GetNumber() == 333)

	// enum value option
	ev := e.GetValues()[0]
	enumValueOption, err := option.GetEnumValueOption(ev, validation.ENUM_VALUE_OPTION_ENUM_VALUE_INFO)
	assert(t, err == nil, err)

	enumValueInfo, ok := enumValueOption.(*validation.TestInfo)
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
