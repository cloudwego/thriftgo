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

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/thriftgo/utils/dir_utils"

	"gopkg.in/yaml.v3"
)

func init() {
	err := LoadConfig()
	if err != nil {
		panic(err)
	}
}

type RawConfig struct {
	Ref   map[string]interface{} `yaml:"ref"`
	Debug bool                   `yaml:"debug"`
}

type Config struct {
	Ref   map[string]*RefConfig `yaml:"ref"`
	Debug bool                  `yaml:"debug"`
}

type RefConfig struct {
	Path       string   `yaml:"path"`
	Structs    []string `yaml:"structs,omitempty"`
	Enums      []string `yaml:"enums,omitempty"`
	Typedefs   []string `yaml:"typedefs,omitempty"`
	Consts     []string `yaml:"consts,omitempty"`
	Unions     []string `yaml:"unions,omitempty"`
	Exceptions []string `yaml:"exceptions,omitempty"`
}

func (r *RefConfig) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Path: %s\n", r.Path))
	if len(r.Structs) > 0 {
		sb.WriteString(fmt.Sprintf("Structs: %v\n", r.Structs))
	}
	if len(r.Enums) > 0 {
		sb.WriteString(fmt.Sprintf("Enums: %v\n", r.Enums))
	}
	if len(r.Typedefs) > 0 {
		sb.WriteString(fmt.Sprintf("Typedefs: %v\n", r.Typedefs))
	}
	if len(r.Consts) > 0 {
		sb.WriteString(fmt.Sprintf("Consts: %v\n", r.Consts))
	}
	if len(r.Unions) > 0 {
		sb.WriteString(fmt.Sprintf("Unions: %v\n", r.Unions))
	}
	if len(r.Exceptions) > 0 {
		sb.WriteString(fmt.Sprintf("Exceptions: %v\n", r.Exceptions))
	}
	return sb.String()
}

func (r *RefConfig) IsAllFieldsEmpty() bool {
	return len(r.Structs) == 0 &&
		len(r.Enums) == 0 &&
		len(r.Typedefs) == 0 &&
		len(r.Consts) == 0 &&
		len(r.Unions) == 0 &&
		len(r.Exceptions) == 0
}

var globalConfig *Config
var useAbs bool = true

func GetRef(name string) *RefConfig {
	if globalConfig == nil {
		return nil
	}
	if useAbs {
		name, _ = filepath.Abs(name)
	}
	refConfig, ok := globalConfig.Ref[name]
	if globalConfig.Debug {
		if ok {
			fmt.Printf("[idl-ref-get]Successfully Get: %s\n", name)
		} else {
			fmt.Printf("[idl-ref-get]Not IDL Ref: %s\n", name)
		}
	}
	return refConfig
}

// LoadConfig by default, config will load only once when the program is invoked, also the same for each plugin
// but for sdk mode, config should be reloaded each time when the sdk is called. so we provide this api and manually call this in sdk mode.
func LoadConfig() error {
	config, err := initConfig()
	if err != nil {
		return errors.New("failed to parse idl ref config: " + err.Error())
	}
	globalConfig = config
	return nil
}

func initConfig() (*Config, error) {
	configPaths := []string{"idl-ref.yml", "idl-ref.yaml"}
	for _, path := range configPaths {
		if dir_utils.HasGlobalWd() {
			dirpath, err := dir_utils.Getwd()
			if err != nil {
				return nil, err
			}
			path = filepath.Join(dirpath, path)
		}
		_, err := os.Stat(path)
		if err == nil {
			return loadConfig(path)
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, nil
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var rawConfig RawConfig
	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, err
	}

	var config Config
	config.Debug = rawConfig.Debug
	config.Ref = map[string]*RefConfig{}
	for k, v := range rawConfig.Ref {
		var rc RefConfig
		// if use absolute path to match idl-ref path and current file path
		// convert the idl path to absoulute path
		if useAbs {
			k, err = dir_utils.ToAbsolute(k)
			if err != nil {
				return nil, err
			}
		}
		if config.Debug {
			fmt.Printf("[idl-ref-register]Path: %s\n", k)
		}
		switch val := v.(type) {
		case map[string]interface{}:
			err = mapToStruct(val, &rc)
		case string:
			rc.Path = val
		default:
			return nil, errors.New("failed to parse yaml")
		}
		config.Ref[k] = &rc
	}
	return &config, nil
}

func mapToStruct(m map[string]interface{}, s interface{}) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return nil
}
