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

// File .
var File = `// Code generated by thriftgo ({{Version}}). DO NOT EDIT.
{{InsertionPoint "bof"}}

package {{.FilePackage}}

import (
	{{InsertionPoint "imports"}}
	{{- if Features.GenerateReflectionInfo}}thriftreflection "github.com/cloudwego/kitex/pkg/reflection/thrift"{{end}}
)

{{template "Constant" .}}

{{- range .Enums}}
{{template "Enum" .}}
{{- end}}

{{- range .Typedefs}}
{{template "Typedef" .}}
{{- end}}

{{- range .Structs}}
{{template "StructLike" .}}
{{- end}}

{{- range .Unions}}
{{template "StructLike" .}}
{{- end}}

{{- range .Exceptions}}
{{template "StructLike" .}}
{{- end}}

{{- range .Services}}
{{template "Service" .}}
{{template "Client" .}}
{{- end}}

{{- range .Services}}
{{template "Processor" .}}
{{- end}}

{{- if Features.GenerateReflectionInfo}}
	var file_{{.IDLName}}_rawDesc = {{.IDLMeta}}
	func init(){
		thriftreflection.RegisterIDL(file_{{.IDLName}}_rawDesc)
	}
{{end}}
{{- InsertionPoint "eof"}}
`
