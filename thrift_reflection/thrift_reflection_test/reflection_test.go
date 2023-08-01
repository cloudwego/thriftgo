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

package thrift_reflection_test_test

import (
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/thrift_reflection/thrift_reflection_test"
)

func TestDescriptor(t *testing.T) {
	// file descriptor
	fd := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl()
	assert(t, fd != nil)
	assert(t, fd.Namespaces["go"] == "thrift_reflection_test")

	// struct descriptor
	person := &thrift_reflection_test.Person{
		Name: "Lee",
		ID: &thrift_reflection_test.IDCard{
			Number: "123",
			Age:    23,
		},
	}
	pd := person.GetDescriptor()

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

	// struct descriptor reflection api
	assert(t, pd.GetGoType() == reflect.TypeOf(thrift_reflection_test.Person{}))

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
	assert(t, stDesc.GetGoType() == reflect.TypeOf(thrift_reflection_test.IDCard{}))

	// enum
	enumDesc := thrift_reflection_test.Gender_MALE.GetDescriptor()
	assert(t, enumDesc != nil)
	assert(t, enumDesc.GetName() == "Gender")
	assert(t, len(enumDesc.GetValues()) == 2)
	maleValue := enumDesc.GetValues()[0]
	assert(t, maleValue.GetValue() == 0)
	assert(t, maleValue.GetName() == "MALE")

	// get enum descriptor from type descriptor
	enumField := pd.GetFieldByName("gender")
	assert(t, enumField.GetType().IsEnum())
	enumDescFromField, err := enumField.GetType().GetEnumDescriptor()
	assert(t, err == nil)
	assert(t, enumDescFromField == enumDesc)

	// typedef
	basicTypedefDesc := thrift_reflection.GetTypedefDescriptorByGoType((*thrift_reflection_test.SpecialString)(nil))
	assert(t, basicTypedefDesc != nil)
	assert(t, basicTypedefDesc.GetAlias() == "SpecialString")
	assert(t, basicTypedefDesc.GetType().IsBasic())
	assert(t, basicTypedefDesc.GetType().GetName() == "string")
	// get typedef descriptor from type descriptor
	typedefField := pd.GetFieldByName("typedefValue")
	assert(t, typedefField.GetType().IsTypedef())
	typeDescFromField, err := typedefField.GetType().GetTypedefDescriptor()
	assert(t, err == nil)
	assert(t, typeDescFromField == basicTypedefDesc)

	// structTypedefDesc := thrift_reflection.GetTypedefDescriptorByGoType((*thrift_reflection_test.SpecialPerson)(nil))
	structTypedefDesc := thrift_reflection.GetTypedefDescriptorByGoType(&thrift_reflection_test.SpecialPerson{})
	assert(t, structTypedefDesc != nil)
	assert(t, structTypedefDesc.GetAlias() == "SpecialPerson")
	assert(t, structTypedefDesc.GetType().IsStruct())
	assert(t, structTypedefDesc.GetType().GetName() == "Person")

	// const
	constDesc := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetConstDescriptor("MY_CONST")
	assert(t, constDesc != nil)
	assert(t, constDesc.GetName() == "MY_CONST")
	constGoType, err := constDesc.GetType().GetGoType()
	assert(t, err == nil)
	assert(t, constGoType == reflect.TypeOf(string("")))

	// union
	unionDesc := thrift_reflection_test.NewMyUnion().GetDescriptor()
	assert(t, unionDesc != nil)
	assert(t, unionDesc.GetName() == "MyUnion")
	unionGoType := unionDesc.GetGoType()
	assert(t, unionGoType == reflect.TypeOf(thrift_reflection_test.MyUnion{}))
	// get union descriptor from type descriptor
	unionField := pd.GetFieldByName("uni")
	assert(t, unionField.GetType().IsUnion())
	unionDescFromField, err := unionField.GetType().GetUnionDescriptor()
	assert(t, err == nil)
	assert(t, unionDescFromField == unionDesc)

	// exception
	exDesc := thrift_reflection_test.NewMyException().GetDescriptor()
	assert(t, exDesc != nil)
	assert(t, exDesc.GetName() == "MyException")
	exGoType := exDesc.GetGoType()
	assert(t, exGoType == reflect.TypeOf(thrift_reflection_test.MyException{}))
	// get exception descriptor from type descriptor
	exceptionField := pd.GetFieldByName("exp")
	assert(t, exceptionField.GetType().IsException())
	exceptionDescFromField, err := exceptionField.GetType().GetExceptionDescriptor()
	assert(t, err == nil)
	assert(t, exceptionDescFromField == exDesc)

	// service and method
	serviceDesc := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetServiceDescriptor("MyService")
	assert(t, serviceDesc.GetName() == "MyService")
	assert(t, serviceDesc.GetMethods()[0].GetName() == "M1")
	assert(t, serviceDesc.GetMethods()[1].GetName() == "M2")

	// check struct default value
	fieldDefaultValue := pd.GetFieldByName("defaultValue")
	assert(t, fieldDefaultValue != nil)
	dv := fieldDefaultValue.GetDefaultValue()
	assert(t, dv != nil)
	assert(t, dv.GetValueString() == "123321")

	fieldConstVal := pd.GetFieldByName("defaultConst")
	assert(t, fieldConstVal != nil)
	cv := fieldConstVal.GetDefaultValue()
	assert(t, cv != nil)
	assert(t, cv.GetValueIdentifier() == "MY_CONST")
	assert(t, fd.GetConstDescriptor(cv.GetValueIdentifier()) == constDesc)
}

func TestLookup(t *testing.T) {
	// lookup struct
	pd := thrift_reflection_test.NewPerson().GetDescriptor()
	assert(t, pd == thrift_reflection.LookupStruct(pd.GetName(), ""))
	assert(t, pd == thrift_reflection.LookupStruct(pd.GetName(), pd.GetFilepath()))
	// lookup const
	cs := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetConstDescriptor("MY_CONST")
	assert(t, cs == thrift_reflection.LookupConst(cs.GetName(), ""))
	assert(t, cs == thrift_reflection.LookupConst(cs.GetName(), cs.GetFilepath()))
	// lookup enum
	en := thrift_reflection_test.Gender_FEMALE.GetDescriptor()
	assert(t, en == thrift_reflection.LookupEnum(en.GetName(), ""))
	assert(t, en == thrift_reflection.LookupEnum(en.GetName(), en.GetFilepath()))
	// lookup typedef
	td := thrift_reflection.GetTypedefDescriptorByGoType(&thrift_reflection_test.SpecialPerson{})
	assert(t, td == thrift_reflection.LookupTypedef(td.GetAlias(), ""))
	assert(t, td == thrift_reflection.LookupTypedef(td.GetAlias(), td.GetFilepath()))
	// lookup union
	un := thrift_reflection_test.NewMyUnion().GetDescriptor()
	assert(t, un == thrift_reflection.LookupUnion(un.GetName(), ""))
	assert(t, un == thrift_reflection.LookupUnion(un.GetName(), td.GetFilepath()))
	// lookup exception
	ex := thrift_reflection_test.NewMyException().GetDescriptor()
	assert(t, ex == thrift_reflection.LookupException(ex.GetName(), ""))
	assert(t, ex == thrift_reflection.LookupException(ex.GetName(), td.GetFilepath()))
	// lookup service
	svc := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetServiceDescriptor("MyService")
	assert(t, svc == thrift_reflection.LookupService(svc.GetName(), ""))
	assert(t, svc == thrift_reflection.LookupService(svc.GetName(), svc.GetFilepath()))
	// lookup method
	md := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetMethodDescriptor("MyService", "M1")
	assert(t, md == thrift_reflection.LookupMethod(md.GetName(), "", ""))
	assert(t, md == thrift_reflection.LookupMethod(md.GetName(), "", md.GetFilepath()))
	assert(t, md == thrift_reflection.LookupMethod(md.GetName(), svc.GetName(), ""))
	assert(t, md == thrift_reflection.LookupMethod(md.GetName(), svc.GetName(), md.GetFilepath()))
}

func TestLookupStruct(t *testing.T) {
	m3 := thrift_reflection_test.GetFileDescriptorForReflectionTestIdl().GetMethodDescriptor("MyService", "M3")
	structs, err := thrift_reflection.LookupIncludedStructsFromMethod(m3)
	assert(t, err == nil)
	assert(t, len(structs) == 12)

	// A1
	structs, err = thrift_reflection.LookupIncludedStructsFromType(m3.Response)
	assert(t, err == nil)
	assert(t, len(structs) == 2)

	// A0
	structs, err = thrift_reflection.LookupIncludedStructsFromType(m3.Args[0].GetType())
	assert(t, err == nil)
	assert(t, len(structs) == 9)

	// A3
	structs, err = thrift_reflection.LookupIncludedStructsFromType(m3.Args[1].GetType())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewA0().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 9)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewA1().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 2)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewA2().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewA3().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewB().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 4)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewB1().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewC().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 4)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewD().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 3)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewD1().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewD2().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewE().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 4)

	structs, err = thrift_reflection.LookupIncludedStructsFromStruct(thrift_reflection_test.NewF().GetDescriptor())
	assert(t, err == nil)
	assert(t, len(structs) == 1)
}

func TestReflection(t *testing.T) {
	p := thrift_reflection_test.NewPerson()
	p.Name = "Lee"

	nameDesc := p.GetDescriptor().GetFieldByName("name")
	assert(t, nameDesc != nil)

	// test get instance value by field descriptor reflection api
	val, err := nameDesc.GetInstanceValue(p)
	assert(t, err == nil)
	stringVal, ok := val.(string)
	assert(t, ok && stringVal == "Lee")

	// test set instance value by field descriptor reflection api
	err = nameDesc.SetInstanceValue(p, "Yun")
	assert(t, err == nil)
	assert(t, p.Name == "Yun")
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
