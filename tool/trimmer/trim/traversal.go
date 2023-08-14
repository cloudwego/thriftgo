package trim

import "github.com/cloudwego/thriftgo/parser"

// traverse and remove the unmarked part of ast
func (t *Trimmer) traversal(ast *parser.Thrift, filename string) {

	var list1 []*parser.Include
	for i := range ast.Includes {
		if t.marks[filename][ast.Includes[i]] || len(ast.Includes[i].Reference.Constants) > 0 {
			t.traversal(ast.Includes[i].Reference, filename)
			list1 = append(list1, ast.Includes[i])
		}
	}
	ast.Includes = list1

	var list2 []*parser.Typedef
	for i := range ast.Typedefs {
		if t.marks[filename][ast.Typedefs[i]] {
			list2 = append(list2, ast.Typedefs[i])
		}
	}
	ast.Typedefs = list2

	var list3 []*parser.Enum
	for i := range ast.Enums {
		if t.marks[filename][ast.Enums[i]] {
			list3 = append(list3, ast.Enums[i])
		}
	}
	ast.Enums = list3

	var list4 []*parser.StructLike
	for i := range ast.Structs {
		if t.marks[filename][ast.Structs[i]] {
			list4 = append(list4, ast.Structs[i])
		}
	}
	ast.Structs = list4

	var list5 []*parser.StructLike
	for i := range ast.Unions {
		if t.marks[filename][ast.Unions[i]] {
			list5 = append(list5, ast.Unions[i])
		}
	}
	ast.Unions = list5

	var list6 []*parser.StructLike
	for i := range ast.Exceptions {
		if t.marks[filename][ast.Exceptions[i]] {
			list6 = append(list6, ast.Exceptions[i])
		}
	}
	ast.Exceptions = list6

	var list7 []*parser.Service
	for i := range ast.Services {
		if t.marks[filename][ast.Services[i]] {
			list7 = append(list7, ast.Services[i])
		}
	}
	ast.Services = list7

}
