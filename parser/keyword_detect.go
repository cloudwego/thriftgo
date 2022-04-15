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

package parser

import (
	"fmt"
	"strings"

	"github.com/cloudwego/thriftgo/pkg/reserved"
)

// DetectKeyword detects if there is any identifier using a reserved
// word in common programming languages.
func DetectKeyword(t *Thrift) (warnings []string) {
	p := func(kind, name string) string {
		return fmt.Sprintf("<%s:%q>", kind, name)
	}
	report := func(langs []string, word string, scope ...string) {
		msg := fmt.Sprintf("%q is a reserved word in %v (%s)",
			word, langs, strings.Join(scope, "."))
		warnings = append(warnings, msg)
	}
	for _, v := range t.Typedefs {
		if langs := reserved.Hit(v.Alias); len(langs) > 0 {
			report(langs, v.Alias, p("typedef", v.Alias))
		}
	}
	for _, v := range t.Constants {
		if langs := reserved.Hit(v.Name); len(langs) > 0 {
			report(langs, v.Name, p("constant", v.Name))
		}
	}
	for _, v := range t.Enums {
		if langs := reserved.Hit(v.Name); len(langs) > 0 {
			report(langs, v.Name, p("enum", v.Name))
		}
	}
	for _, v := range t.GetStructLikes() {
		if langs := reserved.Hit(v.Name); len(langs) > 0 {
			report(langs, v.Name, p(v.Category, v.Name))
		}
		for _, f := range v.Fields {
			if langs := reserved.Hit(f.Name); len(langs) > 0 {
				report(langs, f.Name, p(v.Category, v.Name), p("field", f.Name))
			}
		}
	}
	for _, v := range t.Services {
		if langs := reserved.Hit(v.Name); len(langs) > 0 {
			report(langs, v.Name, p("service", v.Name))
		}
		for _, f := range v.Functions {
			if langs := reserved.Hit(f.Name); len(langs) > 0 {
				report(langs, f.Name, p("service", v.Name), p("function", f.Name))
			}
			for _, a := range f.Arguments {
				if langs := reserved.Hit(a.Name); len(langs) > 0 {
					report(langs, a.Name, p("service", v.Name), p("function", f.Name), p("parameter", a.Name))
				}
			}
			for _, a := range f.Throws {
				if langs := reserved.Hit(a.Name); len(langs) > 0 {
					report(langs, a.Name, p("service", v.Name), p("function", f.Name), p("exception", a.Name))
				}
			}
		}
	}
	return
}
