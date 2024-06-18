package trim

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
)

func TrimBatchContent(mainIDLFilePath string, IDLFileContentMap map[string]string) (map[string]string, error) {
	return TrimBatchContentWithConfig(mainIDLFilePath, IDLFileContentMap, TrimASTArg{
		TrimMethods: nil,
		Preserve:    nil,
		MatchGoName: nil,
	})
}

func TrimBatchContentWithConfig(mainIDLFilePath string, IDLFileContentMap map[string]string, trimArgs TrimASTArg) (map[string]string, error) {
	ast, err := parser.ParseBatchString(mainIDLFilePath, IDLFileContentMap, nil)
	if err != nil {
		return nil, err
	}

	if path := parser.CircleDetect(ast); len(path) > 0 {
		return nil, fmt.Errorf("found include circle:\n\t%s\n", path)
	}
	checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
	_, err = checker.CheckAll(ast)
	if err != nil {
		return nil, err
	}
	err = semantic.ResolveSymbols(ast)
	if err != nil {
		return nil, err
	}

	trimArgs.Ast = ast
	_, _, err = TrimAST(&trimArgs)
	if err != nil {
		return nil, err
	}

	trimmedContent := map[string]string{}

	err = recursiveDump(ast, trimmedContent)
	if err != nil {
		return nil, err
	}
	return trimmedContent, nil
}

func recursiveDump(ast *parser.Thrift, trimmedContent map[string]string) error {
	main, err := dump.DumpIDL(ast)
	if err != nil {
		return err
	}
	trimmedContent[ast.Filename] = main

	for _, inc := range ast.Includes {
		err = recursiveDump(inc.Reference, trimmedContent)
		if err != nil {
			return err
		}
	}
	return nil
}
