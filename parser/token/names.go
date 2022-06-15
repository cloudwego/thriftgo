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

var tokNames = map[Tok]string{
	LexicalError:  "LexicalError",
	EOF:           "EOF",
	BlockComment:  "BlockComment",
	LineComment:   "LineComment",
	UnixComment:   "UnixComment",
	NewLine:       "NewLine",
	Whitespaces:   "Whitespaces",
	Bool:          "Bool",
	Byte:          "Byte",
	I8:            "I8",
	I16:           "I16",
	I32:           "I32",
	I64:           "I64",
	Double:        "Double",
	String:        "String",
	Binary:        "Binary",
	Const:         "Const",
	Oneway:        "Oneway",
	Typedef:       "Typedef",
	Map:           "Map",
	Set:           "Set",
	List:          "List",
	Void:          "Void",
	Throws:        "Throws",
	Exception:     "Exception",
	Extends:       "Extends",
	Required:      "Required",
	Optional:      "Optional",
	Service:       "Service",
	Struct:        "Struct",
	Union:         "Union",
	Enum:          "Enum",
	Include:       "Include",
	CppInclude:    "CppInclude",
	Asterisk:      "Asterisk",
	LBracket:      "LBracket",
	RBracket:      "RBracket",
	LBrace:        "LBrace",
	RBrace:        "RBrace",
	LParenthesis:  "LParenthesis",
	RParenthesis:  "RParenthesis",
	LChevron:      "LChevron",
	RChevron:      "RChevron",
	Equal:         "Equal",
	Comma:         "Comma",
	Colon:         "Colon",
	Semicolon:     "Semicolon",
	StringLiteral: "StringLiteral",
	Identifier:    "Identifier",
	IntLiteral:    "IntLiteral",
	FloatLiteral:  "FloatLiteral",
}
