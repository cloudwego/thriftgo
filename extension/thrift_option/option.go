package thrift_option

import (
	"strings"

	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/utils"
)

var optionDefMap = map[string]bool{
	"_FieldOptions":     true,
	"_StructOptions":    true,
	"_MethodOptions":    true,
	"_ServiceOptions":   true,
	"_EnumOptions":      true,
	"_EnumValueOptions": true,
}

func parseOptionFromKey(in interface {
	GetAnnotations() map[string][]string
	GetFilepath() string
	GetExtra() map[string]string
}, annotationName, optionType string) (option *OptionData, err error) {
	// init FileDescriptor to enable reflection apis

	gd := thrift_reflection.GetGlobalDescriptor(in)
	prefix, optionName := utils.ParseAlias(annotationName)
	optionMeta := &commonOption{
		name:       optionName,
		optionType: optionType,
	}
	optionFD := gd.LookupFD(in.GetFilepath()).GetIncludeFD(prefix)
	if optionFD != nil {
		optionMeta.filepath = optionFD.GetFilepath()
	}

	anno := &AnnotationMeta{
		filepath:    in.GetFilepath(),
		annotations: in.GetAnnotations(),
	}
	return parseOption(gd, anno, optionMeta, true)
}

func parseOptionRuntime(in interface {
	GetAnnotations() map[string][]string
	GetFilepath() string
	GetExtra() map[string]string
}, optionMeta OptionGetter) (val *OptionData, err error,
) {
	anno := &AnnotationMeta{
		filepath:    in.GetFilepath(),
		annotations: in.GetAnnotations(),
	}
	gd := thrift_reflection.GetGlobalDescriptor(in)
	optionData, err := parseOption(gd, anno, optionMeta, false)
	if err != nil {
		return nil, err
	}
	return optionData, nil
}

// todo 添加缓存
func parseOption(gd *thrift_reflection.GlobalDescriptor, anno *AnnotationMeta, optionMeta OptionGetter, mapMode bool) (option *OptionData, err error) {
	optionName := optionMeta.GetName()
	if optionMeta.GetType() == "_StructOptions" && optionDefMap[optionName] {
		return nil, NotAllowError(optionName)
	}

	if optionMeta.GetFilepath() == "" {
		return nil, NotIncludedError(optionName)
	}

	_, opt, ok := getOptionContent(gd, anno, optionMeta)
	if !ok {
		// 传入的 Option 并不能在 Annotation 里找到
		return nil, KeyNotMatchError(optionMeta.GetName())
	}

	optionFilepath := optionMeta.GetFilepath()
	optionType := optionMeta.GetType()

	fieldDesc := gd.LookupFD(optionFilepath).GetStructDescriptor(optionType).GetFieldByName(optionName)
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

func getOptionContent(gd *thrift_reflection.GlobalDescriptor, annotation *AnnotationMeta, optionMeta OptionGetter) (string, string, bool) {
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
					currentFD := gd.LookupFD(annotation.filepath)
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
