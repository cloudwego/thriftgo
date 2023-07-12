package golang

import (
	"github.com/cloudwego/thriftgo/config"
	"github.com/cloudwego/thriftgo/parser"
	"strings"
)

func BuildRefScope(cu *CodeUtils, ast *parser.Thrift) (*Scope, *Scope, error) {
	thriftRef := config.GetRef(ast.Filename)
	// no ref
	if !cu.Features().CodeRef || thriftRef == nil {
		scope, err := BuildScope(cu, ast)
		return scope, nil, err
	}
	// all ref
	if thriftRef != nil && thriftRef.IsAllFieldsEmpty() {
		scope, err := BuildScope(cu, ast)
		scope.setRefImport(thriftRef.Path)
		return nil, scope, err
	}
	// half ref
	scope, err := doBuildScope(cu, ast)
	if err != nil {
		return nil, nil, err
	}
	refScope, err := doBuildScope(cu, ast)
	if err != nil {
		return nil, nil, err
	}
	refScope.setRefImport(thriftRef.Path)
	// do not generate service to remote
	refScope.services = nil
	//grepService(thriftRef.Services, &scope.services, &refScope.services)
	grepStructs(thriftRef.Unions, &scope.unions, &refScope.unions)
	grepStructs(thriftRef.Exceptions, &scope.exceptions, &refScope.exceptions)
	grepStructs(thriftRef.Structs, &scope.structs, &refScope.structs)
	grepConstants(thriftRef.Consts, &scope.constants, &refScope.constants)
	grepTypedefs(thriftRef.Typedefs, &scope.typedefs, &refScope.typedefs)
	grepEnums(thriftRef.Enums, &scope.enums, &refScope.enums)
	// todo clean ref scope import
	return scope, refScope, nil
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
