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
		fp, err = os.Open("idl-ref.yaml")
		if err != nil {
			return
		}
	}
	buf := bufio.NewScanner(fp)
	isFirst := true
	for {
		if !buf.Scan() {
			break
		}
		line := strings.TrimSpace(buf.Text())
		if isFirst {
			if line != "ref:" {
				break
			}
			isFirst = false
			continue
		}
		if strings.Count(line, ":") == 1 {
			arr := strings.Split(line, ":")
			refMap[strings.TrimSpace(arr[0])] = strings.TrimSpace(arr[1])
		}
	}
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
