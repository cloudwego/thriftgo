// Copyright 2022 CloudWeGo Authors
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

package slim

// Extension .
func Extension() []string {
	return []string{
		StructLike,
		Client,
		Processor,
	}
}

// Substitutions.
// Because text.Templates will not substitute an existing template with an empty one,
// we use insertion points to walk around this problem and achieve deleting templates.
var (
	StructLike = `
{{define "StructLike"}}
{{- $TypeName := .GoName}}
{{InsertionPoint .Category .Name}}
{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
type {{$TypeName}} struct {
{{- range .Fields}}
	{{- InsertionPoint $.Category $.Name .Name}}
	{{- if and Features.ReserveComments .ReservedComments}}
	{{.ReservedComments}}
	{{- end}}
	{{(.GoName)}} {{.GoTypeName}} {{GenFieldTags . (InsertionPoint $.Category $.Name .Name "tag")}} 
{{- end}}
	{{- if Features.KeepUnknownFields}}
	{{- UseStdLibrary "unknown"}}
	_unknownFields unknown.Fields
	{{- end}}
}

{{- if Features.GenerateTypeMeta }}
{{- UseStdLibrary "meta"}}
func init() {
	meta.RegisterStruct(New{{$TypeName}}, {{Marshal .}})
}
{{- end}}{{/* if Features.GenerateTypeMeta */}}

func New{{$TypeName}}() *{{$TypeName}} {
	return &{{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}

{{template "FieldGetOrSet" .}}

{{if eq .Category "union"}}
func (p *{{$TypeName}}) CountSetFields{{$TypeName}}() int {
	count := 0
	{{- range .Fields}}
	{{- if SupportIsSet .Field}}
	if p.{{.IsSetter}}() {
		count++
	}
	{{- end}}
	{{- end}}
	return count
}
{{- end}}

{{if Features.KeepUnknownFields}}
func (p *{{$TypeName}}) CarryingUnknownFields() bool {
	return len(p._unknownFields) > 0
}
{{end}}{{/* if Features.KeepUnknownFields */}}

{{template "FieldIsSet" .}}

func (p *{{$TypeName}}) String() string {
	if p == nil {
		return "<nil>"
	}
	{{- UseStdLibrary "fmt"}}
	return fmt.Sprintf("{{$TypeName}}(%+v)", *p)
}

{{- if eq .Category "exception"}}
func (p *{{$TypeName}}) Error() string {
	return p.String()
}
{{- end}}

{{- end}}{{/* define "StructLike" */}}
	`

	Client = `
{{define "Client"}}
{{InsertionPoint "slim.Client"}}
{{end}}{{/* define "Client" */}}`

	Processor = `
{{define "Processor"}}
{{InsertionPoint "slim.Processor"}}
{{$throws := ServiceThrows .}}
{{- if $throws}}
// exceptions of methods in {{.GoName}}.
var (
{{- range $throws}}
_ error = ({{.GoTypeName}})(nil)
{{- end}}{{/* range $throws */}}
)
{{- end}}{{/* if $throws */}}
{{- end}}{{/* define "Processor" */}}
`
)
