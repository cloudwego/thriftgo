package option

import (
	"errors"
	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/utils"
	"sync"
)

var runtimeOptionCache sync.Map

func GetStructOption(s *thrift_reflection.StructDescriptor, os *StructOption) (val interface{}, err error) {
	if optionDefMap[s.GetName()] {
		return nil, errors.New("it's not allowed to parse option from " + s.GetName())
	}
	return getOptionWrapper(s, os)
}

func GetFieldOption(s *thrift_reflection.FieldDescriptor, os *FieldOption) (val interface{}, err error) {
	return getOptionWrapper(s, os)
}

func GetEnumOption(s *thrift_reflection.EnumDescriptor, os *EnumOption) (val interface{}, err error) {
	return getOptionWrapper(s, os)
}

func GetEnumValueOption(s *thrift_reflection.EnumValueDescriptor, os *EnumValueOption) (val interface{}, err error) {
	return getOptionWrapper(s, os)
}
func GetServiceOption(s *thrift_reflection.ServiceDescriptor, os *ServiceOption) (val interface{}, err error) {
	return getOptionWrapper(s, os)
}
func GetMethodOption(s *thrift_reflection.MethodDescriptor, os *MethodOption) (val interface{}, err error) {
	return getOptionWrapper(s, os)
}

func getOptionWrapper(s annotationDescriptor, os OptionDescriptor) (val interface{}, err error) {
	// check from cache
	cacheMap, ok := runtimeOptionCache.Load(s)
	if !ok {
		runtimeOptionCache.Store(s, map[OptionDescriptor]interface{}{})
	}
	cacheMap, ok = runtimeOptionCache.Load(s)
	if !ok {
		return getOption(s, os)
	}
	cache := cacheMap.(map[OptionDescriptor]interface{})
	if cacheVal, ok := cache[os]; ok {
		return cacheVal, nil
	}
	instance, err := getOption(s, os)
	if err == nil {
		cache[os] = instance
	}
	return instance, err
}

type annotationDescriptor interface {
	GetAnnotations() map[string][]string
	GetFilepath() string
}

func getOption(s annotationDescriptor, os OptionDescriptor) (val interface{}, err error) {
	options := s.GetAnnotations()["option"]
	for _, opt := range options {
		optionName, content, ok := ParseOptionStr(opt)
		if ok {
			if checkOptionName(optionName, s.GetFilepath(), os) {
				typeDesc := getOptionTypeDesc(os)
				if typeDesc != nil {
					instance, er := createInstance(typeDesc, content, false)
					if er != nil {
						return nil, errors.New("parse failed for " + optionName + ":" + er.Error())
					}
					return instance, nil
				} else {
					return nil, errors.New("failed to found struct type descriptor")
				}
			}
		} else {
			return nil, errors.New("grammar error:\n" + opt)
		}
	}
	return nil, errors.New("option not exist on current descriptor")
}

func checkOptionName(optionName string, currentPath string, os OptionDescriptor) bool {
	prefix, name := utils.ParseAlias(optionName)
	// check name
	if name == os.GetName() {
		if prefix == "" && os.GetFilepath() == currentPath {
			return true
		}
		if prefix != "" {
			currentFd := thrift_reflection.LookupFD(currentPath)
			if os.GetFilepath() == currentFd.Includes[prefix] {
				return true
			}
		}
	}
	return false
}

func getOptionTypeDesc(os OptionDescriptor) *thrift_reflection.TypeDescriptor {
	fd := thrift_reflection.LookupFD(os.GetFilepath())
	fieldDescriptor := fd.GetStructDescriptor(os.GetType()).GetFieldByName(os.GetName())
	if fieldDescriptor != nil {
		return fieldDescriptor.Type
	}
	return nil
}

/***** Option Runtime Descriptor Types ********/

type OptionDescriptor interface {
	GetName() string
	GetFilepath() string
	GetType() string
}

type OptionInfo struct {
	Name     string
	Filepath string
}

type StructOption struct {
	name     string
	filepath string
}

type EnumOption struct {
	name     string
	filepath string
}

func NewEnumOption(filepath, name string) *EnumOption {
	return &EnumOption{
		name:     name,
		filepath: filepath,
	}
}
func (o *EnumOption) GetName() string {
	return o.name
}
func (o *EnumOption) GetFilepath() string {
	return o.filepath
}
func (o *EnumOption) GetType() string {
	return "_EnumOptions"
}

type MethodOption struct {
	name     string
	filepath string
}

func NewMethodOption(filepath, name string) *MethodOption {
	return &MethodOption{
		name:     name,
		filepath: filepath,
	}
}
func (o *MethodOption) GetName() string {
	return o.name
}
func (o *MethodOption) GetFilepath() string {
	return o.filepath
}
func (o *MethodOption) GetType() string {
	return "_MethodOptions"
}

type ServiceOption struct {
	name     string
	filepath string
}

func NewServiceOption(filepath, name string) *ServiceOption {
	return &ServiceOption{
		name:     name,
		filepath: filepath,
	}
}
func (o *ServiceOption) GetName() string {
	return o.name
}
func (o *ServiceOption) GetFilepath() string {
	return o.filepath
}
func (o *ServiceOption) GetType() string {
	return "_ServiceOptions"
}

type EnumValueOption struct {
	name     string
	filepath string
}

func NewEnumValueOption(filepath, name string) *EnumValueOption {
	return &EnumValueOption{
		name:     name,
		filepath: filepath,
	}
}
func (o *EnumValueOption) GetName() string {
	return o.name
}
func (o *EnumValueOption) GetFilepath() string {
	return o.filepath
}
func (o *EnumValueOption) GetType() string {
	return "_EnumValueOptions"
}

func NewStructOption(filepath, name string) *StructOption {
	return &StructOption{
		name:     name,
		filepath: filepath,
	}
}
func (o *StructOption) GetName() string {
	return o.name
}
func (o *StructOption) GetFilepath() string {
	return o.filepath
}
func (o *StructOption) GetType() string {
	return "_StructOptions"
}

type FieldOption struct {
	name     string
	filepath string
}

func (o *FieldOption) GetName() string {
	return o.name
}
func (o *FieldOption) GetFilepath() string {
	return o.filepath
}
func (o *FieldOption) GetType() string {
	return "_FieldOptions"
}

func NewFieldOption(filepath, name string) *FieldOption {
	return &FieldOption{
		name:     name,
		filepath: filepath,
	}
}
