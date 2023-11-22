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

var DefaultYamlFileName = "trim_config.yaml"

type YamlArguments struct {
	Methods          []string `yaml:"methods,omitempty"`
	Preserve         *bool    `yaml:"preserve,omitempty"`
	PreservedStructs []string `yaml:"preserved_structs,omitempty"`
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
	if cfg.Preserve == nil {
		t := true
		cfg.Preserve = &t
	}
	return &cfg
}
