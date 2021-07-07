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
{{- $Service := .First}}
{{- $Function := .Second}}
{{- with .Second}}
{{- Identify .Name $Service}}(ctx context.Context
	{{- range .Arguments -}}
		, {{GetParamName $Function .Name}} {{.Type | ResolveTypeName}}
	{{- end -}}
		) (
	{{- if not .Void}}r {{.FunctionType | ResolveTypeName}}, {{- end -}}
		err error)
{{- end}}{{/* with .Second */}}
{{- end}}{{/* define "FunctionSignature" */}}
`

// Service .
var Service = `
{{define "Service"}}
{{$BasePrefix := BaseServicePrefix .}}
{{$BaseService := .Extends | GetServiceIdentifier}}
{{$ServiceName := .Name | Identify}}
{{$ClientName := printf "%s%s" $ServiceName "Client"}}
{{InsertionPoint "service" .Name}}
type {{$ServiceName}} interface {
	{{- if .Extends}}
	{{$BasePrefix}}{{$BaseService}}
	{{- end}}
	{{- range .Functions}}
	{{InsertionPoint "service" $.Name .Name}}
	{{template "FunctionSignature" (Pair $ .)}}
	{{- end}}
}

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
{{- $ArgType := GetArgType $ .}} 
{{- $ResType := GetResType $ .}} 
func (p *{{$ClientName}}) {{- template "FunctionSignature" (Pair $ .) -}} {
	var _args {{$ArgType.Name | Identify}}
	{{- range .Arguments}}
	_args.{{Identify .Name $ArgType}} = {{GetParamName $Function .Name}}
	{{- end}}

	{{- if .Void}}
	{{- if .Oneway}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, nil); err != nil {
		return
	}
	{{- else}}
	var _result {{$ResType.Name | Identify}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, &_result); err != nil {
		return
	}
	{{- end}}
	return nil
	{{- else}}{{/* If .Void */}}
	var _result {{$ResType.Name | Identify}}
	if err = p.Client_().Call(ctx, "{{.Name}}", &_args, &_result); err != nil {
		return
	}
	{{- if .Throws}}
	switch {
	{{- range .Throws}}
	case _result.{{Identify .Name $ResType}} != nil:
		return r, _result.{{Identify .Name $ResType}}
	{{- end}}
	}
	{{- end}}
	return _result.GetSuccess(), nil
	{{- end}}{{/* If .Void */}}
}
{{end}}{{/* range .Functions */}}
{{- end}}{{/* define "Service" */}}
`
