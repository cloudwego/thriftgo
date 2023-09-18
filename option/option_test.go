package option

import (
	"github.com/cloudwego/thriftgo/parser"
	"strings"
	"testing"
)

func TestStructOptionWithStructBasic(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test basic option
	opt := options["entity.person_basic_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test struct option
	opt := options["entity.person_struct_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test container option
	opt := options["entity.person_container_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test basic option
	opt := options["validation.person_string_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test container option
	opt := options["validation.person_map_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test enum option
	opt := options["validation.person_enum_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test basic typedef option
	opt1 := options["validation.person_basic_typedef_info"]
	assert(t, opt1 != nil)

	v := opt1.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "hello there")

	// test struct typedef option
	opt2 := options["validation.person_struct_typedef_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	options, err := ParseStructOption(p, ast)
	assert(t, err == nil, err)

	// test basic option
	opt := options["validation.person_struct_default_value_info"]
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	f, ok := p.GetField("name")
	options, err := ParseFieldOption(f, ast)
	assert(t, err == nil, err)

	opt := options["entity.person_field_info"]
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

	// service option
	svc := getServiceFromAst("MyService", ast)
	assert(t, svc != nil)
	options, err := ParseServiceOption(svc, ast)
	assert(t, err == nil, err)

	opt := options["validation.svc_info"]
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valueInfo, ok := v.(map[string]interface{})
	assert(t, ok)
	assert(t, valueInfo["name"] == "ServiceInfoName")
	number, ok := valueInfo["number"].(int16)
	assert(t, ok && number == 666)

	// method option
	method := getMethodFromService("M1", svc)
	assert(t, method != nil)
	methodOptions, err := ParseMethodOption(method, ast)
	assert(t, err == nil, err)

	methodOption := methodOptions["validation.method_info"]
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

	// enum option
	e := getEnumFromAst("MyEnum", ast)
	assert(t, e != nil)
	options, err := ParseEnumOption(e, ast)
	assert(t, err == nil, err)

	opt := options["validation.enum_info"]
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valueInfo, ok := v.(map[string]interface{})
	assert(t, ok)
	assert(t, valueInfo["name"] == "EnumInfoName")
	number, ok := valueInfo["number"].(int16)
	assert(t, ok && number == 333)

	// enum value option
	ev := getEnumValueFromEnum("A", e)
	methodOptions, err := ParseEnumValueOption(ev, ast)
	assert(t, err == nil, err)

	enumValueOption := methodOptions["validation.enum_value_info"]
	assert(t, enumValueOption != nil)

	evv := enumValueOption.GetValue()
	assert(t, err == nil)
	enumValueInfo, ok := evv.(map[string]interface{})
	assert(t, ok)
	assert(t, enumValueInfo["name"] == "EnumValueInfoName")
	enumValueNumber, ok := enumValueInfo["number"].(int16)
	assert(t, ok && enumValueNumber == 222)

}

// 检测各种报错提示场景
func TestOptionError(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test_grammar_error.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	// 错误或者不存在的 Option 名称
	p := getStructFromAst("PersonA", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && err.Error() == "no such option: entity.person_xxx_info", err)

	// 错误的 field value
	p = getStructFromAst("PersonB", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_basic_info:strconv.ParseInt: parsing \"hellostring\": invalid syntax", err)

	// 错误的 field name
	p = getStructFromAst("PersonC", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_struct_info:field not exist:value_xxx", err)

	// 错误的 option 语法
	p = getStructFromAst("PersonD", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && strings.HasPrefix(err.Error(), "grammar error"), err)

	// 错误的 kv 语法
	p = getStructFromAst("PersonE", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && err.Error() == "parse failed for entity.person_container_info:{} not match", err)

	// 没有 include 对应 option 的 IDL
	p = getStructFromAst("PersonF", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, ast)
	assert(t, err != nil && err.Error() == "no such struct found from given include IDLs:validation.person_string_info", err)

}

func TestParseOptionStr(t *testing.T) {
	st1 := "IsOdd=true"
	name, content, ok := ParseOptionStr(st1)
	assert(t, ok && name == "IsOdd" && content == "true")

	st2 := "m2.IsOdd = true"
	name, content, ok = ParseOptionStr(st2)
	assert(t, ok && name == "m2.IsOdd" && content == "true")

	st3 := " IsOdd= true"
	name, content, ok = ParseOptionStr(st3)
	assert(t, ok && name == "IsOdd" && content == "true")

	st4 := "MyStruct={a:b c:d e=f}"
	name, content, ok = ParseOptionStr(st4)
	assert(t, ok && name == "MyStruct" && content == "{a:b c:d e=f}")

	st5 := `MyStruct={
			a:b c:d
			e:f
			g:h
		}
	`
	name, content, ok = ParseOptionStr(st5)
	assert(t, ok && name == "MyStruct" && content == "{    a:b c:d    e:f    g:h   }")
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

func getStructFromAst(name string, ast *parser.Thrift) *parser.StructLike {
	for _, st := range ast.Structs {
		if st.Name == name {
			return st
		}
	}
	return nil
}

func getServiceFromAst(name string, ast *parser.Thrift) *parser.Service {
	for _, svc := range ast.Services {
		if svc.Name == name {
			return svc
		}
	}
	return nil
}

func getMethodFromService(name string, service *parser.Service) *parser.Function {
	for _, m := range service.Functions {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func getEnumFromAst(name string, ast *parser.Thrift) *parser.Enum {
	for _, e := range ast.Enums {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func getEnumValueFromEnum(name string, e *parser.Enum) *parser.EnumValue {
	for _, ev := range e.Values {
		if ev.Name == name {
			return ev
		}
	}
	return nil
}
