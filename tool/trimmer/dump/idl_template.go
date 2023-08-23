// Copyright 2023 CloudWeGo Authors
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

package dump

const IDLTemplate = `
{{- /* INCLUDES */}}
{{- if .Includes}}
{{- range .Includes}}include "{{.Path}}"
{{end -}}
{{"\n"}}
{{- end}}

{{- /* NAMESPACES */}}
{{- if .Namespaces}}
{{- range .Namespaces}}namespace {{.Language}} {{.Name}} {{- template "Annotations" .Annotations}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* CPP INCLUDES */}}
{{- if .CppIncludes}}
{{- range .CppIncludes}}cpp_include "{{.}}"
{{end -}}
{{"\n"}}
{{- end}}

{{- /* TYPEDEF */}}
{{- if .Typedefs}}
{{- range .Typedefs}}
{{- template "Typedef" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* CONST */}}
{{- if .Constants}}
{{- range .Constants}}
{{- template "Constant" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* ENUM */}}
{{- if .Enums}}
{{- range .Enums}}
{{- template "Enum" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* STRUCT */}}
{{- if .Structs}}
{{- range .Structs}}
{{- template "Struct" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* UNION */}}
{{- if .Unions}}
{{- range .Unions}}
{{- template "Union" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* EXCEPTIONS */}}
{{- if .Exceptions}}
{{- range .Exceptions}}
{{- template "Exception" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* SERVICES */}}
{{- if .Services}}
{{- range .Services}}
{{- template "Service" .}}
{{end -}}
{{"\n"}}
{{- end}}

{{- /* end of the file */ -}}
`
