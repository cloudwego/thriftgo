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
