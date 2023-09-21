package thrift_option

import (
	"github.com/cloudwego/thriftgo/parser"
	"testing"
)

func TestStructOptionWithStructBasic(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "entity.person_basic_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test struct option
	opt, err := ParseStructOption(p, "entity.person_struct_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test container option
	opt, err := ParseStructOption(p, "entity.person_container_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "validation.person_string_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test container option
	opt, err := ParseStructOption(p, "validation.person_map_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test enum option
	opt, err := ParseStructOption(p, "validation.person_enum_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test basic typedef option
	opt1, err := ParseStructOption(p, "validation.person_basic_typedef_info", ast)
	assert(t, err == nil)
	assert(t, opt1 != nil)

	v := opt1.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "hello there")

	// test struct typedef option
	opt2, err := ParseStructOption(p, "validation.person_struct_typedef_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)

	// test basic option
	opt, err := ParseStructOption(p, "validation.person_struct_default_value_info", ast)
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

	p := getStructFromAst("Person", ast)
	assert(t, p != nil)
	f, ok := p.GetField("name")

	opt, err := ParseFieldOption(f, "entity.person_field_info", ast)
	assert(t, err == nil)
	assert(t, opt != nil)

	v := opt.GetValue()
	assert(t, err == nil)
	valuestring, ok := v.(string)
	assert(t, ok)
	assert(t, valuestring == "the name of this person")

	optLocal, err := ParseFieldOption(f, "local_field_info", ast)
	assert(t, err == nil)
	assert(t, optLocal != nil)

	v = optLocal.GetValue()
	assert(t, err == nil)
	valuestring, ok = v.(string)
	assert(t, ok)
	assert(t, valuestring == "the ID of this person")

}

func TestServiceAndMethodOption(t *testing.T) {

	ast, err := parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	// service option
	svc := getServiceFromAst("MyService", ast)
	assert(t, svc != nil)

	opt, err := ParseServiceOption(svc, "validation.svc_info", ast)
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
	method := getMethodFromService("M1", svc)
	assert(t, method != nil)

	methodOption, err := ParseMethodOption(method, "validation.method_info", ast)
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

	// enum option
	e := getEnumFromAst("MyEnum", ast)
	assert(t, e != nil)

	opt, err := ParseEnumOption(e, "validation.enum_info", ast)
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
	ev := getEnumValueFromEnum("A", e)

	enumValueOption, err := ParseEnumValueOption(ev, "validation.enum_value_info", ast)
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
