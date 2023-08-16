package dump

import (
	"bytes"
	"github.com/cloudwego/thriftgo/parser"
	"html"
	"html/template"
	"strings"
)

// DumpIDL Dump the ast to idl string
func DumpIDL(ast *parser.Thrift) (string, error) {
	tmpl, _ := template.New("thrift").Funcs(template.FuncMap{"RemoveLastComma": RemoveLastComma,
		"JoinQuotes": JoinQuotes, "ReplaceQuotes": ReplaceQuotes}).
		Parse(IDLTemplate + TypeDefTemplate + AnnotationsTemplate +
			ConstantTemplate + EnumTemplate + ConstValueTemplate + FieldTemplate + StructTemplate + UnionTemplate +
			ExceptionTemplate + ServiceTemplate + FunctionTemplate + SingleLineFieldTemplate + TypeTemplate +
			ConstListTemplate + ConstMapTemplate)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ast); err != nil {
		return "", err
	}
	// deal with \\
	escapedString := strings.Replace(buf.String(), "&#34;", "\\\"", -1)
	outString := strings.Replace(escapedString, "#OUTQUOTES", "\"", -1)
	return html.UnescapeString(outString), nil
}
