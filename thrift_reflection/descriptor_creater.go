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
	"strings"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/utils"
)

func GetFileDescriptor(ast *parser.Thrift) *FileDescriptor {
	services := []*ServiceDescriptor{}

	for _, s := range ast.Services {
		services = append(services, getServiceDescriptor(ast, ast.Filename, s))
	}

	structs := []*StructDescriptor{}
	for _, st := range ast.Structs {
		structs = append(structs, getStructDescriptor(ast, ast.Filename, st))
	}

	exceptions := []*StructDescriptor{}
	for _, ex := range ast.Exceptions {
		exceptions = append(exceptions, getStructDescriptor(ast, ast.Filename, ex))
	}

	unions := []*StructDescriptor{}
	for _, un := range ast.Unions {
		unions = append(unions, getStructDescriptor(ast, ast.Filename, un))
	}

	enums := []*EnumDescriptor{}
	for _, enm := range ast.Enums {
		enums = append(enums, getEnumDescriptor(ast, ast.Filename, enm))
	}

	typedefs := []*TypedefDescriptor{}
	for _, td := range ast.Typedefs {
		typedefs = append(typedefs, getTypedefDescriptor(ast, ast.Filename, td))
	}

	includesMap := map[string]string{}
	for _, inc := range ast.Includes {
		path := inc.GetReference().Filename
		arr := strings.Split(path, "/")
		alias := strings.TrimSuffix(arr[len(arr)-1], ".thrift")
		includesMap[alias] = path
	}

	namespaceMap := map[string]string{}
	for _, ns := range ast.Namespaces {
		namespaceMap[ns.GetLanguage()] = ns.GetName()
	}

	consts := []*ConstDescriptor{}
	for _, c := range ast.Constants {
		consts = append(consts, getConstDescriptor(ast.Filename, c))
	}

	return &FileDescriptor{
		Filepath:   ast.Filename,
		Includes:   includesMap,
		Namespaces: namespaceMap,
		Services:   services,
		Structs:    structs,
		Exceptions: exceptions,
		Typedefs:   typedefs,
		Enums:      enums,
		Unions:     unions,
		Consts:     consts,
	}
}

func getConstDescriptor(path string, c *parser.Constant) *ConstDescriptor {
	return &ConstDescriptor{
		Filepath:    path,
		Name:        c.Name,
		Type:        GetTypeDescriptor(path, c.Type),
		Value:       getConstValueDescriptor(c.Value),
		Annotations: utils.GetAnnotationsAsMap(c.GetAnnotations()),
		Comments:    c.ReservedComments,
		Extra:       nil,
	}
}

func getConstValueDescriptor(cv *parser.ConstValue) *ConstValueDescriptor {
	if cv == nil {
		return nil
	}
	valueType := cv.GetType()
	if valueType == parser.ConstType_ConstInt {
		return &ConstValueDescriptor{
			Type:     ConstValueType_INT,
			ValueInt: cv.GetTypedValue().GetInt(),
		}
	}
	if valueType == parser.ConstType_ConstDouble {
		return &ConstValueDescriptor{
			Type:        ConstValueType_DOUBLE,
			ValueDouble: cv.GetTypedValue().GetDouble(),
		}
	}
	if valueType == parser.ConstType_ConstLiteral {
		return &ConstValueDescriptor{
			Type:        ConstValueType_STRING,
			ValueString: cv.GetTypedValue().GetLiteral(),
		}
	}
	if valueType == parser.ConstType_ConstIdentifier {
		identifier := cv.GetTypedValue().GetIdentifier()
		if identifier == "false" {
			return &ConstValueDescriptor{
				Type:      ConstValueType_BOOL,
				ValueBool: false,
			}
		}
		if identifier == "true" {
			return &ConstValueDescriptor{
				Type:      ConstValueType_BOOL,
				ValueBool: true,
			}
		}
		// lookup const
		return &ConstValueDescriptor{
			Type:            ConstValueType_IDENTIFIER,
			ValueIdentifier: identifier,
		}
	}
	if valueType == parser.ConstType_ConstList {
		vals := []*ConstValueDescriptor{}
		for _, listVal := range cv.GetTypedValue().GetList() {
			vals = append(vals, getConstValueDescriptor(listVal))
		}
		return &ConstValueDescriptor{
			Type:      ConstValueType_LIST,
			ValueList: vals,
		}
	}
	if valueType == parser.ConstType_ConstMap {
		vals := map[*ConstValueDescriptor]*ConstValueDescriptor{}
		for _, mapVal := range cv.GetTypedValue().GetMap() {
			vals[getConstValueDescriptor(mapVal.GetKey())] = getConstValueDescriptor(mapVal.GetValue())
		}
		return &ConstValueDescriptor{
			Type:     ConstValueType_MAP,
			ValueMap: vals,
		}
	}

	return NewConstValueDescriptor()
}

func getMethodDescriptor(ast *parser.Thrift, path string, method *parser.Function) *MethodDescriptor {
	args := []*FieldDescriptor{}
	for _, arg := range method.Arguments {
		args = append(args, getFieldDescriptor(ast, path, arg))
	}

	throws := []*FieldDescriptor{}
	for _, t := range method.Throws {
		throws = append(throws, getFieldDescriptor(ast, path, t))
	}

	return &MethodDescriptor{
		Filepath:        path,
		Name:            method.GetName(),
		Response:        GetTypeDescriptor(path, method.FunctionType),
		Args:            args,
		Annotations:     utils.GetAnnotationsAsMap(method.GetAnnotations()),
		Comments:        method.GetReservedComments(),
		Extra:           nil,
		ThrowExceptions: throws,
		IsOneway:        method.Oneway,
	}
}

func getServiceDescriptor(ast *parser.Thrift, path string, service *parser.Service) *ServiceDescriptor {
	methods := []*MethodDescriptor{}
	for _, method := range service.Functions {
		methods = append(methods, getMethodDescriptor(ast, path, method))
	}

	return &ServiceDescriptor{
		Filepath:    path,
		Name:        service.GetName(),
		Methods:     methods,
		Annotations: utils.GetAnnotationsAsMap(service.GetAnnotations()),
		Comments:    service.GetReservedComments(),
		Extra:       nil,
	}
}

func GetTypeDescriptor(path string, typeStruct *parser.Type) *TypeDescriptor {
	if typeStruct == nil {
		return nil
	}

	return &TypeDescriptor{
		Filepath:  path,
		Name:      typeStruct.Name,
		KeyType:   GetTypeDescriptor(path, typeStruct.KeyType),
		ValueType: GetTypeDescriptor(path, typeStruct.ValueType),
	}
}

func getStructDescriptor(ast *parser.Thrift, path string, structLike *parser.StructLike) *StructDescriptor {
	fields := []*FieldDescriptor{}
	for _, f := range structLike.GetFields() {
		fields = append(fields, getFieldDescriptor(ast, path, f))
	}
	return &StructDescriptor{
		Filepath:    path,
		Name:        structLike.GetName(),
		Fields:      fields,
		Annotations: utils.GetAnnotationsAsMap(structLike.GetAnnotations()),
		Comments:    structLike.GetReservedComments(),
		Extra:       nil,
	}
}

func getTypedefDescriptor(ast *parser.Thrift, path string, td *parser.Typedef) *TypedefDescriptor {
	return &TypedefDescriptor{
		Filepath:    path,
		Type:        GetTypeDescriptor(path, td.GetType()),
		Alias:       td.GetAlias(),
		Annotations: utils.GetAnnotationsAsMap(td.GetAnnotations()),
		Comments:    td.GetReservedComments(),
		Extra:       nil,
	}
}

func getEnumDescriptor(ast *parser.Thrift, path string, enum *parser.Enum) *EnumDescriptor {
	values := []*EnumValueDescriptor{}
	for _, ev := range enum.Values {
		values = append(values, &EnumValueDescriptor{
			Filepath:    path,
			Name:        ev.GetName(),
			Value:       ev.GetValue(),
			Annotations: utils.GetAnnotationsAsMap(ev.GetAnnotations()),
			Comments:    ev.ReservedComments,
			Extra:       nil,
		})
	}

	return &EnumDescriptor{
		Filepath:    path,
		Name:        enum.GetName(),
		Values:      values,
		Annotations: utils.GetAnnotationsAsMap(enum.GetAnnotations()),
		Comments:    enum.GetReservedComments(),
		Extra:       nil,
	}
}

func getFieldDescriptor(ast *parser.Thrift, path string, field *parser.Field) *FieldDescriptor {
	return &FieldDescriptor{
		Filepath:     path,
		Name:         field.GetName(),
		Type:         GetTypeDescriptor(path, field.GetType()),
		Requiredness: field.GetRequiredness().String(),
		ID:           field.GetID(),
		DefaultValue: getConstValueDescriptor(field.GetDefault()),
		Annotations:  utils.GetAnnotationsAsMap(field.Annotations),
		Comments:     field.GetReservedComments(),
	}
}
