// Copyright 2021 CloudWeGo
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

// StructLike is the code template for struct, union, and exception.
var StructLike = `
{{define "StructLike"}}
{{- $TypeName := .Name | Identify}}
{{InsertionPoint .Category .Name}}
type {{$TypeName}} struct {
{{- range .Fields}}
	{{InsertionPoint $.Category $.Name .Name}}
	{{ResolveFieldName .}} {{ResolveFieldTypeName .}} {{GenTags . (InsertionPoint $.Category $.Name .Name "tag")}} 
{{- end}}
	{{if Features.KeepUnknownFields}}_unknownFields unknown.Fields{{end}}
}

func New{{$TypeName}}() *{{$TypeName}} {
	return &{{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}

{{template "FieldGetOrSet" .}}

{{if eq .Category "union"}}
func (p *{{$TypeName}}) CountSetFields{{$TypeName}}() int {
	count := 0
	{{- range .Fields}}
	{{- if SupportIsSet .}}
	if p.IsSet{{ResolveFieldName .}}() {
		count++
	}
	{{- end}}
	{{- end}}
	return count
}
{{- end}}

{{if Features.KeepUnknownFields}}
func (p *{{$TypeName}}) CarryingUnknownFields() bool {
	return len(p._unknownFields) > 0
}
{{end}}{{/* if Features.KeepUnknownFields */}}

var fieldIDToName_{{$TypeName}} = map[int16]string{
{{- range .Fields}}
	{{.ID}}: "{{.Name}}",
{{- end}}{{/* range .Fields */}}
}

{{template "FieldIsSet" .}}

{{template "StructLikeRead" .}}

{{template "StructLikeReadField" .}}

{{template "StructLikeWrite" .}}

{{template "StructLikeWriteField" .}}

func (p *{{$TypeName}}) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{{$TypeName}}(%+v)", *p)
}

{{- if eq .Category "exception"}}
func (p *{{$TypeName}}) Error() string {
	return p.String()
}
{{- end}}

{{- if Features.GenDeepEqual}}
{{- $st := ToStructLike .}}
{{template "StructLikeDeepEqual" $st}}

{{template "StructLikeDeepEqualField" $st}}
{{- end}}

{{- end}}{{/* define "StructLike" */}}
`

// StructLikeDefault is the code template for structure initialization.
var StructLikeDefault = `
{{- define "StructLikeDefault"}}
{{- range .Fields}}
	{{- if HasDefaultValue .}}
		{{ResolveFieldName .}}: {{GetFieldInit .}},
	{{- end}}
{{- end}}
{{- end -}}`

// StructLikeRead .
var StructLikeRead = `
{{define "StructLikeRead"}}
{{- $TypeName := .Name | Identify -}}
func (p *{{$TypeName}}) Read(iprot thrift.TProtocol) (err error) {
	{{if Features.KeepUnknownFields}}var name string{{end}}
	var fieldTypeId thrift.TType
	var fieldId int16
	{{- range .Fields}}
	{{- if IsRequired .}}
	var isset{{ResolveFieldName .}} bool = false
	{{- end}}
	{{- end}}

	if _, err = iprot.ReadStructBegin(); err != nil {
		goto ReadStructBeginError
	}

	for {
		{{if Features.KeepUnknownFields}}name{{else}}_{{end}}, fieldTypeId, fieldId, err = iprot.ReadFieldBegin()
		if err != nil {
		    goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break;
		}
		{{if or (gt (len .Fields) 0) Features.KeepUnknownFields}}
		switch fieldId {
		{{- range .Fields}}
		case {{.ID}}:
			if fieldTypeId == thrift.{{.Type | GetTypeIDConstant }} {
				if err = p.ReadField{{ID .}}(iprot); err != nil {
					goto ReadFieldError
				}
				{{- if IsRequired .}}
				isset{{ResolveFieldName .}} = true
				{{- end}}
			} else {
				if err = iprot.Skip(fieldTypeId); err != nil {
					goto SkipFieldError
				}
			}
		{{- end}}{{/* range .Fields */}}
		default:
			{{- if Features.KeepUnknownFields}}
			if err = p._unknownFields.Append(iprot, name, fieldTypeId, fieldId); err != nil {
				goto UnknownFieldsAppendError
			}
			{{- else}}
		    if err = iprot.Skip(fieldTypeId); err != nil {
				goto SkipFieldError
		    }
			{{- end}}{{/* if Features.KeepUnknownFields */}}
		}
		{{- else -}}
		if err = iprot.Skip(fieldTypeId); err != nil {
		    goto SkipFieldTypeError
		}
		{{- end}}{{/* if len(.Fields) > 0 */}}

		if err = iprot.ReadFieldEnd(); err != nil {
		  goto ReadFieldEndError
		}
	}
	if err = iprot.ReadStructEnd(); err != nil {
		goto ReadStructEndError
	}
	{{ $RequiredFieldNotSetError := false }}
	{{- range .Fields}}
	{{- if IsRequired .}}
	{{ $RequiredFieldNotSetError = true }}
	if !isset{{ResolveFieldName .}} {
		fieldId = {{.ID}}
		goto RequiredFieldNotSetError
	}
	{{- end}}
	{{- end}}
	return nil
ReadStructBeginError:
	return thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)

{{- if gt (len .Fields) 0}}
ReadFieldError:
	return thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_{{$TypeName}}[fieldId]), err)
SkipFieldError:
	return thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
{{- end}}

{{- if Features.KeepUnknownFields}}
UnknownFieldsAppendError:
	return thrift.PrependError(fmt.Sprintf("%T append unknown field(name:%s type:%d id:%d) error: ", p, name, fieldTypeId, fieldId), err)
{{- end}}

{{- if and (eq (len .Fields) 0) (not Features.KeepUnknownFields)}}
SkipFieldTypeError:
	return thrift.PrependError(fmt.Sprintf("%T skip field type %d error", p, fieldTypeId), err)
{{- end}}

ReadFieldEndError:
	return thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
{{- if $RequiredFieldNotSetError}}
RequiredFieldNotSetError:
	return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("Required field %s is not set", fieldIDToName_{{$TypeName}}[fieldId]))
{{- end}}{{/* if $RequiredFieldNotSetError */}}
}
{{- end}}{{/* define "StructLikeRead" */}}
`

// StructLikeReadField .
var StructLikeReadField = `
{{define "StructLikeReadField"}}
{{- $TypeName := .Name | Identify -}}
{{- range .Fields}}
{{$FieldName := ResolveFieldName .}}
func (p *{{$TypeName}}) ReadField{{ID .}}(iprot thrift.TProtocol) error {
	{{- ResetIDGenerator}}
	{{- $ctx := MkRWCtx . "" false false}}
	{{- template "FieldRead" $ctx}}
	return nil
}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* define "StructLikeReadField" */}}
`

// StructLikeWrite .
var StructLikeWrite = `
{{define "StructLikeWrite"}}
{{- $TypeName := .Name | Identify -}}
func (p *{{$TypeName}}) Write(oprot thrift.TProtocol) (err error) {
	{{- if gt (len .Fields) 0 }}
	var fieldId int16
	{{- end}}
	{{- if eq .Category "union"}}
	var c int
	if c = p.CountSetFields{{$TypeName}}(); c != 1 {
		goto CountSetFieldsError
	}
	{{- end}}
	if err = oprot.WriteStructBegin("{{GetStructName .}}"); err != nil {
		goto WriteStructBeginError
	}
	if p != nil {
		{{- range .Fields}}
		if err = p.writeField{{ID .}}(oprot); err != nil {
			fieldId = {{ID .}}
			goto WriteFieldError
		}
		{{- end}}
		{{if Features.KeepUnknownFields}}
		if err = p._unknownFields.Write(oprot); err != nil {
			goto UnknownFieldsWriteError
		}
		{{- end}}
	}
	if err = oprot.WriteFieldStop(); err != nil {
		goto WriteFieldStopError
	}
	if err = oprot.WriteStructEnd(); err != nil {
		goto WriteStructEndError
	}
	return nil
{{- if eq .Category "union"}}
CountSetFieldsError:
	return fmt.Errorf("%T write union: exactly one field must be set (%d set).", p, c)
{{- end}}
WriteStructBeginError:
	return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
{{- if gt (len .Fields) 0 }}
WriteFieldError:
	return thrift.PrependError(fmt.Sprintf("%T write field %d error: ", p, fieldId), err)
{{- end}}
WriteFieldStopError:
	return thrift.PrependError(fmt.Sprintf("%T write field stop error: ", p), err)
WriteStructEndError:
	return thrift.PrependError(fmt.Sprintf("%T write struct end error: ", p), err)
{{- if Features.KeepUnknownFields}}
UnknownFieldsWriteError:
	return thrift.PrependError(fmt.Sprintf("%T write unknown fields error: ", p), err)
{{- end}}{{/* if Features.KeepUnknownFields */}}
}
{{- end}}{{/* define "StructLikeWrite" */}}
`

// StructLikeWriteField .
var StructLikeWriteField = `
{{define "StructLikeWriteField"}}
{{- $TypeName := .Name | Identify -}}
{{- range .Fields}}
{{- $FieldName := ResolveFieldName .}}
{{- $TypeID := .Type | GetTypeIDConstant }}
func (p *{{$TypeName}}) writeField{{ID .}}(oprot thrift.TProtocol) (err error) {
	{{- ResetIDGenerator}}
	{{- if IsOptional .}}
	if p.IsSet{{$FieldName}}() {
	{{- end}}
	if err = oprot.WriteFieldBegin("{{.Name}}", thrift.{{$TypeID}}, {{.ID}}); err != nil {
		goto WriteFieldBeginError
	}
	{{- $ctx := MkRWCtx . "" false false}}
	{{- template "FieldWrite" $ctx}}
	if err = oprot.WriteFieldEnd(); err != nil {
		goto WriteFieldEndError
	}
	{{- if IsOptional .}}
	}
	{{- end}}
	return nil
WriteFieldBeginError:
	return thrift.PrependError(fmt.Sprintf("%T write field {{ID .}} begin error: ", p), err)
WriteFieldEndError:
	return thrift.PrependError(fmt.Sprintf("%T write field {{ID .}} end error: ", p), err)
}
{{end}}{{/* range .Fields */}}
{{- end}}{{/* define "StructLikeWriteField" */}}
`

// FieldGetOrSet .
var FieldGetOrSet = `
{{define "FieldGetOrSet"}}
{{- $TypeName := .Name | Identify -}}
{{- range .Fields}}
{{$FieldName := ResolveFieldName .}}
{{$FieldTypeName := . | ResolveFieldTypeName}}
{{$DefaultVarTypeName := GetDefaultValueTypeName .}}

{{if SupportIsSet .}}
{{$DefaultVarName := printf "%s_%s_%s" $TypeName $FieldName "DEFAULT"}}
var {{$DefaultVarName}} {{$DefaultVarTypeName}}
{{- if .Default}} = {{GetFieldInit .}}{{- end}}

func (p *{{$TypeName}}) Get{{$FieldName}}() {{$DefaultVarTypeName}} {
	if !p.IsSet{{$FieldName}}() {
		return {{$DefaultVarName}}
	}
	{{- if and (NeedRedirect .) (IsBaseType .Type)}}
	return *p.{{$FieldName}}
	{{- else}}
	return p.{{$FieldName}}
	{{- end}}
}

{{- else}}{{/*if SupportIsSet . */}}

func (p *{{$TypeName}}) Get{{$FieldName}}() {{$FieldTypeName}} {
	return p.{{$FieldName}}
}

{{- end}}{{/* if SupportIsSet . */}}
{{- end}}{{/* range .Fields */}}

{{- if Features.GenerateSetter}}
{{range .Fields}}
{{$FieldName := ResolveFieldName .}}
{{$FieldTypeName := . | ResolveFieldTypeName}}
{{$SetterName := GetFieldSetterName .}}
{{- if IsSetterOfResponseType $TypeName $FieldName -}}
func (p *{{$TypeName}}) {{$SetterName}}(x interface{}) {
    p.{{$FieldName}} = x.({{$FieldTypeName}})
}
{{- else -}}
func (p *{{$TypeName}}) {{$SetterName}}(val {{$FieldTypeName}}) {
	p.{{$FieldName}} = val
}
{{- end}}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* if Features.GenerateSetter */}}

{{- end}}{{/* define "FieldGetOrSet" */}}
`

// FieldIsSet .
var FieldIsSet = `
{{define "FieldIsSet"}}
{{- $TypeName := .Name | Identify}}
{{- range .Fields}}
{{- $FieldName := ResolveFieldName .}}
{{- $FieldTypeName := . | ResolveFieldTypeName}}
{{- $DefaultVarName := printf "%s_%s_%s" $TypeName $FieldName "DEFAULT"}}
{{- if SupportIsSet .}}
func (p *{{$TypeName}}) IsSet{{$FieldName}}() bool {
	{{- if HasDefaultValue .}}
		{{- if IsBaseType .Type}}
			{{- if IsBinaryType .Type}}
				return string(p.{{$FieldName}}) != string({{$DefaultVarName}})
			{{- else}}
				return p.{{$FieldName}} != {{$DefaultVarName}}
			{{- end}}
		{{- else}}{{/* container type or struct-like */}}
			return p.{{$FieldName}} != nil
		{{- end}}
	{{- else}}
		return p.{{$FieldName}} != nil
	{{- end}}
}
{{end}}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* define "FieldIsSet" */}}
`

// FieldRead .
var FieldRead = `
{{define "FieldRead"}}
	{{- if IsStructLike .Type}}
		{{- template "FieldReadStructLike" .}}
	{{- else if IsBaseType .Type }}
		{{- template "FieldReadBaseType" .}}
	{{- else}}{{/* IsContainerType */}}
		{{- template "FieldReadContainer" .}}
	{{- end}}
{{- end}}{{/* define "FieldRead" */}}
`

// FieldReadStructLike .
var FieldReadStructLike = `
{{define "FieldReadStructLike"}}
	{{- .Target}} {{if .NeedDecl}}:{{end}}= {{.TypeName | Deref | GetNewFunc}}()
	if err := {{.Target}}.Read(iprot); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadStructLike" */}} 
`

// FieldReadBaseType .
var FieldReadBaseType = `
{{define "FieldReadBaseType"}}
	{{- $DiffType := or (IsEnumType .Type) (IsBinaryType .Type)}}
	{{- if .NeedDecl}}
	var {{.Target}} {{.TypeName}}
	{{- end}}
	if v, err := iprot.Read{{.TypeID}}(); err != nil {
		return err
	} else {
	{{- if .IsPointer}}
		{{- if $DiffType}}
		tmp := {{.TypeName | Deref}}(v)
		{{.Target}} = &tmp
		{{- else -}}
		{{.Target}} = &v
		{{- end}}
	{{- else}}
		{{- if $DiffType}}
		{{.Target}} = {{.TypeName}}(v)
		{{- else}}
		{{.Target}} = v
		{{- end}}
	{{- end}}
	}
{{- end}}{{/* define "FieldReadBaseType" */}}
`

// FieldReadContainer .
var FieldReadContainer = `
{{define "FieldReadContainer"}}
	{{- if eq "Map" .TypeID}}
	     {{- template "FieldReadMap" .}}
	{{- else if eq "List" .TypeID}}
	     {{- template "FieldReadList" .}}
	{{- else}}
	     {{- template "FieldReadSet" .}}
	{{- end}}
{{- end}}{{/* define "FieldReadContainer" */}}
`

// FieldReadMap .
var FieldReadMap = `
{{define "FieldReadMap"}}
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, size)
	for i := 0; i < size; i++ {
		{{- $key := GenID "_key"}}
		{{- $ctx := MkRWCtx (.Type | GetKeyType) $key true true}}
		{{- template "FieldRead" $ctx}}
		{{/* line break */}}
		{{- $val := GenID "_val"}}
		{{- $ctx = MkRWCtx (.Type | GetValType) $val true false}}
		{{- template "FieldRead" $ctx}}

		{{if and (IsStructLike (.Type | GetValType)) Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}}[{{$key}}] = {{$val}}
	}
	if err := iprot.ReadMapEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadMap" */}}
`

// FieldReadSet .
var FieldReadSet = `
{{define "FieldReadSet"}}
	_, size, err := iprot.ReadSetBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, 0, size)
	for i := 0; i < size; i++ {
		{{- $val := GenID "_elem"}}
		{{- $ctx := MkRWCtx (.Type | GetValType) $val true false}}
		{{- template "FieldRead" $ctx}}

		{{if and (IsStructLike (.Type | GetValType)) Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}} = append({{.Target}}, {{$val}})
	}
	if err := iprot.ReadSetEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadSet" */}}
`

// FieldReadList .
var FieldReadList = `
{{define "FieldReadList"}}
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, 0, size)
	for i := 0; i < size; i++ {
		{{- $val := GenID "_elem"}}
		{{- $ctx := MkRWCtx (.Type | GetValType) $val true false}}
		{{- template "FieldRead" $ctx}}

		{{if and (IsStructLike (.Type | GetValType)) Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}} = append({{.Target}}, {{$val}})
	}
	if err := iprot.ReadListEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadList" */}}
`

// FieldWrite .
var FieldWrite = `
{{define "FieldWrite"}}
	{{- if IsStructLike .Type}}
		{{- template "FieldWriteStructLike" . -}}
	{{- else if IsBaseType .Type }}
		{{- template "FieldWriteBaseType" . -}}
	{{- else}}{{/* IsContainerType */}}
		{{- template "FieldWriteContainer" . -}}
	{{- end}}
{{- end}}{{/* define "FieldWrite" */}}
`

// FieldWriteStructLike .
var FieldWriteStructLike = `
{{define "FieldWriteStructLike"}}
	if err := {{.Target}}.Write(oprot); err != nil {
		return err
	}
{{- end}}{{/* define "FieldWriteStructLike" */}}
`

// FieldWriteBaseType .
var FieldWriteBaseType = `
{{define "FieldWriteBaseType"}}
{{- $Value := .Target}}
{{- if .IsPointer}}{{$Value = printf "*%s" $Value}}{{end}}
{{- if IsEnumType .Type}}{{$Value = printf "int32(%s)" $Value}}{{end}}
{{- if IsBinaryType .Type}}{{$Value = printf "[]byte(%s)" $Value}}{{end}}
	if err := oprot.Write{{.TypeID}}({{$Value}}); err != nil {
		return err
	}
{{- end}}{{/* define "FieldWriteBaseType" */}}
`

// FieldWriteContainer .
var FieldWriteContainer = `
{{define "FieldWriteContainer"}}
	{{- if eq "Map" .TypeID}}
		{{- template "FieldWriteMap" .}}
	{{- else if eq "List" .TypeID}}
		{{- template "FieldWriteList" .}}
	{{- else}}
		{{- template "FieldWriteSet" .}}
	{{- end}}
{{- end}}{{/* define "FieldWriteContainer" */}}
`

// FieldWriteMap .
var FieldWriteMap = `
{{define "FieldWriteMap"}}
	if err := oprot.WriteMapBegin(thrift.
		{{- .Type| GetKeyType | GetTypeIDConstant -}}
		, thrift.{{- .Type | GetValType | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
		return err
	}
	for k, v := range {{.Target}}{
		{{$ctx := MkRWCtx (.Type | GetKeyType) "k" false true}}
		{{- template "FieldWrite" $ctx}}
		{{$ctx := MkRWCtx (.Type | GetValType) "v" false false}}
		{{- template "FieldWrite" $ctx}}
	}
	if err := oprot.WriteMapEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldWriteMap" */}}
`

// FieldWriteSet .
var FieldWriteSet = `
{{define "FieldWriteSet"}}
		if err := oprot.WriteSetBegin(thrift.
		{{- .Type | GetValType | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
			return err
		}
		{{- if Features.ValidateSet}}
		{{- $ctx := MkRWCtx2 (.Type | GetValType) "tgt" "src" false false}}
		for i := 0; i < len({{.Target}}); i++ {
			for j := i + 1; j < len({{.Target}}); j++ {
		{{- if Features.GenDeepEqual}}
				if func(tgt, src {{$ctx.TypeName}}) bool {
					{{- template "FieldDeepEqual" $ctx}}
					return true
				}({{.Target}}[i], {{.Target}}[j]) {
		{{- else}}
				if reflect.DeepEqual({{.Target}}[i], {{.Target}}[j]) {
		{{- end}}
					return thrift.PrependError("", fmt.Errorf("%T error writing set field: slice is not unique", {{.Target}}[i]))
				}
			}
		}
		{{- end}}
		for _, v := range {{.Target}} {
			{{- $ctx := MkRWCtx (.Type | GetValType) "v" false false}}
			{{- template "FieldWrite" $ctx}}
		}
		if err := oprot.WriteSetEnd(); err != nil {
			return err
		}
{{- end}}{{/* define "FieldWriteSet" */}}
`

// FieldWriteList .
var FieldWriteList = `
{{define "FieldWriteList"}}
		if err := oprot.WriteListBegin(thrift.
		{{- .Type | GetValType | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
			return err
		}
		for _, v := range {{.Target}} {
			{{- $ctx := MkRWCtx (.Type | GetValType) "v" false false}}
			{{- template "FieldWrite" $ctx}}
		}
		if err := oprot.WriteListEnd(); err != nil {
			return err
		}
{{- end}}{{/* define "FieldWriteList" */}}
`

// StructLikeDeepEqual .
var StructLikeDeepEqual = `
{{define "StructLikeDeepEqual"}}
{{- $TypeName := .Name | Identify}}
func (p *{{$TypeName}}) DeepEqual(ano *{{$TypeName}}) bool {
	if p == ano {
		return true
	} else if p == nil || ano == nil {
		return false
	}
	{{- range .Fields}}
	if !p.Field{{ID .}}DeepEqual(ano.{{. | ResolveFieldName}}) {
		return false
	}
	{{- end}}
	return true
}
{{- end}}{{/* "StructLikeDeepEqual" */}}
`

// StructLikeDeepEqualField .
var StructLikeDeepEqualField = `
{{define "StructLikeDeepEqualField"}}
{{- $TypeName := .Name | Identify}}
{{- range .Fields}}
{{- $field := MkRWCtx . "" false false}}
func (p *{{$TypeName}}) Field{{ID .}}DeepEqual({{$field.Source}} {{$field.TypeName}}) bool {
	{{template "FieldDeepEqual" $field}}
	return true
}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* "StructLikeDeepEqualField" */}}
`

// FieldDeepEqual .
var FieldDeepEqual = `
{{define "FieldDeepEqual"}}
{{- if IsStructLike .Type}}
	{{- template "FieldDeepEqualStructLike" . -}}
{{- else if IsBaseType .Type}}
	{{- template "FieldDeepEqualBase" . -}}
{{- else}}{{/* IsContainerType */}}
	{{- template "FieldDeepEqualContainer" . -}}
{{- end}}
{{- end}}{{/* "FieldDeepEqual" */}}
`

// FieldDeepEqualStructLike .
var FieldDeepEqualStructLike = `
{{define "FieldDeepEqualStructLike"}}
	if !{{.Target}}.DeepEqual({{.Source}}) {
		return false
	}
{{- end}}{{/* "FieldDeepEqualStructLike" */}}
`

// FieldDeepEqualBase .
var FieldDeepEqualBase = `
{{define "FieldDeepEqualBase"}}
	{{- if .IsPointer}}
	if {{.Target}} == {{.Source}} {
		return true
	} else if {{.Target}} == nil || {{.Source}} == nil {
		return false
	}
	{{- end}}
	{{- $tgt := .Target}}
	{{- $src := .Source}}
	{{- if .IsPointer}}{{$tgt = printf "*%s" $tgt}}{{$src = printf "*%s" $src}}{{end}}
	{{- if IsStringType .Type}}
		if strings.Compare({{$tgt}}, {{$src}}) != 0 {
			return false
		}
	{{- else if IsBinaryType .Type}}
		if bytes.Compare({{$tgt}}, {{$src}}) != 0 {
			return false
		}
	{{- else}}{{/* IsFixedLengthType */}}
		if {{$tgt}} != {{$src}} {
			return false
		}
	{{- end}}
{{- end}}{{/* "FieldDeepEqualBase" */}}
`

// FieldDeepEqualContainer .
var FieldDeepEqualContainer = `
{{define "FieldDeepEqualContainer"}}
	{{- if .IsPointer}}
	if {{.Target}} == {{.Source}} {
		return true
	} else if {{.Target}} == nil || {{.Source}} == nil {
		return false
	}
	{{- end}}
	{{- if eq "Map" .TypeID}}
		{{- template "FieldDeepEqualMap" .}}
	{{- else if eq "List" .TypeID}}
		{{- template "FieldDeepEqualList" .}}
	{{- else}}{{/* "Set" */}}
		{{- template "FieldDeepEqualSet" .}}
	{{- end}}
{{- end}}{{/* "FieldDeepEqualContainer" */}}
`

// FieldDeepEqualList .
var FieldDeepEqualList = `
{{define "FieldDeepEqualList"}}
	if len({{.Target}}) != len({{.Source}}) {
		return false
	}
	{{- $src := GenID "_src"}}
	for i, v := range {{.Target}} {
		{{$src}} := {{.Source}}[i]
		{{- $ctx := MkRWCtx2 (.Type | GetValType) "v" $src false false}}
		{{- template "FieldDeepEqual" $ctx}}
	}
{{- end}}{{/* "FieldDeepEqualList" */}}
`

// FieldDeepEqualSet .
var FieldDeepEqualSet = `
{{define "FieldDeepEqualSet"}}
	if len({{.Target}}) != len({{.Source}}) {
		return false
	}
	{{- $src := GenID "_src"}}
	for i, v := range {{.Target}} {
		{{$src}} := {{.Source}}[i]
		{{- $ctx := MkRWCtx2 (.Type | GetValType) "v" $src false false}}
		{{- template "FieldDeepEqual" $ctx}}
	}
{{- end}}{{/* "FieldDeepEqualSet" */}}
`

// FieldDeepEqualMap .
var FieldDeepEqualMap = `
{{define "FieldDeepEqualMap"}}
	if len({{.Target}}) != len({{.Source}}) {
		return false
	}
	{{- $src := GenID "_src"}}
	for k, v := range {{.Target}}{
		{{$src}} := {{.Source}}[k]
		{{$ctx := MkRWCtx2 (.Type | GetValType) "v" $src false false}}
		{{- template "FieldDeepEqual" $ctx}}
	}
{{- end}}{{/* "FieldDeepEqualMap" */}}
`
