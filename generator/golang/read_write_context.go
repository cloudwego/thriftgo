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

	"github.com/cloudwego/thriftgo/parser"
)

// ReadWriteContext contains information for generating codes in ReadField* and
// WriteField* functions. Each context stands for a struct field, a map key, a
// map value, a list elemement, or a set elemement.
type ReadWriteContext struct {
	Type      *parser.Type
	TypeName  string // The type name in Go code
	TypeID    string // For `thrift.TProtocol.(Read|Write)${TypeID}` methods
	IsPointer bool   // Whether the target type is a pointer type in Go

	KeyCtx *ReadWriteContext // sub-context if the type is map
	ValCtx *ReadWriteContext // sub-context if the type is container

	Target   string // The target for assignment
	Source   string // The variable for right hand operand in deep-equal
	NeedDecl bool   // Whether a declaration of target is needed

	ids map[string]int // Prefix => local variable index
}

// GenID returns a local variable with the given name as prefix.
func (c *ReadWriteContext) GenID(prefix string) (name string) {
	name = prefix
	if id := c.ids[prefix]; id > 0 {
		name += fmt.Sprint(id)
	}
	c.ids[prefix]++
	return
}

// WithDecl claims that the context needs a variable declaration.
func (c *ReadWriteContext) WithDecl() *ReadWriteContext {
	c.NeedDecl = true
	return c
}

// WithTarget sets the target name.
func (c *ReadWriteContext) WithTarget(t string) *ReadWriteContext {
	c.Target = t
	return c
}

// WithSource sets the source name.
func (c *ReadWriteContext) WithSource(s string) *ReadWriteContext {
	c.Source = s
	return c
}

func (c *ReadWriteContext) asKeyCtx(cu *CodeUtils) *ReadWriteContext {
	switch c.TypeID {
	case typeids.Struct:
		c.TypeName = cu.Deref(c.TypeName)
		c.IsPointer = false
	case typeids.Binary:
		c.TypeName = "string"
	}
	return c
}

func mkRWCtx(cu *CodeUtils, s *Scope, t *parser.Type, top *ReadWriteContext) (*ReadWriteContext, error) {
	tn, err := cu.resolver.getTypeName(s, t)
	if err != nil {
		return nil, err
	}
	if t.Category.IsStructLike() {
		tn = "*" + tn
	}
	ctx := &ReadWriteContext{
		Type:      t,
		TypeName:  tn,
		TypeID:    cu.GetTypeID(t),
		IsPointer: cu.IsPointerTypeName(tn),
	}
	if top != nil {
		ctx.ids = top.ids // share the namespace for temporary variables
	} else {
		ctx.ids = make(map[string]int)
	}

	// create sub-contexts.
	if t.Category == parser.Category_Map {
		ss, tt, err := cu.GetKeyType(s, t)
		if err != nil {
			return nil, err
		}
		if ctx.KeyCtx, err = mkRWCtx(cu, ss, tt, ctx); err != nil {
			return nil, err
		}
		ctx.KeyCtx.asKeyCtx(cu)
	}

	if t.Category.IsContainerType() {
		ss, tt, err := cu.GetValType(s, t)
		if err != nil {
			return nil, err
		}
		if ctx.ValCtx, err = mkRWCtx(cu, ss, tt, ctx); err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

// MkRWCtx = MakeReadWriteContext.
func (cu *CodeUtils) MkRWCtx(s *parser.StructLike, f *parser.Field) (*ReadWriteContext, error) {
	t, err := cu.ResolveFieldName(s, f)
	if err != nil {
		return nil, fmt.Errorf("MkRWCtx ResolveFieldName%+v: %w", f, err)
	}
	tn, err := cu.ResolveFieldTypeName(f)
	if err != nil {
		return nil, fmt.Errorf("MkRWCtx ResolveFieldTypeName %+v: %w", f, err)
	}

	ctx, err := mkRWCtx(cu, cu.rootScope, f.Type, nil)
	if err != nil {
		return nil, err
	}

	// adjust for fields
	ctx.Target = "p." + t
	ctx.Source = "src"
	ctx.TypeName = tn
	ctx.IsPointer = cu.IsPointerTypeName(tn)
	return ctx, nil
}
