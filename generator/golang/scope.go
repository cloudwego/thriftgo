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
	"strings"

	"github.com/cloudwego/thriftgo/generator/golang/common"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/namespace"
)

// A prefix to denote synthesized identifiers.
const prefix = "$"

func _p(id string) string {
	return prefix + id
}

type synthesized struct {
	ArgType *parser.StructLike
	ResType *parser.StructLike
}

// Scope contains the type symbols defined in a thrift IDL and references of its pkg2paths.
type Scope struct {
	ast       *parser.Thrift
	namespace string

	packages []string            // package name for each include
	includes []*Scope            // scopes of includes
	imports  namespace.Namespace // a namespace to solve package name collision

	globals     namespace.Namespace
	namespaces  map[interface{}]namespace.Namespace // AST node to namespace
	synthesized map[*parser.Service]map[*parser.Function]*synthesized
}

// newScope creates an uninitialized scope from the given IDL.
func newScope(ast *parser.Thrift) *Scope {
	return &Scope{
		ast:       ast,
		namespace: ast.GetNamespaceOrReferenceName("go"),
		imports: namespace.NewNamespace(func(name string, cnt int) string {
			return fmt.Sprintf("%s%d", name, cnt-1) // zero-index
		}),
		globals:     namespace.NewNamespace(namespace.UnderscoreSuffix),
		namespaces:  make(map[interface{}]namespace.Namespace),
		synthesized: make(map[*parser.Service]map[*parser.Function]*synthesized),
	}
}

func (s *Scope) init(cu *CodeUtils) (err error) {
	cu.addImports(s.imports, s.ast)
	s.buildIncludes(cu)
	s.installNames(cu)
	return nil
}

func (s *Scope) buildIncludes(cu *CodeUtils) {
	// the indices of includes must be kept because parser.Reference.Index counts the unused IDLs.
	cnt := len(s.ast.Includes)
	s.includes = make([]*Scope, cnt)
	s.packages = make([]string, cnt)

	for idx, inc := range s.ast.Includes {
		if !inc.GetUsed() {
			continue
		}
		scope, pkg := s.include(cu, inc.Reference)
		s.includes[idx] = scope
		s.packages[idx] = pkg
	}
}

func (s *Scope) include(cu *CodeUtils, t *parser.Thrift) (*Scope, string) {
	scope, err := cu.BuildScope(t)
	if err != nil {
		panic(err)
	}

	pkg, pth := cu.getImport(t)
	pkg = s.imports.Add(pkg, pth)
	return scope, pkg
}

// includeIDL adds an probably new IDL to the include list.
func (s *Scope) includeIDL(cu *CodeUtils, t *parser.Thrift) (pkgName string) {
	_, pth := cu.getImport(t)
	if pkgName = s.imports.Get(pth); pkgName != "" {
		return
	}
	scope, pkg := s.include(cu, t)
	s.includes = append(s.includes, scope)
	s.packages = append(s.packages, pkg)
	return pkg
}

func (s *Scope) addNamespace(obj interface{}) namespace.Namespace {
	ns := namespace.NewNamespace(namespace.UnderscoreSuffix)
	s.namespaces[obj] = ns
	return ns
}

func (s *Scope) installNames(cu *CodeUtils) {
	for _, v := range s.ast.Services {
		s.installNamesForService(cu, v)
	}
	for _, v := range s.ast.GetStructLike() {
		s.installNamesForStructLike(cu, v)
	}
	for _, v := range s.ast.Typedefs {
		s.installNamesForTypedef(cu, v)
	}
	for _, v := range s.ast.Enums {
		s.installNamesForEnum(cu, v)
	}
	for _, v := range s.ast.Constants {
		cn := s.identify(cu, v.Name)
		s.globals.Add(cn, v.Name)
	}
}

func (s *Scope) identify(cu *CodeUtils, raw string) string {
	name, err := cu.identify0(raw)
	if err != nil {
		panic(err)
	}
	if !strings.HasPrefix(raw, prefix) && cu.Features().CompatibleNames {
		if strings.HasPrefix(name, "New") || strings.HasSuffix(name, "Args") || strings.HasSuffix(name, "Result") {
			name += "_"
		}
	}
	return name
}

func (s *Scope) buildSynthesized(v *parser.Service, f *parser.Function) (syn *synthesized, err error) {
	if scope := s.namespaces[v]; scope == nil {
		err = fmt.Errorf("service %+v not defined in %q", v, s.ast.Filename)
		return
	}
	syn = &synthesized{}
	syn.ArgType = &parser.StructLike{
		Category: "struct",
		Name:     f.Name + "_args",
		Fields:   f.Arguments,
	}

	if !f.Oneway {
		syn.ResType = &parser.StructLike{
			Category: "struct",
			Name:     f.Name + "_result",
		}
		if !f.Void {
			syn.ResType.Fields = append(syn.ResType.Fields, &parser.Field{
				ID:           0,
				Name:         "success",
				Requiredness: parser.FieldType_Optional,
				Type:         f.FunctionType,
			})
		}
		syn.ResType.Fields = append(syn.ResType.Fields, f.Throws...)
	}
	return
}

func (s *Scope) installNamesForService(cu *CodeUtils, v *parser.Service) {
	ns := s.addNamespace(v) // namespace of the Service

	// service name
	sn := s.identify(cu, v.Name)
	sn = s.globals.Add(sn, v.Name)
	ss := make(map[*parser.Function]*synthesized)

	// function names
	for _, f := range v.Functions {
		fn := s.identify(cu, f.Name)
		ns.Add(fn, f.Name)
	}

	// install names for argument types and response types
	for _, f := range v.Functions {
		syn, err := s.buildSynthesized(v, f)
		if err != nil {
			panic(err)
		}
		an, rn := v.Name+s.identify(cu, _p(f.Name+"_args")), v.Name+s.identify(cu, _p(f.Name+"_result"))
		s.installNamesForStructLike(cu, syn.ArgType, _p(an))
		if !f.Oneway {
			s.installNamesForStructLike(cu, syn.ResType, _p(rn))
		}
		ss[f] = syn

		s.installNamesForFunction(cu, f)
	}
	if len(v.Functions) > 0 {
		s.synthesized[v] = ss
	}

	// install names for client and processor
	cn := sn + "Client"
	pn := sn + "Processor"
	s.globals.MustReserve(cn, _p("client:"+v.Name))
	s.globals.MustReserve(pn, _p("processor:"+v.Name))
}

// installNamesForFunction builds a namespace for parameters of a Function.
// This function is used to resolve conflicts between parameter, receiver and local variables in generated method.
// Template 'Service' and 'FunctionSignature' depend on this function.
func (s *Scope) installNamesForFunction(cu *CodeUtils, v *parser.Function) {
	ns := s.addNamespace(v) // namespace of the Function

	ns.MustReserve("p", _p("p"))     // the receiver of method
	ns.MustReserve("err", _p("err")) // error
	ns.MustReserve("ctx", _p("ctx")) // first parameter

	if !v.Void {
		ns.MustReserve("r", _p("r"))             // response
		ns.MustReserve("_result", _p("_result")) // a local variable
	}

	for _, a := range v.Arguments {
		name := common.LowerFirstRune(s.identify(cu, a.Name))
		if isKeywords[name] {
			name = "_" + name
		}
		ns.Add(name, a.Name)
	}
}

func (s *Scope) installNamesForTypedef(cu *CodeUtils, t *parser.Typedef) {
	tn := s.identify(cu, t.Alias)
	tn = s.globals.Add(tn, t.Alias)
	if t.Type.Category.IsStructLike() {
		fn := "New" + tn
		s.globals.MustReserve(fn, _p("new:"+t.Alias))
	}
}

func (s *Scope) installNamesForEnum(cu *CodeUtils, e *parser.Enum) {
	en := s.identify(cu, e.Name)
	en = s.globals.Add(en, e.Name)

	ns := s.addNamespace(e)
	for _, v := range e.Values {
		// vn := s.identify(cu, v.Name)
		// ns.add(en+"_"+vn, v.Name)
		ns.Add(en+"_"+v.Name, v.Name)
	}
}

func (s *Scope) installNamesForStructLike(cu *CodeUtils, v *parser.StructLike, usedName ...string) {
	nn := v.Name
	if len(usedName) != 0 {
		nn = usedName[0]
	}
	sn := s.identify(cu, nn)
	sn = s.globals.Add(sn, v.Name)
	s.globals.MustReserve("New"+sn, _p("new:"+nn))

	fids := "fieldIDToName_" + sn
	s.globals.MustReserve(fids, _p("ids:"+nn))

	ns := s.addNamespace(v) // namespace of the Struct-like

	// built-in methods
	funcs := []string{"Read", "Write", "String"}
	if !strings.HasPrefix(v.Name, prefix) {
		if v.Category == "union" {
			funcs = append(funcs, "CountSetFields")
		}
		if v.Category == "exception" {
			funcs = append(funcs, "Error")
		}
		if cu.Features().KeepUnknownFields {
			funcs = append(funcs, "CarryingUnknownFields")
		}
		if cu.Features().GenDeepEqual {
			funcs = append(funcs, "DeepEqual")
		}
	}
	for _, fn := range funcs {
		ns.MustReserve(fn, _p(fn))
	}

	// reserve method names
	for _, f := range v.Fields {
		fn := s.identify(cu, f.Name)
		ns.Add("Get"+fn, _p("get:"+f.Name))
		if cu.Features().GenerateSetter {
			ns.Add("Set"+fn, _p("set:"+f.Name))
		}
		if cu.SupportIsSet(f) {
			ns.Add("IsSet"+fn, _p("isset:"+f.Name))
		}
	}

	// field names
	for _, f := range v.Fields {
		fn := s.identify(cu, f.Name)
		fn = ns.Add(fn, f.Name)
	}
}
