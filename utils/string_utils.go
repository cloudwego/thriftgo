// Copyright 2023 CloudWeGo Authors
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

package utils

import (
	"errors"
	"strings"
)

// ParseArr parses a string such as [xx,xx,xx,xxx] into string arr
func ParseArr(str string) ([]string, error) {
	for {
		newstr := strings.ReplaceAll(str, "\t", " ")
		newstr = strings.ReplaceAll(newstr, "\n", " ")
		newstr = strings.ReplaceAll(newstr, " ,", ",")
		newstr = strings.ReplaceAll(newstr, ", ", ",")
		newstr = strings.ReplaceAll(newstr, " ]", "]")
		newstr = strings.ReplaceAll(newstr, "[ ", "[")
		newstr = strings.ReplaceAll(newstr, "  ", " ")
		newstr = strings.TrimSpace(newstr)
		if len(newstr) == len(str) {
			break
		}
		str = newstr
	}
	if !(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]")) {
		return nil, errors.New("no []")
	}
	str = str[1 : len(str)-1]

	var cb, sb, kstart, kend int
	var key string
	dq, sq := true, true
	result := []string{}
	for i := 0; i < len(str); i++ {
		ch := str[i]
		if ch == '"' {
			dq = !dq
			continue
		}
		if ch == '\'' {
			sq = !sq
			continue
		}
		if ch == '{' {
			cb++
			continue
		}
		if ch == '}' {
			cb--
			continue
		}
		if ch == '[' {
			sb++
			continue
		}
		if ch == ']' {
			sb--
			continue
		}
		if ch == ',' {
			if sb == 0 && cb == 0 && dq && sq {
				kend = i
				key = str[kstart:kend]
				kstart = i + 1
				result = append(result, key)
			}
			continue
		}
	}
	if sb == 0 && cb == 0 && dq && sq {
		kend = len(str)
		if kstart >= kend {
			return nil, errors.New("grammar error")
		}
		key = str[kstart:kend]
		result = append(result, key)
		return result, nil
	} else {
		if dq && sq {
			return nil, errors.New("{} not match")
		} else {
			return nil, errors.New("quote not match")
		}
	}
}

// ParseKV parses a string such as {a:b,c:d} into a string map
func ParseKV(str string) (map[string]string, error) {
	for {
		newstr := strings.ReplaceAll(str, "\t", " ")
		newstr = strings.ReplaceAll(newstr, "\n", " ")
		newstr = strings.ReplaceAll(newstr, " }", "}")
		newstr = strings.ReplaceAll(newstr, "{ ", "{")
		newstr = strings.ReplaceAll(newstr, " :", ":")
		newstr = strings.ReplaceAll(newstr, ": ", ":")
		newstr = strings.ReplaceAll(newstr, "  ", " ")
		newstr = strings.TrimSpace(newstr)
		if len(newstr) == len(str) {
			break
		}
		str = newstr
	}

	if !(strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) {
		return nil, errors.New("no {}")
	}
	str = str[1 : len(str)-1]

	var cb, sb, kstart, kend, vstart, vend int
	dq, sq := true, true
	var key, value string
	result := map[string]string{}
	for i := 0; i < len(str); i++ {
		ch := str[i]
		if ch == '"' {
			dq = !dq
			continue
		}
		if ch == '\'' {
			sq = !sq
			continue
		}
		if ch == '{' {
			cb++
			continue
		}
		if ch == '}' {
			cb--
			continue
		}
		if ch == '[' {
			sb++
			continue
		}
		if ch == ']' {
			sb--
			continue
		}
		if ch == ':' {
			if sb == 0 && cb == 0 && dq && sq {
				kend = i
				// get k
				vstart = i + 1
				key = str[kstart:kend]
			}
			continue
		}
		if ch == ' ' {
			if sb == 0 && cb == 0 && dq && sq {
				vend = i
				if vstart >= vend {
					return nil, errors.New("grammar error")
				}
				kstart = i + 1
				value = str[vstart:vend]
				result[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
			continue
		}
	}
	if sb == 0 && cb == 0 && dq && sq {
		vend = len(str)
		if vstart >= vend {
			return nil, errors.New("grammar error")
		}
		if kstart >= kend {
			return nil, errors.New("grammar error")
		}
		value = str[vstart:vend]
		result[strings.TrimSpace(key)] = strings.TrimSpace(value)
		return result, nil
	} else {
		if dq && sq {
			return nil, errors.New("{} not match")
		} else {
			return nil, errors.New("quote not match")
		}
	}
}

func SplitSubfix(t string) (typ, val string) {
	idx := strings.LastIndex(t, ".")
	if idx == -1 {
		return "", t
	}
	return t[:idx], t[idx+1:]
}
