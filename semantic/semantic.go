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

package semantic

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
)

// Checker reports whether there are semantic errors in the AST and produces
// warning messages for non-fatal errors.
type Checker interface {
	CheckAll(t *parser.Thrift) (warns []string, err error)
}

// Options controls the behavior of the default checker.
type Options struct {
	FixWarnings bool
}

type checker struct {
	Options
}

// NewChecker creates a checker.
func NewChecker(opt Options) Checker {
	return &checker{opt}
}

// CheckAll implements the Checker interface.
func (c *checker) CheckAll(t *parser.Thrift) (warns []string, err error) {
	checks := []func(t *parser.Thrift) ([]string, error){
		c.CheckEnums,
		c.CheckUnions,
		c.CheckFunctions,
	}
	for _, f := range checks {
		ws, err := f(t)
		warns = append(warns, ws...)
		if err != nil {
			return warns, err
		}
	}
	return
}

func (c *checker) CheckEnums(t *parser.Thrift) (warns []string, err error) {
	for _, e := range t.Enums {
		exist := make(map[string]bool)
		v2n := make(map[int64]string)
		for _, v := range e.Values {
			if exist[v.Name] {
				err = fmt.Errorf("enum %s has duplicated value: %s", e.Name, v.Name)
			}
			exist[v.Name] = true
			if n, ok := v2n[v.Value]; ok && n != v.Name {
				err = fmt.Errorf(
					"enum %s: duplicate value %d between '%s' and '%s'",
					e.Name, v.Value, n, v.Name,
				)
			}
			v2n[v.Value] = v.Name
			if err != nil {
				return
			}
		}
	}
	return
}

// CheckUnions checks the semantics of union nodes.
func (c *checker) CheckUnions(t *parser.Thrift) (warns []string, err error) {
	for _, u := range t.Unions {
		var hasDefault bool
		for _, f := range u.Fields {
			if f.Requiredness == parser.FieldType_Required {
				msg := fmt.Sprintf(
					"Union %s field %s: union members must be optional, ignoring specified requiredness.",
					u.Name, f.Name)
				warns = append(warns, msg)
			}

			if f.GetDefault() != nil {
				if hasDefault {
					err = fmt.Errorf("Field %s provides another default value for union %s", f.Name, u.Name)
					return warns, err
				}
			}

			if c.FixWarnings {
				f.Requiredness = parser.FieldType_Optional
			}
		}
	}
	return
}

// CheckFunctions checks the semantics of service functions.
func (c *checker) CheckFunctions(t *parser.Thrift) (warns []string, err error) {
	var argOpt string
	for _, svc := range t.Services {
		for _, f := range svc.Functions {
			for _, a := range f.Arguments {
				if a.Requiredness == parser.FieldType_Optional {
					argOpt = t.Filename + ": optional keyword is ignored in argument lists."
					if c.FixWarnings {
						a.Requiredness = parser.FieldType_Default
					}
				}
			}
		}
	}
	if argOpt != "" {
		warns = append(warns, argOpt)
	}
	return
}
