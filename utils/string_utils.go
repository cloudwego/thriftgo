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
	var dq, sq = true, true
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
	var dq, sq = true, true
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
