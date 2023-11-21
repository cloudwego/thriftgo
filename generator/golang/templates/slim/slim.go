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
	{{- if .IsNested}}
		{{.GoTypeName}} {{GenFieldTags . (InsertionPoint $.Category $.Name .Name "tag")}}
	{{else}}
		{{(.GoName)}} {{.GoTypeName}} {{GenFieldTags . (InsertionPoint $.Category $.Name .Name "tag")}}
	{{- end}}
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

{{if Features.FrugalTag}}
func (p *{{$TypeName}}) InitDefault() {
	*p = {{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}
{{end}}{{/* if Features.FrugalTag */}}

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
	{{- if Features.JSONStringer}}
	{{- UseStdLibrary "json_utils"}}
	JsonBytes , _  := json_utils.JSONFunc(p)
	return string(JsonBytes)	
	{{- else}}
	if p == nil {
		return "<nil>"
	}
	{{- UseStdLibrary "fmt"}}
	return fmt.Sprintf("{{$TypeName}}(%+v)", *p)
	{{- end}}
}

{{- if eq .Category "exception"}}
func (p *{{$TypeName}}) Error() string {
	return p.String()
}
{{- end}}

{{- end}}{{/* define "StructLike" */}}
	`

	Client = `
{{define "ThriftClient"}}
{{InsertionPoint "slim.Client"}}
{{- range .Functions}}
{{- if or .Streaming.ClientStreaming .Streaming.ServerStreaming}}
{{- $arg := index .Arguments 0}}
{{- $ResponseType := .FunctionType.Name}}
type {{.Service.GoName}}_{{.Name}}Server interface {
	{{- UseStdLibrary "streaming" -}}
	streaming.Stream
	{{if .Streaming.ClientStreaming }}
	Recv() (*{{$arg.Type.Name}}, error)
	{{end}}
	{{if .Streaming.ServerStreaming}}
	Send(*{{$ResponseType}}) error
	{{end}}
	{{if and .Streaming.ClientStreaming (not .Streaming.ServerStreaming) }}
	SendAndClose(*{{$ResponseType}}) error
	{{end}}
}
{{- end}}{{/* Streaming */}}
{{- end}}{{/* range .Functions */}}
{{end}}{{/* define "ThriftClient" */}}`

	Processor = `
{{define "ThriftProcessor"}}
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
{{- end}}{{/* define "ThriftProcessor" */}}
`
)
