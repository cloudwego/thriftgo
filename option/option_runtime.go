package option

import (
	"errors"
	"github.com/cloudwego/thriftgo/parser"
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
	value          interface{}
}

func (o *OptionData) GetName() interface{} {
	return o.name
}

func (o *OptionData) GetValue() interface{} {
	return o.value
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
	resultMap := o.value.(map[string]interface{})
	return resultMap[name], nil
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

func ParseFieldOption(field *parser.Field, optionName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(field, ast, optionName, "_FieldOptions")
}

func ParseStructOption(structLike *parser.StructLike, annotationName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(structLike, ast, annotationName, "_StructOptions")
}

func ParseMethodOption(f *parser.Function, optionName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(f, ast, optionName, "_MethodOptions")
}

func ParseServiceOption(s *parser.Service, optionName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(s, ast, optionName, "_ServiceOptions")
}

func ParseEnumOption(e *parser.Enum, optionName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(e, ast, optionName, "_EnumOptions")
}

func ParseEnumValueOption(ev *parser.EnumValue, optionName string, ast *parser.Thrift) (option *OptionData, err error) {
	return parseCommonOption(ev, ast, optionName, "_EnumValueOptions")
}

func GetFieldOption(s *thrift_reflection.FieldDescriptor, os *FieldOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}

func GetEnumOption(s *thrift_reflection.EnumDescriptor, os *EnumOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}

func GetEnumValueOption(s *thrift_reflection.EnumValueDescriptor, os *EnumValueOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}
func GetServiceOption(s *thrift_reflection.ServiceDescriptor, os *ServiceOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}
func GetMethodOption(s *thrift_reflection.MethodDescriptor, os *MethodOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}

func GetStructOption(s *thrift_reflection.StructDescriptor, os *StructOption) (val interface{}, err error) {
	return parseOptionRuntime(s, os)
}
