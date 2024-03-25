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

// StructLike is the code template for struct, union, and exception.
var StructLike = `
{{define "StructLike"}}
{{- $TypeName := .GoName}}
{{InsertionPoint .Category .Name}}
{{- if and Features.ReserveComments .ReservedComments}}{{.ReservedComments}}{{end}}
type {{$TypeName}} struct {
{{- range .Fields}}
	{{- InsertionPoint $.Category $.Name .Name}}
	{{- if and Features.ReserveComments .ReservedComments}}
	{{.ReservedComments}}
	{{- end}}
	{{(.GoName)}} {{.GoTypeName}} {{GenFieldTags . (InsertionPoint $.Category $.Name .Name "tag")}} 
{{- end}}
	{{- if Features.KeepUnknownFields}}
	{{- UseStdLibrary "unknown"}}
	_unknownFields unknown.Fields
	{{- end}}
	{{- if Features.WithFieldMask}}
	{{- UseStdLibrary "fieldmask"}}
	_fieldmask *fieldmask.FieldMask
	{{- end}}
}

{{- if Features.GenerateTypeMeta}}
{{- UseStdLibrary "meta"}}
func init() {
	meta.RegisterStruct(New{{$TypeName}}, {{Marshal .}})
}
{{- end}}{{/* if Features.GenerateTypeMeta */}}

func New{{$TypeName}}() *{{$TypeName}} {
	return &{{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}

{{if Features.FrugalTag}}
func (p *{{$TypeName}}) InitDefault() {
	*p = {{$TypeName}}{
		{{template "StructLikeDefault" .}}
	}
}
{{end}}{{/* if Features.FrugalTag */}}

{{template "FieldGetOrSet" .}}

{{if eq .Category "union"}}
func (p *{{$TypeName}}) CountSetFields{{$TypeName}}() int {
	count := 0
	{{- range .Fields}}
	{{- if SupportIsSet .Field}}
	if p.{{.IsSetter}}() {
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

{{if Features.WithFieldMask}}
func (p *{{$TypeName}}) Get_FieldMask() *fieldmask.FieldMask {
	if p == nil {
		return nil
	}
	return p._fieldmask
}

func (p *{{$TypeName}}) Set_FieldMask(fm *fieldmask.FieldMask) {
	if p == nil {
		return
	}
	p._fieldmask = fm
}

{{- if Features.FieldMaskHalfway}}
func (p *{{$TypeName}}) Pass_FieldMask(fm *fieldmask.FieldMask) {
	if p == nil || p._fieldmask != nil {
		return
	}
	p._fieldmask = fm
}
{{- end}}
{{end}}{{/* if Features.WithFieldMask */}}

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
	{{- if Features.JSONStringer}}
	{{- UseStdLibrary "json_utils"}}
		JsonBytes , _  := json_utils.JSONFunc(p)
		return string(JsonBytes)
	{{- else}}
	if p == nil {
		return "<nil>"
	}
	{{- UseStdLibrary "fmt"}}
	return fmt.Sprintf("{{$TypeName}}(%+v)", *p)
	{{- end}}

}

{{- if eq .Category "exception"}}
func (p *{{$TypeName}}) Error() string {
	return p.String()
}
{{- end}}

{{- if Features.GenDeepEqual}}
{{template "StructLikeDeepEqual" .}}

{{template "StructLikeDeepEqualField" .}}
{{- end}}

{{- end}}{{/* define "StructLike" */}}
`

// StructLikeDefault is the code template for structure initialization.
var StructLikeDefault = `
{{- define "StructLikeDefault"}}
{{- range .Fields}}
	{{- if .IsSetDefault}}
		{{.GoName}}: {{.DefaultValue}},
	{{- end}}
{{- end}}
{{- end -}}`

// StructLikeRead .
var StructLikeRead = `
{{define "StructLikeRead"}}
{{- UseStdLibrary "thrift" "fmt"}}
{{- $TypeName := .GoName}}
func (p *{{$TypeName}}) Read(iprot thrift.TProtocol) (err error) {
	{{if Features.KeepUnknownFields}}var name string{{end}}
	var fieldTypeId thrift.TType
	var fieldId int16
	{{- range .Fields}}
	{{- if .Requiredness.IsRequired}}
	var isset{{.GoName}} bool = false
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
		{{- $isBaseVal := .Type | IsBaseType}}
		case {{.ID}}:
			if fieldTypeId == thrift.{{.Type | GetTypeIDConstant }} {
				if err = p.{{.Reader}}(iprot); err != nil {
					goto ReadFieldError
				}
				{{- if .Requiredness.IsRequired}}
				isset{{.GoName}} = true
				{{- end}}
			} else if err = iprot.Skip(fieldTypeId); err != nil {
				goto SkipFieldError
			}
		{{- end}}{{/* range .Fields */}}
		default:
			{{- template "HandleUnknownFields"}}
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
	{{- if .Requiredness.IsRequired}}
	{{ $RequiredFieldNotSetError = true }}
	if !isset{{.GoName}} {
		fieldId = {{.ID}}
		goto RequiredFieldNotSetError
	}
	{{- end}}
	{{- end}}{{/* range .Fields */}}
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
	return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("required field %s is not set", fieldIDToName_{{$TypeName}}[fieldId]))
{{- end}}{{/* if $RequiredFieldNotSetError */}}
}
{{- end}}{{/* define "StructLikeRead" */}}
`

var HandleUnknownFields = `
{{define "HandleUnknownFields"}}
{{- if Features.KeepUnknownFields}}
if err = p._unknownFields.Append(iprot, name, fieldTypeId, fieldId); err != nil {
	goto UnknownFieldsAppendError
}
{{- else}}
if err = iprot.Skip(fieldTypeId); err != nil {
	goto SkipFieldError
}
{{- end}}{{/* if Features.KeepUnknownFields */}}
{{- end}}{{/* define "HandleUnknownFields" */}}
`

// StructLikeReadField .
var StructLikeReadField = `
{{define "StructLikeReadField"}}
{{- UseStdLibrary "thrift"}}
{{- $TypeName := .GoName}}
{{- range .Fields}}
{{$FieldName := .GoName}}
{{- $isBaseVal := .Type | IsBaseType -}}
func (p *{{$TypeName}}) {{.Reader}}(iprot thrift.TProtocol) error {
	{{- if Features.WithFieldMask}}
	if {{if $isBaseVal}}_{{else}}fm{{end}}, ex := p._fieldmask.Field({{.ID}}); ex {
	{{- end}}
	{{$ctx := (MkRWCtx .).WithFieldMask "fm"}}
	{{- template "FieldRead" $ctx}}
	{{- if Features.WithFieldMask}}
	} else if err := iprot.Skip(thrift.{{.Type | GetTypeIDConstant}}); err != nil {
		return err
	}
	{{- end}}
	return nil
}
{{- end}}{{/* range .Fields */}}
{{- end}}{{/* define "StructLikeReadField" */}}
`

// StructLikeWrite .
var StructLikeWrite = `
{{define "StructLikeWrite"}}
{{- UseStdLibrary "thrift" "fmt"}}
{{- $TypeName := .GoName}}
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
	if err = oprot.WriteStructBegin("{{.Name}}"); err != nil {
		goto WriteStructBeginError
	}
	if p != nil {
		{{- range .Fields}}
		if err = p.{{.Writer}}(oprot); err != nil {
			fieldId = {{.ID}}
			goto WriteFieldError
		}
		
		{{- end}}{{/* range .Fields */}}
		{{- if Features.KeepUnknownFields}}
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
{{- UseStdLibrary "thrift" "fmt"}}
{{- $TypeName := .GoName}}
{{- range .Fields}}
{{- $FieldName := .GoName}}
{{- $IsSetName := .IsSetter}}
{{- $TypeID := .Type | GetTypeIDConstant }}
{{- $isBaseVal := .Type | IsBaseType }}
func (p *{{$TypeName}}) {{.Writer}}(oprot thrift.TProtocol) (err error) {
	{{- if .Requiredness.IsOptional}}
	if p.{{$IsSetName}}() {
	{{- end}}
	{{- if Features.WithFieldMask}}
	{{- if and .Requiredness.IsRequired (not Features.FieldMaskZeroRequired)}}
	{{- if not $isBaseVal}}
	fm, _ := p._fieldmask.Field({{.ID}})
	{{- end}}
	{{- else}}
	if {{if $isBaseVal}}_{{else}}fm{{end}}, ex := p._fieldmask.Field({{.ID}}); ex { 
	{{- end}}
	{{- end}}
	if err = oprot.WriteFieldBegin("{{.Name}}", thrift.{{$TypeID}}, {{.ID}}); err != nil {
		goto WriteFieldBeginError
	}
	{{- $ctx := (MkRWCtx .).WithFieldMask "fm"}}
	{{- template "FieldWrite" $ctx}}
	if err = oprot.WriteFieldEnd(); err != nil {
		goto WriteFieldEndError
	}
	{{- if Features.WithFieldMask}}
	{{- if Features.FieldMaskZeroRequired}}
	} else {
		if err = oprot.WriteFieldBegin("{{.Name}}", thrift.{{$TypeID}}, {{.ID}}); err != nil {
			goto WriteFieldBeginError
		}
		{{ ZeroWriter .Type "oprot" "WriteFieldBeginError" }}
		if err = oprot.WriteFieldEnd(); err != nil {
			goto WriteFieldEndError
		}
	}
	{{- else if not .Requiredness.IsRequired}}
	}
	{{- end}}
	{{- end}}
	{{- if .Requiredness.IsOptional}}
	}
	{{- end}}
	return nil
WriteFieldBeginError:
	return thrift.PrependError(fmt.Sprintf("%T write field {{.ID}} begin error: ", p), err)
WriteFieldEndError:
	return thrift.PrependError(fmt.Sprintf("%T write field {{.ID}} end error: ", p), err)
}
{{end}}{{/* range .Fields */}}
{{- end}}{{/* define "StructLikeWriteField" */}}
`

// FieldGetOrSet .
var FieldGetOrSet = `
{{define "FieldGetOrSet"}}
{{- $TypeName := .GoName}}
{{- range .Fields}}
{{- $FieldName := .GoName}}
{{- $FieldTypeName := .GoTypeName}}
{{- $DefaultVarTypeName := .DefaultTypeName}}
{{- $GetterName := .Getter}}
{{- $SetterName := .Setter}}
{{- $IsSetName := .IsSetter}}

{{if SupportIsSet .Field}}
{{$DefaultVarName := printf "%s_%s_%s" $TypeName $FieldName "DEFAULT"}}
var {{$DefaultVarName}} {{$DefaultVarTypeName}}
{{- if .Default}} = {{.DefaultValue}}{{- end}}

func (p *{{$TypeName}}) {{$GetterName}}() (v {{$DefaultVarTypeName}}) {
	{{- if Features.NilSafe}}
	if p == nil {
		return
	}
	{{- end}}
	if !p.{{$IsSetName}}() {
		return {{$DefaultVarName}}
	}
	{{- if and (NeedRedirect .Field) (IsBaseType .Type)}}
	return *p.{{$FieldName}}
	{{- else}}
	return p.{{$FieldName}}
	{{- end}}
}

{{- else}}{{/*if SupportIsSet . */}}

func (p *{{$TypeName}}) {{$GetterName}}() (v {{$FieldTypeName}}) {
	{{- if Features.NilSafe}}
	if p != nil {
		return p.{{$FieldName}}
	}
	return
	{{- else}}
		return p.{{$FieldName}}
	{{- end}}{{/* if Features.NilSafe */}}
}

{{- end}}{{/* if SupportIsSet . */}}
{{- end}}{{/* range .Fields */}}

{{- if Features.GenerateSetter}}
{{- range .Fields}}
{{- $FieldName := .GoName}}
{{- $FieldTypeName := .GoTypeName}}
{{- $SetterName := .Setter}}
{{- if .IsResponseFieldOfResult}}
func (p *{{$TypeName}}) {{$SetterName}}(x interface{}) {
    p.{{$FieldName}} = x.({{$FieldTypeName}})
}
{{- else}}
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
{{- $TypeName := .GoName}}
{{- range .Fields}}
{{- $FieldName := .GoName}}
{{- $IsSetName := .IsSetter}}
{{- $FieldTypeName := .GoTypeName}}
{{- $DefaultVarName := printf "%s_%s_%s" $TypeName $FieldName "DEFAULT"}}
{{- if SupportIsSet .Field}}
func (p *{{$TypeName}}) {{$IsSetName}}() bool {
	{{- if .IsSetDefault}}
		{{- if IsBaseType .Type}}
			{{- if .Type.Category.IsBinary}}
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
	{{- if .Type.Category.IsStructLike}}
		{{- template "FieldReadStructLike" .}}
	{{- else if .Type.Category.IsContainerType}}
		{{- template "FieldReadContainer" .}}
	{{- else}}{{/* IsBaseType */}}
		{{- template "FieldReadBaseType" .}}
	{{- end}}
{{- end}}{{/* define "FieldRead" */}}
`

// FieldReadStructLike .
var FieldReadStructLike = `
{{define "FieldReadStructLike"}}
	{{- .Target}} {{if .NeedDecl}}:{{end}}= {{.TypeName.Deref.NewFunc}}()
	{{- if and (Features.WithFieldMask) .NeedFieldMask}}
	{{- if Features.FieldMaskHalfway}}
	{{.Target}}.Pass_FieldMask({{.FieldMask}})
	{{- else}}
	{{.Target}}.Set_FieldMask({{.FieldMask}})
	{{- end}}
	{{- end}}
	if err := {{.Target}}.Read(iprot); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadStructLike" */}} 
`

// FieldReadBaseType .
var FieldReadBaseType = `
{{define "FieldReadBaseType"}}
	{{- $DiffType := or .Type.Category.IsEnum .Type.Category.IsBinary}}
	{{- if .NeedDecl}}
	var {{.Target}} {{.TypeName}}
	{{- end}}
	if v, err := iprot.Read{{.TypeID}}(); err != nil {
		return err
	} else {
	{{- if .IsPointer}}
		{{- if $DiffType}}
		tmp := {{.TypeName.Deref}}(v)
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
{{- $isIntKey := .KeyCtx.Type | IsIntType -}}
{{- $isStrKey := .KeyCtx.Type | IsStrType -}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, size)
	for i := 0; i < size; i++ {
		{{- $key := .GenID "_key"}}
		{{- $ctx := .KeyCtx.WithDecl.WithTarget $key}}
		{{- template "FieldRead" $ctx}}
		{{- if Features.WithFieldMask}}
		{{- $curFieldMask = "nfm"}}
		{{- if $isIntKey}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(int({{$key}})); !ex {
			if err := iprot.Skip(thrift.{{.ValCtx.Type | GetTypeIDConstant}}); err != nil {
				return err
			}
			continue
		} else {
		{{- else if $isStrKey}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Str(string({{$key}})); !ex {
			if err := iprot.Skip(thrift.{{.ValCtx.Type | GetTypeIDConstant}}); err != nil {
				return err
			}
			continue
		} else {
		{{- else}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(0); !ex {
			if err := iprot.Skip(thrift.{{.ValCtx.Type | GetTypeIDConstant}}); err != nil {
				return err
			}
			continue
		} else {
		{{- end}}
		{{- end}}{{/* end WithFieldMask */}}
		{{/* line break */}}
		{{- $val := .GenID "_val"}}
		{{- $ctx := (.ValCtx.WithDecl.WithTarget $val).WithFieldMask $curFieldMask}}
		{{- template "FieldRead" $ctx}}

		{{if and .ValCtx.Type.Category.IsStructLike Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}}[{{$key}}] = {{$val}}
		{{- if and Features.WithFieldMask}}
		}
		{{- end}}
	}
	if err := iprot.ReadMapEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadMap" */}}
`

// FieldReadSet .
var FieldReadSet = `
{{define "FieldReadSet"}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
	_, size, err := iprot.ReadSetBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, 0, size)
	for i := 0; i < size; i++ {
		{{- $val := .GenID "_elem"}}
		{{- if Features.WithFieldMask}}
		{{- $curFieldMask = "nfm"}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(i); !ex {
			if err := iprot.Skip(thrift.{{.ValCtx.Type | GetTypeIDConstant}}); err != nil {
				return err
			}
			continue
		} else {
		{{- end}}
		{{- $ctx := (.ValCtx.WithDecl.WithTarget $val).WithFieldMask $curFieldMask}}
		{{template "FieldRead" $ctx}}

		{{if and .ValCtx.Type.Category.IsStructLike Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}} = append({{.Target}}, {{$val}})
		{{- if Features.WithFieldMask}}
		}
		{{- end}}
	}
	if err := iprot.ReadSetEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadSet" */}}
`

// FieldReadList .
var FieldReadList = `
{{define "FieldReadList"}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return err
	}
	{{.Target}} {{if .NeedDecl}}:{{end}}= make({{.TypeName}}, 0, size)
	for i := 0; i < size; i++ {
		{{- $val := .GenID "_elem"}}
		{{- if Features.WithFieldMask}}
		{{- $curFieldMask = "nfm"}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(i); !ex {
			if err := iprot.Skip(thrift.{{.ValCtx.Type | GetTypeIDConstant}}); err != nil {
				return err
			}
			continue
		} else {
		{{- end}}
		{{- $ctx := (.ValCtx.WithDecl.WithTarget $val).WithFieldMask $curFieldMask}}
		{{template "FieldRead" $ctx}}

		{{if and .ValCtx.Type.Category.IsStructLike Features.ValueTypeForSIC}}
			{{$val = printf "*%s" $val}}
		{{end}}

		{{.Target}} = append({{.Target}}, {{$val}})
		{{- if Features.WithFieldMask}}
		}
		{{- end}}
	}
	if err := iprot.ReadListEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldReadList" */}}
`

// FieldWrite .
var FieldWrite = `
{{define "FieldWrite"}}
	{{- if .Type.Category.IsStructLike}}
		{{- template "FieldWriteStructLike" .}}
	{{- else if .Type.Category.IsContainerType}}
		{{- template "FieldWriteContainer" .}}
	{{- else}}{{/* IsBaseType */}}
		{{- template "FieldWriteBaseType" .}}
	{{- end}}
{{- end}}{{/* define "FieldWrite" */}}
`

// FieldWriteStructLike .
var FieldWriteStructLike = `
{{define "FieldWriteStructLike"}}
	{{- if and (Features.WithFieldMask) .NeedFieldMask}}
	{{- if Features.FieldMaskHalfway}}
	{{.Target}}.Pass_FieldMask({{.FieldMask}})
	{{- else}}
	{{.Target}}.Set_FieldMask({{.FieldMask}})
	{{- end}}
	{{- end}}
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
{{- if .Type.Category.IsEnum}}{{$Value = printf "int32(%s)" $Value}}{{end}}
{{- if .Type.Category.IsBinary}}{{$Value = printf "[]byte(%s)" $Value}}{{end}}
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
{{- $isIntKey := .KeyCtx.Type | IsIntType -}}
{{- $isStrKey := .KeyCtx.Type | IsStrType -}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
	{{- if and Features.WithFieldMask (or $isStrKey $isIntKey) }}
	if !{{.FieldMask}}.All() {
		l := len({{.Target}})
		for k := range {{.Target}} {
			{{- if $isIntKey}}
			if _, ex := {{.FieldMask}}.Int(int(k)); !ex {
				l--
			}
			{{- else if $isStrKey}}
			if _, ex := {{.FieldMask}}.Str(string(k)); !ex {
				l--
			}
			{{- end}}
		}
		if err := oprot.WriteMapBegin(thrift.
			{{- .KeyCtx.Type | GetTypeIDConstant -}}
			, thrift.{{- .ValCtx.Type | GetTypeIDConstant -}}
			, l); err != nil {
			return err
		}
	} else {
		if err := oprot.WriteMapBegin(thrift.
			{{- .KeyCtx.Type | GetTypeIDConstant -}}
			, thrift.{{- .ValCtx.Type | GetTypeIDConstant -}}
			, len({{.Target}})); err != nil {
			return err
		}
	}
	{{- else}}
	if err := oprot.WriteMapBegin(thrift.
		{{- .KeyCtx.Type | GetTypeIDConstant -}}
		, thrift.{{- .ValCtx.Type | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
		return err
	}
	{{- end}}
	for k, v := range {{.Target}} {
		{{- if Features.WithFieldMask}}
		{{- $curFieldMask = "nfm"}}
		{{- if $isIntKey}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(int(k)); !ex {
			continue
		} else {
		{{- else if $isStrKey}}
		ks := string(k)
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Str(ks); !ex {
			continue
		} else {
		{{- else}}
		if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(0); !ex {
			continue
		} else {
		{{- end}}
		{{- end}}{{/* end Features.WithFieldMask */}}
		{{- $ctx := .KeyCtx.WithTarget "k" -}}
		{{- template "FieldWrite" $ctx}}
		{{- $ctx := (.ValCtx.WithTarget "v").WithFieldMask $curFieldMask -}}
		{{- template "FieldWrite" $ctx}}
		{{- if and Features.WithFieldMask }}
		}
		{{- end}}
	}
	if err := oprot.WriteMapEnd(); err != nil {
		return err
	}
{{- end}}{{/* define "FieldWriteMap" */}}
`

// FieldWriteSet .
var FieldWriteSet = `
{{define "FieldWriteSet"}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
		{{- if Features.WithFieldMask}}
		if !{{.FieldMask}}.All() {
			l := len({{.Target}})
			for i:=0; i < l; i++ {
				if _, ex := {{.FieldMask}}.Int(i); !ex {
					l--
				}
			}
			if err := oprot.WriteSetBegin(thrift.
			{{- .ValCtx.Type | GetTypeIDConstant -}}
			, l); err != nil {
				return err
			}
		} else {
			if err := oprot.WriteSetBegin(thrift.
			{{- .ValCtx.Type | GetTypeIDConstant -}}
			, len({{.Target}})); err != nil {
				return err
			}
		}
		{{- else}}
		if err := oprot.WriteSetBegin(thrift.
		{{- .ValCtx.Type | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
			return err
		}
		{{- end}}
		{{- if Features.ValidateSet}}
		{{- $ctx := (.ValCtx.WithTarget "tgt").WithSource "src"}}
		for i := 0; i < len({{.Target}}); i++ {
			for j := i + 1; j < len({{.Target}}); j++ {
		{{- if Features.GenDeepEqual}}
				if func(tgt, src {{$ctx.TypeName}}) bool {
					{{- template "FieldDeepEqual" $ctx}}
					return true
				}({{.Target}}[i], {{.Target}}[j]) {
		{{- else}}
				{{- UseStdLibrary "reflect"}}
				if reflect.DeepEqual({{.Target}}[i], {{.Target}}[j]) {
		{{- end}}
					{{- UseStdLibrary "fmt"}}
					return thrift.PrependError("", fmt.Errorf("%T error writing set field: slice is not unique", {{.Target}}[i]))
				}
			}
		}
		{{- end}}
		for {{if Features.WithFieldMask}}i{{else}}_{{end}}, v := range {{.Target}} {
			{{- if Features.WithFieldMask}}
			{{- $curFieldMask = "nfm"}}
			if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(i); !ex {
				continue
			} else {
			{{- end}}
			{{- $ctx := (.ValCtx.WithTarget "v").WithFieldMask $curFieldMask -}}
			{{- template "FieldWrite" $ctx}}
			{{- if Features.WithFieldMask}}
			}
			{{- end}}
		}
		if err := oprot.WriteSetEnd(); err != nil {
			return err
		}
{{- end}}{{/* define "FieldWriteSet" */}}
`

// FieldWriteList .
var FieldWriteList = `
{{define "FieldWriteList"}}
{{- $isBaseVal := .ValCtx.Type | IsBaseType -}}
{{- $curFieldMask := .FieldMask -}}
	{{- if Features.WithFieldMask}}
	if !{{.FieldMask}}.All() {
		l := len({{.Target}})
		for i:=0; i < l; i++ {
			if _, ex := {{.FieldMask}}.Int(i); !ex {
				l--
			}
		}
		if err := oprot.WriteListBegin(thrift.
		{{- .ValCtx.Type | GetTypeIDConstant -}}
		, l); err != nil {
			return err
		}
	} else {
		if err := oprot.WriteListBegin(thrift.
		{{- .ValCtx.Type | GetTypeIDConstant -}}
		, len({{.Target}})); err != nil {
			return err
		}
	}
	{{- else}}
	if err := oprot.WriteListBegin(thrift.
	{{- .ValCtx.Type | GetTypeIDConstant -}}
	, len({{.Target}})); err != nil {
		return err
	}
	{{- end}}
		for {{if Features.WithFieldMask}}i{{else}}_{{end}}, v := range {{.Target}} {
			{{- if Features.WithFieldMask}}
			{{- $curFieldMask = "nfm"}}
			if {{if $isBaseVal}}_{{else}}{{$curFieldMask}}{{end}}, ex := {{.FieldMask}}.Int(i); !ex {
				continue
			} else {
			{{- end}}
			{{- $ctx := (.ValCtx.WithTarget "v").WithFieldMask $curFieldMask -}}
			{{- template "FieldWrite" $ctx}}
			{{- if Features.WithFieldMask}}
			}
			{{- end}}
		}
		if err := oprot.WriteListEnd(); err != nil {
			return err
		}
{{- end}}{{/* define "FieldWriteList" */}}
`
