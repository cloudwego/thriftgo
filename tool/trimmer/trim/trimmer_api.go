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

package trim

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/tool/trimmer/dump"
)

// TrimBatchContent receives a group of thrift idl as map[path]content and mainIDLPath, and return the result as the same format.
func TrimBatchContent(mainIDLFilePath string, IDLFileContentMap map[string]string) (map[string]string, error) {
	return TrimBatchContentWithConfig(mainIDLFilePath, IDLFileContentMap, TrimASTArg{
		TrimMethods: nil,
		Preserve:    nil,
		MatchGoName: nil,
	})
}

// TrimBatchContentWithConfig does the same work with TrimBatchContent, but can extra receive a trimArgs
func TrimBatchContentWithConfig(mainIDLFilePath string, IDLFileContentMap map[string]string, trimArgs TrimASTArg) (map[string]string, error) {
	ast, err := parser.ParseBatchString(mainIDLFilePath, IDLFileContentMap, trimArgs.IncludeDirs)
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
	_, err = TrimAST(&trimArgs)
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
