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

package dump

import (
	"bytes"
	"html"
	"html/template"
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

// DumpIDL Dump the ast to idl string
func DumpIDL(ast *parser.Thrift) (string, error) {
	tmpl, _ := template.New("thrift").Funcs(template.FuncMap{
		"RemoveLastComma": RemoveLastComma,
		"JoinQuotes":      JoinQuotes, "ReplaceQuotes": ReplaceQuotes,
	}).
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
