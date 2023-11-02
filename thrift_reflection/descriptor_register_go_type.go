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

func getReflect(in interface{}) reflect.Type {
	return reflect.TypeOf(in).Elem()
}

func (gd *GlobalDescriptor) registerGoTypes(fd *FileDescriptor, goTypes []interface{}) {
	structList := []*StructDescriptor{}
	structList = append(structList, fd.Structs...)
	structList = append(structList, fd.Unions...)
	structList = append(structList, fd.Exceptions...)
	for idx, s := range structList {
		gd.registerStructGoType(s, getReflect(goTypes[idx]))
	}
	for idx, e := range fd.Enums {
		gd.registerEnumGoType(e, getReflect(goTypes[len(structList)+idx]))
	}
	for idx, t := range fd.Typedefs {
		gd.registerTypedefGoType(t, getReflect(goTypes[len(structList)+len(fd.Enums)+idx]))
	}
}

func (gd *GlobalDescriptor) registerStructGoType(s *StructDescriptor, t reflect.Type) {
	gd.structDes2goType[s] = t
	gd.goType2StructDes[t] = s
}

func (gd *GlobalDescriptor) registerEnumGoType(s *EnumDescriptor, t reflect.Type) {
	gd.enumDes2goType[s] = t
	gd.goType2EnumDes[t] = s
}

func (gd *GlobalDescriptor) registerTypedefGoType(s *TypedefDescriptor, t reflect.Type) {
	gd.typedefDes2goType[s] = t
	gd.goType2TypedefDes[t] = s
}

func (gd *GlobalDescriptor) GetStructDescriptorByGoType(in interface{}) *StructDescriptor {
	if gd.goType2StructDes == nil {
		return nil
	}
	return gd.goType2StructDes[getReflect(in)]
}

func (gd *GlobalDescriptor) GetEnumDescriptorByGoType(in interface{}) *EnumDescriptor {
	if gd.goType2EnumDes == nil {
		return nil
	}
	return gd.goType2EnumDes[getReflect(in)]
}

func (gd *GlobalDescriptor) GetTypedefDescriptorByGoType(in interface{}) *TypedefDescriptor {
	if gd.goType2TypedefDes == nil {
		return nil
	}
	return gd.goType2TypedefDes[getReflect(in)]
}

func GetStructDescriptorByGoType(in interface{}) *StructDescriptor {
	return defaultGlobalDescriptor.GetStructDescriptorByGoType(in)
}

func GetEnumDescriptorByGoType(in interface{}) *EnumDescriptor {
	return defaultGlobalDescriptor.GetEnumDescriptorByGoType(in)
}

func GetTypedefDescriptorByGoType(in interface{}) *TypedefDescriptor {
	return defaultGlobalDescriptor.GetTypedefDescriptorByGoType(in)
}
