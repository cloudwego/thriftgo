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

package reserved

// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#keywords
func init() {
	Register("JavaScript",
		"break",
		"case",
		"catch",
		"class",
		"const",
		"continue",
		"debugger",
		"default",
		"delete",
		"do",
		"else",
		"export",
		"extends",
		"finally",
		"for",
		"function",
		"if",
		"import",
		"in",
		"instanceof",
		"new",
		"return",
		"super",
		"switch",
		"this",
		"throw",
		"try",
		"typeof",
		"var",
		"void",
		"while",
		"with",
		"yield",
		"enum",
		"implements",
		"interface",
		"let",
		"package",
		"private",
		"protected",
		"public",
		"static",
		"yield",
		"await",
		"abstract",
		"boolean",
		"byte",
		"char",
		"double",
		"final",
		"float",
		"goto",
		"int",
		"long",
		"native",
		"short",
		"synchronized",
		"throws",
		"transient",
		"volatile",
		"null",
		"true",
		"false",
	)
}
