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

package templates

import (
	_ "embed"

	"github.com/cloudwego/thriftgo/generator/golang/templates/slim"
)

// Alternative returns all alternative templates.
func Alternative() map[string][]string {
	return map[string][]string{
		"slim": append(Templates(), slim.Extension()...),
	}
}

// templates.
var (
	//go:embed file.go.tmpl
	File string

	//go:embed client.go.tmpl
	Client string

	//go:embed constant.go.tmpl
	Constant string

	//go:embed deep_equal.go.tmpl
	DeepEqual string

	//go:embed enum.go.tmpl
	Enum string

	//go:embed imports.go.tmpl
	Imports string

	//go:embed processor.go.tmpl
	Processor string

	//go:embed service.go.tmpl
	Service string

	//go:embed struct.go.tmpl
	Struct string

	//go:embed typedef.go.tmpl
	Typedef string
)

// Templates returns all templates defined in this package.
func Templates() []string {
	return []string{
		Client,
		Constant,
		DeepEqual,
		Enum,
		File,
		Imports,
		Processor,
		Service,
		Struct,
		Typedef,
	}
}
