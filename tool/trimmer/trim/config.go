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

package trim

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	DefaultYamlFileName       = "trim_config.yaml"
	DefaultIDLComposeFileName = "idl_compose.yaml"
)

// TrimmerYamlArguments is the alias to YamlArguments
type TrimmerYamlArguments YamlArguments

// IDLArguments contains all arguments about the IDL.
// For now, it only contains TrimmerYamlArguments.
type IDLArguments struct {
	Trimmer *TrimmerYamlArguments `yaml:"trimmer,omitempty"`
}

// IDLComposeArguments contains all IDLs and their arguments.
type IDLComposeArguments struct {
	// path is the path of IDL based on working directory
	// e.g. "idl/sample.thrift"
	IDLs map[string]*IDLArguments `yaml:"idls,omitempty"`
}

type YamlArguments struct {
	Methods          []string `yaml:"methods,omitempty"`
	Preserve         *bool    `yaml:"preserve,omitempty"`
	PreservedStructs []string `yaml:"preserved_structs,omitempty"`
	MatchGoName      *bool    `yaml:"match_go_name,omitempty"`
}

// todo: deal with this function with TrimmerYamlArguments
func (arg *YamlArguments) setDefault() {
	if arg.Preserve == nil {
		t := true
		arg.Preserve = &t
	}
	if arg.MatchGoName == nil {
		t := false
		arg.MatchGoName = &t
	}
}

func ParseYamlConfig(path string) *YamlArguments {
	cfg := YamlArguments{}
	dataBytes, err := ioutil.ReadFile(filepath.Join(path, DefaultYamlFileName))
	if err != nil {
		return nil
	}
	fmt.Println("using trim config:", filepath.Join(path, DefaultYamlFileName))
	err = yaml.Unmarshal(dataBytes, &cfg)
	if err != nil {
		fmt.Println("unmarshal yaml config fail:", err)
		return nil
	}
	cfg.setDefault()
	return &cfg
}

func (arg *TrimmerYamlArguments) setDefault() {
	if arg.Preserve == nil {
		t := true
		arg.Preserve = &t
	}
	if arg.MatchGoName == nil {
		t := false
		arg.MatchGoName = &t
	}
}

func ParseIDLComposeConfig(path string) *IDLComposeArguments {
	cfg := IDLComposeArguments{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	fmt.Println("using idl_compose config:", path)
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println("unmarshal idl_compose config fail:", err)
		return nil
	}
	// set default value
	for _, idl := range cfg.IDLs {
		if idl.Trimmer == nil {
			idl.Trimmer = &TrimmerYamlArguments{}
		}
		idl.Trimmer.setDefault()
	}

	return &cfg
}
