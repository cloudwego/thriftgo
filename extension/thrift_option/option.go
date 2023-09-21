package thrift_option

import (
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/utils"
	"strings"
)

var optionDefMap = map[string]bool{
	"_FieldOptions":     true,
	"_StructOptions":    true,
	"_MethodOptions":    true,
	"_ServiceOptions":   true,
	"_EnumOptions":      true,
	"_EnumValueOptions": true,
}

func parseOptionFromAST(in interface{ GetAnnotations() parser.Annotations }, ast *parser.Thrift, annotationName string, optionType string) (option *OptionData, err error) {
	// init FileDescriptor to enable reflection apis
	fd := thrift_reflection.RegisterAST(ast)
	prefix, optionName := utils.ParseAlias(annotationName)
	optionFD := fd.GetIncludeFD(prefix)
	filepath := ""
	if optionFD != nil {
		filepath = optionFD.GetFilepath()
	}
	optionMeta := &commonOption{
		name:       optionName,
		filepath:   filepath,
		optionType: optionType,
	}

	return parseOption(fd.GetFilepath(), utils.GetAnnotationsAsMap(in.GetAnnotations()), optionMeta, true)
}

func parseOptionRuntime(in interface {
	GetAnnotations() map[string][]string
	GetFilepath() string
}, optionMeta OptionGetter) (val *OptionData, err error) {

	optionData, err := parseOption(in.GetFilepath(), in.GetAnnotations(), optionMeta, false)
	if err != nil {
		return nil, err
	}
	return optionData, nil

}

// todo 添加缓存
func parseOption(filepath string, annotations map[string][]string, optionMeta OptionGetter, mapMode bool) (option *OptionData, err error) {

	optionName := optionMeta.GetName()
	if optionMeta.GetType() == "_StructOptions" && optionDefMap[optionName] {
		return nil, NotAllowError(optionName)
	}

	if optionMeta.GetFilepath() == "" {
		return nil, NotIncludedError(optionName)
	}

	// todo 参数生成支持

	anno := &AnnotationMeta{
		filepath:    filepath,
		annotations: annotations,
	}

	_, opt, ok := getOptionContent(anno, optionMeta)
	if !ok {
		// 传入的 Option 并不能在 Annotation 里找到
		return nil, KeyNotMatchError(optionMeta.GetName())
	}

	optionFilepath := optionMeta.GetFilepath()
	optionType := optionMeta.GetType()

	fieldDesc := thrift_reflection.LookupFD(optionFilepath).GetStructDescriptor(optionType).GetFieldByName(optionName)
	if fieldDesc == nil || fieldDesc.GetType() == nil {
		// 传入的 Option 和 Annotation 能匹配到，但并没有实际对应真正的 Option
		return nil, NotExistError(optionName)

	}

	typeDesc := fieldDesc.GetType()

	// format option content
	opt = strings.ReplaceAll(opt, "\n", " ")
	opt = strings.ReplaceAll(opt, "\t", " ")
	opt = strings.TrimSpace(opt)

	mapVal, instanceVal, er := createInstance(typeDesc, opt, mapMode)
	if er != nil {
		return nil, ParseFailedError(optionName, er)
	}
	return &OptionData{optionName, typeDesc, mapVal, instanceVal}, nil

}

func getOptionContent(annotation *AnnotationMeta, optionMeta OptionGetter) (string, string, bool) {
	for annotationKey, values := range annotation.annotations {
		if len(values) > 0 {
			prefix, expectedOptionName := utils.ParseAlias(annotationKey)
			// check name
			if expectedOptionName == optionMeta.GetName() {

				match := false
				if prefix == "" {
					// option and current struct are in the same idl
					// double check their idl filepath
					match = optionMeta.GetFilepath() == annotation.filepath
				} else {
					// option and current struct are not in the same idl
					currentFD := thrift_reflection.LookupFD(annotation.filepath)
					// check if current struct idl include the option's idl
					match = optionMeta.GetFilepath() == currentFD.Includes[prefix]
				}

				if match {
					return annotationKey, values[len(values)-1], true
				}
			}
		}
	}
	return "", "", false
}
