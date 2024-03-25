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
	"errors"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

type commonOption struct {
	name       string
	filepath   string
	optionType string
}

type OptionData struct {
	name           string
	typeDescriptor *thrift_reflection.TypeDescriptor
	mapVal         interface{}
	instanceVal    interface{}
}

func (o *OptionData) GetName() interface{} {
	return o.name
}

func (o *OptionData) GetValue() interface{} {
	return o.mapVal
}

func (o *OptionData) GetInstance() interface{} {
	return o.instanceVal
}

func (o *OptionData) GetFieldValue(name string) (interface{}, error) {
	if !o.typeDescriptor.IsStruct() {
		return nil, errors.New("not struct")
	}
	sd, err := o.typeDescriptor.GetStructDescriptor()
	if err != nil {
		return nil, errors.New("struct descriptor not found")
	}
	f := sd.GetFieldByName(name)
	if f == nil {
		return nil, errors.New("field name not match")
	}
	resultMap := o.mapVal.(map[string]interface{})
	return resultMap[name], nil
}

func (o *OptionData) IsFieldSet(name string) (bool, error) {
	if !o.typeDescriptor.IsStruct() {
		return false, errors.New("not struct")
	}
	sd, err := o.typeDescriptor.GetStructDescriptor()
	if err != nil {
		return false, errors.New("struct descriptor not found")
	}
	f := sd.GetFieldByName(name)
	if f == nil {
		return false, errors.New("field name not match")
	}
	_, ok := o.mapVal.(map[string]interface{})[name]
	return ok, nil
}

type AnnotationMeta struct {
	filepath    string
	annotations map[string][]string
}

type OptionGetter interface {
	GetName() string
	GetFilepath() string
	GetType() string
}

func (o *commonOption) GetName() string {
	return o.name
}

func (o *commonOption) GetFilepath() string {
	return o.filepath
}

func (o *commonOption) GetType() string {
	return o.optionType
}

func newOption(filepath, name, optionType string) *commonOption {
	return &commonOption{
		name:       name,
		filepath:   filepath,
		optionType: optionType,
	}
}

type EnumOption struct {
	*commonOption
}

func NewEnumOption(filepath, name string) *EnumOption {
	return &EnumOption{newOption(filepath, name, "_EnumOptions")}
}

type EnumValueOption struct {
	*commonOption
}

func NewEnumValueOption(filepath, name string) *EnumValueOption {
	return &EnumValueOption{newOption(filepath, name, "_EnumValueOptions")}
}

type MethodOption struct {
	*commonOption
}

func NewMethodOption(filepath, name string) *MethodOption {
	return &MethodOption{newOption(filepath, name, "_MethodOptions")}
}

type ServiceOption struct {
	*commonOption
}

func NewServiceOption(filepath, name string) *ServiceOption {
	return &ServiceOption{newOption(filepath, name, "_ServiceOptions")}
}

type StructOption struct {
	*commonOption
}

func NewStructOption(filepath, name string) *StructOption {
	return &StructOption{newOption(filepath, name, "_StructOptions")}
}

type FieldOption struct {
	*commonOption
}

func NewFieldOption(filepath, name string) *FieldOption {
	return &FieldOption{newOption(filepath, name, "_FieldOptions")}
}

func ParseFieldOption(field *thrift_reflection.FieldDescriptor, optionName string) (option *OptionData, err error) {
	return parseOptionFromKey(field, optionName, "_FieldOptions")
}

func ParseStructOption(structLike *thrift_reflection.StructDescriptor, annotationName string) (option *OptionData, err error) {
	return parseOptionFromKey(structLike, annotationName, "_StructOptions")
}

func ParseMethodOption(f *thrift_reflection.MethodDescriptor, optionName string) (option *OptionData, err error) {
	return parseOptionFromKey(f, optionName, "_MethodOptions")
}

func ParseServiceOption(s *thrift_reflection.ServiceDescriptor, optionName string) (option *OptionData, err error) {
	return parseOptionFromKey(s, optionName, "_ServiceOptions")
}

func ParseEnumOption(e *thrift_reflection.EnumDescriptor, optionName string) (option *OptionData, err error) {
	return parseOptionFromKey(e, optionName, "_EnumOptions")
}

func ParseEnumValueOption(ev *thrift_reflection.EnumValueDescriptor, optionName string) (option *OptionData, err error) {
	return parseOptionFromKey(ev, optionName, "_EnumValueOptions")
}

func GetFieldOption(s *thrift_reflection.FieldDescriptor, os *FieldOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}

func GetEnumOption(s *thrift_reflection.EnumDescriptor, os *EnumOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}

func GetEnumValueOption(s *thrift_reflection.EnumValueDescriptor, os *EnumValueOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}

func GetServiceOption(s *thrift_reflection.ServiceDescriptor, os *ServiceOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}

func GetMethodOption(s *thrift_reflection.MethodDescriptor, os *MethodOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}

func GetStructOption(s *thrift_reflection.StructDescriptor, os *StructOption) (val *OptionData, err error) {
	return parseOptionRuntime(s, os)
}
