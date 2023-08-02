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
	"fmt"
	"reflect"

	"github.com/cloudwego/thriftgo/parser"
)

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

func RegisterStructGoType(s *StructDescriptor, t reflect.Type) {
	structDes2goType[s] = t
	goType2StructDes[t] = s
}

func RegisterEnumGoType(s *EnumDescriptor, t reflect.Type) {
	enumDes2goType[s] = t
	goType2EnumDes[t] = s
}

func RegisterTypedefGoType(s *TypedefDescriptor, t reflect.Type) {
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

var globalFD = map[string]*FileDescriptor{}

func LookupFD(filepath string) *FileDescriptor {
	return globalFD[filepath]
}

func LookupEnum(name, filepath string) *EnumDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetEnumDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetEnumDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupConst(name, filepath string) *ConstDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetConstDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetConstDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupTypedef(alias, filepath string) *TypedefDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetTypedefDescriptor(alias)
	}
	for _, fd := range globalFD {
		val := fd.GetTypedefDescriptor(alias)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupStruct(name, filepath string) *StructDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetStructDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetStructDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupUnion(name, filepath string) *StructDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetUnionDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetUnionDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupException(name, filepath string) *StructDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetExceptionDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetExceptionDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupService(name, filepath string) *ServiceDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetServiceDescriptor(name)
	}
	for _, fd := range globalFD {
		val := fd.GetServiceDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupMethod(method, service, filepath string) *MethodDescriptor {
	if filepath != "" {
		return LookupFD(filepath).GetMethodDescriptor(service, method)
	}
	for _, fd := range globalFD {
		val := fd.GetMethodDescriptor(service, method)
		if val != nil {
			return val
		}
	}
	return nil
}

func LookupIncludedStructsFromMethod(method *MethodDescriptor) ([]*StructDescriptor, error) {
	structMap := map[*StructDescriptor]bool{}
	typeArr := []*TypeDescriptor{}
	typeArr = append(typeArr, method.GetResponse())
	for _, arg := range method.GetArgs() {
		typeArr = append(typeArr, arg.GetType())
	}
	for _, typeDesc := range typeArr {
		err := lookupIncludedStructsFromType(typeDesc, structMap)
		if err != nil {
			return nil, err
		}
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

// LookupIncludedStructsFromStruct finds all struct descriptor included by this structDescriptor (and current struct descriptor is also included in the return result)
func LookupIncludedStructsFromStruct(sd *StructDescriptor) ([]*StructDescriptor, error) {
	structMap := map[*StructDescriptor]bool{}
	err := lookupIncludedStructsFromStruct(sd, structMap)
	if err != nil {
		return nil, err
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

func LookupIncludedStructsFromType(td *TypeDescriptor) ([]*StructDescriptor, error) {
	structMap := map[*StructDescriptor]bool{}
	err := lookupIncludedStructsFromType(td, structMap)
	if err != nil {
		return nil, err
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

func lookupIncludedStructsFromType(typeDesc *TypeDescriptor, structMap map[*StructDescriptor]bool) error {
	if typeDesc.IsStruct() {
		stDesc, err := typeDesc.GetStructDescriptor()
		if err != nil {
			return err
		}
		return lookupIncludedStructsFromStruct(stDesc, structMap)
	}
	if typeDesc.IsContainer() {
		err := lookupIncludedStructsFromType(typeDesc.GetValueType(), structMap)
		if err != nil {
			return err
		}
		if typeDesc.IsMap() {
			er := lookupIncludedStructsFromType(typeDesc.GetKeyType(), structMap)
			if er != nil {
				return er
			}
		}
	}
	return nil
}

func lookupIncludedStructsFromStruct(sd *StructDescriptor, structMap map[*StructDescriptor]bool) error {
	if structMap[sd] {
		return nil
	}
	structMap[sd] = true
	for _, f := range sd.GetFields() {
		typeDesc := f.GetType()
		err := lookupIncludedStructsFromType(typeDesc, structMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func BuildFileDescriptor(bytes []byte, goTypes []interface{}) *FileDescriptor {
	fd, err := RegisterIDL(bytes)
	fmt.Println(fd.Filepath)
	if err != nil {
		panic(err)
	}
	structList := []*StructDescriptor{}
	structList = append(structList, fd.Structs...)
	structList = append(structList, fd.Unions...)
	structList = append(structList, fd.Exceptions...)
	for idx, s := range structList {
		RegisterStructGoType(s, getReflect(goTypes[idx]))
	}
	for idx, e := range fd.Enums {
		RegisterEnumGoType(e, getReflect(goTypes[len(structList)+idx]))
	}
	for idx, t := range fd.Typedefs {
		RegisterTypedefGoType(t, getReflect(goTypes[len(structList)+len(fd.Enums)+idx]))
	}
	return fd
}

func RegisterIDL(bytes []byte) (*FileDescriptor, error) {
	fd, err := Unmarshal(bytes)
	if err != nil {
		return nil, err
	}
	RegisterFd(fd)
	return fd, err
}

func RegisterFd(fd *FileDescriptor) {
	globalFD[fd.Filepath] = fd
}

func RegisterAST(ast *parser.Thrift) *FileDescriptor {
	fd, ok := globalFD[ast.Filename]
	if ok {
		return fd
	}
	fd = GetFileDescriptor(ast)
	RegisterFd(fd)
	for _, inc := range ast.Includes {
		RegisterAST(inc.GetReference())
	}
	return fd
}
