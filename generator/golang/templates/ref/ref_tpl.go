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

package ref

var FileRef = `// Code generated by thriftgo ({{Version}}). DO NOT EDIT.
{{InsertionPoint "bof"}}
package {{.FilePackage}}
{{- $RefPackage := .RefPackage}}
import (
	{{InsertionPoint "imports"}}
	{{define "Imports"}}
	{{end}}
	{{.RefPackage}} "{{.RefPath}}"
)

` + constRef + `

{{- range .Structs}}
` + structRef + `
{{- end}}

` + enumRef + `

` + typedefRef + `

{{- range .Unions}}
` + structRef + `
{{- end}}

{{- range .Exceptions}}
` + structRef + `
{{- end}}

{{- range .Services}}
` + processorRef + `
{{- end}}

{{- InsertionPoint "eof"}}
`

var processorRef = `

{{- $BasePrefix := ServicePrefix .Base}}
{{- $BaseService := ServiceName .Base}}
{{- $ServiceName := .GoName}}

type {{$ServiceName}} = {{$RefPackage}}.{{$ServiceName}}


{{- $ClientName := printf "%s%s" $ServiceName "Client"}}
type {{$ClientName}} = {{$RefPackage}}.{{$ClientName}}

var New{{$ClientName}}Factory = {{$RefPackage}}.New{{$ClientName}}Factory

var New{{$ClientName}}Protocol = {{$RefPackage}}.New{{$ClientName}}Protocol

var New{{$ClientName}} = {{$RefPackage}}.New{{$ClientName}}

{{- $ProcessorName := printf "%s%s" $ServiceName "Processor"}}
type {{$ProcessorName}} = {{$RefPackage}}.{{$ProcessorName}}

var New{{$ProcessorName}} = {{$RefPackage}}.New{{$ProcessorName}}

{{- range .Functions}}
{{$ArgsType := .ArgType.GoName}}
type {{$ArgsType}} = {{$RefPackage}}.{{$ArgsType}}
var New{{$ArgsType}} = {{$RefPackage}}.New{{$ArgsType}}
	{{- range .ArgType.Fields}}
		{{- $FieldName := .GoName}}
		{{if SupportIsSet .Field}}
		{{$DefaultVarName := printf "%s_%s_%s" $ArgsType $FieldName "DEFAULT"}}
		var {{$DefaultVarName}} = {{$RefPackage}}.{{$DefaultVarName}}
		{{- end}}	
	{{- end}}
{{- if not .Oneway}}
{{$ResType := .ResType.GoName}}
type {{$ResType}} = {{$RefPackage}}.{{$ResType}}
var New{{$ResType}} = {{$RefPackage}}.New{{$ResType}}
	{{- range .ResType.Fields}}
		{{- $FieldName := .GoName}}
		{{if SupportIsSet .Field}}
		{{$DefaultVarName := printf "%s_%s_%s" $ResType $FieldName "DEFAULT"}}
		var {{$DefaultVarName}} = {{$RefPackage}}.{{$DefaultVarName}}
		{{- end}}	
	{{- end}}
{{- end}}

{{- end}}{{/* range .Functions */}}

`

var constRef = `{{- $Consts := .Constants.GoConstants}}
{{- if $Consts}}
const (
	{{- range $Consts}}
	{{.GoName}} = {{$RefPackage}}.{{.GoName}} 
	{{- end}}{{/* range $Consts */}}
)
{{- end}}

{{- $NonConsts := .Constants.GoVariables}}
{{- if $NonConsts}}
var (
	{{- range $NonConsts }}
	{{.GoName}} = {{$RefPackage}}.{{.GoName}} 
	{{- end}}
)
{{- end}}
`

var structRef = `
	{{- $TypeName := .GoName}}
	type {{$TypeName}}= {{$RefPackage}}.{{$TypeName}}
	var New{{$TypeName}} = {{$RefPackage}}.New{{$TypeName}}
	{{- range .Fields}}
		{{- $FieldName := .GoName}}
		{{- $DefaultVarTypeName := .DefaultTypeName}}
		{{if SupportIsSet .Field}}
		{{$DefaultVarName := printf "%s_%s_%s" $TypeName $FieldName "DEFAULT"}}
		{{- if Features.CodeRefSlim }}
			
		
		{{- else }}
			var {{$DefaultVarName}} = {{$RefPackage}}.{{$DefaultVarName}}
		{{- end }}
		{{- end}}	
	{{- end}}
`

var enumRef = `
{{- range .Enums}}
	{{- $EnumType := .GoName}}
	{{- $TypeName := .GoName}}
	type {{$TypeName}}= {{$RefPackage}}.{{$TypeName}}
	var {{$EnumType}}FromString = {{$RefPackage}}.{{$EnumType}}FromString
	var {{$EnumType}}Ptr = {{$RefPackage}}.{{$EnumType}}Ptr
	{{- if Features.CodeRefSlim }}
	const (
		{{- range .Values}}
		{{- if and Features.ReserveComments .ReservedComments}}
		{{.ReservedComments}}{{end}}
		{{.GoName}} {{$EnumType}} = {{.Value}}
		{{- end}}
	)
	{{- else }}
	const (
		{{- range .Values}}
		{{.GoName}} = {{$RefPackage}}.{{.GoName}}
		{{- end}}
	)
	{{- end }}
{{- end}}
`

var typedefRef = `
{{- range .Typedefs}}
	{{- $NewTypeName := .GoName}}
	type {{$NewTypeName}}= {{$RefPackage}}.{{$NewTypeName}}
	{{if .Type.Category.IsStructLike}} 
	var New{{$NewTypeName}} = {{$RefPackage}}.New{{$NewTypeName}}
	{{- end}}
{{- end}}
`
