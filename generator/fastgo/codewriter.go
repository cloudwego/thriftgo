/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fastgo

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

type codewriter struct {
	*bytes.Buffer

	pkgs map[string]string // import -> alias
}

func newCodewriter() *codewriter {
	return &codewriter{
		Buffer: &bytes.Buffer{},
		pkgs:   make(map[string]string),
	}
}

func (w *codewriter) UsePkg(s, a string) {
	if path.Base(s) == a {
		w.pkgs[s] = ""
	} else {
		w.pkgs[s] = a
	}
}

func (w *codewriter) Imports() string {
	pp0 := make([]string, 0, len(w.pkgs))
	pp1 := make([]string, 0, len(w.pkgs)) // for cloudwego
	for pkg, _ := range w.pkgs {          // grouping
		if strings.HasPrefix(pkg, cloudwegoRepoPrefix) {
			pp1 = append(pp1, pkg)
		} else {
			pp0 = append(pp0, pkg)
		}
	}

	// check if need an empty line between groups
	if len(pp0) != 0 && len(pp1) > 0 {
		pp0 = append(pp0, "")
	}

	// no imports?
	pp0 = append(pp0, pp1...)
	if len(pp0) == 0 {
		return ""
	}

	// only imports one pkg?
	if len(pp0) == 1 {
		return fmt.Sprintf("import %s %q", w.pkgs[pp0[0]], pp0[0])
	}

	// more than one imports
	s := &strings.Builder{}
	fmt.Fprintln(s, "import (")
	for _, p := range pp0 {
		if p == "" {
			fmt.Fprintln(s, "")
		} else {
			fmt.Fprintf(s, "%s %q\n", w.pkgs[p], p)
		}
	}
	fmt.Fprintln(s, ")")
	return s.String()
}

func (w *codewriter) f(format string, a ...interface{}) {
	fmt.Fprintf(w, format, a...)

	// always newline for each call
	if len(format) == 0 || format[len(format)-1] != '\n' {
		w.WriteByte('\n')
	}
}
