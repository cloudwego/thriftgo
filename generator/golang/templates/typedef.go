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

// Typedef .
var Typedef = `
{{define "Typedef"}}
{{- $NewTypeName := .GoName}}
{{- $OldTypeName := .GoTypeName}}
{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
{{- if Features.TypedefAsTypeAlias }}
type {{$NewTypeName}} = {{$OldTypeName}}
{{- else}}
type {{$NewTypeName}} {{$OldTypeName}}
{{- end}}

{{if .Type.Category.IsStructLike}} 
func New{{$NewTypeName}}() *{{$NewTypeName}} {
	return (*{{$NewTypeName}})({{$OldTypeName.NewFunc}}())
}
{{- end}}{{/* if .Type.Category.IsStructLike */}} 
{{- end}}{{/* define "Typedef" */}}
`
