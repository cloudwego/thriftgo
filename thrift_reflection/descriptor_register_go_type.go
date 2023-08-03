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

import "reflect"

var (
	structDes2goType  = map[*StructDescriptor]reflect.Type{}
	enumDes2goType    = map[*EnumDescriptor]reflect.Type{}
	typedefDes2goType = map[*TypedefDescriptor]reflect.Type{}

	goType2StructDes  = map[reflect.Type]*StructDescriptor{}
	goType2EnumDes    = map[reflect.Type]*EnumDescriptor{}
	goType2TypedefDes = map[reflect.Type]*TypedefDescriptor{}
)

func getReflect(in interface{}) reflect.Type {
	return reflect.TypeOf(in).Elem()
}

func registerGoTypes(fd *FileDescriptor, goTypes []interface{}) {
	structList := []*StructDescriptor{}
	structList = append(structList, fd.Structs...)
	structList = append(structList, fd.Unions...)
	structList = append(structList, fd.Exceptions...)
	for idx, s := range structList {
		registerStructGoType(s, getReflect(goTypes[idx]))
	}
	for idx, e := range fd.Enums {
		registerEnumGoType(e, getReflect(goTypes[len(structList)+idx]))
	}
	for idx, t := range fd.Typedefs {
		registerTypedefGoType(t, getReflect(goTypes[len(structList)+len(fd.Enums)+idx]))
	}
}

func registerStructGoType(s *StructDescriptor, t reflect.Type) {
	structDes2goType[s] = t
	goType2StructDes[t] = s
}

func registerEnumGoType(s *EnumDescriptor, t reflect.Type) {
	enumDes2goType[s] = t
	goType2EnumDes[t] = s
}

func registerTypedefGoType(s *TypedefDescriptor, t reflect.Type) {
	typedefDes2goType[s] = t
	goType2TypedefDes[t] = s
}

func GetStructDescriptorByGoType(in interface{}) *StructDescriptor {
	return goType2StructDes[getReflect(in)]
}

func GetEnumDescriptorByGoType(in interface{}) *EnumDescriptor {
	return goType2EnumDes[getReflect(in)]
}

func GetTypedefDescriptorByGoType(in interface{}) *TypedefDescriptor {
	return goType2TypedefDes[getReflect(in)]
}
