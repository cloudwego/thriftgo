package dump

import (
	"bytes"
	"github.com/cloudwego/thriftgo/parser"
	"text/template"
)

// DumpIDL Dump the ast to idl string
func DumpIDL(ast *parser.Thrift) (string, error) {
	tmpl, _ := template.New("thrift").Parse(IDLTemplate + TypeDefTemplate + AnnotationsTemplate +
		ConstantTemplate + EnumTemplate + ConstValueTemplate + FieldTemplate + StructTemplate + UnionTemplate +
		ExceptionTemplate + ServiceTemplate + FunctionTemplate + SingleLineFieldTemplate + TypeTemplate)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ast); err != nil {
		return "", err
	}
	return buf.String(), nil
}
