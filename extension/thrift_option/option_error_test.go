package thrift_option

import (
	"errors"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
	"testing"
)

// 检测各种报错提示场景
func TestOptionError(t *testing.T) {
	ast, err := parser.ParseFile("option_idl/test_grammar_error.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	p := getStructFromAst("PersonA", ast)
	assert(t, p != nil)

	// 错误或者不存在的 Option 名称
	_, err = ParseStructOption(p, "abc", ast)
	// todo 这里需要展示前缀？
	assert(t, err != nil && errors.Is(err, ErrKeyNotMatch), err)

	// 错误或者不存在的 Option 名称
	_, err = ParseStructOption(p, "entity.person_xxx_info", ast)
	// todo 这里需要展示前缀？
	assert(t, err != nil && errors.Is(err, ErrNotExistOption), err)

	// 错误的 field value
	p = getStructFromAst("PersonB", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, "entity.person_basic_info", ast)
	assert(t, err != nil && errors.Is(err, ErrParseFailed), err)

	// 错误的 field name
	p = getStructFromAst("PersonC", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, "entity.person_struct_info", ast)
	// todo 具体的 parse field 可以以后增加测试校验
	assert(t, err != nil && errors.Is(err, ErrParseFailed), err)

	// 错误的 kv 语法
	p = getStructFromAst("PersonE", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, "entity.person_container_info", ast)
	assert(t, err != nil && errors.Is(err, ErrParseFailed), err)

	// 没有 include 对应 option 的 IDL
	p = getStructFromAst("PersonF", ast)
	assert(t, p != nil)
	_, err = ParseStructOption(p, "validation.person_string_info", ast)
	assert(t, err != nil && errors.Is(err, ErrNotIncluded), err)

}

func TestGrammarCheck(t *testing.T) {

	// 测试有 option 解析错误等各种情况的 IDL
	ast, err := parser.ParseFile("option_idl/test_grammar_error.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	err = CheckOptionGrammar(ast)
	test.Assert(t, err != nil)

	// 测试 option 写法都正常的 IDL （忽略 option 没有匹配到的情况）
	ast, err = parser.ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)
	err = CheckOptionGrammar(ast)
	test.Assert(t, err == nil)
}
