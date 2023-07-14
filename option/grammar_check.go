package option

import (
	"errors"
	"github.com/cloudwego/thriftgo/parser"
)

func CheckOptionGrammar(ast *parser.Thrift) error {
	for _, s := range ast.Structs {
		if optionDefMap[s.Name] {
			continue
		}
		_, err := ParseStructOption(s, ast)
		if err != nil {
			return errors.New("Option Check:" + s.Name + " failed:" + err.Error())
		}
		for _, f := range s.Fields {
			_, er := ParseFieldOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}

	}
	for _, s := range ast.Services {
		_, err := ParseServiceOption(s, ast)
		if err != nil {
			return errors.New("Option Check:" + s.Name + " failed:" + err.Error())
		}
		for _, f := range s.Functions {
			_, er := ParseMethodOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}
	}
	for _, en := range ast.Enums {
		_, err := ParseEnumOption(en, ast)
		if err != nil {
			return errors.New("Option Check:" + en.Name + " failed:" + err.Error())
		}
		for _, f := range en.Values {
			_, er := ParseEnumValueOption(f, ast)
			if er != nil {
				return errors.New("Option Check:" + f.Name + " failed:" + er.Error())
			}
		}
	}
	return nil
}
