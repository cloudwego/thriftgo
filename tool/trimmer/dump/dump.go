package dump

import (
	"bytes"
	"github.com/cloudwego/thriftgo/parser"
	"text/template"
)

const thriftTemplate = `
{{- /* 输出依赖文件 */}}
{{range .Includes}}include "{{.Path}}"
{{end}}
{{- /* 输出命名空间 */}}
{{range .Namespaces}}namespace {{.Language}} {{.Name}}
{{end}}
{{- /* 输出类型定义 */}}
{{range .Typedefs}}typedef {{.Type.Name}} {{.Alias}}{{if .ReservedComments}} {{.ReservedComments}}{{end}}
{{end}}
{{- /* 输出常量定义 */}}{{range .Constants}}const {{.Type.Name}} {{.Name}} = {{if .Value.TypedValue.Double}}{{.Value.TypedValue.Double}}{{end}}
{{- /* 常量-double */}}{{if .Value.TypedValue.Int}}{{.Value.TypedValue.Int}}{{end}}
{{- /* 常量-literal */}}{{if .Value.TypedValue.Literal}}{{.Value.TypedValue.Literal}}{{end}}
{{- /* 常量-identifier */}}{{if .Value.TypedValue.Identifier}}{{.Value.TypedValue.Identifier}}{{end}}
{{- /* 常量-list */}}{{if .Value.TypedValue.List}}{{.Value.TypedValue.List}}{{end}}
{{- /* 常量-map */}}{{if .Value.TypedValue.Map}}{{.Value.TypedValue.Map}}{{end}}
{{end}}
{{- /* 输出枚举类型定义 */}}
{{range .Enums}}enum {{.Name}} {
{{range .Values}}	{{.Name}}{{if .Value}} = {{.Value}}{{end}}
{{end}}}{{end}}
{{- /* 输出结构体定义 */}}
{{range .Structs}}
{{.Category}} {{.Name}} { {{range .Fields}}
	{{.ID}}:{{if .Requiredness.IsRequired}} required{{else if .Requiredness.IsOptional}} optional{{end}} {{.Type.Name}} {{.Name}} {{if .Default}}= 
{{- /* 常量-double */}}{{if .Default.TypedValue.Int}}{{.Default.TypedValue.Int}}{{end}}
{{- /* 常量-literal */}}{{if .Default.TypedValue.Literal}}{{.Default.TypedValue.Literal}}{{end}}
{{- /* 常量-identifier */}}{{if .Default.TypedValue.Identifier}}{{.Default.TypedValue.Identifier}}{{end}}
{{- /* 常量-list */}}{{if .Default.TypedValue.List}}{{.Default.TypedValue.List}}{{end}}
{{- /* 常量-map */}}{{if .Default.TypedValue.Map}}{{.Default.TypedValue.Map}}{{end}}
{{end}}{{end}}
}
{{end}}
{{- /* 输出异常定义 */}}
{{range .Exceptions}}
{{.Category}} {{.Name}} { {{range .Fields}}
	{{.ID}}:{{if .Requiredness.IsRequired}} required{{else if .Requiredness.IsOptional}} optional{{end}} {{.Type.Name}} {{.Name}} {{if .Default}}=
{{- /* 常量-double */}}{{if .Default.TypedValue.Int}} {{.Default.TypedValue.Int}}{{end}}
{{- /* 常量-literal */}}{{if .Default.TypedValue.Literal}} {{.Default.TypedValue.Literal}}{{end}}
{{- /* 常量-identifier */}}{{if .Default.TypedValue.Identifier}} {{.Default.TypedValue.Identifier}}{{end}}
{{- /* 常量-list */}}{{if .Default.TypedValue.List}} {{.Default.TypedValue.List}}{{end}}
{{- /* 常量-map */}}{{if .Default.TypedValue.Map}} {{.Default.TypedValue.Map}}{{end}}{{end}}{{end}}
}
{{end}}
{{- /* 输出联合体定义 */}}
{{range .Unions}}
    {{- /* 输出注解 */}}
    {{range .Annotations}}
        {{.Key}}({{range .Values}}"{{.}}"{{end}});
    {{end}}
    union {{.Name}} {
        {{range .Fields}}
            {{- /* 输出注解 */}}
            {{range .Annotations}}
                {{.Key}}({{range .Values}}"{{.}}"{{end}});
            {{end}}
            {{.Type.Name}} {{.Name}} (id: {{.ID}}){{if .Requiredness.IsRequired}}required{{else if .Requiredness.IsOptional}}optional{{end}} {{if .Default}}= {{.Default.TypedValue}}{{end}}{{if .ReservedComments}} {{.ReservedComments}}{{end}};
        {{end}}
    }{{if .ReservedComments}} {{.ReservedComments}}{{end}};
{{end}}

{{- /* 输出服务定义 */}}
{{range .Services}}
    {{- /* 输出注解 */}}
    {{range .Annotations}}
        {{.Key}}({{range .Values}}"{{.}}"{{end}});
    {{end}}
    service {{.Name}} {
        {{- /* 输出服务方法定义 */}}
        {{range .Functions}}
            {{- /* 输出注解 */}}
            {{range .Annotations}}
                {{.Key}}({{range .Values}}"{{.}}"{{end}});
            {{end}}
            {{.FunctionType.Name}} {{.Name}}(
                {{- /* 输出服务方法参数 */}}
                {{range .Arguments}}
                    {{- /* 输出注解 */}}
                    {{range .Annotations}}
                        {{.Key}}({{range .Values}}"{{.}}"{{end}});
                    {{end}}
                    {{.Type.Name}} {{.Name}} (id: {{.ID}}){{if .Requiredness.IsRequired}} required{{else if .Requiredness.IsOptional}} optional{{end}},{{if .Default}} default {{.Default.TypedValue}},{{end}}{{if .ReservedComments}} {{.ReservedComments}} {{end}}
                {{end}}
            ){{if .Throws}} throws ({{range .Throws}}{{.Type.Name}} {{.Name}},{{end}}){{end}}{{if .ReservedComments}} {{.ReservedComments}}{{end}};
        {{end}}
    }{{if .ReservedComments}} {{.ReservedComments}}{{end}};
{{end}}
    `

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
