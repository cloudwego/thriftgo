package trim

import "github.com/cloudwego/thriftgo/parser"

// mark the used part of ast
func (t *Trimmer) markAST(ast *parser.Thrift) {
	t.marks[ast.Filename] = make(map[interface{}]bool)
	for _, service := range ast.Services {
		t.markService(service, ast, ast.Filename)
	}
}

func (t *Trimmer) markService(svc *parser.Service, ast *parser.Thrift, filename string) {
	if t.marks[filename][svc] {
		return
	}

	t.marks[filename][svc] = true
	if svc.Extends != "" {
		// handle extension
		theInclude := ast.Includes[svc.Reference.Index]
		t.marks[filename][theInclude] = true
		for _, service := range theInclude.Reference.Services {
			if service.Name == svc.Reference.Name {
				t.markService(service, theInclude.Reference, filename)
				break
			}
		}
	}

	for _, function := range svc.Functions {
		for _, arg := range function.Arguments {
			t.markType(arg.Type, ast, filename)
		}
		for _, throw := range function.Throws {
			t.markType(throw.Type, ast, filename)
		}
		if !function.Void {
			t.markType(function.FunctionType, ast, filename)
		}
	}
}

func (t *Trimmer) markType(theType *parser.Type, ast *parser.Thrift, filename string) {
	// plain type
	if theType.Category <= 8 && theType.IsTypedef == nil {
		return
	}

	if theType.KeyType != nil {
		t.markType(theType.KeyType, ast, filename)
	}
	if theType.ValueType != nil {
		t.markType(theType.ValueType, ast, filename)
	}
	if theType.IsTypedef != nil {
		t.markTypeDef(theType, ast, filename)
		return
	}

	baseAST := ast
	if theType.Reference != nil {
		// if referenced, redirect to included ast
		baseAST = ast.Includes[theType.Reference.Index].Reference
		t.marks[filename][ast.Includes[theType.Reference.Index]] = true
	}
	if theType.Category.IsStruct() {
		for _, str := range baseAST.Structs {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsException() {
		for _, str := range baseAST.Exceptions {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsUnion() {
		for _, str := range baseAST.Unions {
			if str.Name == theType.Name || (theType.Reference != nil && str.Name == theType.Reference.Name) {
				t.markStructLike(str, baseAST, filename)
				break
			}
		}
	} else if theType.Category.IsEnum() {
		for _, enum := range baseAST.Enums {
			if enum.Name == theType.Name || (theType.Reference != nil && enum.Name == theType.Reference.Name) {
				t.markEnum(enum, filename)
				break
			}
		}
	}
}

func (t *Trimmer) markStructLike(str *parser.StructLike, ast *parser.Thrift, filename string) {
	if t.marks[filename][str] {
		return
	}
	t.marks[filename][str] = true
	for _, field := range str.Fields {
		t.markType(field.Type, ast, filename)
	}
}

func (t *Trimmer) markEnum(enum *parser.Enum, filename string) {
	t.marks[filename][enum] = true
}

func (t *Trimmer) markTypeDef(theType *parser.Type, ast *parser.Thrift, filename string) {
	if theType.IsTypedef == nil {
		return
	}

	for _, typedef := range ast.Typedefs {
		if typedef.Alias == theType.Name {
			if !t.marks[filename][typedef] {
				t.marks[filename][typedef] = true
				t.markType(typedef.Type, ast, filename)
			}
			return
		}
	}
}
