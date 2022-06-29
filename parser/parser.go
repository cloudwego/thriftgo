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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/thriftgo/parser/token"
)

func exists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

func search(file, dir string, includeDirs []string) (string, error) {
	ps := []string{filepath.Join(dir, file)}
	for _, inc := range includeDirs {
		ps = append(ps, filepath.Join(inc, file))
	}
	for _, p := range ps {
		if exists(p) {
			if filepath.IsAbs(p) {
				return p, nil
			}
			if v, err := filepath.Abs(p); err == nil {
				return v, nil
			}
			return p, nil
		}
	}
	return file, &os.PathError{Op: "search", Path: file, Err: os.ErrNotExist}
}

// ParseFile parses a thrift file and returns an AST.
// If recursive is true, then the include IDLs are parsed recursively as well.
func ParseFile(path string, includeDirs []string, recursive bool) (*Thrift, error) {
	if recursive {
		parsed := make(map[string]*Thrift)
		return parseFileRecursively(path, "", includeDirs, parsed)
	}
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	p := newParser(path, src)
	return p.parse()
}

func parseFileRecursively(file, dir string, includeDirs []string, parsed map[string]*Thrift) (*Thrift, error) {
	path, err := search(file, dir, includeDirs)
	if err != nil {
		return nil, err
	}
	if t, ok := parsed[path]; ok {
		return t, nil
	}
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	p := newParser(path, src)
	t, err := p.parse()
	if err != nil {
		return nil, err
	}
	parsed[path] = t
	dir = filepath.Dir(path)
	for _, inc := range t.Includes {
		t, err := parseFileRecursively(inc.Path, dir, includeDirs, parsed)
		if err != nil {
			return nil, err
		}
		inc.Reference = t
	}
	return t, nil
}

// ParseString parses the thrift file path and file content then return an AST.
func ParseString(path, content string) (*Thrift, error) {
	src := strings.NewReader(content)
	p := newParser(path, src)
	return p.parse()
}

type parser struct {
	lexer *token.Tokenizer
	ast   *Thrift
	last  token.Token // the last token read
	next  token.Token // the next token to be read
	pAnno *Annotations

	comments []string
	comspans []token.Span

	// TODO: remove this by designing the grammar
	newlineNumberBeforeNext int
}

func newParser(filename string, src io.Reader) *parser {
	return &parser{
		lexer: token.NewTokenizer(src),
		ast: &Thrift{
			Filename: filename,
		},
	}
}

func meet(actual, expected token.Tok, more []token.Tok) bool {
	if actual == expected {
		return true
	}
	for _, t := range more {
		if t == actual {
			return true
		}
	}
	return false
}

func (p *parser) expect(t token.Tok, more ...token.Tok) {
	move := func() token.Token {
		t := p.lexer.Next()
		return t
	}

	if t == token.NewLine && p.newlineNumberBeforeNext > 0 {
		p.newlineNumberBeforeNext--
		return
	}

	if !meet(p.next.Tok, t, more) {
		panic(p.syntaxError(t, more))
	}

	p.last = p.next
	p.newlineNumberBeforeNext = 0
	for {
		p.next = move()
		switch p.next.Tok {
		case token.LexicalError:
			panic(p.lexicalError())
		case token.EOF:
			return
		case token.BlockComment, token.LineComment, token.UnixComment:
			if p.last.Span.End != 0 {
				ln1 := p.lexer.LineSpan(p.last.Span).End
				ln2 := p.lexer.LineSpan(p.next.Span).Beg
				if ln1 == ln2 {
					// discard non-prefix comments
					continue
				}
			}
			p.comments = append(p.comments, p.next.AsString())
			p.comspans = append(p.comspans, p.lexer.LineSpan(p.next.Span))
		case token.Whitespaces:
		case token.NewLine:
			p.newlineNumberBeforeNext++
		default:
			return
		}
	}
}

func (p *parser) parse() (ast *Thrift, err error) {
	defer func() {
		if x := recover(); x != nil {
			switch v := x.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("panic: %+v", v)
			}
		}
	}()
	p.expect(p.next.Tok) // read the first token
	p.parserDocument()
	return p.ast, nil
}

// Document <- Header* Definition* !.
func (p *parser) parserDocument() {
	p.parseHeaders()
	p.parseDefinitions()
	p.expect(token.EOF)
}

// Header <- (Include / CppInclude / Namespace) Newline
func (p *parser) parseHeaders() {
	for {
		switch p.next.Tok {
		case token.Include:
			// Include    <- 'include' Literal
			p.expect(token.Include)
			p.expect(token.StringLiteral)
			p.ast.Includes = append(p.ast.Includes, &Include{
				Path: p.last.Unquote(),
			})

		case token.CppInclude:
			// CppInclude <- 'cppinclude' Literal
			p.expect(token.CppInclude)
			p.expect(token.StringLiteral)
			p.ast.CppIncludes = append(p.ast.CppIncludes, p.last.AsString())

		case token.Namespace:
			// Namespace  <- 'namespace' NamespaceScope Identifier Annotations?
			p.expect(token.Namespace)
			p.expect(token.Identifier, token.Asterisk)

			ns := &Namespace{}
			ns.Language = p.last.AsString()

			p.expect(token.Identifier)
			ns.Name = p.last.AsString()

			if p.next.Tok == token.LParenthesis {
				ns.Annotations = p.parseAnnotations()
			}
			p.ast.Namespaces = append(p.ast.Namespaces, ns)

		default:
			return
		}
	}
}

// Definition <- (Const / Typedef / Enum / Service / Struct / Union / Exception) Annotations? NewLine
func (p *parser) parseDefinitions() {
	for {
		switch p.next.Tok {
		case token.Const:
			p.parseConst()
		case token.Typedef:
			p.parseTypedef()
		case token.Enum:
			p.parseEnum()
		case token.Service:
			p.parseService()
		case token.Struct, token.Union, token.Exception:
			p.parseStruct()
		default:
			return
		}
		if p.next.Tok == token.LParenthesis {
			if p.pAnno != nil {
				*p.pAnno = p.parseAnnotations()
			}
		}
		if p.next.Tok == token.Semicolon {
			p.expect(token.Semicolon)
		}
		p.expect(token.NewLine)
	}
}

// Const <- 'const' FieldType Identifier '=' ConstValue
func (p *parser) parseConst() {
	comment := p.prefixComment()
	p.expect(token.Const)

	typ := p.parseFieldType()

	p.expect(token.Identifier)
	id := p.last.AsString()

	p.expect(token.Equal)
	cv := p.parseConstValue()

	c := &Constant{
		Name:             id,
		Type:             typ,
		Value:            cv,
		ReservedComments: comment,
	}
	p.ast.Constants = append(p.ast.Constants, c)
	p.pAnno = &c.Annotations
}

// Typedef <- 'typedef' FieldType Identifier
func (p *parser) parseTypedef() {
	comment := p.prefixComment()
	p.expect(token.Typedef)

	typ := p.parseFieldType()

	p.expect(token.Identifier)
	id := p.last.AsString()

	td := &Typedef{
		Type:             typ,
		Alias:            id,
		ReservedComments: comment,
	}
	p.ast.Typedefs = append(p.ast.Typedefs, td)
	p.pAnno = &td.Annotations
}

// Enum <- 'enum' Identifier '{' (Identifier ('=' IntLiteral )? Annotations? ListSeparator? NewLine)* '}'
func (p *parser) parseEnum() {
	comment := p.prefixComment()
	p.expect(token.Enum)
	p.expect(token.Identifier)
	e := &Enum{
		Name:             p.last.AsString(),
		ReservedComments: comment,
	}
	p.expect(token.LBrace)
	for p.next.Tok == token.Identifier {
		comment := p.prefixComment()
		p.expect(token.Identifier)
		ev := &EnumValue{
			Name:             p.last.AsString(),
			ReservedComments: comment,
		}

		if p.next.Tok == token.Equal {
			p.expect(token.Equal)
			p.expect(token.IntLiteral)
			ev.Value = p.last.AsInt()
		} else {
			if len(e.Values) == 0 {
				ev.Value = 0
			} else {
				ev.Value = e.Values[len(e.Values)-1].Value + 1
			}
		}

		if p.next.Tok == token.LParenthesis {
			ev.Annotations = p.parseAnnotations()
		}
		p.consumeListSeparator()
		e.Values = append(e.Values, ev)
	}
	p.expect(token.RBrace)
	p.pAnno = &e.Annotations
	p.ast.Enums = append(p.ast.Enums, e)
}

// Service <- 'service' Identifier ( EXTENDS Identifier )? '{' Function* '}'
func (p *parser) parseService() {
	comment := p.prefixComment()
	p.expect(token.Service)
	p.expect(token.Identifier)
	svc := &Service{
		Name:             p.last.AsString(),
		ReservedComments: comment,
	}
	if p.next.Tok == token.Extends {
		p.expect(token.Extends)
		p.expect(token.Identifier)
		svc.Extends = p.last.AsString()
	}

	p.expect(token.LBrace)
	for p.next.Tok != token.RBrace {
		fn := p.parseFunction()
		svc.Functions = append(svc.Functions, fn)
	}
	p.expect(token.RBrace)

	p.pAnno = &svc.Annotations
	p.ast.Services = append(p.ast.Services, svc)
}

// Function <- 'oneway'? FunctionType Identifier '(' Field* ')' Throws? Annotations? ListSeparator? NewLine
func (p *parser) parseFunction() *Function {
	comment := p.prefixComment()
	fn := &Function{
		ReservedComments: comment,
	}
	if p.next.Tok == token.Oneway {
		p.expect(token.Oneway)
		fn.Oneway = true
	}
	// FunctionType <- 'void' / FieldType
	if p.next.Tok == token.Void {
		p.expect(token.Void)
		fn.Void = true
	} else {
		fn.FunctionType = p.parseFieldType()
	}
	p.expect(token.Identifier)
	fn.Name = p.last.AsString()

	// Arguments
	p.expect(token.LParenthesis)
	for p.next.Tok != token.RParenthesis {
		arg := p.parseField()
		fn.Arguments = append(fn.Arguments, arg)
	}
	p.expect(token.RParenthesis)

	// Throws
	if p.next.Tok == token.Throws {
		p.expect(token.Throws)
		p.expect(token.LParenthesis)
		for p.next.Tok != token.RParenthesis {
			ex := p.parseField()
			fn.Throws = append(fn.Throws, ex)
		}
		p.expect(token.RParenthesis)
	}

	if p.next.Tok == token.LParenthesis {
		fn.Annotations = p.parseAnnotations()
	}
	p.consumeListSeparator()
	return fn
}

// Struct    <- 'struct' Identifier '{' Field* '}'
// Union     <- 'union' Identifier '{' Field* '}'
// Exception <- 'exception' Identifier '{' Field* '}'
func (p *parser) parseStruct() {
	comment := p.prefixComment()
	s := &StructLike{
		ReservedComments: comment,
	}
	p.expect(token.Struct, token.Union, token.Exception)
	s.Category = p.last.AsString()

	p.expect(token.Identifier)
	s.Name = p.last.AsString()

	p.expect(token.LBrace)
	for p.next.Tok != token.RBrace {
		f := p.parseField()
		s.Fields = append(s.Fields, f)
	}
	p.expect(token.RBrace)
	p.pAnno = &s.Annotations
	switch s.Category {
	case "struct":
		p.ast.Structs = append(p.ast.Structs, s)
	case "union":
		p.ast.Unions = append(p.ast.Unions, s)
	case "exception":
		p.ast.Exceptions = append(p.ast.Exceptions, s)
	default:
		panic("?")
	}
}

// Field <- FieldId? FieldReq? FieldType Identifier ('=' ConstValue)? Annotations? ListSeparator? NewLine
func (p *parser) parseField() *Field {
	comment := p.prefixComment()
	f := &Field{
		ReservedComments: comment,
	}

	// FieldId  <- IntLiteral ':'
	if p.next.Tok == token.IntLiteral {
		p.expect(token.IntLiteral)
		f.ID = int32(p.last.AsInt())
		p.expect(token.Colon)
	}

	// FieldReq <- 'required' / 'optional'
	if p.next.Tok == token.Required || p.next.Tok == token.Optional {
		p.expect(token.Required, token.Optional)
		if p.last.Tok == token.Required {
			f.Requiredness = FieldType_Required
		} else {
			f.Requiredness = FieldType_Optional
		}
	}
	f.Type = p.parseFieldType()

	p.expect(token.Identifier)
	f.Name = p.last.AsString()

	if p.next.Tok == token.Equal {
		p.expect(token.Equal)
		f.Default = p.parseConstValue()
	}
	if p.next.Tok == token.LParenthesis {
		f.Annotations = p.parseAnnotations()
	}
	p.consumeListSeparator()
	return f
}

// FieldType      <- (ContainerType / BaseType / Identifier) Annotations?
// BaseType       <- 'bool' / 'byte' / 'i8' / 'i16' / 'i32' / 'i64' / 'double' / 'string' / 'binary'
// ContainerType  <- MapType / SetType / ListType
// MapType        <- 'map' CppType? '<' FieldType ',' FieldType '>'
// SetType        <- 'set' CppType? '<' FieldType '>'
// ListType       <- 'list' '<' FieldType '>' CppType?
func (p *parser) parseFieldType() (typ *Type) {
	p.expect(token.Identifier,
		token.Bool, token.Byte,
		token.I8, token.I16, token.I32, token.I64,
		token.Double, token.String, token.Binary,
		token.Map, token.Set, token.List)

	defer func() {
		if typ != nil && p.next.Tok == token.LParenthesis {
			typ.Annotations = p.parseAnnotations()
		}
	}()
	switch p.last.Tok {
	case token.Identifier,
		token.Bool, token.Byte,
		token.I8, token.I16, token.I32, token.I64,
		token.Double, token.String, token.Binary:
		return &Type{Name: p.last.AsString()}

	case token.Map:
		typ = &Type{Name: p.last.AsString()}
		p.checkCPPType()
		p.expect(token.LChevron)
		typ.KeyType = p.parseFieldType()
		p.expect(token.Comma)
		typ.ValueType = p.parseFieldType()
		p.expect(token.RChevron)
		return typ

	case token.Set:
		typ = &Type{Name: p.last.AsString()}
		p.checkCPPType()
		p.expect(token.LChevron)
		typ.ValueType = p.parseFieldType()
		p.expect(token.RChevron)
		return typ

	case token.List:
		typ = &Type{Name: p.last.AsString()}
		p.expect(token.LChevron)
		typ.ValueType = p.parseFieldType()
		p.expect(token.RChevron)
		p.checkCPPType()
		return typ

	default:
		panic("?")
	}
}

// for compatibility
func (p *parser) checkCPPType() {
	if p.next.Tok == token.CppType {
		p.expect(token.CppType)
		p.expect(token.StringLiteral)
	}
}

// ConstValue <- FloatLiteral / IntLiteral / Literal / Identifier / ConstList / ConstMap
// ConstList  <- '[' (ConstValue ListSeparator?)* ']'
// ConstMap   <- '{' (ConstValue ':' ConstValue ListSeparator?)* '}'
func (p *parser) parseConstValue() *ConstValue {
	p.expect(token.FloatLiteral, token.IntLiteral, token.StringLiteral, token.Identifier, token.LBracket, token.LBrace)

	switch p.last.Tok {
	case token.FloatLiteral:
		d := p.last.AsFloat()
		return &ConstValue{
			Type: ConstType_ConstDouble,
			TypedValue: &ConstTypedValue{
				Double: &d,
			},
		}

	case token.IntLiteral:
		i := p.last.AsInt()
		return &ConstValue{
			Type: ConstType_ConstInt,
			TypedValue: &ConstTypedValue{
				Int: &i,
			},
		}

	case token.StringLiteral:
		s := p.last.Unquote()
		return &ConstValue{
			Type: ConstType_ConstLiteral,
			TypedValue: &ConstTypedValue{
				Literal: &s,
			},
		}

	case token.Identifier:
		id := p.last.AsString()
		return &ConstValue{
			Type: ConstType_ConstIdentifier,
			TypedValue: &ConstTypedValue{
				Identifier: &id,
			},
		}

	case token.LBracket:
		cvs := []*ConstValue{} // important: can't not be nil
		for p.next.Tok != token.RBracket {
			cv := p.parseConstValue()
			p.consumeListSeparator()
			cvs = append(cvs, cv)
		}
		p.expect(token.RBracket)
		return &ConstValue{
			Type: ConstType_ConstList,
			TypedValue: &ConstTypedValue{
				List: cvs,
			},
		}

	case token.LBrace:
		mcvs := []*MapConstValue{} // important: can't not be nil
		for p.next.Tok != token.RBrace {
			k := p.parseConstValue()
			p.expect(token.Colon)
			v := p.parseConstValue()
			p.consumeListSeparator()
			mcvs = append(mcvs, &MapConstValue{Key: k, Value: v})
		}
		p.expect(token.RBrace)
		return &ConstValue{
			Type: ConstType_ConstMap,
			TypedValue: &ConstTypedValue{
				Map: mcvs,
			},
		}

	default:
		panic("?")
	}
}

// Annotations <- '(' Annotation+ ')'
// Annotation  <- Identifier '=' Literal ListSeparator?
func (p *parser) parseAnnotations() (as Annotations) {
	p.expect(token.LParenthesis)
	for p.next.Tok != token.RParenthesis {
		p.expect(token.Identifier)
		key := p.last.AsString()
		p.expect(token.Equal)
		p.expect(token.StringLiteral)
		val := p.last.Unquote()
		as.Append(key, val)
		p.consumeListSeparator()
	}
	p.expect(token.RParenthesis)
	return
}

// ListSeparator  <- ',' / ';'
func (p *parser) consumeListSeparator() {
	if p.next.Tok == token.Comma || p.next.Tok == token.Semicolon {
		p.expect(token.Comma, token.Semicolon)
	}
}

func (p *parser) prefixComment() (res string) {
	line := p.lexer.Pos2Pos(p.next.Span.Beg).Line
	i := len(p.comspans) - 1
	for i >= 0 && p.comspans[i].End >= line-1 {
		// adjoining comments
		line = p.comspans[i].Beg
		i--
	}
	res = strings.Join(p.comments[i+1:], "\n")
	p.comments = p.comments[:i+1]
	p.comspans = p.comspans[:i+1]
	return
}

func (p *parser) syntaxError(t token.Tok, more []token.Tok) error {
	pos := p.lexer.Pos2Pos(p.next.Span.Beg)
	expected := append([]token.Tok{t}, more...)
	return fmt.Errorf("%q:%d:%d expect %+v, got %s",
		p.ast.Filename, pos.Line, pos.Offset, expected, &p.next)
}

func (p *parser) lexicalError() error {
	pos := p.lexer.Pos2Pos(p.next.Span.Beg)
	return fmt.Errorf("%q:%d:%d unknown token: %q",
		p.ast.Filename, pos.Line, pos.Offset, string(p.next.Data))
}
