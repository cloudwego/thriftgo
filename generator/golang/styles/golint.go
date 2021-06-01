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

package styles

import (
	"strings"
	"unicode"

	"github.com/cloudwego/thriftgo/generator/golang/common"
)

// GoLint .
type GoLint struct {
	nolint bool
}

// Identify implements NamingStyle.
func (g *GoLint) Identify(name string) (string, error) {
	return g.convertName(name), nil
}

// UseInitialisms implements NamingStyle.
func (g *GoLint) UseInitialisms(enable bool) {
	g.nolint = !enable
}

// convertName convert name to thrift name, but will not add underline
func (g *GoLint) convertName(name string) string {
	if name == "" {
		return ""
	}
	var result string
	words := strings.Split(name, "_")
	for i, str := range words {
		if str == "" {
			words[i] = "_"
		} else if str[0] >= 'A' && str[0] <= 'Z' && i != 0 {
			words[i] = "_" + str
		} else if str[0] >= '0' && str[0] <= '9' && i != 0 {
			words[i] = "_" + str
		} else {
			words[i] = common.UpperFirstRune(str)
		}
	}
	result = strings.Join(words, "")
	return g.lintName(result)
}

// lintName returns a different name if it should be different.
// Adapted from https://github.com/golang/lint/blob/master/lint.go#L700.
func (g *GoLint) lintName(name string) string {
	if g.nolint {
		return name
	}

	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); common.IsCommonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}
