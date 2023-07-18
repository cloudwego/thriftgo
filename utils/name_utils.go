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

package utils

import (
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

var basicMap = map[string]bool{
	"i8":     true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"double": true,
	"string": true,
	"byte":   true,
	"binary": true,
	"bool":   true,
}

var containerMap = map[string]bool{
	"set":  true,
	"list": true,
	"map":  true,
}

func IsBasic(name string) bool {
	return basicMap[name]
}

func IsContainer(name string) bool {
	return containerMap[name]
}

func GetAnnotationsAsMap(annotations parser.Annotations) map[string][]string {
	annotationsMap := map[string][]string{}
	for _, annotation := range annotations {
		annotationsMap[annotation.Key] = annotation.Values
	}
	return annotationsMap
}

func ParseAlias(tname string) (prefix, name string) {
	if strings.Contains(tname, ".") {
		arr := strings.Split(tname, ".")
		realName := arr[len(arr)-1]
		prefix = strings.TrimSuffix(tname, "."+realName)
		tname = realName
	}
	return prefix, tname
}

func ParsePrefix(filepath string) (prefix string) {
	if strings.Contains(filepath, "/") {
		arr := strings.Split(filepath, "/")
		filename := arr[len(arr)-1]
		return strings.TrimSuffix(filename, ".thrift")
	}
	return strings.TrimSuffix(filepath, ".thrift")
}
