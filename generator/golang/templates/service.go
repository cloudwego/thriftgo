// Copyright 2021 CloudWeGo Authors
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

package templates

// FunctionSignature .
var FunctionSignature = `
{{define "FunctionSignature"}}
{{- $Function := .}}
{{- if or .Streaming.ClientStreaming .Streaming.ServerStreaming}}
	{{- $arg := index .Arguments 0}}
	{{- .GoName}}(
	{{- if and .Streaming.ServerStreaming (not .Streaming.ClientStreaming) -}}
		req *{{$arg.Type}}, 
	{{- end -}}
		stream {{.Service.GoName}}_{{.Name}}Server) (err error)
{{- else -}}
	{{- UseStdLibrary "context" -}}
	{{- .GoName}}(ctx context.Context
	{{- range .Arguments -}}
		, {{.GoName}} {{.GoTypeName}}
	{{- end -}}
		) (
	{{- if not .Void}}r {{.ResponseGoTypeName}}, {{- end -}}
		err error)
{{- end -}}{{/* end if streaming */}}
{{- end}}{{/* define "FunctionSignature" */}}
`

// Service .
var Service = `
{{define "ThriftService"}}
{{- $BasePrefix := ServicePrefix .Base}}
{{- $BaseService := ServiceName .Base}}
{{- $ServiceName := .GoName}}
{{InsertionPoint "service" .Name}}
{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
type {{$ServiceName}} interface {
	{{- if .Extends}}
	{{$BasePrefix}}{{$BaseService}}
	{{- end}}
	{{- range .Functions}}
	{{InsertionPoint "service" $.Name .Name}}
	{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
	{{template "FunctionSignature" .}}
	{{- end}}
}
{{- end}}{{/* define "ThriftService" */}}
`
