package dump

const TypeTemplate = `
{{define "Type"}}
{{- if eq .Name "list" -}}
list<{{- template "Type" .ValueType}}>
{{- else if eq .Name "map" -}}
map<{{- template "Type" .KeyType}}, {{- template "Type" .ValueType}}>
{{- else if eq .Name "set" -}}
set<{{- template "Type" .ValueType}}>
{{- else -}}
{{.Name}}
{{- end}}
{{- template "Annotations" .Annotations -}}
{{end}}
`
const TypeDefTemplate = `
{{define "Typedef"}}
{{- if .ReservedComments -}}{{- .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "typedef"}} {{template "Type" .Type}} {{.Alias}} {{template "Annotations" .Annotations}}
{{- end -}}
`
const ConstValueTemplate = `
{{define "ConstValue"}}
{{- if .TypedValue.Double}}{{.TypedValue.Double}}{{end}}
{{- if .TypedValue.Int}}{{.TypedValue.Int}}{{end}}
{{- if .TypedValue.Literal}}{{.TypedValue.Literal}}{{end}}
{{- if .TypedValue.Identifier}}{{.TypedValue.Identifier}}{{end}}
{{- if .TypedValue.List}}{{.TypedValue.List}}{{end}}
{{- if .TypedValue.Map}}{{.TypedValue.Map}}{{end}}
{{- end -}}
`
const ConstantTemplate = `
{{define "Constant"}}
{{- if .ReservedComments -}}{{- .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "const"}} {{template "Type" .Type}} {{.Name}} {{"= "}}
{{- template "ConstValue" .Value -}}
{{- template "Annotations" .Annotations -}}
{{- end -}}
`
const EnumTemplate = `
{{define "Enum"}}
{{- if .ReservedComments -}}{{- .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "enum"}} {{.Name}} {
	{{range .Values}}
	{{- if .ReservedComments -}}{{- .ReservedComments -}}{{- end}}
	{{.Name}} = {{.Value}} {{template "Annotations" .Annotations -}}
	{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`
const FieldTemplate = `
{{define "Field"}}
{{- if .ReservedComments}}{{"    "}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{"    "}}{{.ID}}{{":"}}
{{- if .Requiredness.IsRequired}} required
{{- else if .Requiredness.IsOptional}} optional{{end}}{{" "}}
{{- template "Type" .Type}} {{.Name}}
{{- if .Default}} = {{template "ConstValue" .Default -}}{{- end -}}
{{- template "Annotations" .Annotations -}}
{{- end -}}
`
const StructTemplate = `
{{define "Struct"}}
{{- if .ReservedComments}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{- "struct"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`
const UnionTemplate = `
{{define "Union"}}
{{- if .ReservedComments}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{- "union"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`
const ExceptionTemplate = `
{{define "Exception"}}
{{- if .ReservedComments}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{- "exception"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`
const ServiceTemplate = `
{{define "Service"}}
{{- if .ReservedComments}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{- "service"}} {{.Name}}{{if .Extends}} extends {{.Extends}}{{end}} {
{{- range .Functions}}
{{template "Function" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`
const FunctionTemplate = `
{{define "Function"}}
{{- if .ReservedComments}}{{"    "}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{"    "}}{{if .Oneway}}oneway {{end}}{{template "Type" .FunctionType}} {{.Name}}(
{{- range .Arguments}}
{{- template "SingleLineField" .}}{{", "}}
{{- end}})
{{- if .Throws}}
{{- "throws" }} (
{{- range .Throws}}
{{- template "SingleLineField" .}}{{", "}}
{{- end}})
{{- end}}
{{- end -}}
`

// SingleLineFieldTemplate for args and throws of functions
const SingleLineFieldTemplate = `
{{define "SingleLineField"}}
{{- if .ReservedComments}}{{"    "}}{{.ReservedComments}}{{"\n"}}{{end -}}
{{.ID}}{{":"}}
{{- if .Requiredness.IsRequired}} required
{{- else if .Requiredness.IsOptional}} optional{{end}}{{" "}}
{{- template "Type" .Type}}{{- template "Annotations" .Type.Annotations }} {{.Name}}
{{- if .Default}} = {{template "ConstValue" .Default -}}{{- end -}}
{{- template "Annotations" .Annotations -}}
{{- end -}}
`
const AnnotationsTemplate = `
{{define "Annotations"}}
{{- if . -}}(
{{- range .}}
{{- $key := .Key -}}
{{- range .Values -}}
 {{$key}} = "{{.}}", {{- end -}}
{{- end -}}){{- end -}}
{{- end -}}
`
