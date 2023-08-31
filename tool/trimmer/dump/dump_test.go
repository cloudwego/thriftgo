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

package dump

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestDumpSingle(t *testing.T) {
	filename := filepath.Join("..", "test_cases", "sample1.thrift")
	ast, err := parser.ParseFile(filename, []string{"test_cases"}, true)
	test.Assert(t, err == nil, err)
	_, err = DumpIDL(ast)
	test.Assert(t, err == nil, err)
}

func TestDumpMany(t *testing.T) {
	dir := filepath.Join("..", "test_cases")
	testDir(dir, t)
}

func testDir(dir string, t *testing.T) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return
	}
	var thriftFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".thrift") {
			filePath := filepath.Join(dir, file.Name())
			thriftFiles = append(thriftFiles, filePath)
		}
		if file.IsDir() {
			testDir(filepath.Join(dir, file.Name()), t)
		}
	}

	for _, f := range thriftFiles {
		ast, err := parser.ParseFile(f, []string{"test_cases"}, true)
		test.Assert(t, err == nil, err)
		out, err := DumpIDL(ast)
		test.Assert(t, err == nil, err)
		test.Assert(t, out != "", out)
	}
}
