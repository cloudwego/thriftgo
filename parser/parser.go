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

// Package parser parses a thrift IDL file with its dependencies into an abstract syntax tree.
// The acceptable IDL grammar is defined in the 'thrift.peg' file.
package parser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NOTSET is a value to express 'not set'.
const NOTSET = -999999

type parser struct {
	ThriftIDL
	Thrift
	IncludeDirs               []string
	Annotations               *Annotations
	DefinitionReservedComment string
}

func exists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

func search(file, dir string, includeDirs []string) (string, error) {
	ps := []string{file, filepath.Join(dir, file)}
	for _, inc := range includeDirs {
		ps = append(ps, filepath.Join(inc, file))
	}
	for _, p := range ps {
		if exists(p) {
			return normalizeFilename(p), nil
		}
	}
	return file, &os.PathError{Op: "search", Path: file, Err: os.ErrNotExist}
}

// ParseBatchString parses a group of string content and returns an AST.
// IDLContent is a map, which's key is IDLPath and value is IDL content.
func ParseBatchString(mainIDLFilePath string, IDLFileContentMap map[string]string, includeDirs []string) (*Thrift, error) {
	thriftMap := make(map[string]*Thrift)
	return parseBatchStringRecursively(mainIDLFilePath, includeDirs, thriftMap, IDLFileContentMap)
}

// doParseBatchString
func parseBatchStringRecursively(path string, includeDirs []string, thriftMap map[string]*Thrift, IDLFileContentMap map[string]string) (*Thrift, error) {

	bs, ok := IDLFileContentMap[path]
	if !ok {
		return nil, fmt.Errorf("no idl found for: %s\n", path)
	}
	if t, ok := thriftMap[path]; ok {
		return t, nil
	}

	t, err := parseString(path, bs, includeDirs)
	if err != nil {
		return nil, fmt.Errorf("parse %s err: %w\n", path, err)
	}
	thriftMap[path] = t
	dir := filepath.Dir(path)
	for _, inc := range t.Includes {
		incPath := filepath.Join(dir, inc.Path)
		t, err := parseBatchStringRecursively(incPath, includeDirs, thriftMap, IDLFileContentMap)
		if err != nil {
			return nil, err
		}
		inc.Reference = t
	}
	return t, nil
}

// ParseFile parses a thrift file and returns an AST.
// If recursive is true, then the include IDLs are parsed recursively as well.
func ParseFile(path string, includeDirs []string, recursive bool) (*Thrift, error) {
	if recursive {
		thriftMap := make(map[string]*Thrift)
		dir := filepath.Dir(normalizeFilename(path))
		return parseFileRecursively(path, dir, includeDirs, thriftMap)
	}
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseString(path, string(bs), includeDirs)
}

func parseFileRecursively(file, dir string, includeDirs []string, thriftMap map[string]*Thrift) (*Thrift, error) {
	path, err := search(file, dir, includeDirs)
	if err != nil {
		return nil, err
	}
	if t, ok := thriftMap[path]; ok {
		return t, nil
	}
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t, err := parseString(path, string(bs), includeDirs)
	if err != nil {
		return nil, fmt.Errorf("parse %s err: %w", path, err)
	}
	thriftMap[path] = t
	dir = filepath.Dir(path)
	for _, inc := range t.Includes {
		t, err := parseFileRecursively(inc.Path, dir, includeDirs, thriftMap)
		if err != nil {
			return nil, err
		}
		inc.Reference = t
	}
	return t, nil
}

// ParseString parses the thrift file path and file content then return an AST.
func ParseString(path, content string) (*Thrift, error) {
	return parseString(path, content, nil)
}

func parseString(path, content string, includeDirs []string) (*Thrift, error) {
	p := &parser{
		IncludeDirs: includeDirs,
	}
	p.Filename = path
	p.Buffer = content
	p.Init()
	if err := p.ThriftIDL.Parse(); err != nil {
		return nil, err
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return &p.Thrift, nil
}

func (p *parser) parse() (err error) {
	root := p.AST()
	if root == nil || root.pegRule != ruleDocument {
		return errors.New("not document")
	}
	// Header* Definition* !.
	for n := root.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleSkip, ruleSkipLine:
			continue
		case ruleHeader:
			if err := p.parseHeader(n); err != nil {
				return err
			}
		case ruleDefinition:
			if err := p.parseDefinition(n); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown rule: " + rul3s[n.pegRule])
		}
	}
	return nil
}

func (p *parser) pegText(node *node32) string {
	for n := node; n != nil; n = n.next {
		if s := p.pegText(n.up); s != "" {
			return s
		}
		if n.pegRule != rulePegText {
			continue
		}

		quote := p.buffer[n.begin-1]
		runes := make([]rune, 0, n.end-n.begin)
		for i := n.begin; i < n.end-1; i++ {
			r := p.buffer[i]

			// unescape \' for single quoted, \" for double quoted
			if r == '\\' {
				switch p.buffer[i+1] {
				case '\\':
					i++
					runes = append(runes, r)
				case quote:
					continue
				}
			}
			runes = append(runes, r)
		}
		runes = append(runes, p.buffer[n.end-1])
		return string(runes)
	}
	return ""
}

func (p *parser) parseHeader(node *node32) (err error) {
	node, err = checkrule(node, ruleHeader)
	if err != nil {
		return err
	}
	// Skip (Include / CppInclude / Namespace) SkipLine
	if node.pegRule == ruleSkip {
		node = node.next
	}
	switch node.pegRule {
	case ruleInclude:
		if err := p.parseInclude(node); err != nil {
			return err
		}
	case ruleNamespace:
		if err := p.parseNamespace(node); err != nil {
			return err
		}
	case ruleCppInclude:
		if err := p.parseCppInclude(node); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown rule: " + rul3s[node.pegRule])
	}
	return nil
}

func (p *parser) parseInclude(node *node32) (err error) {
	node, err = checkrule(node, ruleInclude)
	if err != nil {
		return err
	}
	filename := p.pegText(node)
	if filename == "" {
		return
	}
	for _, inc := range p.Includes {
		if inc.Path == filename {
			return
		}
	}
	p.Includes = append(p.Includes, &Include{Path: filename})
	return nil
}

func (p *parser) parseCppInclude(node *node32) (err error) {
	node, err = checkrule(node, ruleCppInclude)
	if err != nil {
		return err
	}
	p.CppIncludes = append(p.CppIncludes, p.pegText(node))
	return nil
}

func (p *parser) parseNamespace(node *node32) (err error) {
	ns := Namespace{}
	node, err = checkrule(node, ruleNamespace)
	if err != nil {
		return err
	}
	node = node.next // ignore "namespace"
	ns.Language = p.pegText(node.up)
	node = node.next
	ns.Name = p.pegText(node.up)
	node = node.next
	if node != nil && node.pegRule == ruleAnnotations { // ANNOTATIONS
		ns.Annotations, err = p.parseAnnotations(node)
		if err != nil {
			return err
		}
	}
	p.Namespaces = append(p.Namespaces, &ns)
	return nil
}

func (p *parser) parseDefinition(node *node32) (err error) {
	node, err = checkrule(node, ruleDefinition)
	if err != nil {
		return err
	}
	p.DefinitionReservedComment = ""
	// ReservedComments Skip (Const / Typedef / Enum / Struct / Union / Service / Exception) Annotations? SkipLine
	if node.pegRule == ruleReservedComments {
		reservedComments, err := p.parseReservedComments(node)
		if err != nil {
			return err
		}
		p.DefinitionReservedComment = reservedComments
		node = node.next
	}
	// skip Skip
	if node.pegRule == ruleSkip {
		node = node.next
	}
	switch node.pegRule {
	case ruleConst:
		if err := p.parseConst(node); err != nil {
			return err
		}
	case ruleTypedef:
		if err := p.parseTypedef(node); err != nil {
			return err
		}
	case ruleEnum:
		if err := p.parseEnum(node); err != nil {
			return err
		}
	case ruleUnion:
		if err := p.parseUnion(node); err != nil {
			return err
		}
	case ruleStruct:
		if err := p.parseStruct(node); err != nil {
			return err
		}
	case ruleException:
		if err := p.parseException(node); err != nil {
			return err
		}
	case ruleService:
		if err := p.parseService(node); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown rule: " + rul3s[node.pegRule])
	}
	node = node.next
	if node != nil && node.pegRule == ruleAnnotations {
		ann, err := p.parseAnnotations(node)
		if err != nil {
			return err
		}
		*p.Annotations = ann
	}
	return nil
}

func (p *parser) parseConst(node *node32) (err error) {
	node, err = checkrule(node, ruleConst)
	if err != nil {
		return err
	}
	// CONST FieldType Identifier EQUAL ConstValue ListSeparator?
	node = node.next // ignore CONST
	ft, err := p.parseFieldType(node)
	if err != nil {
		return err
	}
	node = node.next
	name := p.pegText(node)
	node = node.next.next // ignore EQUAL
	value, err := p.parseConstValue(node)
	if err != nil {
		return err
	}
	c := &Constant{Name: name, Type: ft, Value: value}
	c.ReservedComments = p.DefinitionReservedComment
	p.Constants = append(p.Constants, c)
	p.Annotations = &c.Annotations
	return nil
}

func (p *parser) parseFieldType(node *node32) (typ *Type, err error) {
	node, err = checkrule(node, ruleFieldType)
	if err != nil {
		return nil, err
	}
	// ContainerType / BaseType / Identifier
	switch node.pegRule {
	case ruleContainerType:
		typ, err = p.parseContainerType(node)
		if err != nil {
			return typ, err
		}
	case ruleIdentifier, ruleBaseType:
		typ = &Type{Name: p.pegText(node)}
	default:
		return typ, fmt.Errorf("unknown rule: " + rul3s[node.pegRule])
	}
	node = node.next
	if node != nil && node.pegRule == ruleAnnotations {
		typ.Annotations, err = p.parseAnnotations(node)
		if err != nil {
			return typ, err
		}
	}
	return typ, nil
}

func (p *parser) parseContainerType(node *node32) (typ *Type, err error) {
	node, err = checkrule(node, ruleContainerType)
	if err != nil {
		return &Type{}, err
	}
	// MapType / SetType / ListType
	switch node.pegRule {
	case ruleMapType: // MAP CppType? LPOINT FieldType COMMA FieldType RPOINT
		node = node.up.next // ignore MAP LPOINT
		var cppType string
		if node.pegRule == ruleCppType {
			cppType = p.pegText(node.up.next) // ignore CPPTYPE
			node = node.next
		}
		node = node.next // ignore LPOINT
		kt, err := p.parseFieldType(node)
		if err != nil {
			return kt, err
		}
		node = node.next.next // ignore COMMA
		vt, err := p.parseFieldType(node)
		if err != nil {
			return vt, err
		}
		return &Type{Name: "map", KeyType: kt, ValueType: vt, CppType: cppType}, nil
	case ruleSetType: // SET CppType? LPOINT FieldType RPOINT
		node = node.up.next // ignore SET
		var cppType string
		if node.pegRule == ruleCppType {
			cppType = p.pegText(node.up.next) // ignore CPPTYPE
			node = node.next
		}
		node = node.next // ignore LPOINT
		vt, err := p.parseFieldType(node)
		if err != nil {
			return vt, err
		}
		return &Type{Name: "set", ValueType: vt, CppType: cppType}, nil
	case ruleListType: // LIST LPOINT FieldType RPOINT CppType?
		node = node.up.next.next // ignore LIST LPOINT
		vt, err := p.parseFieldType(node)
		if err != nil {
			return vt, err
		}
		node = node.next.next // ignore RPOINT
		var cppType string
		if node != nil && node.pegRule == ruleCppType {
			cppType = p.pegText(node.up.next) // ignore CPPTYPE
		}
		return &Type{Name: "list", ValueType: vt, CppType: cppType}, nil
	default:
		return &Type{}, fmt.Errorf("unknown rule: " + rul3s[node.pegRule])
	}
}

func (p *parser) parseConstValue(node *node32) (cv *ConstValue, err error) {
	node, err = checkrule(node, ruleConstValue)
	if err != nil {
		return nil, err
	}
	// DoubleConstant / IntConstant / Literal / Identifier / ConstList / ConstMap
	switch node.pegRule {
	case ruleDoubleConstant:
		double, _ := strconv.ParseFloat(p.pegText(node), 64)
		return &ConstValue{Type: ConstType_ConstDouble, TypedValue: &ConstTypedValue{Double: &double}}, nil
	case ruleIntConstant:
		i, err := strconv.ParseInt(p.pegText(node), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("parseConstValue failed at '%s': %w", p.pegText(node), err)
		}
		return &ConstValue{Type: ConstType_ConstInt, TypedValue: &ConstTypedValue{Int: &i}}, nil
	case ruleLiteral:
		literal := p.pegText(node)
		return &ConstValue{Type: ConstType_ConstLiteral, TypedValue: &ConstTypedValue{Literal: &literal}}, nil
	case ruleIdentifier:
		identifier := p.pegText(node)
		return &ConstValue{Type: ConstType_ConstIdentifier, TypedValue: &ConstTypedValue{Identifier: &identifier}}, nil
	case ruleConstList:
		// LBRK (ConstValue ListSeparator?)* RBRK
		ret := []*ConstValue{} // important: can't not be nil
		for n := node.up; n != nil; n = n.next {
			if n.pegRule == ruleConstValue {
				val, err := p.parseConstValue(n)
				if err != nil {
					return nil, err
				}
				ret = append(ret, val)
			}
		}
		return &ConstValue{Type: ConstType_ConstList, TypedValue: &ConstTypedValue{List: ret}}, nil
	case ruleConstMap:
		node = node.up
		// LWING (ConstValue COLON ConstValue ListSeparator?)* RWING
		ret := []*MapConstValue{} // important: can't not be nil
		for n := node.next; n != nil; n = n.next {
			if n.pegRule != ruleConstValue {
				continue
			}
			k, err := p.parseConstValue(n)
			if err != nil {
				return nil, err
			}
			n = n.next.next // ignore COLON
			v, err := p.parseConstValue(n)
			if err != nil {
				return nil, err
			}
			ret = append(ret, &MapConstValue{Key: k, Value: v})
		}
		return &ConstValue{Type: ConstType_ConstMap, TypedValue: &ConstTypedValue{Map: ret}}, nil
	default:
		return nil, fmt.Errorf("unknown rule: " + rul3s[node.pegRule])
	}
}

func (p *parser) parseTypedef(node *node32) (err error) {
	node, err = checkrule(node, ruleTypedef)
	if err != nil {
		return err
	}
	// TYPEDEF FieldType Identifier
	node = node.next // ignore TYPEDEF
	ft, err := p.parseFieldType(node)
	if err != nil {
		return err
	}
	var typd Typedef
	typd.Type = ft
	node = node.next
	typd.Alias = p.pegText(node)
	typd.ReservedComments = p.DefinitionReservedComment
	p.Typedefs = append(p.Typedefs, &typd)
	p.Annotations = &typd.Annotations
	return nil
}

func (p *parser) parseEnum(node *node32) (err error) {
	node, err = checkrule(node, ruleEnum)
	if err != nil {
		return err
	}
	// ENUM Identifier LWING ( ReservedComments Identifier (EQUAL IntConstant)? Annotations? ListSeparator? ReservedEndLineComments SkipLine)* RWING
	node = node.next // ignore ENUM
	name := p.pegText(node)
	var values []*EnumValue
	for n := node.next.next; n != nil; n = n.next {
		valueComments := ""
		if n.pegRule == ruleReservedComments {
			valueComments, err = p.parseReservedComments(n)
			if err != nil {
				return err
			}
			n = n.next
		}
		if n.pegRule == ruleIdentifier {
			var v EnumValue
			v.ReservedComments = valueComments
			v.Name = p.pegText(n)
			if n.next.pegRule == ruleEQUAL {
				n = n.next.next
				v.Value, _ = strconv.ParseInt(p.pegText(n), 0, 64)
			} else {
				if len(values) == 0 {
					v.Value = 0
				} else {
					v.Value = values[len(values)-1].Value + 1
				}
			}

			if n.next.pegRule == ruleAnnotations {
				n = n.next
				v.Annotations, err = p.parseAnnotations(n)
				if err != nil {
					return err
				}
			}
			if n.next.pegRule == ruleListSeparator {
				n = n.next
			}
			if n.next.pegRule == ruleReservedEndLineComments && v.ReservedComments == "" {
				n = n.next
				v.ReservedComments, err = p.parseReservedEndLineComments(n)
				if err != nil {
					return err
				}
			}
			values = append(values, &v)
		}
	}
	e := &Enum{Name: name, Values: values}
	e.ReservedComments = p.DefinitionReservedComment
	p.Enums = append(p.Enums, e)
	p.Annotations = &e.Annotations
	return nil
}

func (p *parser) parseUnion(node *node32) (err error) {
	node, err = checkrule(node, ruleUnion)
	if err != nil {
		return err
	}
	// UNION Identifier LWING Field* RWING
	node = node.next // ignore UNION
	name := p.pegText(node)
	node = node.next
	var fields []*Field
	for n := node.next; n != nil; n = n.next {
		switch n.pegRule {
		case ruleField:
			field, err := p.parseField(n)
			if err != nil {
				return err
			}
			if field.ID == NOTSET {
				if len(fields) > 0 {
					field.ID = fields[len(fields)-1].ID + 1
				} else {
					field.ID = 1
				}
			}
			fields = append(fields, field)
		}
	}
	u := &StructLike{Category: "union", Name: name, Fields: fields}
	u.ReservedComments = p.DefinitionReservedComment
	p.Unions = append(p.Unions, u)
	p.Annotations = &u.Annotations
	return nil
}

func (p *parser) parseStruct(node *node32) (err error) {
	node, err = checkrule(node, ruleStruct)
	if err != nil {
		return err
	}
	// STRUCT Identifier LWING Field* RWING
	node = node.next // ignore STRUCT
	name := p.pegText(node)
	node = node.next
	var fields []*Field
	for n := node.next; n != nil; n = n.next {
		switch n.pegRule {
		case ruleField:
			field, err := p.parseField(n)
			if err != nil {
				return err
			}
			if field.ID == NOTSET {
				if len(fields) > 0 {
					field.ID = fields[len(fields)-1].ID + 1
				} else {
					field.ID = 1
				}
			}
			fields = append(fields, field)
		}
	}
	s := &StructLike{Category: "struct", Name: name, Fields: fields}
	s.ReservedComments = p.DefinitionReservedComment
	p.Structs = append(p.Structs, s)
	p.Annotations = &s.Annotations
	return nil
}

func (p *parser) parseException(node *node32) (err error) {
	node, err = checkrule(node, ruleException)
	if err != nil {
		return err
	}
	// EXCEPTION Identifier LWING Field* RWING
	node = node.next // ignore EXCEPTION
	name := p.pegText(node)
	var fields []*Field
	for n := node.next; n != nil; n = n.next {
		if n.pegRule == ruleField {
			field, err := p.parseField(n)
			if err != nil {
				return err
			}
			if field.ID == NOTSET {
				if len(fields) > 0 {
					field.ID = fields[len(fields)-1].ID + 1
				} else {
					field.ID = 1
				}
			}
			fields = append(fields, field)
		}
	}
	e := &StructLike{Category: "exception", Name: name, Fields: fields}
	e.ReservedComments = p.DefinitionReservedComment
	p.Exceptions = append(p.Exceptions, e)
	p.Annotations = &e.Annotations
	return nil
}

func (p *parser) parseField(node *node32) (field *Field, err error) {
	node, err = checkrule(node, ruleField)
	if err != nil {
		return nil, err
	}
	// ReservedComments Skip FieldId? FieldReq? FieldType Identifier (EQUAL ConstValue)? Annotations? ListSeparator? ReservedEndLineComments
	var f Field
	f.ID = NOTSET
	for ; node != nil; node = node.next {
		switch node.pegRule {
		case ruleSkip, ruleSkipLine:
			continue
		case ruleReservedComments:
			reservedComments, err := p.parseReservedComments(node)
			if err != nil {
				return nil, err
			}
			f.ReservedComments = reservedComments
		case ruleReservedEndLineComments:
			// endline comment's priority is lower than the header comment
			reservedComments, err := p.parseReservedEndLineComments(node)
			if err != nil {
				return nil, err
			}
			if f.ReservedComments == "" {
				f.ReservedComments = reservedComments
			}
		case ruleFieldId:
			i, _ := strconv.Atoi(p.pegText(node))
			f.ID = int32(i)
		case ruleFieldReq:
			require := p.pegText(node)
			f.Requiredness = FieldType_Default
			if require == "required" {
				f.Requiredness = FieldType_Required
			} else if require == "optional" {
				f.Requiredness = FieldType_Optional
			}
		case ruleFieldType:
			f.Type, err = p.parseFieldType(node)
			if err != nil {
				return nil, err
			}
		case ruleIdentifier:
			f.Name = p.pegText(node)
		case ruleEQUAL:
			node = node.next // ignore EQUAL
			f.Default, err = p.parseConstValue(node)
			if err != nil {
				return nil, err
			}
		case ruleAnnotations:
			f.Annotations, err = p.parseAnnotations(node)
			if err != nil {
				return nil, err
			}
		}
	}
	return &f, nil
}

func (p *parser) parseAnnotations(node *node32) ([]*Annotation, error) {
	// LPAR Annotation+ RPAR
	var err error
	var ret Annotations
	node, err = checkrule(node, ruleAnnotations)
	if err != nil {
		return nil, err
	}
	for node = node.next; node != nil; node = node.next {
		if node.pegRule == ruleAnnotation {
			k, v, err := p.parseAnnotation(node)
			if err != nil {
				return nil, err
			}
			ret.Append(k, v)
		}
	}
	return ret, nil
}

func (p *parser) parseAnnotation(node *node32) (k, v string, err error) {
	// Identifier EQUAL Literal ListSeparator?
	node, err = checkrule(node, ruleAnnotation)
	if err != nil {
		return "", "", err
	}

	k = p.pegText(node) // Identifier
	node = node.next.next
	v = p.pegText(node) // Literal
	return k, v, nil
}

func (p *parser) parseService(node *node32) (err error) {
	node, err = checkrule(node, ruleService)
	if err != nil {
		return err
	}
	// SERVICE Identifier ( EXTENDS Identifier )? LWING Function* RWING
	var s Service
	node = node.next // ignore SERVICE
	s.Name = p.pegText(node)
	node = node.next
	if node.pegRule == ruleEXTENDS {
		s.Extends = p.pegText(node)
		node = node.next
	}
	for node = node.next; node != nil; node = node.next {
		if node.pegRule == ruleFunction {
			fu, err := p.parseFunction(node)
			if err != nil {
				return err
			}
			s.Functions = append(s.Functions, fu)
		}
	}
	s.ReservedComments = p.DefinitionReservedComment
	p.Services = append(p.Services, &s)
	p.Annotations = &s.Annotations
	return nil
}

func (p *parser) parseFunction(node *node32) (fu *Function, err error) {
	node, err = checkrule(node, ruleFunction)
	if err != nil {
		return nil, err
	}
	// ReservedComments ONEWAY? FunctionType Identifier LPAR Field* RPAR Throws? Annotations? ListSeparator?
	var f Function
	for ; node != nil; node = node.next {
		switch node.pegRule {
		case ruleReservedComments:
			reservedComments, err := p.parseReservedComments(node)
			if err != nil {
				return nil, err
			}
			f.ReservedComments = reservedComments
		case ruleONEWAY:
			f.Oneway = true
		case ruleFunctionType:
			n := node.up
			if n.pegRule == ruleFieldType {
				f.FunctionType, err = p.parseFieldType(n)
				if err != nil {
					return nil, err
				}
			} else if n.pegRule == ruleVOID {
				f.Void = true
				f.FunctionType = &Type{Name: "void"}
			}
		case ruleIdentifier:
			f.Name = p.pegText(node)
		case ruleField:
			field, err := p.parseField(node)
			if err != nil {
				return nil, err
			}
			f.Arguments = addField(f.Arguments, field)
		case ruleThrows:
			f.Throws, err = p.parseThrows(node)
			if err != nil {
				return nil, err
			}
		case ruleAnnotations:
			f.Annotations, err = p.parseAnnotations(node)
			if err != nil {
				return nil, err
			}
		}
	}
	return &f, nil
}

func (p *parser) parseThrows(node *node32) (fs []*Field, err error) {
	node, err = checkrule(node, ruleThrows)
	if err != nil {
		return nil, err
	}
	// THROWS LPAR Field* RPAR
	node = node.next // ignore THROWS
	var fields []*Field
	for ; node != nil; node = node.next {
		if node.pegRule == ruleField {
			field, err := p.parseField(node)
			if err != nil {
				return nil, err
			}
			field.Requiredness = FieldType_Optional
			fields = addField(fields, field)
		}
	}
	return fields, nil
}

func (p *parser) parseReservedComments(node *node32) (ReservedComments string, err error) {
	node, err = checkrule(node, ruleReservedComments)
	if err != nil {
		return "", err
	}
	// Skip
	node = node.up
	// (Space / Comment)*
	var comments []string
	for ; node != nil; node = node.next {
		if node.pegRule == ruleComment {
			comment := strings.TrimRight(string(p.buffer[int(node.up.begin):int(node.up.end)]), "\r\n")
			if strings.HasPrefix(comment, "#") {
				comment = strings.TrimPrefix(comment, "#")
				comment = "//" + comment
			}
			if comment != "" {
				comments = append(comments, comment)
			}
		}
	}
	return strings.Join(comments, "\n"), nil
}

func (p *parser) parseReservedEndLineComments(node *node32) (ReservedComments string, err error) {
	node, err = checkrule(node, ruleReservedEndLineComments)
	if err != nil {
		return "", err
	}
	// Skip
	node = node.up
	// (Space / Comment)*
	var comments []string
	for ; node != nil; node = node.next {
		if node.pegRule == ruleComment {
			comment := strings.TrimRight(string(p.buffer[int(node.up.begin):int(node.up.end)]), "\r\n")
			if strings.HasPrefix(comment, "#") {
				comment = strings.TrimPrefix(comment, "#")
				comment = "//" + comment
			}
			if comment != "" {
				comments = append(comments, comment)
			}
		}
	}
	return strings.Join(comments, "\n"), nil
}
