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

package templates

// Client .
var Client = `
{{define "ThriftClient"}}
{{- if not Features.NoProcessor}}
{{- UseStdLibrary "thrift"}}
{{- $BasePrefix := ServicePrefix .Base}}
{{- $BaseService := ServiceName .Base}}
{{- $ServiceName := .GoName}}
{{- $ClientName := printf "%s%s" $ServiceName "Client"}}
type {{$ClientName}} struct {
	{{- if .Extends}}
	*{{$BasePrefix}}{{$BaseService}}Client
	{{- else}}
	c thrift.TClient
	{{- end}}
}

func New{{$ClientName}}Factory(t thrift.TTransport, f thrift.TProtocolFactory) *{{$ClientName}} {
	return &{{$ClientName}}{
		{{- if .Extends}}
			{{$BaseService}}Client: {{$BasePrefix}}New{{$BaseService}}ClientFactory(t, f),
		{{- else}}
			c: thrift.NewTStandardClient(f.GetProtocol(t), f.GetProtocol(t)),
		{{- end}}
	}
}

func New{{$ClientName}}Protocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *{{$ClientName}} {
	return &{{$ClientName}}{
		{{- if .Extends}}
			{{$BaseService}}Client: {{$BasePrefix}}New{{$BaseService}}ClientProtocol(t, iprot, oprot),
		{{- else}}
			c: thrift.NewTStandardClient(iprot, oprot),
		{{- end}}
	}
}

func New{{$ClientName}}(c thrift.TClient) *{{$ClientName}}{
	return &{{$ClientName}}{
		{{- if .Extends}}
			{{$BaseService}}Client: {{$BasePrefix}}New{{$BaseService}}Client(c),
		{{- else}}
			c: c,
		{{- end}}
	}
}

{{if not .Extends}}
func (p *{{$ClientName}}) Client_() thrift.TClient {
	return p.c
}
{{end}}

{{- range .Functions}}
{{- $Function := .}} 
{{- $ArgType := .ArgType}} 
{{- $ResType := .ResType}} 
func (p *{{$ClientName}}) {{- template "FunctionSignature" . -}} {
	{{if .Streaming.IsStreaming -}}
	panic("streaming method {{$ServiceName}}.{{.Name}}(mode = {{.Streaming.Mode}}) not available, please use Kitex Thrift Streaming Client.")
	{{else -}}
	var _args {{$ArgType.GoName}}
	{{- range .Arguments}}
	_args.{{($ArgType.Field .Name).GoName}} = {{.GoName}}
	{{- end}}

	{{- if .Void}}
	{{- if .Oneway}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, nil); err != nil {
		return
	}
	{{- else}}
	var _result {{$ResType.GoName}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, &_result); err != nil {
		return
	}
	{{- if .Throws}}
	switch {
	{{- range .Throws}}
	case _result.{{($ResType.Field .Name).GoName}} != nil:
		return _result.{{($ResType.Field .Name).GoName}}
	{{- end}}
	}
	{{- end}}

	{{- end}}
	return nil
	{{- else}}{{/* If .Void */}}
	var _result {{$ResType.GoName}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, &_result); err != nil {
		return
	}
	{{- if .Throws}}
	switch {
	{{- range .Throws}}
	case _result.{{($ResType.Field .Name).GoName}} != nil:
		return r, _result.{{($ResType.Field .Name).GoName}}
	{{- end}}
	}
	{{- end}}
	return _result.GetSuccess(), nil
	{{- end}}{{/* If .Void */}}
	{{- end}}{{/* If .Streaming.IsStreaming */ -}}
}
{{- end}}{{/* range .Functions */}}
{{- end}}{{/* if not Features.NoProcessor */}}

{{- range .Functions}}
{{- if or .Streaming.ClientStreaming .Streaming.ServerStreaming}}
{{- $arg := index .Arguments 0}}
{{- if Features.StreamX}}{{/* StreamX */}}
{{- UseStdLibrary "streaming" -}}
{{- if and .Streaming.ClientStreaming .Streaming.ServerStreaming}}
type {{.Service.GoName}}_{{.Name}}Server streaming.BidiStreamingServer[{{NotPtr $arg.GoTypeName}},{{NotPtr .ResponseGoTypeName}}]
{{- else if .Streaming.ClientStreaming}}
type {{.Service.GoName}}_{{.Name}}Server streaming.ClientStreamingServer[{{NotPtr $arg.GoTypeName}},{{NotPtr .ResponseGoTypeName}}]
{{- else}}
type {{.Service.GoName}}_{{.Name}}Server streaming.ServerStreamingServer[{{NotPtr .ResponseGoTypeName}}]
{{- end}}
{{- else}}
type {{.Service.GoName}}_{{.Name}}Server interface {
	{{- UseStdLibrary "streaming" -}}
	streaming.Stream
	{{if .Streaming.ClientStreaming }}
	Recv() ({{$arg.GoTypeName}}, error)
	{{end}}
	{{if .Streaming.ServerStreaming}}
	Send({{.ResponseGoTypeName}}) error
	{{end}}
	{{if and .Streaming.ClientStreaming (not .Streaming.ServerStreaming) }}
	SendAndClose({{.ResponseGoTypeName}}) error
	{{end}}
}
{{- end}}{{/* StreamX */}}
{{- end}}{{/* Streaming */}}
{{- end}}{{/* range .Functions */}}
{{- end}}{{/* define "ThriftClient" */}}
`
