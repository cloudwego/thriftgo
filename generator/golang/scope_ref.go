// Copyright 2023 CloudWeGo Authors
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
	"errors"
	"strings"

	"github.com/cloudwego/thriftgo/config"
	"github.com/cloudwego/thriftgo/parser"
)

func (s *Scope) GetFirstDescriptor() string {
	for _, st := range s.ast.Structs {
		return st.Name
	}
	for _, st := range s.ast.Enums {
		return st.Name
	}
	for _, st := range s.ast.Typedefs {
		return st.Alias
	}
	for _, st := range s.ast.Constants {
		return st.Name
	}
	for _, st := range s.ast.Exceptions {
		return st.Name
	}
	for _, st := range s.ast.Unions {
		return st.Name
	}
	for _, st := range s.ast.Services {
		return st.Name
	}
	return ""
}

func BuildRefScope(cu *CodeUtils, ast *parser.Thrift) (*Scope, *Scope, error) {
	thriftRef := config.GetRef(ast.Filename)
	enableCodeRef := cu.Features().CodeRef || cu.Features().CodeRefSlim
	scope, err := BuildScope(cu, ast)
	if err != nil {
		return nil, nil, err
	}
	// no ref
	if !enableCodeRef || thriftRef == nil {
		return scope, nil, err
	}
	// all ref
	if thriftRef != nil && thriftRef.IsAllFieldsEmpty() {
		scope.setRefImport(thriftRef.Path)
		return nil, scope, err
	}
	return nil, nil, errors.New("config not support this feature currently")
	// todo not support now
	// half ref
	// we will change the fields from scope, we can't use BuildScope() to create scope because that function will put scope into a cache map.
	//localScope, err := doBuildScope(cu, ast)
	//if err != nil {
	//	return nil, nil, err
	//}
	//refScope, err := doBuildScope(cu, ast)
	//if err != nil {
	//	return nil, nil, err
	//}
	//refScope.setRefImport(thriftRef.Path)
	//// do not generate service to remote
	//refScope.services = nil
	//// grepService(thriftRef.Services, &localScope.services, &refScope.services)
	//grepStructs(thriftRef.Unions, &localScope.unions, &refScope.unions)
	//grepStructs(thriftRef.Exceptions, &localScope.exceptions, &refScope.exceptions)
	//grepStructs(thriftRef.Structs, &localScope.structs, &refScope.structs)
	//grepConstants(thriftRef.Consts, &localScope.constants, &refScope.constants)
	//grepTypedefs(thriftRef.Typedefs, &localScope.typedefs, &refScope.typedefs)
	//grepEnums(thriftRef.Enums, &localScope.enums, &refScope.enums)
	//// todo clean ref scope import
	//return localScope, refScope, nil
}

func isContains(sa []string, s string) bool {
	for _, str := range sa {
		if str == "*" {
			return true
		}
		if strings.HasPrefix(str, "*") && strings.HasSuffix(str, "*") {
			// *XXX* 模糊匹配
			if strings.Contains(s, str[1:len(str)-1]) {
				return true
			}
		} else if strings.HasPrefix(str, "*") {
			// *XXX 后缀模糊匹配
			if strings.HasSuffix(s, str[1:]) {
				return true
			}
		} else if strings.HasSuffix(str, "*") {
			// XXX* 前缀模糊匹配
			if strings.HasPrefix(s, str[:len(str)-1]) {
				return true
			}
		} else if str == s {
			// 严格匹配
			return true
		}
	}
	return false
}

func grepStructs(refNames []string, localArr, refArr *[]*StructLike) {
	*refArr = []*StructLike{}
	for i := 0; i < len(*localArr); i++ {
		elem := (*localArr)[i]
		if isContains(refNames, elem.GetName()) {
			*localArr = append((*localArr)[:i], (*localArr)[i+1:]...)
			*refArr = append(*refArr, elem)
			i--
		}
	}
}

func grepEnums(refNames []string, localArr, refArr *[]*Enum) {
	*refArr = []*Enum{}
	for i := 0; i < len(*localArr); i++ {
		elem := (*localArr)[i]
		if isContains(refNames, elem.GetName()) {
			*localArr = append((*localArr)[:i], (*localArr)[i+1:]...)
			*refArr = append(*refArr, elem)
			i--
		}
	}
}

func grepConstants(refNames []string, localArr, refArr *[]*Constant) {
	*refArr = []*Constant{}
	for i := 0; i < len(*localArr); i++ {
		elem := (*localArr)[i]
		if isContains(refNames, elem.GetName()) {
			*localArr = append((*localArr)[:i], (*localArr)[i+1:]...)
			*refArr = append(*refArr, elem)
			i--
		}
	}
}

func grepTypedefs(refNames []string, localArr, refArr *[]*Typedef) {
	*refArr = []*Typedef{}
	for i := 0; i < len(*localArr); i++ {
		elem := (*localArr)[i]
		if isContains(refNames, elem.GetName()) {
			*localArr = append((*localArr)[:i], (*localArr)[i+1:]...)
			*refArr = append(*refArr, elem)
			i--
		}
	}
}
