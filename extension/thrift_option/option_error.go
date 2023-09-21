package thrift_option

import (
	"errors"
	"github.com/cloudwego/thriftgo/parser"
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
	for _, s := range ast.Structs {
		if optionDefMap[s.Name] {
			continue
		}
		for _, an := range s.Annotations {
			_, err := ParseStructOption(s, an.GetKey(), ast)
			if hasOptionCompileError(err) {
				return err
			}
		}

		for _, f := range s.Fields {
			for _, fan := range f.Annotations {
				_, err := ParseFieldOption(f, fan.GetKey(), ast)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}

	}
	for _, s := range ast.Services {
		for _, san := range s.Annotations {
			_, err := ParseServiceOption(s, san.GetKey(), ast)
			if hasOptionCompileError(err) {
				return err
			}
		}
		for _, f := range s.Functions {
			for _, fa := range f.Annotations {
				_, err := ParseMethodOption(f, fa.GetKey(), ast)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}
	}
	for _, en := range ast.Enums {
		for _, an := range en.Annotations {
			_, err := ParseEnumOption(en, an.GetKey(), ast)
			if hasOptionCompileError(err) {
				return err
			}
		}
		for _, f := range en.Values {
			for _, an := range f.Annotations {
				_, err := ParseEnumValueOption(f, an.GetKey(), ast)
				if hasOptionCompileError(err) {
					return err
				}
			}
		}
	}
	return nil
}
