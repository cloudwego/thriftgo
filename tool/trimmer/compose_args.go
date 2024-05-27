// Copyright 2024 CloudWeGo Authors
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

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var DefaultIDLComposeFileName = "idl_compose.yaml"

type IDLComposeArguments struct {
	IDLs map[string]*IDLArguments `yaml:"idls,omitempty"`
}

type IDLArguments struct {
	Trimmer *TrimmerYamlArguments `yaml:"trimmer,omitempty"`
}

type TrimmerYamlArguments struct {
	Methods          []string `yaml:"methods,omitempty"`
	Preserve         *bool    `yaml:"preserve,omitempty"`
	PreservedStructs []string `yaml:"preserved_structs,omitempty"`
	MatchGoName      *bool    `yaml:"match_go_name,omitempty"`
}

func ParseYamlConfig(path string) *IDLComposeArguments {
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
		if idl.Trimmer.Preserve == nil {
			t := true
			idl.Trimmer.Preserve = &t
		}
		if idl.Trimmer.MatchGoName == nil {
			t := false
			idl.Trimmer.MatchGoName = &t
		}
	}

	return &cfg
}

type ComposeArguments struct {
	Config string
}

func (a *ComposeArguments) BuildFlags() *flag.FlagSet {
	f := flag.NewFlagSet("trimmer compose", flag.ContinueOnError)

	f.StringVar(&a.Config, "c", "", "")
	f.StringVar(&a.Config, "config", "", "")

	// todo: write help
	f.Usage = help
	return f
}

func (a *ComposeArguments) Parse(argv []string) error {
	f := a.BuildFlags()
	if err := f.Parse(argv[1:]); err != nil {
		return err
	}
	return nil
}

func (a *ComposeArguments) run() error {
	if a.Config == "" {
		a.Config = DefaultIDLComposeFileName
	}

	cfg := ParseYamlConfig(a.Config)
	if cfg == nil {
		// todo: more detailed information
		return errors.New("failed to parse idl_compose config")
	}
	ancestor, err := createAncestor(cfg)
	if err != nil {
		return err
	}
	// todo: deal with ancestor
	ancestor.String()
	return nil
}

const (
	ancestorFmt = `namespace go test.all

%s

%s
`
)

func createAncestor(cfg *IDLComposeArguments) (*parser.Thrift, error) {
	var svcs []string
	for path := range cfg.IDLs {
		// todo: deal with includes
		child, err := parser.ParseFile(path, nil, true)
		if err != nil {
			// todo: deal with error
			return nil, err
		}
		child.ForEachService(func(v *parser.Service) bool {
			svcs = append(svcs, v.Name)
			return true
		})
	}
	ancestorStr := fmt.Sprintf(ancestorFmt, getImports(), getFakeServices())
	return parser.ParseString("./all.thrift", ancestorStr)
}

func getImports() string {
	return ""
}

func getFakeServices() string {
	return ""
}
