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

package parser

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDocument
	ruleHeader
	ruleInclude
	ruleCppInclude
	ruleNamespace
	ruleNamespaceScope
	ruleDefinition
	ruleConst
	ruleTypedef
	ruleEnum
	ruleStruct
	ruleUnion
	ruleService
	ruleException
	ruleField
	ruleFieldId
	ruleFieldReq
	ruleFunction
	ruleFunctionType
	ruleThrows
	ruleFieldType
	ruleBaseType
	ruleContainerType
	ruleMapType
	ruleSetType
	ruleListType
	ruleCppType
	ruleConstValue
	ruleIntConstant
	ruleDoubleConstant
	ruleExponent
	ruleAnnotations
	ruleConstList
	ruleConstMap
	ruleEscapeLiteralChar
	ruleLiteral
	ruleIdentifier
	ruleListSeparator
	ruleLetter
	ruleLetterOrDigit
	ruleDigit
	ruleSkip
	ruleSpace
	ruleLongComment
	ruleLineComment
	ruleUnixComment
	ruleBOOL
	ruleBYTE
	ruleI8
	ruleI16
	ruleI32
	ruleI64
	ruleDOUBLE
	ruleSTRING
	ruleBINARY
	ruleCONST
	ruleONEWAY
	ruleTYPEDEF
	ruleMAP
	ruleSET
	ruleLIST
	ruleVOID
	ruleTHROWS
	ruleEXCEPTION
	ruleEXTENDS
	ruleSERVICE
	ruleSTRUCT
	ruleUNION
	ruleENUM
	ruleINCLUDE
	ruleCPPINCLUDE
	ruleNAMESPACE
	ruleCPPTYPE
	ruleLBRK
	ruleRBRK
	ruleLWING
	ruleRWING
	ruleEQUAL
	ruleLPOINT
	ruleRPOINT
	ruleCOMMA
	ruleLPAR
	ruleRPAR
	ruleCOLON
	rulePegText

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"Document",
	"Header",
	"Include",
	"CppInclude",
	"Namespace",
	"NamespaceScope",
	"Definition",
	"Const",
	"Typedef",
	"Enum",
	"Struct",
	"Union",
	"Service",
	"Exception",
	"Field",
	"FieldId",
	"FieldReq",
	"Function",
	"FunctionType",
	"Throws",
	"FieldType",
	"BaseType",
	"ContainerType",
	"MapType",
	"SetType",
	"ListType",
	"CppType",
	"ConstValue",
	"IntConstant",
	"DoubleConstant",
	"Exponent",
	"Annotations",
	"ConstList",
	"ConstMap",
	"EscapeLiteralChar",
	"Literal",
	"Identifier",
	"ListSeparator",
	"Letter",
	"LetterOrDigit",
	"Digit",
	"Skip",
	"Space",
	"LongComment",
	"LineComment",
	"UnixComment",
	"BOOL",
	"BYTE",
	"I8",
	"I16",
	"I32",
	"I64",
	"DOUBLE",
	"STRING",
	"BINARY",
	"CONST",
	"ONEWAY",
	"TYPEDEF",
	"MAP",
	"SET",
	"LIST",
	"VOID",
	"THROWS",
	"EXCEPTION",
	"EXTENDS",
	"SERVICE",
	"STRUCT",
	"UNION",
	"ENUM",
	"INCLUDE",
	"CPPINCLUDE",
	"NAMESPACE",
	"CPPTYPE",
	"LBRK",
	"RBRK",
	"LWING",
	"RWING",
	"EQUAL",
	"LPOINT",
	"RPOINT",
	"COMMA",
	"LPAR",
	"RPAR",
	"COLON",
	"PegText",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type ThriftIDL struct {
	Buffer string
	buffer []rune
	rules  [86]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *ThriftIDL
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *ThriftIDL) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *ThriftIDL) Highlighter() {
	p.PrintSyntax()
}

func (p *ThriftIDL) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Document <- <(Skip Header* Definition* !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleSkip]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !_rules[ruleHeader]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
			l4:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !_rules[ruleDefinition]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !matchDot() {
						goto l6
					}
					goto l0
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				depth--
				add(ruleDocument, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Header <- <(Include / CppInclude / Namespace)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				{
					position9, tokenIndex9, depth9 := position, tokenIndex, depth
					if !_rules[ruleInclude]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !_rules[ruleCppInclude]() {
						goto l11
					}
					goto l9
				l11:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !_rules[ruleNamespace]() {
						goto l7
					}
				}
			l9:
				depth--
				add(ruleHeader, position8)
			}
			return true
		l7:
			position, tokenIndex, depth = position7, tokenIndex7, depth7
			return false
		},
		/* 2 Include <- <(INCLUDE Literal)> */
		func() bool {
			position12, tokenIndex12, depth12 := position, tokenIndex, depth
			{
				position13 := position
				depth++
				if !_rules[ruleINCLUDE]() {
					goto l12
				}
				if !_rules[ruleLiteral]() {
					goto l12
				}
				depth--
				add(ruleInclude, position13)
			}
			return true
		l12:
			position, tokenIndex, depth = position12, tokenIndex12, depth12
			return false
		},
		/* 3 CppInclude <- <(CPPINCLUDE Literal)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if !_rules[ruleCPPINCLUDE]() {
					goto l14
				}
				if !_rules[ruleLiteral]() {
					goto l14
				}
				depth--
				add(ruleCppInclude, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 4 Namespace <- <(NAMESPACE NamespaceScope Identifier Annotations?)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if !_rules[ruleNAMESPACE]() {
					goto l16
				}
				if !_rules[ruleNamespaceScope]() {
					goto l16
				}
				if !_rules[ruleIdentifier]() {
					goto l16
				}
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l18
					}
					goto l19
				l18:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
				}
			l19:
				depth--
				add(ruleNamespace, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 5 NamespaceScope <- <((<'*'> Skip) / Identifier)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					{
						position24 := position
						depth++
						if buffer[position] != rune('*') {
							goto l23
						}
						position++
						depth--
						add(rulePegText, position24)
					}
					if !_rules[ruleSkip]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruleIdentifier]() {
						goto l20
					}
				}
			l22:
				depth--
				add(ruleNamespaceScope, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 6 Definition <- <((Const / Typedef / Enum / Struct / Union / Service / Exception) Annotations?)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[ruleConst]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleTypedef]() {
						goto l29
					}
					goto l27
				l29:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleEnum]() {
						goto l30
					}
					goto l27
				l30:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleStruct]() {
						goto l31
					}
					goto l27
				l31:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleUnion]() {
						goto l32
					}
					goto l27
				l32:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleService]() {
						goto l33
					}
					goto l27
				l33:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleException]() {
						goto l25
					}
				}
			l27:
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l34
					}
					goto l35
				l34:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
				}
			l35:
				depth--
				add(ruleDefinition, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 7 Const <- <(CONST FieldType Identifier EQUAL ConstValue ListSeparator?)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleCONST]() {
					goto l36
				}
				if !_rules[ruleFieldType]() {
					goto l36
				}
				if !_rules[ruleIdentifier]() {
					goto l36
				}
				if !_rules[ruleEQUAL]() {
					goto l36
				}
				if !_rules[ruleConstValue]() {
					goto l36
				}
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !_rules[ruleListSeparator]() {
						goto l38
					}
					goto l39
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
			l39:
				depth--
				add(ruleConst, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 8 Typedef <- <(TYPEDEF (BaseType / ContainerType / Identifier) Identifier Annotations?)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleTYPEDEF]() {
					goto l40
				}
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleBaseType]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleContainerType]() {
						goto l44
					}
					goto l42
				l44:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleIdentifier]() {
						goto l40
					}
				}
			l42:
				if !_rules[ruleIdentifier]() {
					goto l40
				}
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l45
					}
					goto l46
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
			l46:
				depth--
				add(ruleTypedef, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 9 Enum <- <(ENUM Identifier LWING (Identifier (EQUAL IntConstant)? Annotations? ListSeparator?)* RWING)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[ruleENUM]() {
					goto l47
				}
				if !_rules[ruleIdentifier]() {
					goto l47
				}
				if !_rules[ruleLWING]() {
					goto l47
				}
			l49:
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !_rules[ruleIdentifier]() {
						goto l50
					}
					{
						position51, tokenIndex51, depth51 := position, tokenIndex, depth
						if !_rules[ruleEQUAL]() {
							goto l51
						}
						if !_rules[ruleIntConstant]() {
							goto l51
						}
						goto l52
					l51:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
					}
				l52:
					{
						position53, tokenIndex53, depth53 := position, tokenIndex, depth
						if !_rules[ruleAnnotations]() {
							goto l53
						}
						goto l54
					l53:
						position, tokenIndex, depth = position53, tokenIndex53, depth53
					}
				l54:
					{
						position55, tokenIndex55, depth55 := position, tokenIndex, depth
						if !_rules[ruleListSeparator]() {
							goto l55
						}
						goto l56
					l55:
						position, tokenIndex, depth = position55, tokenIndex55, depth55
					}
				l56:
					goto l49
				l50:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
				}
				if !_rules[ruleRWING]() {
					goto l47
				}
				depth--
				add(ruleEnum, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 10 Struct <- <(STRUCT Identifier LWING Field* RWING)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if !_rules[ruleSTRUCT]() {
					goto l57
				}
				if !_rules[ruleIdentifier]() {
					goto l57
				}
				if !_rules[ruleLWING]() {
					goto l57
				}
			l59:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if !_rules[ruleField]() {
						goto l60
					}
					goto l59
				l60:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
				}
				if !_rules[ruleRWING]() {
					goto l57
				}
				depth--
				add(ruleStruct, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 11 Union <- <(UNION Identifier LWING Field* RWING)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				if !_rules[ruleUNION]() {
					goto l61
				}
				if !_rules[ruleIdentifier]() {
					goto l61
				}
				if !_rules[ruleLWING]() {
					goto l61
				}
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !_rules[ruleField]() {
						goto l64
					}
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				if !_rules[ruleRWING]() {
					goto l61
				}
				depth--
				add(ruleUnion, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 12 Service <- <(SERVICE Identifier (EXTENDS Identifier)? LWING Function* RWING)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if !_rules[ruleSERVICE]() {
					goto l65
				}
				if !_rules[ruleIdentifier]() {
					goto l65
				}
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if !_rules[ruleEXTENDS]() {
						goto l67
					}
					if !_rules[ruleIdentifier]() {
						goto l67
					}
					goto l68
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
			l68:
				if !_rules[ruleLWING]() {
					goto l65
				}
			l69:
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
					if !_rules[ruleFunction]() {
						goto l70
					}
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				if !_rules[ruleRWING]() {
					goto l65
				}
				depth--
				add(ruleService, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 13 Exception <- <(EXCEPTION Identifier LWING Field* RWING)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if !_rules[ruleEXCEPTION]() {
					goto l71
				}
				if !_rules[ruleIdentifier]() {
					goto l71
				}
				if !_rules[ruleLWING]() {
					goto l71
				}
			l73:
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if !_rules[ruleField]() {
						goto l74
					}
					goto l73
				l74:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
				}
				if !_rules[ruleRWING]() {
					goto l71
				}
				depth--
				add(ruleException, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 14 Field <- <(FieldId? FieldReq? FieldType Identifier (EQUAL ConstValue)? Annotations? ListSeparator?)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if !_rules[ruleFieldId]() {
						goto l77
					}
					goto l78
				l77:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
				}
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !_rules[ruleFieldReq]() {
						goto l79
					}
					goto l80
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
			l80:
				if !_rules[ruleFieldType]() {
					goto l75
				}
				if !_rules[ruleIdentifier]() {
					goto l75
				}
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if !_rules[ruleEQUAL]() {
						goto l81
					}
					if !_rules[ruleConstValue]() {
						goto l81
					}
					goto l82
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
			l82:
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l83
					}
					goto l84
				l83:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
				}
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if !_rules[ruleListSeparator]() {
						goto l85
					}
					goto l86
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
			l86:
				depth--
				add(ruleField, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 15 FieldId <- <(IntConstant ':' Skip)> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if !_rules[ruleIntConstant]() {
					goto l87
				}
				if buffer[position] != rune(':') {
					goto l87
				}
				position++
				if !_rules[ruleSkip]() {
					goto l87
				}
				depth--
				add(ruleFieldId, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 16 FieldReq <- <(<(('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd') / ('o' 'p' 't' 'i' 'o' 'n' 'a' 'l'))> Skip)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				{
					position91 := position
					depth++
					{
						position92, tokenIndex92, depth92 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l93
						}
						position++
						if buffer[position] != rune('e') {
							goto l93
						}
						position++
						if buffer[position] != rune('q') {
							goto l93
						}
						position++
						if buffer[position] != rune('u') {
							goto l93
						}
						position++
						if buffer[position] != rune('i') {
							goto l93
						}
						position++
						if buffer[position] != rune('r') {
							goto l93
						}
						position++
						if buffer[position] != rune('e') {
							goto l93
						}
						position++
						if buffer[position] != rune('d') {
							goto l93
						}
						position++
						goto l92
					l93:
						position, tokenIndex, depth = position92, tokenIndex92, depth92
						if buffer[position] != rune('o') {
							goto l89
						}
						position++
						if buffer[position] != rune('p') {
							goto l89
						}
						position++
						if buffer[position] != rune('t') {
							goto l89
						}
						position++
						if buffer[position] != rune('i') {
							goto l89
						}
						position++
						if buffer[position] != rune('o') {
							goto l89
						}
						position++
						if buffer[position] != rune('n') {
							goto l89
						}
						position++
						if buffer[position] != rune('a') {
							goto l89
						}
						position++
						if buffer[position] != rune('l') {
							goto l89
						}
						position++
					}
				l92:
					depth--
					add(rulePegText, position91)
				}
				if !_rules[ruleSkip]() {
					goto l89
				}
				depth--
				add(ruleFieldReq, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 17 Function <- <(ONEWAY? FunctionType Identifier LPAR Field* RPAR Throws? Annotations? ListSeparator?)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleONEWAY]() {
						goto l96
					}
					goto l97
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
			l97:
				if !_rules[ruleFunctionType]() {
					goto l94
				}
				if !_rules[ruleIdentifier]() {
					goto l94
				}
				if !_rules[ruleLPAR]() {
					goto l94
				}
			l98:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if !_rules[ruleField]() {
						goto l99
					}
					goto l98
				l99:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
				}
				if !_rules[ruleRPAR]() {
					goto l94
				}
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if !_rules[ruleThrows]() {
						goto l100
					}
					goto l101
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
			l101:
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l102
					}
					goto l103
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
			l103:
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !_rules[ruleListSeparator]() {
						goto l104
					}
					goto l105
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
			l105:
				depth--
				add(ruleFunction, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 18 FunctionType <- <(VOID / FieldType)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleVOID]() {
						goto l109
					}
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleFieldType]() {
						goto l106
					}
				}
			l108:
				depth--
				add(ruleFunctionType, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 19 Throws <- <(THROWS LPAR Field* RPAR)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if !_rules[ruleTHROWS]() {
					goto l110
				}
				if !_rules[ruleLPAR]() {
					goto l110
				}
			l112:
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if !_rules[ruleField]() {
						goto l113
					}
					goto l112
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
				if !_rules[ruleRPAR]() {
					goto l110
				}
				depth--
				add(ruleThrows, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 20 FieldType <- <((ContainerType / BaseType / Identifier) Annotations?)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleContainerType]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleBaseType]() {
						goto l118
					}
					goto l116
				l118:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleIdentifier]() {
						goto l114
					}
				}
			l116:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[ruleAnnotations]() {
						goto l119
					}
					goto l120
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
			l120:
				depth--
				add(ruleFieldType, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 21 BaseType <- <(BOOL / BYTE / I8 / I16 / I32 / I64 / DOUBLE / STRING / (BINARY Skip))> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if !_rules[ruleBOOL]() {
						goto l124
					}
					goto l123
				l124:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleBYTE]() {
						goto l125
					}
					goto l123
				l125:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleI8]() {
						goto l126
					}
					goto l123
				l126:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleI16]() {
						goto l127
					}
					goto l123
				l127:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleI32]() {
						goto l128
					}
					goto l123
				l128:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleI64]() {
						goto l129
					}
					goto l123
				l129:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleDOUBLE]() {
						goto l130
					}
					goto l123
				l130:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleSTRING]() {
						goto l131
					}
					goto l123
				l131:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleBINARY]() {
						goto l121
					}
					if !_rules[ruleSkip]() {
						goto l121
					}
				}
			l123:
				depth--
				add(ruleBaseType, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 22 ContainerType <- <(MapType / SetType / ListType)> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if !_rules[ruleMapType]() {
						goto l135
					}
					goto l134
				l135:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					if !_rules[ruleSetType]() {
						goto l136
					}
					goto l134
				l136:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					if !_rules[ruleListType]() {
						goto l132
					}
				}
			l134:
				depth--
				add(ruleContainerType, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 23 MapType <- <(MAP CppType? LPOINT FieldType COMMA FieldType RPOINT)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if !_rules[ruleMAP]() {
					goto l137
				}
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if !_rules[ruleCppType]() {
						goto l139
					}
					goto l140
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
			l140:
				if !_rules[ruleLPOINT]() {
					goto l137
				}
				if !_rules[ruleFieldType]() {
					goto l137
				}
				if !_rules[ruleCOMMA]() {
					goto l137
				}
				if !_rules[ruleFieldType]() {
					goto l137
				}
				if !_rules[ruleRPOINT]() {
					goto l137
				}
				depth--
				add(ruleMapType, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 24 SetType <- <(SET CppType? LPOINT FieldType RPOINT)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if !_rules[ruleSET]() {
					goto l141
				}
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if !_rules[ruleCppType]() {
						goto l143
					}
					goto l144
				l143:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
				}
			l144:
				if !_rules[ruleLPOINT]() {
					goto l141
				}
				if !_rules[ruleFieldType]() {
					goto l141
				}
				if !_rules[ruleRPOINT]() {
					goto l141
				}
				depth--
				add(ruleSetType, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 25 ListType <- <(LIST LPOINT FieldType RPOINT CppType?)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				if !_rules[ruleLIST]() {
					goto l145
				}
				if !_rules[ruleLPOINT]() {
					goto l145
				}
				if !_rules[ruleFieldType]() {
					goto l145
				}
				if !_rules[ruleRPOINT]() {
					goto l145
				}
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[ruleCppType]() {
						goto l147
					}
					goto l148
				l147:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
				}
			l148:
				depth--
				add(ruleListType, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 26 CppType <- <(CPPTYPE Literal Skip)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				if !_rules[ruleCPPTYPE]() {
					goto l149
				}
				if !_rules[ruleLiteral]() {
					goto l149
				}
				if !_rules[ruleSkip]() {
					goto l149
				}
				depth--
				add(ruleCppType, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 27 ConstValue <- <(DoubleConstant / IntConstant / Literal / Identifier / ConstList / ConstMap)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleDoubleConstant]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleIntConstant]() {
						goto l155
					}
					goto l153
				l155:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleLiteral]() {
						goto l156
					}
					goto l153
				l156:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleIdentifier]() {
						goto l157
					}
					goto l153
				l157:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleConstList]() {
						goto l158
					}
					goto l153
				l158:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleConstMap]() {
						goto l151
					}
				}
			l153:
				depth--
				add(ruleConstValue, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 28 IntConstant <- <(<(('0' 'x' ([0-9] / [A-Z] / [a-z])+) / ('0' 'o' Digit+) / (('+' / '-')? Digit+))> Skip)> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				{
					position161 := position
					depth++
					{
						position162, tokenIndex162, depth162 := position, tokenIndex, depth
						if buffer[position] != rune('0') {
							goto l163
						}
						position++
						if buffer[position] != rune('x') {
							goto l163
						}
						position++
						{
							position166, tokenIndex166, depth166 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l167
							}
							position++
							goto l166
						l167:
							position, tokenIndex, depth = position166, tokenIndex166, depth166
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l168
							}
							position++
							goto l166
						l168:
							position, tokenIndex, depth = position166, tokenIndex166, depth166
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l163
							}
							position++
						}
					l166:
					l164:
						{
							position165, tokenIndex165, depth165 := position, tokenIndex, depth
							{
								position169, tokenIndex169, depth169 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l170
								}
								position++
								goto l169
							l170:
								position, tokenIndex, depth = position169, tokenIndex169, depth169
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l171
								}
								position++
								goto l169
							l171:
								position, tokenIndex, depth = position169, tokenIndex169, depth169
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l165
								}
								position++
							}
						l169:
							goto l164
						l165:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
						}
						goto l162
					l163:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						if buffer[position] != rune('0') {
							goto l172
						}
						position++
						if buffer[position] != rune('o') {
							goto l172
						}
						position++
						if !_rules[ruleDigit]() {
							goto l172
						}
					l173:
						{
							position174, tokenIndex174, depth174 := position, tokenIndex, depth
							if !_rules[ruleDigit]() {
								goto l174
							}
							goto l173
						l174:
							position, tokenIndex, depth = position174, tokenIndex174, depth174
						}
						goto l162
					l172:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						{
							position175, tokenIndex175, depth175 := position, tokenIndex, depth
							{
								position177, tokenIndex177, depth177 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l178
								}
								position++
								goto l177
							l178:
								position, tokenIndex, depth = position177, tokenIndex177, depth177
								if buffer[position] != rune('-') {
									goto l175
								}
								position++
							}
						l177:
							goto l176
						l175:
							position, tokenIndex, depth = position175, tokenIndex175, depth175
						}
					l176:
						if !_rules[ruleDigit]() {
							goto l159
						}
					l179:
						{
							position180, tokenIndex180, depth180 := position, tokenIndex, depth
							if !_rules[ruleDigit]() {
								goto l180
							}
							goto l179
						l180:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
						}
					}
				l162:
					depth--
					add(rulePegText, position161)
				}
				if !_rules[ruleSkip]() {
					goto l159
				}
				depth--
				add(ruleIntConstant, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 29 DoubleConstant <- <(<(('+' / '-')? ((Digit* '.' Digit+ Exponent?) / (Digit+ Exponent)))> Skip)> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				{
					position183 := position
					depth++
					{
						position184, tokenIndex184, depth184 := position, tokenIndex, depth
						{
							position186, tokenIndex186, depth186 := position, tokenIndex, depth
							if buffer[position] != rune('+') {
								goto l187
							}
							position++
							goto l186
						l187:
							position, tokenIndex, depth = position186, tokenIndex186, depth186
							if buffer[position] != rune('-') {
								goto l184
							}
							position++
						}
					l186:
						goto l185
					l184:
						position, tokenIndex, depth = position184, tokenIndex184, depth184
					}
				l185:
					{
						position188, tokenIndex188, depth188 := position, tokenIndex, depth
					l190:
						{
							position191, tokenIndex191, depth191 := position, tokenIndex, depth
							if !_rules[ruleDigit]() {
								goto l191
							}
							goto l190
						l191:
							position, tokenIndex, depth = position191, tokenIndex191, depth191
						}
						if buffer[position] != rune('.') {
							goto l189
						}
						position++
						if !_rules[ruleDigit]() {
							goto l189
						}
					l192:
						{
							position193, tokenIndex193, depth193 := position, tokenIndex, depth
							if !_rules[ruleDigit]() {
								goto l193
							}
							goto l192
						l193:
							position, tokenIndex, depth = position193, tokenIndex193, depth193
						}
						{
							position194, tokenIndex194, depth194 := position, tokenIndex, depth
							if !_rules[ruleExponent]() {
								goto l194
							}
							goto l195
						l194:
							position, tokenIndex, depth = position194, tokenIndex194, depth194
						}
					l195:
						goto l188
					l189:
						position, tokenIndex, depth = position188, tokenIndex188, depth188
						if !_rules[ruleDigit]() {
							goto l181
						}
					l196:
						{
							position197, tokenIndex197, depth197 := position, tokenIndex, depth
							if !_rules[ruleDigit]() {
								goto l197
							}
							goto l196
						l197:
							position, tokenIndex, depth = position197, tokenIndex197, depth197
						}
						if !_rules[ruleExponent]() {
							goto l181
						}
					}
				l188:
					depth--
					add(rulePegText, position183)
				}
				if !_rules[ruleSkip]() {
					goto l181
				}
				depth--
				add(ruleDoubleConstant, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 30 Exponent <- <(('e' / 'E') IntConstant)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l201
					}
					position++
					goto l200
				l201:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
					if buffer[position] != rune('E') {
						goto l198
					}
					position++
				}
			l200:
				if !_rules[ruleIntConstant]() {
					goto l198
				}
				depth--
				add(ruleExponent, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 31 Annotations <- <(LPAR (Identifier EQUAL Literal ListSeparator?)+ RPAR)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if !_rules[ruleLPAR]() {
					goto l202
				}
				if !_rules[ruleIdentifier]() {
					goto l202
				}
				if !_rules[ruleEQUAL]() {
					goto l202
				}
				if !_rules[ruleLiteral]() {
					goto l202
				}
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if !_rules[ruleListSeparator]() {
						goto l206
					}
					goto l207
				l206:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
				}
			l207:
			l204:
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if !_rules[ruleIdentifier]() {
						goto l205
					}
					if !_rules[ruleEQUAL]() {
						goto l205
					}
					if !_rules[ruleLiteral]() {
						goto l205
					}
					{
						position208, tokenIndex208, depth208 := position, tokenIndex, depth
						if !_rules[ruleListSeparator]() {
							goto l208
						}
						goto l209
					l208:
						position, tokenIndex, depth = position208, tokenIndex208, depth208
					}
				l209:
					goto l204
				l205:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
				}
				if !_rules[ruleRPAR]() {
					goto l202
				}
				depth--
				add(ruleAnnotations, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 32 ConstList <- <(LBRK (ConstValue ListSeparator?)* RBRK)> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if !_rules[ruleLBRK]() {
					goto l210
				}
			l212:
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if !_rules[ruleConstValue]() {
						goto l213
					}
					{
						position214, tokenIndex214, depth214 := position, tokenIndex, depth
						if !_rules[ruleListSeparator]() {
							goto l214
						}
						goto l215
					l214:
						position, tokenIndex, depth = position214, tokenIndex214, depth214
					}
				l215:
					goto l212
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
				if !_rules[ruleRBRK]() {
					goto l210
				}
				depth--
				add(ruleConstList, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 33 ConstMap <- <(LWING (ConstValue COLON ConstValue ListSeparator?)* RWING)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				if !_rules[ruleLWING]() {
					goto l216
				}
			l218:
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[ruleConstValue]() {
						goto l219
					}
					if !_rules[ruleCOLON]() {
						goto l219
					}
					if !_rules[ruleConstValue]() {
						goto l219
					}
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						if !_rules[ruleListSeparator]() {
							goto l220
						}
						goto l221
					l220:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
					}
				l221:
					goto l218
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				if !_rules[ruleRWING]() {
					goto l216
				}
				depth--
				add(ruleConstMap, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 34 EscapeLiteralChar <- <('\\' ('"' / '\''))> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l222
				}
				position++
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('"') {
						goto l225
					}
					position++
					goto l224
				l225:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
					if buffer[position] != rune('\'') {
						goto l222
					}
					position++
				}
			l224:
				depth--
				add(ruleEscapeLiteralChar, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 35 Literal <- <(('"' <(EscapeLiteralChar / (!'"' .))*> '"' Skip) / ('\'' <(EscapeLiteralChar / (!'\'' .))*> '\'' Skip))> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if buffer[position] != rune('"') {
						goto l229
					}
					position++
					{
						position230 := position
						depth++
					l231:
						{
							position232, tokenIndex232, depth232 := position, tokenIndex, depth
							{
								position233, tokenIndex233, depth233 := position, tokenIndex, depth
								if !_rules[ruleEscapeLiteralChar]() {
									goto l234
								}
								goto l233
							l234:
								position, tokenIndex, depth = position233, tokenIndex233, depth233
								{
									position235, tokenIndex235, depth235 := position, tokenIndex, depth
									if buffer[position] != rune('"') {
										goto l235
									}
									position++
									goto l232
								l235:
									position, tokenIndex, depth = position235, tokenIndex235, depth235
								}
								if !matchDot() {
									goto l232
								}
							}
						l233:
							goto l231
						l232:
							position, tokenIndex, depth = position232, tokenIndex232, depth232
						}
						depth--
						add(rulePegText, position230)
					}
					if buffer[position] != rune('"') {
						goto l229
					}
					position++
					if !_rules[ruleSkip]() {
						goto l229
					}
					goto l228
				l229:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
					if buffer[position] != rune('\'') {
						goto l226
					}
					position++
					{
						position236 := position
						depth++
					l237:
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if !_rules[ruleEscapeLiteralChar]() {
									goto l240
								}
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								{
									position241, tokenIndex241, depth241 := position, tokenIndex, depth
									if buffer[position] != rune('\'') {
										goto l241
									}
									position++
									goto l238
								l241:
									position, tokenIndex, depth = position241, tokenIndex241, depth241
								}
								if !matchDot() {
									goto l238
								}
							}
						l239:
							goto l237
						l238:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
						}
						depth--
						add(rulePegText, position236)
					}
					if buffer[position] != rune('\'') {
						goto l226
					}
					position++
					if !_rules[ruleSkip]() {
						goto l226
					}
				}
			l228:
				depth--
				add(ruleLiteral, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 36 Identifier <- <(<(Letter (Letter / Digit / '.')*)> Skip)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				{
					position244 := position
					depth++
					if !_rules[ruleLetter]() {
						goto l242
					}
				l245:
					{
						position246, tokenIndex246, depth246 := position, tokenIndex, depth
						{
							position247, tokenIndex247, depth247 := position, tokenIndex, depth
							if !_rules[ruleLetter]() {
								goto l248
							}
							goto l247
						l248:
							position, tokenIndex, depth = position247, tokenIndex247, depth247
							if !_rules[ruleDigit]() {
								goto l249
							}
							goto l247
						l249:
							position, tokenIndex, depth = position247, tokenIndex247, depth247
							if buffer[position] != rune('.') {
								goto l246
							}
							position++
						}
					l247:
						goto l245
					l246:
						position, tokenIndex, depth = position246, tokenIndex246, depth246
					}
					depth--
					add(rulePegText, position244)
				}
				if !_rules[ruleSkip]() {
					goto l242
				}
				depth--
				add(ruleIdentifier, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 37 ListSeparator <- <((',' / ';') Skip)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l253
					}
					position++
					goto l252
				l253:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
					if buffer[position] != rune(';') {
						goto l250
					}
					position++
				}
			l252:
				if !_rules[ruleSkip]() {
					goto l250
				}
				depth--
				add(ruleListSeparator, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 38 Letter <- <([A-Z] / [a-z] / '_')> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l257
					}
					position++
					goto l256
				l257:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l258
					}
					position++
					goto l256
				l258:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if buffer[position] != rune('_') {
						goto l254
					}
					position++
				}
			l256:
				depth--
				add(ruleLetter, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 39 LetterOrDigit <- <([a-z] / [A-Z] / [0-9] / ('_' / '$'))> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l262
					}
					position++
					goto l261
				l262:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l263
					}
					position++
					goto l261
				l263:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l264
					}
					position++
					goto l261
				l264:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					{
						position265, tokenIndex265, depth265 := position, tokenIndex, depth
						if buffer[position] != rune('_') {
							goto l266
						}
						position++
						goto l265
					l266:
						position, tokenIndex, depth = position265, tokenIndex265, depth265
						if buffer[position] != rune('$') {
							goto l259
						}
						position++
					}
				l265:
				}
			l261:
				depth--
				add(ruleLetterOrDigit, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 40 Digit <- <[0-9]> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l267
				}
				position++
				depth--
				add(ruleDigit, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 41 Skip <- <(Space / LongComment / LineComment / UnixComment)*> */
		func() bool {
			{
				position270 := position
				depth++
			l271:
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					{
						position273, tokenIndex273, depth273 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l274
						}
						goto l273
					l274:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if !_rules[ruleLongComment]() {
							goto l275
						}
						goto l273
					l275:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if !_rules[ruleLineComment]() {
							goto l276
						}
						goto l273
					l276:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if !_rules[ruleUnixComment]() {
							goto l272
						}
					}
				l273:
					goto l271
				l272:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
				}
				depth--
				add(ruleSkip, position270)
			}
			return true
		},
		/* 42 Space <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position277, tokenIndex277, depth277 := position, tokenIndex, depth
			{
				position278 := position
				depth++
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l282
					}
					position++
					goto l281
				l282:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
					if buffer[position] != rune('\t') {
						goto l283
					}
					position++
					goto l281
				l283:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
					if buffer[position] != rune('\r') {
						goto l284
					}
					position++
					goto l281
				l284:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
					if buffer[position] != rune('\n') {
						goto l277
					}
					position++
				}
			l281:
			l279:
				{
					position280, tokenIndex280, depth280 := position, tokenIndex, depth
					{
						position285, tokenIndex285, depth285 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l286
						}
						position++
						goto l285
					l286:
						position, tokenIndex, depth = position285, tokenIndex285, depth285
						if buffer[position] != rune('\t') {
							goto l287
						}
						position++
						goto l285
					l287:
						position, tokenIndex, depth = position285, tokenIndex285, depth285
						if buffer[position] != rune('\r') {
							goto l288
						}
						position++
						goto l285
					l288:
						position, tokenIndex, depth = position285, tokenIndex285, depth285
						if buffer[position] != rune('\n') {
							goto l280
						}
						position++
					}
				l285:
					goto l279
				l280:
					position, tokenIndex, depth = position280, tokenIndex280, depth280
				}
				depth--
				add(ruleSpace, position278)
			}
			return true
		l277:
			position, tokenIndex, depth = position277, tokenIndex277, depth277
			return false
		},
		/* 43 LongComment <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				if buffer[position] != rune('/') {
					goto l289
				}
				position++
				if buffer[position] != rune('*') {
					goto l289
				}
				position++
			l291:
				{
					position292, tokenIndex292, depth292 := position, tokenIndex, depth
					{
						position293, tokenIndex293, depth293 := position, tokenIndex, depth
						if buffer[position] != rune('*') {
							goto l293
						}
						position++
						if buffer[position] != rune('/') {
							goto l293
						}
						position++
						goto l292
					l293:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
					}
					if !matchDot() {
						goto l292
					}
					goto l291
				l292:
					position, tokenIndex, depth = position292, tokenIndex292, depth292
				}
				if buffer[position] != rune('*') {
					goto l289
				}
				position++
				if buffer[position] != rune('/') {
					goto l289
				}
				position++
				depth--
				add(ruleLongComment, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 44 LineComment <- <('/' '/' (!('\r' / '\n') .)*)> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
				if buffer[position] != rune('/') {
					goto l294
				}
				position++
				if buffer[position] != rune('/') {
					goto l294
				}
				position++
			l296:
				{
					position297, tokenIndex297, depth297 := position, tokenIndex, depth
					{
						position298, tokenIndex298, depth298 := position, tokenIndex, depth
						{
							position299, tokenIndex299, depth299 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l300
							}
							position++
							goto l299
						l300:
							position, tokenIndex, depth = position299, tokenIndex299, depth299
							if buffer[position] != rune('\n') {
								goto l298
							}
							position++
						}
					l299:
						goto l297
					l298:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
					}
					if !matchDot() {
						goto l297
					}
					goto l296
				l297:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
				}
				depth--
				add(ruleLineComment, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 45 UnixComment <- <('#' (!('\r' / '\n') .)*)> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if buffer[position] != rune('#') {
					goto l301
				}
				position++
			l303:
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					{
						position305, tokenIndex305, depth305 := position, tokenIndex, depth
						{
							position306, tokenIndex306, depth306 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l307
							}
							position++
							goto l306
						l307:
							position, tokenIndex, depth = position306, tokenIndex306, depth306
							if buffer[position] != rune('\n') {
								goto l305
							}
							position++
						}
					l306:
						goto l304
					l305:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
					}
					if !matchDot() {
						goto l304
					}
					goto l303
				l304:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
				}
				depth--
				add(ruleUnixComment, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 46 BOOL <- <(<('b' 'o' 'o' 'l')> !LetterOrDigit Skip)> */
		func() bool {
			position308, tokenIndex308, depth308 := position, tokenIndex, depth
			{
				position309 := position
				depth++
				{
					position310 := position
					depth++
					if buffer[position] != rune('b') {
						goto l308
					}
					position++
					if buffer[position] != rune('o') {
						goto l308
					}
					position++
					if buffer[position] != rune('o') {
						goto l308
					}
					position++
					if buffer[position] != rune('l') {
						goto l308
					}
					position++
					depth--
					add(rulePegText, position310)
				}
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l311
					}
					goto l308
				l311:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
				}
				if !_rules[ruleSkip]() {
					goto l308
				}
				depth--
				add(ruleBOOL, position309)
			}
			return true
		l308:
			position, tokenIndex, depth = position308, tokenIndex308, depth308
			return false
		},
		/* 47 BYTE <- <(<('b' 'y' 't' 'e')> !LetterOrDigit Skip)> */
		func() bool {
			position312, tokenIndex312, depth312 := position, tokenIndex, depth
			{
				position313 := position
				depth++
				{
					position314 := position
					depth++
					if buffer[position] != rune('b') {
						goto l312
					}
					position++
					if buffer[position] != rune('y') {
						goto l312
					}
					position++
					if buffer[position] != rune('t') {
						goto l312
					}
					position++
					if buffer[position] != rune('e') {
						goto l312
					}
					position++
					depth--
					add(rulePegText, position314)
				}
				{
					position315, tokenIndex315, depth315 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l315
					}
					goto l312
				l315:
					position, tokenIndex, depth = position315, tokenIndex315, depth315
				}
				if !_rules[ruleSkip]() {
					goto l312
				}
				depth--
				add(ruleBYTE, position313)
			}
			return true
		l312:
			position, tokenIndex, depth = position312, tokenIndex312, depth312
			return false
		},
		/* 48 I8 <- <(<('i' '8')> !LetterOrDigit Skip)> */
		func() bool {
			position316, tokenIndex316, depth316 := position, tokenIndex, depth
			{
				position317 := position
				depth++
				{
					position318 := position
					depth++
					if buffer[position] != rune('i') {
						goto l316
					}
					position++
					if buffer[position] != rune('8') {
						goto l316
					}
					position++
					depth--
					add(rulePegText, position318)
				}
				{
					position319, tokenIndex319, depth319 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l319
					}
					goto l316
				l319:
					position, tokenIndex, depth = position319, tokenIndex319, depth319
				}
				if !_rules[ruleSkip]() {
					goto l316
				}
				depth--
				add(ruleI8, position317)
			}
			return true
		l316:
			position, tokenIndex, depth = position316, tokenIndex316, depth316
			return false
		},
		/* 49 I16 <- <(<('i' '1' '6')> !LetterOrDigit Skip)> */
		func() bool {
			position320, tokenIndex320, depth320 := position, tokenIndex, depth
			{
				position321 := position
				depth++
				{
					position322 := position
					depth++
					if buffer[position] != rune('i') {
						goto l320
					}
					position++
					if buffer[position] != rune('1') {
						goto l320
					}
					position++
					if buffer[position] != rune('6') {
						goto l320
					}
					position++
					depth--
					add(rulePegText, position322)
				}
				{
					position323, tokenIndex323, depth323 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l323
					}
					goto l320
				l323:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
				}
				if !_rules[ruleSkip]() {
					goto l320
				}
				depth--
				add(ruleI16, position321)
			}
			return true
		l320:
			position, tokenIndex, depth = position320, tokenIndex320, depth320
			return false
		},
		/* 50 I32 <- <(<('i' '3' '2')> !LetterOrDigit Skip)> */
		func() bool {
			position324, tokenIndex324, depth324 := position, tokenIndex, depth
			{
				position325 := position
				depth++
				{
					position326 := position
					depth++
					if buffer[position] != rune('i') {
						goto l324
					}
					position++
					if buffer[position] != rune('3') {
						goto l324
					}
					position++
					if buffer[position] != rune('2') {
						goto l324
					}
					position++
					depth--
					add(rulePegText, position326)
				}
				{
					position327, tokenIndex327, depth327 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l327
					}
					goto l324
				l327:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
				}
				if !_rules[ruleSkip]() {
					goto l324
				}
				depth--
				add(ruleI32, position325)
			}
			return true
		l324:
			position, tokenIndex, depth = position324, tokenIndex324, depth324
			return false
		},
		/* 51 I64 <- <(<('i' '6' '4')> !LetterOrDigit Skip)> */
		func() bool {
			position328, tokenIndex328, depth328 := position, tokenIndex, depth
			{
				position329 := position
				depth++
				{
					position330 := position
					depth++
					if buffer[position] != rune('i') {
						goto l328
					}
					position++
					if buffer[position] != rune('6') {
						goto l328
					}
					position++
					if buffer[position] != rune('4') {
						goto l328
					}
					position++
					depth--
					add(rulePegText, position330)
				}
				{
					position331, tokenIndex331, depth331 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l331
					}
					goto l328
				l331:
					position, tokenIndex, depth = position331, tokenIndex331, depth331
				}
				if !_rules[ruleSkip]() {
					goto l328
				}
				depth--
				add(ruleI64, position329)
			}
			return true
		l328:
			position, tokenIndex, depth = position328, tokenIndex328, depth328
			return false
		},
		/* 52 DOUBLE <- <(<('d' 'o' 'u' 'b' 'l' 'e')> !LetterOrDigit Skip)> */
		func() bool {
			position332, tokenIndex332, depth332 := position, tokenIndex, depth
			{
				position333 := position
				depth++
				{
					position334 := position
					depth++
					if buffer[position] != rune('d') {
						goto l332
					}
					position++
					if buffer[position] != rune('o') {
						goto l332
					}
					position++
					if buffer[position] != rune('u') {
						goto l332
					}
					position++
					if buffer[position] != rune('b') {
						goto l332
					}
					position++
					if buffer[position] != rune('l') {
						goto l332
					}
					position++
					if buffer[position] != rune('e') {
						goto l332
					}
					position++
					depth--
					add(rulePegText, position334)
				}
				{
					position335, tokenIndex335, depth335 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l335
					}
					goto l332
				l335:
					position, tokenIndex, depth = position335, tokenIndex335, depth335
				}
				if !_rules[ruleSkip]() {
					goto l332
				}
				depth--
				add(ruleDOUBLE, position333)
			}
			return true
		l332:
			position, tokenIndex, depth = position332, tokenIndex332, depth332
			return false
		},
		/* 53 STRING <- <(<('s' 't' 'r' 'i' 'n' 'g')> !LetterOrDigit Skip)> */
		func() bool {
			position336, tokenIndex336, depth336 := position, tokenIndex, depth
			{
				position337 := position
				depth++
				{
					position338 := position
					depth++
					if buffer[position] != rune('s') {
						goto l336
					}
					position++
					if buffer[position] != rune('t') {
						goto l336
					}
					position++
					if buffer[position] != rune('r') {
						goto l336
					}
					position++
					if buffer[position] != rune('i') {
						goto l336
					}
					position++
					if buffer[position] != rune('n') {
						goto l336
					}
					position++
					if buffer[position] != rune('g') {
						goto l336
					}
					position++
					depth--
					add(rulePegText, position338)
				}
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l339
					}
					goto l336
				l339:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
				}
				if !_rules[ruleSkip]() {
					goto l336
				}
				depth--
				add(ruleSTRING, position337)
			}
			return true
		l336:
			position, tokenIndex, depth = position336, tokenIndex336, depth336
			return false
		},
		/* 54 BINARY <- <(<('b' 'i' 'n' 'a' 'r' 'y')> !LetterOrDigit Skip)> */
		func() bool {
			position340, tokenIndex340, depth340 := position, tokenIndex, depth
			{
				position341 := position
				depth++
				{
					position342 := position
					depth++
					if buffer[position] != rune('b') {
						goto l340
					}
					position++
					if buffer[position] != rune('i') {
						goto l340
					}
					position++
					if buffer[position] != rune('n') {
						goto l340
					}
					position++
					if buffer[position] != rune('a') {
						goto l340
					}
					position++
					if buffer[position] != rune('r') {
						goto l340
					}
					position++
					if buffer[position] != rune('y') {
						goto l340
					}
					position++
					depth--
					add(rulePegText, position342)
				}
				{
					position343, tokenIndex343, depth343 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l343
					}
					goto l340
				l343:
					position, tokenIndex, depth = position343, tokenIndex343, depth343
				}
				if !_rules[ruleSkip]() {
					goto l340
				}
				depth--
				add(ruleBINARY, position341)
			}
			return true
		l340:
			position, tokenIndex, depth = position340, tokenIndex340, depth340
			return false
		},
		/* 55 CONST <- <('c' 'o' 'n' 's' 't' !LetterOrDigit Skip)> */
		func() bool {
			position344, tokenIndex344, depth344 := position, tokenIndex, depth
			{
				position345 := position
				depth++
				if buffer[position] != rune('c') {
					goto l344
				}
				position++
				if buffer[position] != rune('o') {
					goto l344
				}
				position++
				if buffer[position] != rune('n') {
					goto l344
				}
				position++
				if buffer[position] != rune('s') {
					goto l344
				}
				position++
				if buffer[position] != rune('t') {
					goto l344
				}
				position++
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l346
					}
					goto l344
				l346:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
				}
				if !_rules[ruleSkip]() {
					goto l344
				}
				depth--
				add(ruleCONST, position345)
			}
			return true
		l344:
			position, tokenIndex, depth = position344, tokenIndex344, depth344
			return false
		},
		/* 56 ONEWAY <- <('o' 'n' 'e' 'w' 'a' 'y' !LetterOrDigit Skip)> */
		func() bool {
			position347, tokenIndex347, depth347 := position, tokenIndex, depth
			{
				position348 := position
				depth++
				if buffer[position] != rune('o') {
					goto l347
				}
				position++
				if buffer[position] != rune('n') {
					goto l347
				}
				position++
				if buffer[position] != rune('e') {
					goto l347
				}
				position++
				if buffer[position] != rune('w') {
					goto l347
				}
				position++
				if buffer[position] != rune('a') {
					goto l347
				}
				position++
				if buffer[position] != rune('y') {
					goto l347
				}
				position++
				{
					position349, tokenIndex349, depth349 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l349
					}
					goto l347
				l349:
					position, tokenIndex, depth = position349, tokenIndex349, depth349
				}
				if !_rules[ruleSkip]() {
					goto l347
				}
				depth--
				add(ruleONEWAY, position348)
			}
			return true
		l347:
			position, tokenIndex, depth = position347, tokenIndex347, depth347
			return false
		},
		/* 57 TYPEDEF <- <('t' 'y' 'p' 'e' 'd' 'e' 'f' !LetterOrDigit Skip)> */
		func() bool {
			position350, tokenIndex350, depth350 := position, tokenIndex, depth
			{
				position351 := position
				depth++
				if buffer[position] != rune('t') {
					goto l350
				}
				position++
				if buffer[position] != rune('y') {
					goto l350
				}
				position++
				if buffer[position] != rune('p') {
					goto l350
				}
				position++
				if buffer[position] != rune('e') {
					goto l350
				}
				position++
				if buffer[position] != rune('d') {
					goto l350
				}
				position++
				if buffer[position] != rune('e') {
					goto l350
				}
				position++
				if buffer[position] != rune('f') {
					goto l350
				}
				position++
				{
					position352, tokenIndex352, depth352 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l352
					}
					goto l350
				l352:
					position, tokenIndex, depth = position352, tokenIndex352, depth352
				}
				if !_rules[ruleSkip]() {
					goto l350
				}
				depth--
				add(ruleTYPEDEF, position351)
			}
			return true
		l350:
			position, tokenIndex, depth = position350, tokenIndex350, depth350
			return false
		},
		/* 58 MAP <- <('m' 'a' 'p' !LetterOrDigit Skip)> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
				depth++
				if buffer[position] != rune('m') {
					goto l353
				}
				position++
				if buffer[position] != rune('a') {
					goto l353
				}
				position++
				if buffer[position] != rune('p') {
					goto l353
				}
				position++
				{
					position355, tokenIndex355, depth355 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l355
					}
					goto l353
				l355:
					position, tokenIndex, depth = position355, tokenIndex355, depth355
				}
				if !_rules[ruleSkip]() {
					goto l353
				}
				depth--
				add(ruleMAP, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 59 SET <- <('s' 'e' 't' !LetterOrDigit Skip)> */
		func() bool {
			position356, tokenIndex356, depth356 := position, tokenIndex, depth
			{
				position357 := position
				depth++
				if buffer[position] != rune('s') {
					goto l356
				}
				position++
				if buffer[position] != rune('e') {
					goto l356
				}
				position++
				if buffer[position] != rune('t') {
					goto l356
				}
				position++
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l358
					}
					goto l356
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
				if !_rules[ruleSkip]() {
					goto l356
				}
				depth--
				add(ruleSET, position357)
			}
			return true
		l356:
			position, tokenIndex, depth = position356, tokenIndex356, depth356
			return false
		},
		/* 60 LIST <- <('l' 'i' 's' 't' !LetterOrDigit Skip)> */
		func() bool {
			position359, tokenIndex359, depth359 := position, tokenIndex, depth
			{
				position360 := position
				depth++
				if buffer[position] != rune('l') {
					goto l359
				}
				position++
				if buffer[position] != rune('i') {
					goto l359
				}
				position++
				if buffer[position] != rune('s') {
					goto l359
				}
				position++
				if buffer[position] != rune('t') {
					goto l359
				}
				position++
				{
					position361, tokenIndex361, depth361 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l361
					}
					goto l359
				l361:
					position, tokenIndex, depth = position361, tokenIndex361, depth361
				}
				if !_rules[ruleSkip]() {
					goto l359
				}
				depth--
				add(ruleLIST, position360)
			}
			return true
		l359:
			position, tokenIndex, depth = position359, tokenIndex359, depth359
			return false
		},
		/* 61 VOID <- <('v' 'o' 'i' 'd' !LetterOrDigit Skip)> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				if buffer[position] != rune('v') {
					goto l362
				}
				position++
				if buffer[position] != rune('o') {
					goto l362
				}
				position++
				if buffer[position] != rune('i') {
					goto l362
				}
				position++
				if buffer[position] != rune('d') {
					goto l362
				}
				position++
				{
					position364, tokenIndex364, depth364 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l364
					}
					goto l362
				l364:
					position, tokenIndex, depth = position364, tokenIndex364, depth364
				}
				if !_rules[ruleSkip]() {
					goto l362
				}
				depth--
				add(ruleVOID, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 62 THROWS <- <('t' 'h' 'r' 'o' 'w' 's' !LetterOrDigit Skip)> */
		func() bool {
			position365, tokenIndex365, depth365 := position, tokenIndex, depth
			{
				position366 := position
				depth++
				if buffer[position] != rune('t') {
					goto l365
				}
				position++
				if buffer[position] != rune('h') {
					goto l365
				}
				position++
				if buffer[position] != rune('r') {
					goto l365
				}
				position++
				if buffer[position] != rune('o') {
					goto l365
				}
				position++
				if buffer[position] != rune('w') {
					goto l365
				}
				position++
				if buffer[position] != rune('s') {
					goto l365
				}
				position++
				{
					position367, tokenIndex367, depth367 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l367
					}
					goto l365
				l367:
					position, tokenIndex, depth = position367, tokenIndex367, depth367
				}
				if !_rules[ruleSkip]() {
					goto l365
				}
				depth--
				add(ruleTHROWS, position366)
			}
			return true
		l365:
			position, tokenIndex, depth = position365, tokenIndex365, depth365
			return false
		},
		/* 63 EXCEPTION <- <('e' 'x' 'c' 'e' 'p' 't' 'i' 'o' 'n' !LetterOrDigit Skip)> */
		func() bool {
			position368, tokenIndex368, depth368 := position, tokenIndex, depth
			{
				position369 := position
				depth++
				if buffer[position] != rune('e') {
					goto l368
				}
				position++
				if buffer[position] != rune('x') {
					goto l368
				}
				position++
				if buffer[position] != rune('c') {
					goto l368
				}
				position++
				if buffer[position] != rune('e') {
					goto l368
				}
				position++
				if buffer[position] != rune('p') {
					goto l368
				}
				position++
				if buffer[position] != rune('t') {
					goto l368
				}
				position++
				if buffer[position] != rune('i') {
					goto l368
				}
				position++
				if buffer[position] != rune('o') {
					goto l368
				}
				position++
				if buffer[position] != rune('n') {
					goto l368
				}
				position++
				{
					position370, tokenIndex370, depth370 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l370
					}
					goto l368
				l370:
					position, tokenIndex, depth = position370, tokenIndex370, depth370
				}
				if !_rules[ruleSkip]() {
					goto l368
				}
				depth--
				add(ruleEXCEPTION, position369)
			}
			return true
		l368:
			position, tokenIndex, depth = position368, tokenIndex368, depth368
			return false
		},
		/* 64 EXTENDS <- <('e' 'x' 't' 'e' 'n' 'd' 's' !LetterOrDigit Skip)> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				if buffer[position] != rune('e') {
					goto l371
				}
				position++
				if buffer[position] != rune('x') {
					goto l371
				}
				position++
				if buffer[position] != rune('t') {
					goto l371
				}
				position++
				if buffer[position] != rune('e') {
					goto l371
				}
				position++
				if buffer[position] != rune('n') {
					goto l371
				}
				position++
				if buffer[position] != rune('d') {
					goto l371
				}
				position++
				if buffer[position] != rune('s') {
					goto l371
				}
				position++
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l373
					}
					goto l371
				l373:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
				}
				if !_rules[ruleSkip]() {
					goto l371
				}
				depth--
				add(ruleEXTENDS, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
			return false
		},
		/* 65 SERVICE <- <('s' 'e' 'r' 'v' 'i' 'c' 'e' !LetterOrDigit Skip)> */
		func() bool {
			position374, tokenIndex374, depth374 := position, tokenIndex, depth
			{
				position375 := position
				depth++
				if buffer[position] != rune('s') {
					goto l374
				}
				position++
				if buffer[position] != rune('e') {
					goto l374
				}
				position++
				if buffer[position] != rune('r') {
					goto l374
				}
				position++
				if buffer[position] != rune('v') {
					goto l374
				}
				position++
				if buffer[position] != rune('i') {
					goto l374
				}
				position++
				if buffer[position] != rune('c') {
					goto l374
				}
				position++
				if buffer[position] != rune('e') {
					goto l374
				}
				position++
				{
					position376, tokenIndex376, depth376 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l376
					}
					goto l374
				l376:
					position, tokenIndex, depth = position376, tokenIndex376, depth376
				}
				if !_rules[ruleSkip]() {
					goto l374
				}
				depth--
				add(ruleSERVICE, position375)
			}
			return true
		l374:
			position, tokenIndex, depth = position374, tokenIndex374, depth374
			return false
		},
		/* 66 STRUCT <- <('s' 't' 'r' 'u' 'c' 't' !LetterOrDigit Skip)> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				if buffer[position] != rune('s') {
					goto l377
				}
				position++
				if buffer[position] != rune('t') {
					goto l377
				}
				position++
				if buffer[position] != rune('r') {
					goto l377
				}
				position++
				if buffer[position] != rune('u') {
					goto l377
				}
				position++
				if buffer[position] != rune('c') {
					goto l377
				}
				position++
				if buffer[position] != rune('t') {
					goto l377
				}
				position++
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l379
					}
					goto l377
				l379:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
				}
				if !_rules[ruleSkip]() {
					goto l377
				}
				depth--
				add(ruleSTRUCT, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 67 UNION <- <('u' 'n' 'i' 'o' 'n' !LetterOrDigit Skip)> */
		func() bool {
			position380, tokenIndex380, depth380 := position, tokenIndex, depth
			{
				position381 := position
				depth++
				if buffer[position] != rune('u') {
					goto l380
				}
				position++
				if buffer[position] != rune('n') {
					goto l380
				}
				position++
				if buffer[position] != rune('i') {
					goto l380
				}
				position++
				if buffer[position] != rune('o') {
					goto l380
				}
				position++
				if buffer[position] != rune('n') {
					goto l380
				}
				position++
				{
					position382, tokenIndex382, depth382 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l382
					}
					goto l380
				l382:
					position, tokenIndex, depth = position382, tokenIndex382, depth382
				}
				if !_rules[ruleSkip]() {
					goto l380
				}
				depth--
				add(ruleUNION, position381)
			}
			return true
		l380:
			position, tokenIndex, depth = position380, tokenIndex380, depth380
			return false
		},
		/* 68 ENUM <- <('e' 'n' 'u' 'm' !LetterOrDigit Skip)> */
		func() bool {
			position383, tokenIndex383, depth383 := position, tokenIndex, depth
			{
				position384 := position
				depth++
				if buffer[position] != rune('e') {
					goto l383
				}
				position++
				if buffer[position] != rune('n') {
					goto l383
				}
				position++
				if buffer[position] != rune('u') {
					goto l383
				}
				position++
				if buffer[position] != rune('m') {
					goto l383
				}
				position++
				{
					position385, tokenIndex385, depth385 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l385
					}
					goto l383
				l385:
					position, tokenIndex, depth = position385, tokenIndex385, depth385
				}
				if !_rules[ruleSkip]() {
					goto l383
				}
				depth--
				add(ruleENUM, position384)
			}
			return true
		l383:
			position, tokenIndex, depth = position383, tokenIndex383, depth383
			return false
		},
		/* 69 INCLUDE <- <('i' 'n' 'c' 'l' 'u' 'd' 'e' !LetterOrDigit Skip)> */
		func() bool {
			position386, tokenIndex386, depth386 := position, tokenIndex, depth
			{
				position387 := position
				depth++
				if buffer[position] != rune('i') {
					goto l386
				}
				position++
				if buffer[position] != rune('n') {
					goto l386
				}
				position++
				if buffer[position] != rune('c') {
					goto l386
				}
				position++
				if buffer[position] != rune('l') {
					goto l386
				}
				position++
				if buffer[position] != rune('u') {
					goto l386
				}
				position++
				if buffer[position] != rune('d') {
					goto l386
				}
				position++
				if buffer[position] != rune('e') {
					goto l386
				}
				position++
				{
					position388, tokenIndex388, depth388 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l388
					}
					goto l386
				l388:
					position, tokenIndex, depth = position388, tokenIndex388, depth388
				}
				if !_rules[ruleSkip]() {
					goto l386
				}
				depth--
				add(ruleINCLUDE, position387)
			}
			return true
		l386:
			position, tokenIndex, depth = position386, tokenIndex386, depth386
			return false
		},
		/* 70 CPPINCLUDE <- <('c' 'p' 'p' '_' 'i' 'n' 'c' 'l' 'u' 'd' 'e' !LetterOrDigit Skip)> */
		func() bool {
			position389, tokenIndex389, depth389 := position, tokenIndex, depth
			{
				position390 := position
				depth++
				if buffer[position] != rune('c') {
					goto l389
				}
				position++
				if buffer[position] != rune('p') {
					goto l389
				}
				position++
				if buffer[position] != rune('p') {
					goto l389
				}
				position++
				if buffer[position] != rune('_') {
					goto l389
				}
				position++
				if buffer[position] != rune('i') {
					goto l389
				}
				position++
				if buffer[position] != rune('n') {
					goto l389
				}
				position++
				if buffer[position] != rune('c') {
					goto l389
				}
				position++
				if buffer[position] != rune('l') {
					goto l389
				}
				position++
				if buffer[position] != rune('u') {
					goto l389
				}
				position++
				if buffer[position] != rune('d') {
					goto l389
				}
				position++
				if buffer[position] != rune('e') {
					goto l389
				}
				position++
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l391
					}
					goto l389
				l391:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
				}
				if !_rules[ruleSkip]() {
					goto l389
				}
				depth--
				add(ruleCPPINCLUDE, position390)
			}
			return true
		l389:
			position, tokenIndex, depth = position389, tokenIndex389, depth389
			return false
		},
		/* 71 NAMESPACE <- <('n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !LetterOrDigit Skip)> */
		func() bool {
			position392, tokenIndex392, depth392 := position, tokenIndex, depth
			{
				position393 := position
				depth++
				if buffer[position] != rune('n') {
					goto l392
				}
				position++
				if buffer[position] != rune('a') {
					goto l392
				}
				position++
				if buffer[position] != rune('m') {
					goto l392
				}
				position++
				if buffer[position] != rune('e') {
					goto l392
				}
				position++
				if buffer[position] != rune('s') {
					goto l392
				}
				position++
				if buffer[position] != rune('p') {
					goto l392
				}
				position++
				if buffer[position] != rune('a') {
					goto l392
				}
				position++
				if buffer[position] != rune('c') {
					goto l392
				}
				position++
				if buffer[position] != rune('e') {
					goto l392
				}
				position++
				{
					position394, tokenIndex394, depth394 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l394
					}
					goto l392
				l394:
					position, tokenIndex, depth = position394, tokenIndex394, depth394
				}
				if !_rules[ruleSkip]() {
					goto l392
				}
				depth--
				add(ruleNAMESPACE, position393)
			}
			return true
		l392:
			position, tokenIndex, depth = position392, tokenIndex392, depth392
			return false
		},
		/* 72 CPPTYPE <- <('c' 'p' 'p' '_' 't' 'y' 'p' 'e' !LetterOrDigit Skip)> */
		func() bool {
			position395, tokenIndex395, depth395 := position, tokenIndex, depth
			{
				position396 := position
				depth++
				if buffer[position] != rune('c') {
					goto l395
				}
				position++
				if buffer[position] != rune('p') {
					goto l395
				}
				position++
				if buffer[position] != rune('p') {
					goto l395
				}
				position++
				if buffer[position] != rune('_') {
					goto l395
				}
				position++
				if buffer[position] != rune('t') {
					goto l395
				}
				position++
				if buffer[position] != rune('y') {
					goto l395
				}
				position++
				if buffer[position] != rune('p') {
					goto l395
				}
				position++
				if buffer[position] != rune('e') {
					goto l395
				}
				position++
				{
					position397, tokenIndex397, depth397 := position, tokenIndex, depth
					if !_rules[ruleLetterOrDigit]() {
						goto l397
					}
					goto l395
				l397:
					position, tokenIndex, depth = position397, tokenIndex397, depth397
				}
				if !_rules[ruleSkip]() {
					goto l395
				}
				depth--
				add(ruleCPPTYPE, position396)
			}
			return true
		l395:
			position, tokenIndex, depth = position395, tokenIndex395, depth395
			return false
		},
		/* 73 LBRK <- <('[' Skip)> */
		func() bool {
			position398, tokenIndex398, depth398 := position, tokenIndex, depth
			{
				position399 := position
				depth++
				if buffer[position] != rune('[') {
					goto l398
				}
				position++
				if !_rules[ruleSkip]() {
					goto l398
				}
				depth--
				add(ruleLBRK, position399)
			}
			return true
		l398:
			position, tokenIndex, depth = position398, tokenIndex398, depth398
			return false
		},
		/* 74 RBRK <- <(']' Skip)> */
		func() bool {
			position400, tokenIndex400, depth400 := position, tokenIndex, depth
			{
				position401 := position
				depth++
				if buffer[position] != rune(']') {
					goto l400
				}
				position++
				if !_rules[ruleSkip]() {
					goto l400
				}
				depth--
				add(ruleRBRK, position401)
			}
			return true
		l400:
			position, tokenIndex, depth = position400, tokenIndex400, depth400
			return false
		},
		/* 75 LWING <- <('{' Skip)> */
		func() bool {
			position402, tokenIndex402, depth402 := position, tokenIndex, depth
			{
				position403 := position
				depth++
				if buffer[position] != rune('{') {
					goto l402
				}
				position++
				if !_rules[ruleSkip]() {
					goto l402
				}
				depth--
				add(ruleLWING, position403)
			}
			return true
		l402:
			position, tokenIndex, depth = position402, tokenIndex402, depth402
			return false
		},
		/* 76 RWING <- <('}' Skip)> */
		func() bool {
			position404, tokenIndex404, depth404 := position, tokenIndex, depth
			{
				position405 := position
				depth++
				if buffer[position] != rune('}') {
					goto l404
				}
				position++
				if !_rules[ruleSkip]() {
					goto l404
				}
				depth--
				add(ruleRWING, position405)
			}
			return true
		l404:
			position, tokenIndex, depth = position404, tokenIndex404, depth404
			return false
		},
		/* 77 EQUAL <- <('=' Skip)> */
		func() bool {
			position406, tokenIndex406, depth406 := position, tokenIndex, depth
			{
				position407 := position
				depth++
				if buffer[position] != rune('=') {
					goto l406
				}
				position++
				if !_rules[ruleSkip]() {
					goto l406
				}
				depth--
				add(ruleEQUAL, position407)
			}
			return true
		l406:
			position, tokenIndex, depth = position406, tokenIndex406, depth406
			return false
		},
		/* 78 LPOINT <- <('<' Skip)> */
		func() bool {
			position408, tokenIndex408, depth408 := position, tokenIndex, depth
			{
				position409 := position
				depth++
				if buffer[position] != rune('<') {
					goto l408
				}
				position++
				if !_rules[ruleSkip]() {
					goto l408
				}
				depth--
				add(ruleLPOINT, position409)
			}
			return true
		l408:
			position, tokenIndex, depth = position408, tokenIndex408, depth408
			return false
		},
		/* 79 RPOINT <- <('>' Skip)> */
		func() bool {
			position410, tokenIndex410, depth410 := position, tokenIndex, depth
			{
				position411 := position
				depth++
				if buffer[position] != rune('>') {
					goto l410
				}
				position++
				if !_rules[ruleSkip]() {
					goto l410
				}
				depth--
				add(ruleRPOINT, position411)
			}
			return true
		l410:
			position, tokenIndex, depth = position410, tokenIndex410, depth410
			return false
		},
		/* 80 COMMA <- <(',' Skip)> */
		func() bool {
			position412, tokenIndex412, depth412 := position, tokenIndex, depth
			{
				position413 := position
				depth++
				if buffer[position] != rune(',') {
					goto l412
				}
				position++
				if !_rules[ruleSkip]() {
					goto l412
				}
				depth--
				add(ruleCOMMA, position413)
			}
			return true
		l412:
			position, tokenIndex, depth = position412, tokenIndex412, depth412
			return false
		},
		/* 81 LPAR <- <('(' Skip)> */
		func() bool {
			position414, tokenIndex414, depth414 := position, tokenIndex, depth
			{
				position415 := position
				depth++
				if buffer[position] != rune('(') {
					goto l414
				}
				position++
				if !_rules[ruleSkip]() {
					goto l414
				}
				depth--
				add(ruleLPAR, position415)
			}
			return true
		l414:
			position, tokenIndex, depth = position414, tokenIndex414, depth414
			return false
		},
		/* 82 RPAR <- <(')' Skip)> */
		func() bool {
			position416, tokenIndex416, depth416 := position, tokenIndex, depth
			{
				position417 := position
				depth++
				if buffer[position] != rune(')') {
					goto l416
				}
				position++
				if !_rules[ruleSkip]() {
					goto l416
				}
				depth--
				add(ruleRPAR, position417)
			}
			return true
		l416:
			position, tokenIndex, depth = position416, tokenIndex416, depth416
			return false
		},
		/* 83 COLON <- <(':' Skip)> */
		func() bool {
			position418, tokenIndex418, depth418 := position, tokenIndex, depth
			{
				position419 := position
				depth++
				if buffer[position] != rune(':') {
					goto l418
				}
				position++
				if !_rules[ruleSkip]() {
					goto l418
				}
				depth--
				add(ruleCOLON, position419)
			}
			return true
		l418:
			position, tokenIndex, depth = position418, tokenIndex418, depth418
			return false
		},
		nil,
	}
	p.rules = _rules
}
