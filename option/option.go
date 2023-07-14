package option

import (
	"errors"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/utils"
	"regexp"
	"strings"
	"sync"
)

var optionDefMap = map[string]bool{
	"_FieldOptions":     true,
	"_StructOptions":    true,
	"_MethodOptions":    true,
	"_ServiceOptions":   true,
	"_EnumOptions":      true,
	"_EnumValueOptions": true,
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

type OptionMap = map[string]*OptionData

var optionCache sync.Map

func parseOptionWithCache(data annotationData, ast *parser.Thrift, fromStruct string) (OptionMap, error) {
	omCache, ok := optionCache.Load(data)
	if ok {
		return omCache.(OptionMap), nil
	}
	om, err := parseOption(data, ast, fromStruct)
	if err == nil {
		optionCache.Store(data, om)
	}
	return om, err
}

func ParseFieldOption(field *parser.Field, ast *parser.Thrift) (optionMap OptionMap, err error) {
	return parseOptionWithCache(field, ast, "_FieldOptions")
}

func ParseStructOption(structLike *parser.StructLike, ast *parser.Thrift) (optionMap OptionMap, err error) {
	if optionDefMap[structLike.GetName()] {
		return nil, errors.New("it's not allowed to parse option from " + structLike.GetName())
	}
	return parseOptionWithCache(structLike, ast, "_StructOptions")
}

func ParseMethodOption(f *parser.Function, ast *parser.Thrift) (optionMap OptionMap, err error) {
	return parseOptionWithCache(f, ast, "_MethodOptions")
}

func ParseServiceOption(s *parser.Service, ast *parser.Thrift) (optionMap OptionMap, err error) {
	return parseOptionWithCache(s, ast, "_ServiceOptions")
}

func ParseEnumOption(e *parser.Enum, ast *parser.Thrift) (optionMap OptionMap, err error) {
	return parseOptionWithCache(e, ast, "_EnumOptions")
}

func ParseEnumValueOption(ev *parser.EnumValue, ast *parser.Thrift) (optionMap OptionMap, err error) {
	return parseOptionWithCache(ev, ast, "_EnumValueOptions")
}

type annotationData interface {
	GetAnnotations() (v parser.Annotations)
}

func parseOption(dataSource annotationData, ast *parser.Thrift, fromStruct string) (optionMap OptionMap, err error) {
	annotations := dataSource.GetAnnotations()
	options := utils.GetAnnotationsAsMap(annotations)["option"]
	if len(options) == 0 {
		return nil, nil
	}

	// 初始化 FileDescriptor 注册器
	fd := thrift_reflection.RegisterAST(ast)

	optionDataMap := map[string]*OptionData{}
	for _, opt := range options {
		name, content, ok := ParseOptionStr(opt)
		if !ok {
			return nil, errors.New("grammar error:\n" + opt)
		}
		fieldTypeDesc, er := findOptionType(name, fromStruct, fd)
		if er != nil {
			return nil, er
		}
		res, er := createInstance(fieldTypeDesc, content, true)
		if er != nil {
			return nil, errors.New("parse failed for " + name + ":" + er.Error())
		}
		optionDataMap[name] = &OptionData{name, fieldTypeDesc, res}
	}
	return optionDataMap, nil
}

func findOptionType(name string, fromStruct string, currentFd *thrift_reflection.FileDescriptor) (*thrift_reflection.TypeDescriptor, error) {
	prefix, tname := utils.ParseAlias(name)
	targetFd := currentFd.GetIncludeFD(prefix)
	optionStructDesc := targetFd.GetStructDescriptor(fromStruct)
	if optionStructDesc == nil {
		return nil, errors.New("no such struct found from given include IDLs:" + name)
	}
	fieldDesc := optionStructDesc.GetFieldByName(tname)
	if fieldDesc == nil {
		return nil, errors.New("no such option: " + name)
	}
	return fieldDesc.GetType(), nil
}

// ParseOptionStr parse an annotation to name and content
func ParseOptionStr(opt string) (name, content string, ok bool) {
	opt = strings.ReplaceAll(opt, "\n", " ")
	opt = strings.ReplaceAll(opt, "\t", " ")
	opt = strings.TrimSpace(opt)
	pattern := `(\w+(?:\.\w+)*)\s*=\s*(.*)$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(opt)
	if len(matches) == 3 {
		return matches[1], matches[2], true
	}
	return "", "", false
}
