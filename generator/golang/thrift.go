// Copyright 2021 CloudWeGo Authors
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

package golang

import (
	"fmt"
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

// GetTypeIDConstant returns the thrift type ID literal for the given type which
// is suitable to concate with "thrift." to produce a valid type ID constant.
func GetTypeIDConstant(t *parser.Type) string {
	tid := GetTypeID(t)
	tid = strings.ToUpper(tid)
	if tid == "BINARY" {
		tid = "STRING"
	}
	return tid
}

// GetTypeID returns the thrift type ID literal for the given type which is suitable
// to concate with "Read" or "Write" to produce a valid method name in the TProtocol
// interface. Note that enum types results in I32.
func GetTypeID(t *parser.Type) string {
	// Bool|Byte|I16|I32|I64|Double|String|Binary|Set|List|Map|Struct
	return category2TypeID[t.Category]
}

// IsBaseType determines whether the given type is a base type.
func IsBaseType(t *parser.Type) bool {
	if t.Category.IsBaseType() || t.Category == parser.Category_Enum {
		return true
	}
	if t.Category == parser.Category_Typedef {
		panic(fmt.Sprintf("unexpected typedef category: %+v", t))
	}
	return false
}

func checkErrorTPL(assign string, err string) string {
	return "if err := " + assign + "; err != nil {\n goto " + err + "\n}\n"
}

// IsBaseType determines whether the given type is a base type.
func ZeroWriter(t *parser.Type, oprot string, err string) string {
	switch t.GetCategory() {
	case parser.Category_Bool:
		return checkErrorTPL(oprot+".WriteBool(false)", err)
	case parser.Category_Byte:
		return checkErrorTPL(oprot+".WriteByte(0)", err)
	case parser.Category_I16:
		return checkErrorTPL(oprot+".WriteI16(0)", err)
	case parser.Category_Enum, parser.Category_I32:
		return checkErrorTPL(oprot+".WriteI32(0)", err)
	case parser.Category_I64:
		return checkErrorTPL(oprot+".WriteI64(0)", err)
	case parser.Category_Double:
		return checkErrorTPL(oprot+".WriteDouble(0)", err)
	case parser.Category_String:
		return checkErrorTPL(oprot+".WriteString(\"\")", err)
	case parser.Category_Binary:
		return checkErrorTPL(oprot+".WriteBinary([]byte{})", err)
	case parser.Category_Map:
		return checkErrorTPL(oprot+".WriteMapBegin(thrift."+GetTypeIDConstant(t.GetKeyType())+
			",thrift."+GetTypeIDConstant(t.GetValueType())+",0)", err) + checkErrorTPL(oprot+".WriteMapEnd()", err)
	case parser.Category_List:
		return checkErrorTPL(oprot+".WriteListBegin(thrift."+GetTypeIDConstant(t.GetValueType())+
			",0)", err) + checkErrorTPL(oprot+".WriteListEnd()", err)
	case parser.Category_Set:
		return checkErrorTPL(oprot+".WriteSetBegin(thrift."+GetTypeIDConstant(t.GetValueType())+
			",0)", err) + checkErrorTPL(oprot+".WriteSetEnd()", err)
	case parser.Category_Struct:
		return checkErrorTPL(oprot+".WriteStructBegin(\"\")", err) + checkErrorTPL(oprot+".WriteFieldStop()", err) +
			checkErrorTPL(oprot+".WriteStructEnd()", err)
	default:
		panic("unsuported type zero writer for" + t.Name)
	}
}

// IsIntType determines whether the given type is a Int type.
func IsIntType(t *parser.Type) bool {
	switch t.Category {
	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64, parser.Category_Enum:
		return true
	default:
		return false
	}
}

// IsStrType determines whether the given type is a Str type.
func IsStrType(t *parser.Type) bool {
	switch t.Category {
	case parser.Category_String, parser.Category_Binary:
		return true
	default:
		return false
	}
}

// NeedRedirect deterimines whether the given field should result in a pointer type.
// Condition: struct-like || (optional non-binary base type without default vlaue).
func NeedRedirect(f *parser.Field) bool {
	if f.Type.Category.IsStructLike() {
		return true
	}

	if f.Requiredness.IsOptional() && !f.IsSetDefault() {
		if f.Type.Category.IsBinary() {
			// binary types produce slice types
			return false
		}
		return IsBaseType(f.Type)
	}

	return false
}

// IsConstantInGo tells whether a constant in thrift IDL results in a constant in go.
func IsConstantInGo(v *parser.Constant) bool {
	c := v.Type.Category
	if c.IsBaseType() && c != parser.Category_Binary {
		return true
	}
	return c == parser.Category_Enum
}

// IsFixedLengthType determines whether the given type is a fixed length type.
func IsFixedLengthType(t *parser.Type) bool {
	return parser.Category_Bool <= t.Category && t.Category <= parser.Category_Double
}

// SupportIsSet determines whether a field supports IsSet query.
func SupportIsSet(f *parser.Field) bool {
	return f.Type.Category.IsStructLike() || f.Requiredness.IsOptional()
}
