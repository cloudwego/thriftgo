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

package golang

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudwego/thriftgo/generator/golang/styles"
)

// Features controls the behavior of CodeUtils.
type Features struct {
	MarshalEnumToText  bool `json_enum_as_text:"Generate MarshalText for enum values"`
	GenerateSetter     bool `gen_setter:"Generate Set* methods for fields"`
	GenDatabaseTag     bool `gen_db_tag:"Generate 'db:$field' tag"`
	GenOmitEmptyTag    bool `omitempty_for_optional:"Generate 'omitempty' tags for optional fields."`
	TypedefAsTypeAlias bool `use_type_alias:"Generate type alias for typedef instead of type define."`
	ValidateSet        bool `validate_set:"Generate codes to validate the uniqueness of set elements."`
	ValueTypeForSIC    bool `value_type_in_container:"Genenerate value type for struct-like in container instead of pointer type."`
	ScanValueForEnum   bool `scan_value_for_enum:"Generate Scan and Value methods for enums to implement interfaces in std sql library."`
	ReorderFields      bool `reorder_fields:"Reorder fields of structs to improve memory usage."`
	TypedEnumString    bool `typed_enum_string:"Add type prefix to the string representation of enum values."`
	KeepUnknownFields  bool `keep_unknown_fields:"Genenerate codes to store unrecognized fields in structs."`
	GenDeepEqual       bool `gen_deep_equal:"Generate DeepEqual function for struct/union/exception."`
	CompatibleNames    bool `compatible_names:"Add a '_' suffix if an name has a prefix 'New' or suffix 'Args' or 'Result'."`
	ReserveComments    bool `reserve_comments:"Reserve comments of definitions in thrift file"`
	NilSafe            bool `nil_safe:"Generate nil-safe getters."`
	FrugalTag          bool `frugal_tag:"Generate 'frugal' tags."`
	EscapeDoubleInTag  bool `unescape_double_quote:"Unescape the double quotes in literals when generating go tags."`
	GenerateTypeMeta   bool `gen_type_meta:"Generate and register type meta for structures."`
	GenerateJSONTag    bool `gen_json_tag:"Generate struct with 'json' tag"`
	SnakeTyleJSONTag   bool `snake_style_json_tag:"Generate snake style json tag"`
}

var defaultFeatures = Features{
	MarshalEnumToText:  false,
	GenerateSetter:     false,
	GenDatabaseTag:     false,
	GenOmitEmptyTag:    true,
	TypedefAsTypeAlias: true,
	ValidateSet:        true,
	ValueTypeForSIC:    false,
	ScanValueForEnum:   true,
	ReorderFields:      false,
	TypedEnumString:    false,
	KeepUnknownFields:  false,
	GenDeepEqual:       false,
	CompatibleNames:    false,
	ReserveComments:    false,
	NilSafe:            false,
	FrugalTag:          false,
	EscapeDoubleInTag:  true,
	GenerateTypeMeta:   false,
	GenerateJSONTag:    true,
	SnakeTyleJSONTag:   false,
}

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
			cu.UsePackage(DefaultThriftLib, value)
			return nil
		},
	},
	{
		name: "use_package",
		desc: "Specify an import path replacement. Form: 'path=repl', (e.g. 'database/sql/driver=example.com/my/dirver')",
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
	{
		name: "template",
		desc: "Specify a different template to generate codes. (current available templates: 'slim')",
		action: func(value string, cu *CodeUtils) error {
			return cu.UseTemplate(value)
		},
	},
}

// creates parameters by reflection.
func (fs Features) params() (ps []*param) {
	t := reflect.TypeOf(fs)
	x := reflect.ValueOf(fs)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n := strings.SplitN(string(f.Tag), ":", 2)[0]
		v := f.Tag.Get(n)
		if !x.Field(i).IsZero() {
			v += " (Enabled by default)"
		}

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
func (cu *CodeUtils) HandleOptions(args []string) error {
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
				err := p.action(value, cu)
				if err != nil {
					return err
				}
				cu.Info("option:", a)
				continue next
			}
		}
		cu.Info("unsupported option:", a)
	}
	return nil
}
