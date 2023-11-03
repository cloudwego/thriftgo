package thrift_option

import (
	"encoding/hex"
	"errors"
	"github.com/cloudwego/thriftgo/thrift_reflection"
	"github.com/cloudwego/thriftgo/utils"
	"reflect"
	"strconv"
	"strings"
)

func trimQuote(value string) string {
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		value = value[1 : len(value)-1]
	}
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}
	return value
}

func createInstance(td *thrift_reflection.TypeDescriptor, content string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	content = trimQuote(content)
	if td.IsBasic() {
		val, err := createBasic(td.GetName(), content)
		return val, val, err
	}
	if td.IsContainer() {
		return createContainer(td, content, mapMode)
	}
	if td.IsStruct() {
		return creatStruct(td, content, mapMode)
	}
	if td.IsEnum() {
		return createEnum(td, content, mapMode)
	}
	if td.IsTypedef() {
		return createTypedef(td, content, mapMode)
	}
	return nil, nil, errors.New("unknown type")
}

func createEnum(td *thrift_reflection.TypeDescriptor, content string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	enumDesc, err := td.GetEnumDescriptor()
	if err != nil {
		return nil, nil, err
	}
	for _, value := range enumDesc.GetValues() {
		if content == value.GetName() {
			val := value.GetValue()
			if mapMode {
				return val, nil, nil
			} else {
				enumGoType, er := td.GetGoType()
				if er != nil {
					return nil, nil, er
				}
				enumInstance := reflect.New(enumGoType).Elem()
				enumVal := reflect.ValueOf(val)
				enumInstance.Set(enumVal.Convert(enumGoType))
				return val, enumInstance.Interface(), nil
			}
		}
	}
	return nil, nil, errors.New("enum value " + content + " not found for" + enumDesc.GetName())
}

func createTypedef(td *thrift_reflection.TypeDescriptor, content string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	tdDesc, err := td.GetTypedefDescriptor()
	if err != nil {
		return nil, nil, err
	}
	return createInstance(tdDesc.GetType(), content, mapMode)
}

type quadruple struct {
	idx           int
	key           string
	mapValue      interface{}
	instanceValue interface{}
}

func creatStruct(td *thrift_reflection.TypeDescriptor, content string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	des, err := td.GetStructDescriptor()
	if err != nil {
		return nil, nil, err
	}
	kv, err := utils.ParseKV(content)
	if err != nil {
		return nil, nil, err
	}
	// 检查 kv 里是否有非法字段
	for k := range kv {
		if des.GetFieldByName(k) == nil {
			return nil, nil, errors.New("field not exist:" + k)
		}
	}
	triples := []*quadruple{}
	for idx, fd := range des.GetFields() {
		value, ok := kv[fd.GetName()]
		if !ok {
			// 当 option 里没对字段赋值时，使用 default value
			if fd.GetDefaultValue() != nil {
				value = fd.GetDefaultValue().GetValueAsString()
			} else {
				continue
			}
		}
		mv, iv, er := createInstance(fd.GetType(), value, mapMode)
		if er != nil {
			return nil, nil, er
		}
		triples = append(triples, &quadruple{
			idx:           idx,
			key:           fd.GetName(),
			mapValue:      mv,
			instanceValue: iv,
		})
	}

	resultMap := map[string]interface{}{}
	for _, t := range triples {
		resultMap[t.key] = t.mapValue
	}

	if !mapMode {
		goType := des.GetGoType()
		structPtr := reflect.New(goType)
		structEntity := structPtr.Elem()
		if !structEntity.IsValid() {
			return nil, nil, errors.New("invalid")
		}
		for _, t := range triples {
			reflectField := structEntity.Field(t.idx)
			reflectField.Set(reflect.ValueOf(t.instanceValue))
		}
		return resultMap, structPtr.Interface(), nil
	}
	return resultMap, nil, nil
}

func createBasic(name, value string) (interface{}, error) {
	switch name {
	case "bool":
		i, er := strconv.ParseBool(value)
		if er != nil {
			return nil, er
		}
		return i, nil
	case "byte":
		i, er := strconv.ParseInt(value, 10, 8)
		if er != nil {
			return nil, er
		}
		return int8(i), nil
	case "i8":
		i, er := strconv.ParseInt(value, 10, 8)
		if er != nil {
			return nil, er
		}
		return int8(i), nil
	case "i16":
		i, er := strconv.ParseInt(value, 10, 16)
		if er != nil {
			return nil, er
		}
		return int16(i), nil
	case "i32":
		i, er := strconv.ParseInt(value, 10, 32)
		if er != nil {
			return nil, er
		}
		return int32(i), nil
	case "i64":
		i, er := strconv.ParseInt(value, 10, 64)
		if er != nil {
			return nil, er
		}
		return i, nil
	case "double":
		i, er := strconv.ParseFloat(value, 64)
		if er != nil {
			return nil, er
		}
		return i, nil
	case "binary":
		i, er := hex.DecodeString(value)
		if er != nil {
			return nil, er
		}
		return i, nil
	case "string":
		return value, nil
	default:
		return nil, errors.New("unsupported basic type: " + name)
	}
}

func createList(td *thrift_reflection.TypeDescriptor, value string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	arr, err := utils.ParseArr(value)
	if err != nil {
		return nil, nil, errors.New(err.Error() + " when parse " + td.Name)
	}

	resultsMap := []interface{}{}
	resultsInstance := []interface{}{}
	for _, elm := range arr {
		mv, iv, er := createInstance(td.GetValueType(), elm, mapMode)
		if er != nil {
			return nil, nil, er
		}
		resultsMap = append(resultsMap, mv)
		resultsInstance = append(resultsInstance, iv)
	}

	if mapMode {
		return resultsMap, nil, nil
	} else {
		listType, er := td.GetGoType()
		if er != nil {
			return nil, nil, er
		}
		listValue := reflect.MakeSlice(listType, 0, 0)
		for _, elmInstance := range resultsInstance {
			listValue = reflect.Append(listValue, reflect.ValueOf(elmInstance))
		}
		return resultsMap, listValue.Interface(), nil
	}
}

func createMap(td *thrift_reflection.TypeDescriptor, value string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	kvMap, err := utils.ParseKV(value)
	if err != nil {
		return nil, nil, errors.New(err.Error() + " when parse map " + td.Name)
	}

	resultMap := map[interface{}]interface{}{}
	resultInstances := map[interface{}]interface{}{}
	for k, v := range kvMap {
		kmv, kiv, er := createInstance(td.GetKeyType(), k, mapMode)
		if er != nil {
			return nil, nil, er
		}
		vmv, viv, er := createInstance(td.GetValueType(), v, mapMode)
		if er != nil {
			return nil, nil, er
		}
		resultMap[kmv] = vmv
		resultInstances[kiv] = viv
	}

	if mapMode {
		return resultMap, nil, nil
	} else {
		mapType, er := td.GetGoType()
		if er != nil {
			return nil, nil, er
		}
		mapValue := reflect.MakeMap(mapType)

		for k, v := range resultInstances {
			mapValue.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		return resultMap, mapValue.Interface(), nil
	}
}

func createContainer(td *thrift_reflection.TypeDescriptor, value string, mapMode bool) (mapVal, instanceVal interface{}, e error) {
	typeName := td.GetName()
	if typeName == "map" {
		return createMap(td, value, mapMode)
	}
	if typeName == "list" || typeName == "set" {
		return createList(td, value, mapMode)
	}
	return nil, nil, errors.New("illegal container type")
}
