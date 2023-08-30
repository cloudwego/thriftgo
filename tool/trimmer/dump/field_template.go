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
	"fmt"
	"strings"
)

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
{{- if .ReservedComments -}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "typedef"}} {{template "Type" .Type}} {{.Alias}} {{template "Annotations" .Annotations}}
{{- end -}}
`

const ConstValueTemplate = `
{{define "ConstValue"}}
{{- if .TypedValue.Double}}{{.TypedValue.Double}}{{end}}
{{- if .TypedValue.Int}}{{.TypedValue.Int}}{{end}}
{{- if .TypedValue.Literal}}"{{.TypedValue.Literal}}"{{end}}
{{- if .TypedValue.Identifier}}{{.TypedValue.Identifier}}{{end}}
{{- if .TypedValue.IsSetList}}{{template "ConstList" .TypedValue.List}}{{end}}
{{- if .TypedValue.IsSetMap}}{{template "ConstMap" .TypedValue.Map}}{{end}}
{{- end -}}
`

const ConstListTemplate = `
{{define "ConstList"}}
{{- "[" -}}
{{- range $index, $element := .}}
{{- if $index}}, {{end -}}{{- template "ConstValue" $element -}}
{{- end -}}
{{- "]" -}}
{{- end -}}
`

const ConstMapTemplate = `
{{define "ConstMap"}}
{{- "{" -}}
{{- range $index, $element := . -}}
{{- if $index}}, {{end -}}{{- template "ConstValue" .Key -}}:{{- template "ConstValue" .Value -}}
{{- end -}}
{{- "}" -}}
{{- end -}}
`

const ConstantTemplate = `
{{define "Constant"}}
{{- if .ReservedComments -}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "const"}} {{template "Type" .Type}} {{.Name}} {{"= "}}
{{- template "ConstValue" .Value -}}
{{- template "Annotations" .Annotations -}}
{{- end -}}
`

const EnumTemplate = `
{{define "Enum"}}
{{- if .ReservedComments -}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{- end -}}
{{- "enum"}} {{.Name}} {
	{{- range .Values}}
	{{- if .ReservedComments -}}{{- ReplaceQuotes .ReservedComments -}}{{- end}}
	{{.Name}} = {{.Value}} {{template "Annotations" .Annotations -}}
	{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`

const FieldTemplate = `
{{define "Field"}}
{{- if .ReservedComments}}{{"    "}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
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
{{- if .ReservedComments}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
{{- "struct"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`

const UnionTemplate = `
{{define "Union"}}
{{- if .ReservedComments}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
{{- "union"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`

const ExceptionTemplate = `
{{define "Exception"}}
{{- if .ReservedComments}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
{{- "exception"}} {{.Name}} {
{{- range .Fields}}
{{template "Field" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`

const ServiceTemplate = `
{{define "Service"}}
{{- if .ReservedComments}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
{{- "service"}} {{.Name}}{{if .Extends}} extends {{.Extends}}{{end}} {
{{- range .Functions}}
{{template "Function" .}}
{{- end}}
} {{template "Annotations" .Annotations -}}{{"\n"}}
{{- end -}}
`

const FunctionTemplate = `
{{define "Function"}}
{{- if .ReservedComments}}{{"    "}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
{{"    "}}{{if .Oneway}}oneway {{end}}{{template "Type" .FunctionType}} {{.Name}}(
{{- range $index, $element := .Arguments}}
{{- if $index}}, {{end -}}{{- template "SingleLineField" .}}
{{- end}})
{{- if .Throws}}
{{- "throws" }} (
{{- range $index, $element := .Throws}}
{{- if $index}},{{end -}}{{- template "SingleLineField" .}}
{{- end}})
{{- end}}
{{- end -}}
`

// SingleLineFieldTemplate for args and throws of functions
const SingleLineFieldTemplate = `
{{define "SingleLineField"}}
{{- if .ReservedComments}}{{"    "}}{{- ReplaceQuotes .ReservedComments -}}{{"\n"}}{{end -}}
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
{{- if . -}}
{{- $result := "" -}}
{{- range .}}
{{- $key := .Key -}}
{{- range .Values -}}
{{ $value := JoinQuotes . }}
{{- $result = printf "%s%s = %s, " $result $key $value -}}{{- end -}}
{{- end -}}
	({{- $result | RemoveLastComma -}})
{{- end -}}
{{- end -}}
`

func RemoveLastComma(s string) string {
	return strings.TrimRight(s, ", ")
}

func JoinQuotes(s string) string {
	return fmt.Sprintf("%s", "#OUTQUOTES"+s+"#OUTQUOTES")
}

func ReplaceQuotes(s string) string {
	out := strings.Replace(s, "\"", "#OUTQUOTES", -1)
	return out
}
