package parser

import (
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func CheckOptionGrammar(ast *Thrift) error {
	for _, s := range ast.Structs {
		if optionDefMap[s.Name] {
			continue
		}
		_, err := ParseStructOption(s, ast)
		if err != nil {
			return errors.New("Option Check:" + s.Name + " failed:" + err.Error())
		}
		for _, f := range s.Fields {
			_, er := ParseFieldOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}
	}
	for _, s := range ast.Services {
		_, err := ParseServiceOption(s, ast)
		if err != nil {
			return errors.New("Option Check:" + s.Name + " failed:" + err.Error())
		}
		for _, f := range s.Functions {
			_, er := ParseMethodOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}
	}
	for _, en := range ast.Enums {
		_, err := ParseEnumOption(en, ast)
		if err != nil {
			return errors.New("Option Check:" + en.Name + " failed:" + err.Error())
		}
		for _, f := range en.Values {
			_, er := ParseEnumValueOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}
	}
	return nil
}

var optionDefMap = map[string]bool{
	"_FieldOptions":     true,
	"_StructOptions":    true,
	"_MethodOptions":    true,
	"_ServiceOptions":   true,
	"_EnumOptions":      true,
	"_EnumValueOptions": true,
}

var basicMap = map[string]bool{
	"i8":     true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"double": true,
	"string": true,
	"byte":   true,
	"binary": true,
	"bool":   true,
}

var containerMap = map[string]bool{
	"set":  true,
	"list": true,
	"map":  true,
}

func ParseFieldOption(field *Field, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	return parseOption(field.GetAnnotations(), ast, "_FieldOptions")
}

func ParseStructOption(structLike *StructLike, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	if optionDefMap[structLike.Name] {
		return nil, errors.New("Getting Option from " + structLike.Name + " is not allowed.")
	}
	return parseOption(structLike.GetAnnotations(), ast, "_StructOptions")
}

func ParseMethodOption(f *Function, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	return parseOption(f.GetAnnotations(), ast, "_MethodOptions")
}

func ParseServiceOption(s *Service, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	return parseOption(s.GetAnnotations(), ast, "_ServiceOptions")
}

func ParseEnumOption(e *Enum, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	return parseOption(e.GetAnnotations(), ast, "_EnumOptions")
}

func ParseEnumValueOption(ev *EnumValue, ast *Thrift) (optionMap map[string]*OptionData, err error) {
	return parseOption(ev.GetAnnotations(), ast, "_EnumValueOptions")
}

type structInfo struct {
	isStruct   bool
	name       string
	structLike *StructLike
	fromAst    *Thrift
}

type optionDef struct {
	isBasic       bool
	isContainer   bool
	isStruct      bool
	name          string
	structLike    *StructLike
	fromAst       *Thrift
	structType    *Type
	structTypeAst *Thrift
}

func parseOption(annotations Annotations, ast *Thrift, fromStruct string) (optionMap map[string]*OptionData, err error) {
	options := []string{}
	for _, an := range annotations {
		if an.GetKey() == "option" {
			options = append(options, an.GetValues()...)
		}
	}
	if len(options) == 0 {
		return nil, nil
	}
	optionDataMap := map[string]*OptionData{}
	for _, opt := range options {
		name, content, ok := parseOptionStr(opt)
		if !ok {
			return nil, errors.New("grammar error:\n" + opt)
		}
		od, ok := getOptionType(name, fromStruct, ast)
		if !ok {
			return nil, errors.New("no option def for " + name)
		}
		res, er := fillContent(od, content)
		if er != nil {
			return nil, errors.New("parse failed for " + name + ":" + er.Error())
		}
		optionDataMap[name] = &OptionData{name, od.isStruct,
			res, od.structLike,
		}
	}
	return optionDataMap, nil
}

// parseOptionStr parse an annotation to name and content
func parseOptionStr(opt string) (name, content string, ok bool) {
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

// getOptionType find the real type's structLike
func getOptionType(name string, fromStruct string, ast *Thrift) (*optionDef, bool) {
	prefix, simpleName := parsePrefixAndName(name)
	if prefix != "" {
		prefix += "."
	}
	so, err := getStruct(prefix+fromStruct, ast)
	if err != nil {
		return nil, false
	}
	f := getField(simpleName, so.structLike)
	if f == nil {
		return nil, false
	}
	// search struct
	sl, err := getStruct(f.Type.GetName(), so.fromAst)
	if err != nil {
		return nil, false
	}
	return newOptionDef(sl, f.GetType(), so.fromAst), true
}

func newOptionDef(sl *structInfo, structType *Type, structTypeAst *Thrift) *optionDef {
	isBasic := basicMap[sl.name]
	isContainer := containerMap[sl.name]
	return &optionDef{
		isBasic:       isBasic,
		isContainer:   isContainer,
		isStruct:      sl.isStruct,
		name:          sl.name,
		structLike:    sl.structLike,
		fromAst:       sl.fromAst,
		structType:    structType,
		structTypeAst: structTypeAst,
	}
}

// fillContent fill a string content into the given structLike. If it's a basic type, then return interface{}. if it's a struct, return map[string]interface{}, else return container type
func fillContent(od *optionDef, content string) (interface{}, error) {
	if od.isBasic {
		return convertBasic(od, content)
	}
	if od.isContainer {
		return convertContainer(od, content)
	}
	if od.isStruct {
		return convertStruct(od, content)
	}
	return nil, errors.New("unknown option def")

}

func convertBasic(od *optionDef, value string) (interface{}, error) {
	switch od.name {
	case "bool":
		return strconv.ParseBool(value)
	case "byte":
		i, err := strconv.ParseInt(value, 10, 8)
		return int8(i), err
	case "i8":
		i, err := strconv.ParseInt(value, 10, 8)
		return int8(i), err
	case "i16":
		i, err := strconv.ParseInt(value, 10, 16)
		return int16(i), err
	case "i32":
		i, err := strconv.ParseInt(value, 10, 32)
		return int32(i), err
	case "i64":
		return strconv.ParseInt(value, 10, 64)
	case "double":
		return strconv.ParseFloat(value, 64)
	case "binary":
		return hex.DecodeString(value)
	case "string":
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			return value[1 : len(value)-1], nil
		}
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			return value[1 : len(value)-1], nil
		}
		return value, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

func convertContainer(od *optionDef, value string) (interface{}, error) {
	tname := od.name
	if tname == "list" || tname == "set" {
		arr, err := parseArr(value)
		if err != nil {
			return nil, errors.New(err.Error() + " when parse " + tname)
		}
		valueSli, err := getStruct(od.structType.GetValueType().GetName(), od.structTypeAst)
		if err != nil {
			return nil, errors.New(err.Error() + " when parse " + tname)
		}
		odv := newOptionDef(valueSli, od.structType.GetValueType(), od.structTypeAst)

		results := []interface{}{}
		for _, content := range arr {
			res, er := fillContent(odv, content)
			if er != nil {
				return nil, er
			}
			results = append(results, res)
		}
		return results, nil

	}
	if tname == "map" {
		kvMap, err := parseKV(value)
		if err != nil {
			return nil, errors.New(err.Error() + " when parse " + tname)
		}
		keySli, err := getStruct(od.structType.GetKeyType().GetName(), od.structTypeAst)
		if err != nil {
			return nil, errors.New(err.Error() + " when parse " + tname)
		}
		odk := newOptionDef(keySli, od.structType.GetValueType(), od.structTypeAst)

		valueSli, err := getStruct(od.structType.GetValueType().GetName(), od.structTypeAst)
		if err != nil {
			return nil, errors.New(err.Error() + " when parse " + tname)
		}
		odv := newOptionDef(valueSli, od.structType.GetValueType(), od.structTypeAst)

		resultMap := map[interface{}]interface{}{}
		for k, v := range kvMap {
			key, er := fillContent(odk, k)
			if er != nil {
				return nil, er
			}
			va, er := fillContent(odv, v)
			if er != nil {
				return nil, er
			}
			resultMap[key] = va
		}
		return resultMap, nil
	}
	return nil, errors.New("unsupported type")
}

func convertStruct(od *optionDef, value string) (map[string]interface{}, error) {
	tname := od.name
	kvMap, err := parseKV(value)
	if err != nil {
		return nil, errors.New(err.Error() + " when parse " + tname + ", input:\n" + value)
	}
	resultMap := map[string]interface{}{}
	for k, v := range kvMap {
		f := getField(k, od.structLike)
		if f == nil {
			return nil, errors.New(od.name + ":+can't find field for " + k)
		}
		fieldStruct, er := getStruct(f.GetType().GetName(), od.fromAst)
		if er != nil {
			return nil, errors.New("can't find type:" + f.GetType().GetName() + " for " + f.GetName())
		}
		odf := newOptionDef(fieldStruct, f.GetType(), od.fromAst)
		result, err := fillContent(odf, v)
		if err != nil {
			return nil, err
		}
		resultMap[k] = result
	}
	return resultMap, nil
}

func parseArr(str string) ([]string, error) {
	for {
		newstr := strings.ReplaceAll(str, "\t", " ")
		newstr = strings.ReplaceAll(newstr, "\n", " ")
		newstr = strings.ReplaceAll(newstr, " ,", ",")
		newstr = strings.ReplaceAll(newstr, ", ", ",")
		newstr = strings.ReplaceAll(newstr, " ]", "]")
		newstr = strings.ReplaceAll(newstr, "[ ", "[")
		newstr = strings.ReplaceAll(newstr, "  ", " ")
		newstr = strings.TrimSpace(newstr)
		if len(newstr) == len(str) {
			break
		}
		str = newstr
	}
	if !(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]")) {
		return nil, errors.New("no []")
	}
	str = str[1 : len(str)-1]

	var cb, sb, kstart, kend int
	var key string
	var dq, sq = true, true
	result := []string{}
	for i := 0; i < len(str); i++ {
		ch := str[i]
		if ch == '"' {
			dq = !dq
			continue
		}
		if ch == '\'' {
			sq = !sq
			continue
		}
		if ch == '{' {
			cb++
			continue
		}
		if ch == '}' {
			cb--
			continue
		}
		if ch == '[' {
			sb++
			continue
		}
		if ch == ']' {
			sb--
			continue
		}
		if ch == ',' {
			if sb == 0 && cb == 0 && dq && sq {
				kend = i
				key = str[kstart:kend]
				kstart = i + 1
				result = append(result, key)
			}
			continue
		}
	}
	if sb == 0 && cb == 0 && dq && sq {
		kend = len(str)
		if kstart >= kend {
			return nil, errors.New("grammar error")
		}
		key = str[kstart:kend]
		result = append(result, key)
		return result, nil
	} else {
		if dq && sq {
			return nil, errors.New("{} not match")
		} else {
			return nil, errors.New("quote not match")
		}
	}
}

func parseKV(str string) (map[string]string, error) {

	for {
		newstr := strings.ReplaceAll(str, "\t", " ")
		newstr = strings.ReplaceAll(newstr, "\n", " ")
		newstr = strings.ReplaceAll(newstr, " }", "}")
		newstr = strings.ReplaceAll(newstr, "{ ", "{")
		newstr = strings.ReplaceAll(newstr, " :", ":")
		newstr = strings.ReplaceAll(newstr, ": ", ":")
		newstr = strings.ReplaceAll(newstr, "  ", " ")
		newstr = strings.TrimSpace(newstr)
		if len(newstr) == len(str) {
			break
		}
		str = newstr
	}

	if !(strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) {
		return nil, errors.New("no {}")
	}
	str = str[1 : len(str)-1]

	var cb, sb, kstart, kend, vstart, vend int
	var dq, sq = true, true
	var key, value string
	result := map[string]string{}
	for i := 0; i < len(str); i++ {
		ch := str[i]
		if ch == '"' {
			dq = !dq
			continue
		}
		if ch == '\'' {
			sq = !sq
			continue
		}
		if ch == '{' {
			cb++
			continue
		}
		if ch == '}' {
			cb--
			continue
		}
		if ch == '[' {
			sb++
			continue
		}
		if ch == ']' {
			sb--
			continue
		}
		if ch == ':' {
			if sb == 0 && cb == 0 && dq && sq {
				kend = i
				// get k
				vstart = i + 1
				key = str[kstart:kend]
			}
			continue
		}
		if ch == ' ' {
			if sb == 0 && cb == 0 && dq && sq {
				vend = i
				if vstart >= vend {
					return nil, errors.New("grammar error")
				}
				kstart = i + 1
				value = str[vstart:vend]
				result[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
			continue
		}
	}
	if sb == 0 && cb == 0 && dq && sq {
		vend = len(str)
		if vstart >= vend {
			return nil, errors.New("grammar error")
		}
		if kstart >= kend {
			return nil, errors.New("grammar error")
		}
		value = str[vstart:vend]
		result[strings.TrimSpace(key)] = strings.TrimSpace(value)
		return result, nil
	} else {
		if dq && sq {
			return nil, errors.New("{} not match")

		} else {
			return nil, errors.New("quote not match")
		}
	}

}

type OptionData struct {
	optionName string
	isStruct   bool
	result     interface{}
	structLike *StructLike
}

func (o *OptionData) GetName() string {
	return o.optionName
}

func (o *OptionData) GetValue() interface{} {
	return o.result
}

func (o *OptionData) GetFieldValue(name string) (interface{}, error) {
	if !o.isStruct {
		return nil, errors.New("not struct")
	}
	f := getField(name, o.structLike)
	if f == nil {
		return nil, errors.New("field name not match")
	}
	resultMap := o.result.(map[string]interface{})
	return resultMap[f.GetName()], nil
}

// get StructLike from current ast
func getStruct(name string, ast *Thrift) (*structInfo, error) {
	if basicMap[name] || containerMap[name] {
		return &structInfo{
			isStruct:   false,
			name:       name,
			structLike: nil,
			fromAst:    nil,
		}, nil
	}
	if ast == nil {
		return nil, errors.New("no ast to find struct")
	}
	prefix, name := parsePrefixAndName(name)
	if prefix != "" {
		ast = getIncludeThrift(prefix, ast)
	}
	if ast == nil {
		return nil, errors.New("no ast to find struct")
	}
	for _, s := range ast.Structs {
		if s.Name == name {
			return &structInfo{true, s.Name, s, ast}, nil
		}
	}
	return nil, errors.New("struct not found")
}

func getIncludeThrift(prefix string, ast *Thrift) *Thrift {
	for _, inc := range ast.Includes {
		if strings.HasSuffix(inc.Path, prefix+".thrift") {
			return inc.Reference
		}
	}
	return nil
}

func getField(name string, structLike *StructLike) *Field {
	for _, field := range structLike.Fields {
		if field.GetName() == name {
			return field
		}
	}
	return nil
}

func parsePrefixAndName(name string) (string, string) {
	if strings.Contains(name, ".") {
		nameArr := strings.Split(name, ".")
		structName := nameArr[len(nameArr)-1]
		prefix := strings.TrimSuffix(name, "."+structName)
		return prefix, structName
	} else {
		return "", name
	}
}
