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

// Extension.
func Extension() []string {
	return []string{
		StructLikeRead,
		StructLikeReadField,
		StructLikeWrite,
		StructLikeWriteField,
		Client,
		Processor,
	}
}

// Substitutions.
// Because text.Templates will not substitute an existing template with an empty one,
// we use insertion points to walk around this problem and achieve deleting templates.
var (
	StructLikeRead       = `{{define "StructLikeRead"}}{{InsertionPoint "slim1"}}{{end}}`
	StructLikeReadField  = `{{define "StructLikeReadField"}}{{InsertionPoint "slim2"}}{{end}}`
	StructLikeWrite      = `{{define "StructLikeWrite"}}{{InsertionPoint "slim3"}}{{end}}`
	StructLikeWriteField = `{{define "StructLikeWriteField"}}{{InsertionPoint "sim4"}}{{end}}`
	Client               = `{{define "Client"}}{{InsertionPoint "slim5"}}{{end}}`
	Processor            = `
{{define "Processor"}}
{{- range .Functions}}
{{$ArgsType := .ArgType}}
{{template "StructLike" $ArgsType}}
{{- if not .Oneway}}
	{{$ResType := .ResType}}
	{{template "StructLike" $ResType}}
{{- end}}
{{- end}}{{/* range .Functions */}}
{{- end}}{{/* define "Processor" */}}
	`
)
