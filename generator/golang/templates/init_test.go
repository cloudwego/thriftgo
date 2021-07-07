// Copyright 2021 CloudWeGo Authors
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

package templates_test

import (
	"fmt"
	"log"
	"testing"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/generator/golang/templates"
)

// forceSingleDefinition panics if the given template contains multiple definition.
// It is used to keep maintainability of templates.
// The result is the name defined in the given template.
func forceSingleDefinition(text string, funcMap template.FuncMap) string {
	name := "force-single-definition"
	tpl := template.New(name).Funcs(funcMap)
	tpl = template.Must(tpl.Parse(text))
	if tpls := tpl.Templates(); len(tpls) != 2 {
		err := fmt.Errorf(
			"templates must have only one definition: `\n----%s----\n`",
			text)
		panic(err)
	} else {
		if tpls[0].Name() == name {
			return tpls[1].Name()
		}
		return tpls[0].Name()
	}
}

var logFunc backend.LogFunc

func init() {
	logFunc.Info = func(v ...interface{}) {
		log.Print(v...)
	}
	logFunc.Warn = func(v ...interface{}) {
		log.Print(v...)
	}
	logFunc.MultiWarn = func(warns []string) {
		log.Print(warns)
	}
}

func TestDefinitionNumber(t *testing.T) {
	utils := golang.NewCodeUtils(logFunc)
	funcs := utils.BuildFuncMap()
	for _, tpl := range templates.Templates() {
		if tpl != templates.File {
			forceSingleDefinition(tpl, funcs)
		}
	}
}
