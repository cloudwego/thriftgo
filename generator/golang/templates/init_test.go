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
	"log"
	"testing"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/generator/golang/templates"
)

func validate(pkg, text string, funcMap template.FuncMap) {
	name := "template-validation"
	tpl := template.New(name).Funcs(funcMap)
	template.Must(tpl.Parse(text))
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

func TestTemplateValidation(t *testing.T) {
	utils := golang.NewCodeUtils(logFunc)
	funcs := utils.BuildFuncMap()
	for _, tpl := range templates.Templates() {
		if tpl != templates.File {
			validate("", tpl, funcs)
		}
	}
	for pkg, tpls := range templates.Alternative() {
		for _, tpl := range tpls {
			if tpl != templates.File {
				validate(pkg, tpl, funcs)
			}
		}
	}
}
