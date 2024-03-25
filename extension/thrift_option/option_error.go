// Copyright 2024 CloudWeGo Authors
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

package thrift_option

import (
	"errors"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

var (
	ErrKeyNotMatch    = errors.New("key not matched")
	ErrNotExistOption = errors.New("option does not exist")
	ErrParseFailed    = errors.New("failed to parse option")
	ErrNotAllowOption = errors.New("not allowed to parse option")
	ErrNotIncluded    = errors.New("no such prefix found in the given include IDLs")
)

type OptionError struct {
	optionName string
	detail     error
	reason     error
}

func ParseFailedError(optionName string, reason error) error {
	return &OptionError{
		optionName: optionName,
		detail:     ErrParseFailed,
		reason:     reason,
	}
}

func NotIncludedError(optionName string) error {
	return &OptionError{
		optionName: optionName,
		detail:     ErrNotIncluded,
	}
}

func NotAllowError(optionName string) error {
	return &OptionError{
		optionName: optionName,
		detail:     ErrNotAllowOption,
	}
}

func KeyNotMatchError(optionName string) error {
	return &OptionError{
		optionName: optionName,
		detail:     ErrKeyNotMatch,
	}
}

func NotExistError(optionName string) error {
	return &OptionError{
		optionName: optionName,
		detail:     ErrNotExistOption,
	}
}

func (e *OptionError) Error() string {
	if e.reason != nil {
		return e.optionName + ":" + e.detail.Error() + ":" + e.reason.Error()
	}
	return e.optionName + ":" + e.detail.Error()
}

func (e *OptionError) Unwrap() error {
	return e.detail
}

func hasOptionCompileError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrKeyNotMatch) {
		return false
	}
	if errors.Is(err, ErrNotExistOption) {
		return false
	}
	if errors.Is(err, ErrNotIncluded) {
		return false
	}
	return true
}

func CheckOptionGrammar(ast *parser.Thrift) error {
	_, fd := thrift_reflection.RegisterAST(ast)

	for _, s := range fd.Structs {
		if optionDefMap[s.Name] {
			continue
		}
		for an := range s.Annotations {
			_, err := ParseStructOption(s, an)
			if hasOptionCompileError(err) {
				return err
			}
		}

		for _, f := range s.Fields {
			for fan := range f.Annotations {
				_, err := ParseFieldOption(f, fan)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}

	}
	for _, s := range fd.Services {
		for san := range s.Annotations {
			_, err := ParseServiceOption(s, san)
			if hasOptionCompileError(err) {
				return err
			}
		}
		for _, f := range s.Methods {
			for fa := range f.Annotations {
				_, err := ParseMethodOption(f, fa)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}
	}
	for _, en := range fd.Enums {
		for an := range en.Annotations {
			_, err := ParseEnumOption(en, an)
			if hasOptionCompileError(err) {
				return err
			}
		}
		for _, f := range en.Values {
			for an := range f.Annotations {
				_, err := ParseEnumValueOption(f, an)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}
	}
	return nil
}
