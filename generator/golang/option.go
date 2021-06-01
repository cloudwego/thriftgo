// Copyright 2021 CloudWeGo
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
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudwego/thriftgo/generator/golang/styles"
)

type param struct {
	name   string
	desc   string
	action func(value string, cu *CodeUtils) error
}

func (p *param) match(value string) bool {
	return strings.HasPrefix(value, p.name)
}

var codeUtilsParams = []*param{
	{
		name: "thrift_import_path",
		desc: "Override thrift package import path (default:github.com/apache/thrift/lib/go/thrift)",
		action: func(value string, cu *CodeUtils) error {
			cu.UsePackage("thrift", value)
			return nil
		},
	},
	{
		name: "use_package",
		desc: "Specify an import path for a package. Form: 'pkg=path'",
		action: func(value string, cu *CodeUtils) error {
			parts := strings.SplitN(value, "=", 2)
			if len(parts) < 2 {
				return fmt.Errorf("invalid argument for use_package: '%s'", value)
			}
			cu.UsePackage(parts[0], parts[1])
			return nil
		},
	},
	{
		name: "naming_style",
		desc: fmt.Sprintf(
			"Set the naming style for identifiers: %s. Default is 'thriftgo'.",
			strings.Join(styles.NamingStyles(), ", ")),
		action: func(value string, cu *CodeUtils) error {
			style := styles.NewNamingStyle(value)
			if style == nil {
				return fmt.Errorf("unsupported naming style: '%s'", value)
			}
			cu.SetNamingStyle(style)
			return nil
		},
	},
	{
		name: "ignore_initialisms",
		desc: "Disable spelling correction of initialisms (e.g. 'URL')",
		action: func(value string, cu *CodeUtils) error {
			ignore, err := checkBool("ignore_initialisms", value)
			if err != nil {
				return err
			}
			cu.UseInitialisms(!ignore)
			return nil
		},
	},
	{
		name: "package_prefix",
		desc: "Specify a package prefix for all generated codes.",
		action: func(value string, cu *CodeUtils) error {
			cu.SetPackagePrefix(value)
			return nil
		},
	},
}

// creates parameters by reflection.
func (fs Features) params() (ps []*param) {
	t := reflect.TypeOf(fs)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n := strings.SplitN(string(f.Tag), ":", 2)[0]
		v := f.Tag.Get(n)

		name, nth := n, i // for closure
		p := &param{
			name: n,
			desc: v,
			action: func(value string, cu *CodeUtils) error {
				val, err := checkBool(name, value)
				if err != nil {
					return err
				}
				x := cu.Features()
				field := reflect.ValueOf(&x).Elem().Field(nth)
				field.SetBool(val)
				cu.SetFeatures(x)
				return nil
			},
		}
		ps = append(ps, p)
	}
	return
}

var allParams = append(codeUtilsParams, defaultFeatures.params()...)

func checkBool(name, value string) (bool, error) {
	switch value {
	case "", "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("%s: expect a bool value or empty string, got '%s'",
		name, value)
}

// HandleOptions updates the CodeUtils with options.
func (utils *CodeUtils) HandleOptions(args []string) error {
	var name, value string
next:
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		switch len(parts) {
		case 0:
			continue
		case 1:
			name, value = parts[0], ""
		case 2:
			name, value = parts[0], parts[1]
		}

		for _, p := range allParams {
			if p.match(name) {
				err := p.action(value, utils)
				if err != nil {
					return err
				}
				utils.Info("option:", a)
				continue next
			}
		}
		utils.Info("unsupported option:", a)
	}
	return nil
}
