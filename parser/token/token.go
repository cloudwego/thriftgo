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

package token

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Tok .
type Tok int

// LexicalError .
const LexicalError = Tok(-1)

// Toks .
const (
	EOF           Tok = iota
	BlockComment      // /\*(?:\*[^/]|[^*])*\*/
	LineComment       // //[^\n]*
	UnixComment       // #[^\n]*
	Whitespaces       //  \t\v\r
	NewLine           // \r?\n
	Bool              // bool
	Byte              // byte
	I8                // i8
	I16               // i16
	I32               // i32
	I64               // i64
	Double            // double
	String            // string
	Binary            // binary
	Const             // const
	Oneway            // oneway
	Typedef           // typedef
	Map               // map
	Set               // set
	List              // list
	Void              // void
	Throws            // throws
	Exception         // exception
	Extends           // extends
	Required          // required
	Optional          // optional
	Service           // service
	Struct            // struct
	Union             // union
	Enum              // enum
	Include           // include
	CppInclude        // cpp_include
	Namespace         // namespace
	Asterisk          // *
	LBracket          // [
	RBracket          // ]
	LBrace            // {
	RBrace            // }
	LParenthesis      // (
	RParenthesis      // )
	LChevron          // <
	RChevron          // >
	Equal             // =
	Comma             // ,
	Colon             // :
	Semicolon         // ;
	StringLiteral     // '(?:\\'|[^'])*'|"(?:\\"|[^"])*"
	Identifier        // [_a-zA-Z][_a-zA-Z0-9]*(?:\.[_a-zA-Z][_a-zA-Z0-9]*)*
	IntLiteral        // [-+]?(?:0x[0-9a-fA-F]+|0o[0-7]+|0|[1-9][0-9]*)
	FloatLiteral      //
	// 	`[-+]?` + // Sign
	// 		(`(?:` +
	// 			(`(?:0|[1-9][0-9]*)?` + // integer?
	// 				`\.` + // .
	// 				`[0-9]+` + // digits
	// 				`(?:[eE][0-9]+)?`) + // exponent?
	// 			`|` +
	// 			(`(?:0|[1-9][0-9]*)` + // integer
	// 				`(?:[eE][0-9]+)`) + // exponent
	// 			`)`),
)

func (t Tok) String() string {
	if str, ok := tokRepr[t]; ok {
		return str
	}
	return fmt.Sprintf("<Tok %d>", t)
}

// Token is the set of lexical tokens of the thrift IDL.
type Token struct {
	Tok
	Span
	Data []byte
}

func (t *Token) String() string {
	if len(t.Data) > 0 {
		return fmt.Sprintf("%s(%d): %q",
			tokNames[t.Tok], t.Tok, string(t.Data))
	}
	return fmt.Sprintf("%s(%d)", tokNames[t.Tok], t.Tok)
}

// AsString returns the data of the token as a string.
func (t *Token) AsString() string {
	return string(t.Data)
}

// AsInt returns the data of the token as a int64.
func (t *Token) AsInt() int64 {
	// TODO: check error
	i64, _ := strconv.ParseInt(string(t.Data), 0, 64)
	return i64
}

// AsFloat returns the data of the token as a float64.
func (t *Token) AsFloat() float64 {
	// TODO: check error
	f64, _ := strconv.ParseFloat(string(t.Data), 64)
	return f64
}

// Unquote interprets the token's data as a quote string and returns
// the string it quotes.
func (t *Token) Unquote() string {
	res := make([]byte, 0, len(t.Data)-2)
	for i := 1; i < len(t.Data)-1; i++ {
		if t.Data[i] == t.Data[0] && t.Data[i-1] == '\\' {
			res[len(res)-1] = t.Data[i]
		} else {
			res = append(res, t.Data[i])
		}
	}
	return string(res)
}

// Span represents a segment in a stream.
type Span struct {
	Beg int
	End int
}

// Pos represents a position in the source file.
type Pos struct {
	Line   int // start from 1
	Offset int // start from 1
}

// Tokenizer is a lexical analyzer for thrift IDL.
type Tokenizer struct {
	src *bufio.Reader
	Span
	buf []byte
	nlp []int // newline positions
}

// NewTokenizer returns a Tokenizer over the given source.
func NewTokenizer(src io.Reader) *Tokenizer {
	p := new(Tokenizer)
	p.src = bufio.NewReader(src)
	p.nlp = append(p.nlp, -1)
	return p
}

// Pos2Pos converts an index in the byte stream to a Pos.
func (p *Tokenizer) Pos2Pos(pos int) Pos {
	for ln, idx := range p.nlp {
		if idx >= pos {
			return Pos{
				Line:   ln,
				Offset: pos - idx,
			}
		}
	}
	return Pos{
		Line:   len(p.nlp),
		Offset: pos - p.nlp[len(p.nlp)-1],
	}
}

// LineSpan returns the first and last line that the given span exists.
func (p *Tokenizer) LineSpan(s Span) Span {
	return Span{
		Beg: p.Pos2Pos(s.Beg).Line,
		End: p.Pos2Pos(s.End).Line,
	}
}

func (p *Tokenizer) reset() {
	p.Span.Beg = p.End
	p.buf = p.buf[0:0]
}

func (p *Tokenizer) data() []byte {
	res := make([]byte, len(p.buf))
	copy(res, p.buf)
	return res
}

func (p *Tokenizer) unread(c byte) {
	if err := p.src.UnreadByte(); err != nil {
		panic(err)
	}
	p.Span.End--
	p.buf = p.buf[:len(p.buf)-1]
	if c == '\n' {
		p.nlp = p.nlp[:len(p.nlp)-1]
	}
}

func (p *Tokenizer) nextc() (c byte, eof bool) {
	b, err := p.src.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			eof = true
			p.nlp = append(p.nlp, p.Span.End)
			return
		}
		panic(fmt.Errorf("ReadByte: %w", err))
	}
	if b == '\n' {
		p.nlp = append(p.nlp, p.Span.End)
	}
	c = b
	p.buf = append(p.buf, c)
	p.Span.End++
	return
}

func (p *Tokenizer) lexicalError() Token {
	return p.token(LexicalError)
}

func (p *Tokenizer) token(t Tok) Token {
	return Token{Tok: t, Span: p.Span, Data: p.data()}
}

// Next returns the next token from the source stream.
func (p *Tokenizer) Next() Token {
	p.reset()
	c, eof := p.nextc()
	if eof {
		return Token{
			Tok:  EOF,
			Span: p.Span,
		}
	}
	switch c {
	case '/':
		c2, eof := p.nextc()
		switch c2 {
		case '*':
			gotStar := false
			for {
				if c, eof = p.nextc(); eof {
					break
				}
				if c == '/' && gotStar {
					return p.token(BlockComment)
				}
				gotStar = c == '*'
			}
			return p.lexicalError()
		case '/':
			for ; !eof && c != '\n'; c, eof = p.nextc() {
			}
			if c == '\n' {
				p.unread(c)
			}
			return p.token(LineComment)
		default:
			if !eof {
				p.unread(c2)
			}
			return p.lexicalError()
		}
	case '#':
		for ; !eof && c != '\n'; c, eof = p.nextc() {
		}
		if c == '\n' {
			p.unread(c)
		}
		return p.token(UnixComment)
	case ' ', '\t', '\v', '\r':
		for !eof && (c == ' ' || c == '\t' || c == '\v' || c == '\r') {
			c, eof = p.nextc()
		}
		if !eof {
			p.unread(c)
		}
		return p.token(Whitespaces)
	case '\n':
		return p.token(NewLine)
	case '*':
		return p.token(Asterisk)
	case '[':
		return p.token(LBracket)
	case ']':
		return p.token(RBracket)
	case '{':
		return p.token(LBrace)
	case '}':
		return p.token(RBrace)
	case '(':
		return p.token(LParenthesis)
	case ')':
		return p.token(RParenthesis)
	case '<':
		return p.token(LChevron)
	case '>':
		return p.token(RChevron)
	case '=':
		return p.token(Equal)
	case ',':
		return p.token(Comma)
	case ':':
		return p.token(Colon)
	case ';':
		return p.token(Semicolon)
	case '\'', '"':
		d := c
		c, eof = p.nextc()
		for !eof && c != d {
			if c == '\\' {
				if _, eof = p.nextc(); eof {
					break
				}
			}
			c, eof = p.nextc()
		}
		if eof {
			return p.lexicalError()
		}
		return p.token(StringLiteral)
	default:
		if maybeKeywordOfID(c) {
			return p.tryKeywordOrID(c)
		}
		if maybeNumber(c) {
			return p.tryNumber(c)
		}
		return p.lexicalError()
	}
}

func (p *Tokenizer) tryKeywordOrID(last byte) Token {
	// [_a-zA-Z][_a-zA-Z0-9]*(?:\.[_a-zA-Z][_a-zA-Z0-9]*)*
	for {
		c, eof := p.nextc()
		if eof {
			break
		}
		if c == '_' || '0' <= c && c <= '9' ||
			'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
			last = c
			continue
		}
		if c == '.' {
			if last == '.' {
				break
			}
			last = c
			continue
		}
		p.unread(c)
		break
	}
	if last == '.' {
		return p.lexicalError()
	}
	wd := string(p.buf)
	if tok, ok := keywords[wd]; ok {
		return Token{Tok: tok, Span: p.Span, Data: p.data()}
	}
	return p.token(Identifier)
}

func (p *Tokenizer) tryNumber(c byte) Token {
	// [-+]? // Sign
	//      (0x[0-9a-fA-F]+|0o[0-7]+|0|[1-9][0-9]*) // integer
	// 		(0|[1-9][0-9]*)?\.[0-9]*([eE][0-9]+)?   // float
	// 		(0|[1-9][0-9]*)([eE][0-9]+)             // float
	var eof, isFloat bool
	if c == '-' || c == '+' {
		if c, eof = p.nextc(); eof {
			return p.lexicalError()
		}
	}
	if c == '0' {
		if p.pick(is('x')) {
			// 0[xX][0-9a-fA-F]+
			if p.pickAll(isHexical) < 1 {
				return p.lexicalError()
			}
			goto ReturnInteger
		} else if p.pick(is('o')) {
			// 0o[0-7]+
			if p.pickAll(isOctal) < 1 {
				return p.lexicalError()
			}
			goto ReturnInteger
		} else {
			// 0
			goto TryGetFloatPart
		}
	} else if '1' <= c && c <= '9' {
		p.pickAll(isDigit) // [1-9][0-9]*
		goto TryGetFloatPart
	} else if c == '.' {
		p.pickAll(isDigit) // \.[0-9]*
		if p.pick(is('e', 'E')) {
			// \.[0-9]*[eE][0-9]+
			if p.pickAll(isDigit) < 1 {
				return p.lexicalError()
			}
		}
		goto ReturnFloat
	} else {
		return p.lexicalError()
	}

TryGetFloatPart:
	if p.pick(is('.')) {
		p.pickAll(isDigit) // ~\.[0-9]*
		isFloat = true
	}
	if p.pick(is('e', 'E')) {
		// ~\.[0-9]*[eE][0-9]+
		// ~[eE][0-9]+
		if p.pickAll(isDigit) < 1 {
			return p.lexicalError()
		}
		isFloat = true
	}
	if isFloat {
		goto ReturnFloat
	} else {
		goto ReturnInteger
	}
ReturnInteger:
	if p.pickAll(isDigit) > 0 {
		return p.lexicalError() // invalid integer literal: "000123"
	}
	return p.token(IntLiteral)
ReturnFloat:
	return p.token(FloatLiteral)
}

func (p *Tokenizer) pick(accept func(byte) bool) bool {
	if c, eof := p.nextc(); !eof {
		if accept(c) {
			return true
		}
		p.unread(c)
	}
	return false
}

func (p *Tokenizer) pickAll(accept func(byte) bool) (cnt int) {
	for p.pick(accept) {
		cnt++
	}
	return
}

func maybeKeywordOfID(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_'
}

func maybeNumber(c byte) bool {
	return c == '-' || c == '+' || '0' <= c && c <= '9' || c == '.'
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isOctal(c byte) bool {
	return '0' <= c && c <= '7'
}

func isHexical(c byte) bool {
	return '0' <= c && c <= '9' ||
		'a' <= c && c <= 'f' ||
		'A' <= c && c <= 'F'
}

func is(cs ...byte) func(byte) bool {
	return func(x byte) bool {
		for _, c := range cs {
			if c == x {
				return true
			}
		}
		return false
	}
}

var keywords = map[string]Tok{
	"bool":        Bool,
	"byte":        Byte,
	"i8":          I8,
	"i16":         I16,
	"i32":         I32,
	"i64":         I64,
	"double":      Double,
	"string":      String,
	"binary":      Binary,
	"const":       Const,
	"oneway":      Oneway,
	"typedef":     Typedef,
	"map":         Map,
	"set":         Set,
	"list":        List,
	"void":        Void,
	"throws":      Throws,
	"exception":   Exception,
	"extends":     Extends,
	"required":    Required,
	"optional":    Optional,
	"service":     Service,
	"struct":      Struct,
	"union":       Union,
	"enum":        Enum,
	"include":     Include,
	"cpp_include": CppInclude,
	"namespace":   Namespace,
}
