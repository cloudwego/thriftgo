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

	content := getOptionContent(gd, anno, optionMeta)
	if content == "" {
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

	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\t", " ")
	content = strings.TrimSpace(content)
	mapVal, instanceVal, er := createInstance(typeDesc, content, mapMode)
	if er != nil {
		return nil, ParseFailedError(optionName, er)
	}

	return &OptionData{optionName, typeDesc, mapVal, instanceVal}, nil
}

type subValue struct {
	path  string
	value string
}

func getOptionContent(gd *thrift_reflection.GlobalDescriptor, annotation *AnnotationMeta, optionMeta OptionGetter) string {

	opts := []*subValue{}
	for annotationKey, vals := range annotation.annotations {
		if len(vals) < 1 {
			continue
		}
		value := vals[len(vals)-1]
		prefix, expectedOptionName, subpath, ok := parseOptionName(annotationKey)
		if !ok {
			continue
		}
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
				if subpath == "" {
					return value
				}
				opts = append(opts, &subValue{
					path:  subpath,
					value: value,
				})
			}
		}
	}

	if len(opts) == 0 {
		return ""
	}
	// todo 一个 prefix 的时候 panic
	return buildTree(opts)
}

func parseOptionName(tname string) (prefix, name, subpath string, ok bool) {
	arr := strings.Split(tname, ".")
	if len(arr) == 1 {
		return "", tname, "", false
	}
	if len(arr) == 2 {
		return arr[0], arr[1], "", true
	}
	return arr[0], arr[1], strings.Join(arr[2:], "."), true
}

func buildTree(opts []*subValue) string {
	tree := make(map[string]interface{})

	for _, opt := range opts {
		path := strings.Split(opt.path, ".")
		value := opt.value

		current := tree
		for i, comp := range path {
			if i == len(path)-1 {
				current[comp] = value
			} else {
				if _, ok := current[comp]; !ok {
					current[comp] = make(map[string]interface{})
				}
				if _, ok := current[comp].(map[string]interface{}); ok {
					current = current[comp].(map[string]interface{})
				} else {
					current[comp] = make(map[string]interface{})
					current = current[comp].(map[string]interface{})
				}
			}
		}
	}

	return formatTree(tree)
}

func formatTree(tree map[string]interface{}) string {
	output := "{"
	for key, value := range tree {
		output += key + ":"

		switch v := value.(type) {
		case string:
			output += `"` + v + `"`
		case map[string]interface{}:
			output += formatTree(v)
		}

		output += ","
	}

	output = strings.TrimSuffix(output, ",")
	output += "}"

	return output
}
