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

// GetConstGlobals returns all global variables defined in the root scope that result in constant value.
func (cu *CodeUtils) GetConstGlobals(ast *parser.Thrift) (cs []*parser.Constant) {
	for _, c := range ast.Constants {
		if cu.IsConstantInGo(c) {
			cs = append(cs, c)
		}
	}
	return
}

// GetNonConstGlobals returns all global variables defined in the root scope that result in non-constant values.
func (cu *CodeUtils) GetNonConstGlobals(ast *parser.Thrift) (ncs []*parser.Constant) {
	for _, c := range ast.Constants {
		if !cu.IsConstantInGo(c) {
			ncs = append(ncs, c)
		}
	}
	return
}

// NeedRedirect deterimines whether the given field should result in a pointer type.
// Condition: struct-like || (optional non-binary base type without default vlaue).
func (cu *CodeUtils) NeedRedirect(f *parser.Field) bool {
	if f.Type.Category.IsStructLike() {
		return true
	}

	if f.Requiredness.IsOptional() && !f.IsSetDefault() {
		if f.Type.Category.IsBinary() {
			// binary types produce slice types
			return false
		}
		return cu.IsBaseType(f.Type)
	}

	return false
}

// IsConstantInGo tells whether a constant in thrift IDL results in a constant in go.
func (cu *CodeUtils) IsConstantInGo(v *parser.Constant) bool {
	c := v.Type.Category
	if c.IsBaseType() && c != parser.Category_Binary {
		return true
	}
	return c == parser.Category_Enum
}

// IsBaseType determines whether the given type is a base type.
func (cu *CodeUtils) IsBaseType(t *parser.Type) bool {
	if t.Category.IsBaseType() || t.Category == parser.Category_Enum {
		return true
	}
	if t.Category == parser.Category_Typedef {
		panic(fmt.Sprintf("unexpected typedef category: %+v", t))
	}
	return false
}

// IsFixedLengthType determines whether the given type is a fixed length type.
func (cu *CodeUtils) IsFixedLengthType(t *parser.Type) bool {
	return parser.Category_Bool <= t.Category && t.Category <= parser.Category_Double
}

// IsSetterOfResponseType determines whether the field in th given type is the Success field of a response wrapper type.
func (cu *CodeUtils) IsSetterOfResponseType(typeName, fieldName string) (bool, error) {
	yes := strings.HasSuffix(typeName, "Result") && fieldName == "Success"
	return yes, nil
}

// SupportIsSet determines whether a field supports IsSet query.
func (cu *CodeUtils) SupportIsSet(f *parser.Field) bool {
	return f.Type.Category.IsStructLike() || f.Requiredness.IsOptional()
}

// IsPointerTypeName reports whether the type name has a prefix "*".
func (cu *CodeUtils) IsPointerTypeName(typeName string) bool {
	return strings.HasPrefix(typeName, "*")
}
