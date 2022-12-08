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

package golang

import (
	"bufio"
	"os"
	"strings"

	"github.com/cloudwego/thriftgo/generator/golang/templates/ref"
)

var refMap = map[string]string{}

func init() {
	fp, err := os.Open("idl-ref.yml")
	if err != nil {
		return
	}
	readFile(fp)
}

func readFile(fp *os.File) {
	buf := bufio.NewScanner(fp)
	isFirst := true
	for {
		if !buf.Scan() {
			break
		}
		line := strings.TrimSpace(buf.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if isFirst {
			if line != "ref:" {
				break
			}
			isFirst = false
			continue
		}
		if strings.Count(line, ":") == 1 {
			arr := strings.Split(line, ":")
			refMap[trimInput(arr[0])] = trimInput(arr[1])
		}
	}
}

func trimInput(input string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(input, `'`, ""), `"`, ""))
}

func DoRef(path string) (bool, string) {
	if refpath := refMap[path]; refpath != "" {
		return true, refpath
	}
	return false, ""
}

func TemplatesRef() []string {
	return []string{
		ref.FileRef,
	}
}
