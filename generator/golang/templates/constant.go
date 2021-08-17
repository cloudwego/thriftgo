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

// Constant .
var Constant = `
{{define "Constant"}}
{{- $Consts := .Constants.GoConstants}}
{{- if $Consts}}
const (
	{{- range $Consts}}
	{{InsertionPoint "constant" .Name}}
	{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
	{{.GoName}} = {{.Initialization}}
	{{- end}}{{/* range $Consts */}}
	{{InsertionPoint "constants"}}
)
{{- end}}{{/* fi $Consts */}}

{{- $NonConsts := .Constants.GoVariables}}
{{- if $NonConsts}}
var (
	{{- range $NonConsts }}
	{{InsertionPoint "constant" .Name }}
	{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
	{{.GoName}} = {{.Initialization}}
	{{- end}}
	{{InsertionPoint "variables"}}
)
{{- end}}
{{end}}{{- /* define "Constant" */ -}}
`
