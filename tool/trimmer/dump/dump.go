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
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

var UseOldDumpFunction bool

// DumpIDL Dump the ast to idl string very fast
func DumpIDL(ast *parser.Thrift) (string, error) {
	// 等新的 DumpIDL 稳定运行之后彻底去掉旧的，现在如果遇到问题手动设置 UseOldDumpFunction 来回滚
	if UseOldDumpFunction {
		return DumpIDL_V1(ast)
	}
	var sb stringBuilder

	for _, include := range ast.Includes {
		sb.writeString(fmt.Sprintf("include \"%s\"\n", include.Path))
	}

	if len(ast.Includes) > 0 {
		sb.writeString("\n")
	}

	for _, ns := range ast.Namespaces {
		sb.writeString(fmt.Sprintf("namespace %s %s", ns.Language, ns.Name))
		printAnnotation(&sb, ns.Annotations)
		sb.writeString("\n")
	}

	if len(ast.Namespaces) > 0 {
		sb.writeString("\n")
	}

	for _, include := range ast.CppIncludes {
		sb.writeString(fmt.Sprintf("cpp_include \"%s\"\n", include))
	}

	if len(ast.CppIncludes) > 0 {
		sb.writeString("\n")
	}

	for _, td := range ast.Typedefs {
		printComment(&sb, td.ReservedComments, "")
		sb.writeString(fmt.Sprintf("typedef %s", typeName(td.Type)))
		sb.writeString(" " + td.Alias + " ")
		printAnnotation(&sb, td.Annotations)
		sb.writeString("\n")
	}

	if len(ast.Typedefs) > 0 {
		sb.writeString("\n")
	}

	for _, c := range ast.Constants {
		printComment(&sb, c.ReservedComments, "")
		sb.writeString(fmt.Sprintf("const %s %s = ", typeName(c.Type), c.Name))
		ctv := c.Value.TypedValue
		printConstTypedValue(&sb, ctv)
		printAnnotation(&sb, c.Annotations)
		sb.writeString("\n")
	}

	if len(ast.Constants) > 0 {
		sb.writeString("\n")
	}

	for _, enm := range ast.Enums {
		printComment(&sb, enm.ReservedComments, "")
		sb.writeString(fmt.Sprintf("enum %s ", enm.Name))
		sb.writeString("{\n")
		for i, ev := range enm.Values {
			printComment(&sb, ev.ReservedComments, "    ")
			sb.writeString(fmt.Sprintf("    %s = %d ", ev.Name, ev.Value))
			printAnnotation(&sb, ev.Annotations)
			sb.writeString("\n")
			if i != len(enm.Values)-1 {
				sb.writeString("\n")
			}
		}
		sb.writeString("} ")
		printAnnotation(&sb, enm.Annotations)
		sb.writeString("\n\n")
	}

	if len(ast.Enums) > 0 {
		sb.writeString("\n")
	}

	for _, s := range ast.Structs {
		printStruct(&sb, s, "struct")
	}

	if len(ast.Structs) > 0 {
		sb.writeString("\n")
	}

	for _, s := range ast.Unions {
		printStruct(&sb, s, "union")
	}

	if len(ast.Unions) > 0 {
		sb.writeString("\n")
	}

	for _, s := range ast.Exceptions {
		printStruct(&sb, s, "exception")
	}

	if len(ast.Exceptions) > 0 {
		sb.writeString("\n")
	}

	for _, svc := range ast.Services {
		printComment(&sb, svc.ReservedComments, "")
		sb.writeString(fmt.Sprintf("service %s ", svc.Name))
		if svc.Extends != "" {
			sb.writeString("extends " + svc.Extends + " ")
		}
		sb.writeString("{\n")
		for _, f := range svc.Functions {
			printComment(&sb, f.ReservedComments, "    ")
			sb.writeString("    ")
			if f.Oneway {
				sb.writeString("oneway ")
			}
			sb.writeString(fmt.Sprintf("%s %s", typeName(f.FunctionType), f.Name))
			sb.writeString("(")
			for i, ag := range f.Arguments {
				required := ""
				if ag.Requiredness.IsOptional() {
					required = "optional "
				} else if ag.Requiredness.IsRequired() {
					required = "required "
				}
				sb.writeString(fmt.Sprintf("%d: %s%s %s", ag.ID, required, typeName(ag.Type), ag.Name))
				if i != len(f.Arguments)-1 {
					sb.writeString(", ")
				}
			}
			sb.writeString(")")
			if len(f.Throws) > 0 {
				sb.writeString("throws ")
				sb.writeString("(")
				for i, th := range f.Throws {
					required := ""
					if th.Requiredness.IsOptional() {
						required = "optional "
					} else if th.Requiredness.IsRequired() {
						required = "required "
					}
					sb.writeString(fmt.Sprintf("%d: %s%s %s", th.ID, required, typeName(th.Type), th.Name))
					if i != len(f.Arguments)-1 {
						sb.writeString(", ")
					}
				}
				sb.writeString(")")
			}
			printAnnotation(&sb, f.Annotations)
			sb.writeString("\n")
		}

		sb.writeString("} ")
		printAnnotation(&sb, svc.Annotations)
		sb.writeString("\n\n")
	}

	escapedString := sb.String()
	// 把 " 替换为 \"
	escapedString = strings.Replace(escapedString, "##34;", `\"`, -1)
	// 如果本身就有 \"，上面的情况就会变成 \\"，给转回 \"
	escapedString = strings.Replace(escapedString, `\\"`, `\"`, -1)
	// tag 的前后符号统一采用 "
	outString := strings.Replace(escapedString, "#OUTQUOTES", "\"", -1)
	return html.UnescapeString(outString), nil
}

func typeName(t *parser.Type) string {
	if t == nil {
		return ""
	}

	name := t.Name
	if t.KeyType != nil && t.ValueType != nil {
		name = fmt.Sprintf("%s<%s,%s>", t.Name, typeName(t.KeyType), typeName(t.ValueType))
	} else if t.ValueType != nil && t.KeyType == nil {
		name = fmt.Sprintf("%s<%s>", t.Name, typeName(t.ValueType))
	}

	if t.Annotations != nil {
		var sb stringBuilder
		printAnnotation(&sb, t.Annotations)
		name = name + sb.String()
	}
	return name
}

type stringBuilder struct {
	buffer strings.Builder
}

func (s *stringBuilder) writeString(str string) {
	if strings.Contains(str, "&") {
		// 将 & 转义为 &amp;
		str = strings.ReplaceAll(str, "&", "&amp;")
	}
	s.buffer.WriteString(str)
}

func (s *stringBuilder) String() string {
	return s.buffer.String()
}

func joinQuotes(s string) string {
	return fmt.Sprintf("%s", "#OUTQUOTES"+s+"#OUTQUOTES")
}

func replaceQuotes(s string) string {
	out := strings.Replace(s, "\"", "#OUTQUOTES", -1)
	return out
}

func printAnnotation(sb *stringBuilder, a parser.Annotations) {
	if len(a) == 0 {
		return
	}
	sb.writeString("(")
	for i, anno := range a {
		for ii, v := range anno.Values {
			val := strings.ReplaceAll(joinQuotes(v), `"`, "##34;")

			sb.writeString(fmt.Sprintf("%s = %s", anno.Key, val))
			if i != len(a)-1 || ii != len(anno.Values)-1 {
				sb.writeString(", ")
			}
		}
	}
	sb.writeString(")")
}

func printComment(sb *stringBuilder, comment, prefix string) {
	if len(strings.TrimSpace(comment)) > 0 {
		sb.writeString(prefix + replaceQuotes(comment) + "\n")
	}
}

func printStruct(sb *stringBuilder, s *parser.StructLike, structType string) {
	printComment(sb, s.ReservedComments, "")
	sb.writeString(fmt.Sprintf("%s %s ", structType, s.Name))
	sb.writeString("{\n")
	for _, f := range s.Fields {
		printComment(sb, f.ReservedComments, "    ")
		required := ""
		if f.Requiredness.IsOptional() {
			required = "optional "
		} else if f.Requiredness.IsRequired() {
			required = "required "
		}
		sb.writeString(fmt.Sprintf("    %d: %s%s", f.ID, required, typeName(f.Type)))
		sb.writeString(fmt.Sprintf(" %s", f.Name))

		if f.Default != nil {
			sb.writeString(" = ")
			printConstTypedValue(sb, f.Default.TypedValue)
		}
		printAnnotation(sb, f.Annotations)

		sb.writeString("\n")
	}
	sb.writeString("} ")
	printAnnotation(sb, s.Annotations)
	sb.writeString("\n\n")
}

func printConstTypedValue(sb *stringBuilder, ctv *parser.ConstTypedValue) {
	if ctv.Double != nil {
		sb.writeString(strconv.FormatFloat(*ctv.Double, 'f', -1, 64))
	} else if ctv.Int != nil {
		sb.writeString(fmt.Sprintf("%d", *ctv.Int))
	} else if ctv.Literal != nil {
		val := *ctv.Literal
		val = strings.ReplaceAll(joinQuotes(val), `"`, "##34;")
		sb.writeString(fmt.Sprintf("%s", val))
	} else if ctv.Identifier != nil {
		sb.writeString(fmt.Sprintf("%s", *ctv.Identifier))
	} else if ctv.IsSetList() {
		sb.writeString("[")
		for i, v := range ctv.List {
			printConstTypedValue(sb, v.TypedValue)
			if i != len(ctv.List)-1 {
				sb.writeString(", ")
			}
		}
		sb.writeString("]")
	} else if ctv.IsSetMap() {
		sb.writeString("{")
		for i, pair := range ctv.Map {
			sb.writeString("\n\t")
			printConstTypedValue(sb, pair.Key.TypedValue)
			sb.writeString(": ")
			printConstTypedValue(sb, pair.Value.TypedValue)
			if i != len(ctv.Map)-1 {
				sb.writeString(", ")
			}
		}
		sb.writeString("\n")
		sb.writeString("}")
	}
}

// DumpIDL_V1 Deprecated DumpIDL_V1 Dump the ast to idl string by go template, use DumpIDL() for better speed.
func DumpIDL_V1(ast *parser.Thrift) (string, error) {
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
	escapedString := buf.String()
	// 把 " 替换为 \"
	escapedString = strings.Replace(escapedString, "&#34;", `\"`, -1)
	// 如果本身就有 \"，上面的情况就会变成 \\"，给转回 \"
	escapedString = strings.Replace(escapedString, `\\"`, `\"`, -1)
	// tag 的前后符号统一采用 "
	outString := strings.Replace(escapedString, "#OUTQUOTES", "\"", -1)
	return html.UnescapeString(outString), nil
}
