// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package golang

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
)

// FrugalResolver resolves type names for frugal.
type FrugalResolver struct {
	root *Scope
	util *CodeUtils
}

// NewFrugalResolver creates a new FrugalResolver with the given scope.
func NewFrugalResolver(root *Scope, cu *CodeUtils) *FrugalResolver {
	return &FrugalResolver{
		root: root,
		util: cu,
	}
}

// ResolveFrugalTypeName returns a legal type name in frugal for the given AST type.
func (r *FrugalResolver) ResolveFrugalTypeName(t *parser.Type) (TypeName, error) {
	name, err := r.getTypeName(r.root, t)
	return TypeName(name), err
}

func getUnderlay(g *Scope, t *parser.Type) (*Scope, *parser.Type, string) {
	name := t.Name
	if ref := t.GetReference(); ref != nil {
		g = g.includes[ref.Index].Scope
		name = ref.Name
	}
	if t.IsTypedef != nil && *t.IsTypedef {
		if typedef := g.Typedef(name); typedef != nil {
			return getUnderlay(g, typedef.Type)
		}
	}

	return g, t, name
}

func (r *FrugalResolver) getTypeName(g *Scope, t *parser.Type) (name string, err error) {
	if _, ok := baseTypes[t.Name]; ok {
		return t.Name, nil
	}
	g, ut, name := getUnderlay(g, t)
	t = ut
	if isContainerTypes[name] {
		return r.getContainerTypeName(g, t)
	}

	if name == "" {
		return "", fmt.Errorf("getTypeName failed: type[%v] file[%s]", t, g.ast.Filename)
	}

	if !ut.Category.IsBaseType() && !ut.Category.IsContainerType() {
		name = g.globals.Get(name)
	}

	if g.namespace != r.root.namespace && ut.Category.IsStructLike() {
		pkg, _ := r.util.Import(g.ast)
		name = pkg + "." + name
	}
	return
}

func (r *FrugalResolver) getContainerTypeName(g *Scope, t *parser.Type) (name string, err error) {
	if t.Name == "map" {
		var k string
		if t.KeyType.Category == parser.Category_Binary {
			k = "string" // 'binary => string' for key type in map
		} else {
			k, err = r.getTypeName(g, t.KeyType)
			if err != nil {
				return "", fmt.Errorf("resolve key type of '%s' failed: %w", t, err)
			}
		}
		v, err := r.getTypeName(g, t.ValueType)
		if err != nil {
			return "", fmt.Errorf("resolve value type of '%s' failed: %w", t, err)
		}
		name = fmt.Sprintf("map<%s:%s>", k, v)
	} else {
		v, err := r.getTypeName(g, t.ValueType)
		if err != nil {
			return "", fmt.Errorf("resolve value type of '%s' failed: %w", t, err)
		}
		if t.Name == "list" {
			name = "list<" + v + ">"
		} else {
			name = "set<" + v + ">"
		}
	}

	return name, nil
}
