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

// Enum .
var Enum = `
{{define "Enum"}}
{{- $EnumType := .GoName}}
{{InsertionPoint "enum" .Name}}
{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
type {{$EnumType}} int{{if Features.EnumAsINT32}}32{{else}}64{{end}}

const (
	{{- range .Values}}
	{{- if and Features.ReserveComments .ReservedComments}}
	{{.ReservedComments}}{{end}}
	{{.GoName}} {{$EnumType}} = {{.Value}}
	{{- end}}
)

func (p {{$EnumType}}) String() string {
	switch p {
	{{- range .Values}}
	case {{.GoName}}:
		return "{{.GoLiteral}}"
	{{- end}}
	}
	return "<UNSET>"
}

func {{$EnumType}}FromString(s string) ({{$EnumType}}, error) {
	switch s {
	{{- range .Values}}
	case "{{.GoLiteral}}":
		return {{.GoName}}, nil
	{{- end}}
	}
	{{- UseStdLibrary "fmt"}}
	return {{$EnumType}}(0), fmt.Errorf("not a valid {{$EnumType}} string")
}

func {{$EnumType}}Ptr(v {{$EnumType}} ) *{{$EnumType}}  { return &v }

{{- if or Features.MarshalEnumToText Features.MarshalEnum}}

func (p {{$EnumType}}) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

{{end}}{{/* if or Features.MarshalEnumToText Features.MarshalEnum */}}

{{- if or Features.MarshalEnumToText Features.UnmarshalEnum}}

func (p *{{$EnumType}}) UnmarshalText(text []byte) error {
	q, err := {{$EnumType}}FromString(string(text))
	if err != nil {
		return err
	}
	*p = q
	return nil
}
{{end}}{{/* if or Features.MarshalEnumToText Features.UnmarshalEnum */}}

{{- if Features.ScanValueForEnum}}
{{- UseStdLibrary "sql" "driver"}}
func (p *{{$EnumType}}) Scan(value interface{}) (err error) {
	var result sql.NullInt{{if Features.EnumAsINT32}}32{{else}}64{{end}}
	err = result.Scan(value)
	*p = {{$EnumType}}(result.Int{{if Features.EnumAsINT32}}32{{else}}64{{end}})
	return
}

func (p *{{$EnumType}}) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return int{{if Features.EnumAsINT32}}32{{else}}64{{end}}(*p), nil
}
{{- end}}{{/* if .Features.ScanValueForEnum */}}
{{end}}
`
