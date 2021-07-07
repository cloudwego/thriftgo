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
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang/styles"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/namespace"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
)

// Default libraries.
const (
	DefaultThriftLib  = "github.com/apache/thrift/lib/go/thrift"
	DefaultUnknownLib = "github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
)

// Errors.
var (
	ErrRootScopeNotSet = errors.New("root scope is not set")
)

// CodeUtils contains a set of functions used in the template execution.
type CodeUtils struct {
	backend.LogFunc
	packagePrefix string            // Package prefix for all generated codes.
	customized    map[string]string // Customized imports, package => import path.
	features      Features          // Available features.
	namingStyle   styles.Naming     // Naming style.
	doInitialisms bool              // Make initialisms setting kept event naming style changes.
	libNotUsed    map[string]bool

	resolver  *resolver
	rootScope *Scope

	scopeCache map[*parser.Thrift]*Scope
}

// NewCodeUtils creates a new CodeUtils.
func NewCodeUtils(log backend.LogFunc) *CodeUtils {
	cu := &CodeUtils{
		LogFunc:     log,
		customized:  make(map[string]string),
		features:    defaultFeatures,
		namingStyle: styles.NewNamingStyle("thriftgo"),
		scopeCache:  make(map[*parser.Thrift]*Scope),
	}
	return cu
}

// GetPackagePrefix sets the package prefix in generated codes.
func (cu *CodeUtils) GetPackagePrefix() (pp string) {
	return cu.packagePrefix
}

// SetPackagePrefix sets the package prefix in generated codes.
func (cu *CodeUtils) SetPackagePrefix(pp string) {
	cu.packagePrefix = pp
}

// UsePackage forces the generated codes to use the specific package.
func (cu *CodeUtils) UsePackage(pkg, path string) {
	cu.customized[pkg] = path
}

// NamingStyle returns the current naming style.
func (cu *CodeUtils) NamingStyle() styles.Naming {
	return cu.namingStyle
}

// SetNamingStyle sets the naming style.
func (cu *CodeUtils) SetNamingStyle(style styles.Naming) {
	cu.namingStyle = style
	cu.namingStyle.UseInitialisms(cu.doInitialisms)
}

// UseInitialisms sets the naming style's initialisms option.
func (cu *CodeUtils) UseInitialisms(enable bool) {
	cu.doInitialisms = enable
	cu.namingStyle.UseInitialisms(cu.doInitialisms)
}

// SetFeatures sets the feature set.
func (cu *CodeUtils) SetFeatures(fs Features) {
	cu.features = fs
}

// Features returns the current settings of generator features.
func (cu *CodeUtils) Features() Features {
	return cu.features
}

// SetRootScope sets the root scope.
func (cu *CodeUtils) SetRootScope(s *Scope) {
	cu.rootScope = s
	cu.resolver = &resolver{util: cu, root: s}

	// The usage of these three libraries depends on specific code generation
	// and feature settings which requires tedisous type cheking before rendering.
	// So we register them into the namespace by addImports and mark them not-used,
	// use the UseStdLibrary to confirm if they are actually used.
	cu.libNotUsed = map[string]bool{
		"strings": true,
		"bytes":   true,
		"reflect": true,
	}

	if cu.Features().ReorderFields {
		for _, x := range s.ast.GetStructLike() {
			diff := reorderFields(cu, x)
			if diff != nil && diff.original != diff.arranged {
				cu.Info(fmt.Sprintf("<reorder>(%s) %s: %d -> %d: %.2f%%",
					s.ast.Filename, x.Name, diff.original, diff.arranged, diff.percent()))
			}
		}
	}
}

// BuildFuncMap builds a function map for templates.
func (cu *CodeUtils) BuildFuncMap() template.FuncMap {
	m := map[string]interface{}{
		"ToUpper":        strings.ToUpper,
		"ToLower":        strings.ToLower,
		"InsertionPoint": plugin.InsertionPoint,
		"Pair": func(a, b interface{}) *pair {
			return &pair{First: a, Second: b}
		},
	}
	v := reflect.ValueOf(cu)
	t := v.Type()
	for i := 0; i < v.NumMethod(); i++ {
		f := t.Method(i)
		g := v.Method(i)
		if f.Type.NumOut() != 1 && f.Type.NumOut() != 2 || !cu.IsExported(f.Name) {
			continue
		}
		m[f.Name] = g.Interface()
	}
	delete(m, "HandleOptions")
	delete(m, "ResolveImports")
	delete(m, "BuildFuncMap")
	delete(m, "SetRootScope")
	return m
}

// Debug prints the given values with println.
func (cu *CodeUtils) Debug(vs ...interface{}) string {
	var ss []string
	for _, v := range vs {
		ss = append(ss, fmt.Sprintf("%T(%+v)", v, v))
	}
	println("[DEBUG]", strings.Join(ss, " "))
	return ""
}

// ParseNamespace retrieves informations from the given AST and returns a
// reference name in the IDL, a package name for generated codes and an import path.
func (cu *CodeUtils) ParseNamespace(ast *parser.Thrift) (ref, pkg, pth string) {
	ref = filepath.Base(ast.Filename)
	ref = strings.TrimSuffix(ref, filepath.Ext(ref))
	ns := ast.GetNamespaceOrReferenceName("go")
	pkg = cu.NamespaceToPackage(ns)
	pth = cu.NamespaceToImportPath(ns)
	return
}

// BuildScope creates a scope of the AST with its includes processed recursively.
func (cu *CodeUtils) BuildScope(ast *parser.Thrift) (*Scope, error) {
	if scope, ok := cu.scopeCache[ast]; ok {
		return scope, nil
	}
	scope := newScope(ast)
	err := scope.init(cu)
	if err != nil {
		return nil, fmt.Errorf("process '%s' failed: %w", ast.Filename, err)
	}
	cu.scopeCache[ast] = scope
	return scope, nil
}

// ResolveImports returns a map of import path to alias built from the include list
// of the IDL. An alias may be an empty string to indicate no alias is need for the
// import path.
func (cu *CodeUtils) ResolveImports() (map[string]string, error) {
	if cu.rootScope == nil {
		return nil, ErrRootScopeNotSet
	}

	imports := make(map[string]string)
	cu.rootScope.imports.Iterate(func(alias, path string) bool {
		if cu.libNotUsed[alias] {
			return true // skip
		}
		if alias == path || strings.HasSuffix(path, "/"+alias) {
			imports[path] = ""
		} else {
			imports[path] = alias
		}
		return true
	})
	return imports, nil
}

// GetDefaultValueTypeName returns a type name suitable for the default value of the given field.
func (cu *CodeUtils) GetDefaultValueTypeName(f *parser.Field) (string, error) {
	t, err := cu.ResolveFieldTypeName(f)
	if err != nil {
		return "", err
	}
	if cu.IsBaseType(f.Type) {
		t = cu.Deref(t)
	}
	return t, nil
}

// GetFieldInit returns the initialization code for a field.
// The given field must have a default value.
func (cu *CodeUtils) GetFieldInit(f *parser.Field) (string, error) {
	return cu.GetConstInit(f.Name, f.Type, f.Default)
}

// GetConstInit returns the initialization code for a constant.
func (cu *CodeUtils) GetConstInit(name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	return cu.resolver.resolveConst(cu.rootScope, name, t, v)
}

// Identify converts an raw name from IDL into an exported identifier in go.
// This function accept an optional scope argument to specify a sub-namespace
// instead of using the global namespace.
func (cu *CodeUtils) Identify(raw string, scope ...interface{}) (string, error) {
	switch len(scope) {
	case 0:
		return cu.rootScope.globals.Get(raw), nil
	case 1:
		ns := cu.rootScope.namespaces[scope[0]]
		if ns == nil {
			return "", fmt.Errorf("%+v is not defined in %q", scope[0], cu.rootScope.ast.Filename)
		}
		return ns.Get(raw), nil
	default:
		return "", fmt.Errorf("invalid scope count: %d", len(scope))
	}
}

// identify0 converts an raw name from IDL into an exported identifier in go.
func (cu *CodeUtils) identify0(name string) (s string, err error) {
	s = strings.TrimPrefix(name, prefix)
	s, err = cu.namingStyle.Identify(s)
	if err != nil {
		return "", err
	}
	return s, nil
}

// GetParamName returns a valid name for a parameter.
func (cu *CodeUtils) GetParamName(f *parser.Function, p string) (string, error) {
	ns := cu.rootScope.namespaces[f]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", f, cu.rootScope.ast.Filename)
	}
	return ns.Get(p), nil
}

// ResolveFieldName returns the resolved name of the given field.
func (cu *CodeUtils) ResolveFieldName(s *parser.StructLike, f *parser.Field) (string, error) {
	ns := cu.rootScope.namespaces[s]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", s, cu.rootScope.ast.Filename)
	}
	return ns.Get(f.Name), nil
}

// GetFieldGetterName returns a name of the getter function for the given field.
func (cu *CodeUtils) GetFieldGetterName(s *parser.StructLike, f *parser.Field) (string, error) {
	ns := cu.rootScope.namespaces[s]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", s, cu.rootScope.ast.Filename)
	}
	return ns.Get(_p("get:" + f.Name)), nil
}

// GetFieldSetterName returns a name of the setter function for the given field.
func (cu *CodeUtils) GetFieldSetterName(s *parser.StructLike, f *parser.Field) (string, error) {
	ns := cu.rootScope.namespaces[s]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", s, cu.rootScope.ast.Filename)
	}
	return ns.Get(_p("set:" + f.Name)), nil
}

// GetFieldIsSetName returns a name of the isset function for the given field.
func (cu *CodeUtils) GetFieldIsSetName(s *parser.StructLike, f *parser.Field) (string, error) {
	ns := cu.rootScope.namespaces[s]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", s, cu.rootScope.ast.Filename)
	}
	return ns.Get(_p("isset:" + f.Name)), nil
}

// MakeEnumValueName returns a legal identifier for a enum value.
func (cu *CodeUtils) MakeEnumValueName(e *parser.Enum, v *parser.EnumValue) (string, error) {
	ns := cu.rootScope.namespaces[e]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", e, cu.rootScope.ast.Filename)
	}
	return ns.Get(v.Name), nil
}

// GetEnumValueLiteral returns the literal representation of the given enum value.
func (cu *CodeUtils) GetEnumValueLiteral(e *parser.Enum, v *parser.EnumValue) (string, error) {
	ns := cu.rootScope.namespaces[e]
	if ns == nil {
		return "", fmt.Errorf("%+v not defined in %q", e, cu.rootScope.ast.Filename)
	}
	if cu.Features().TypedEnumString {
		return ns.Get(v.Name), nil
	}
	return v.Name, nil
}

// ResolveFieldTypeName returns a legal type name in go for the given field.
func (cu *CodeUtils) ResolveFieldTypeName(f *parser.Field) (string, error) {
	tn, err := cu.resolver.getTypeName(cu.rootScope, f.Type)
	if err != nil {
		return "", err
	}

	if cu.NeedRedirect(f) {
		return "*" + cu.Deref(tn), nil
	}
	return tn, nil
}

// ResolveTypeName returns a legal type name in go for the given AST type.
func (cu *CodeUtils) ResolveTypeName(t *parser.Type) (string, error) {
	tn, err := cu.resolver.getTypeName(cu.rootScope, t)
	if err != nil {
		return "", err
	}
	if t.Category.IsStructLike() {
		return "*" + tn, nil
	}
	return tn, nil
}

// BaseServicePrefix returns the package prefix of the base of the given service.
// If the given service has no base service or the base service is in current IDL,
// then the package prefix will be "".
func (cu *CodeUtils) BaseServicePrefix(t *parser.Service) (string, error) {
	if t.Extends != "" {
		if cu.rootScope == nil {
			return "", ErrRootScopeNotSet
		}

		if ref := t.GetReference(); ref != nil {
			return cu.rootScope.packages[ref.GetIndex()] + ".", nil
		}
	}
	return "", nil
}

// GetArgType returns a parser.StructLike for a given function's parameters.
func (cu *CodeUtils) GetArgType(svc *parser.Service, f *parser.Function) (*parser.StructLike, error) {
	syn := cu.rootScope.synthesized[svc][f]
	if syn == nil {
		return nil, fmt.Errorf("no arg type for %q.%q", svc.Name, f.Name)
	}
	return syn.ArgType, nil
}

// GetResType returns a parser.StructLike for a given function's result type.
func (cu *CodeUtils) GetResType(svc *parser.Service, f *parser.Function) (*parser.StructLike, error) {
	syn := cu.rootScope.synthesized[svc][f]
	if syn == nil {
		return nil, fmt.Errorf("no res type for %q.%q", svc.Name, f.Name)
	}
	return syn.ResType, nil
}

// GetKeyType returns the key type of the given type. T must be a map type.
func (cu *CodeUtils) GetKeyType(s *Scope, t *parser.Type) (*Scope, *parser.Type, error) {
	if t.Category != parser.Category_Map {
		return nil, nil, fmt.Errorf("expect map type, got: '%s'", t)
	}
	g, x, err := semantic.Deref(s.ast, t)
	if err != nil {
		return nil, nil, err
	}
	return cu.scopeCache[g], x.KeyType, nil
}

// GetValType returns the value type of the given type. T must be a container type.
func (cu *CodeUtils) GetValType(s *Scope, t *parser.Type) (*Scope, *parser.Type, error) {
	if !t.Category.IsContainerType() {
		return nil, nil, fmt.Errorf("expect container type, got: '%s'", t)
	}
	g, x, err := semantic.Deref(s.ast, t)
	if err != nil {
		return nil, nil, err
	}
	return cu.scopeCache[g], x.ValueType, nil
}

// MakeTemplateData returns an object that contains essential information
// for rendering the templates.
func (cu *CodeUtils) MakeTemplateData(ast *parser.Thrift) (interface{}, error) {
	scope, err := cu.BuildScope(ast)
	if err != nil {
		return nil, err
	}
	cu.SetRootScope(scope)

	return ast, nil
}

func (cu *CodeUtils) addImports(ns namespace.Namespace, ast *parser.Thrift) {
	for pkg, path := range cu.customized {
		ns.Add(pkg, path)
	}

	if len(ast.Enums) > 0 {
		ns.Add("fmt", "fmt")
		if cu.Features().ScanValueForEnum {
			ns.Add("driver", "database/sql/driver")
			ns.Add("sql", "database/sql")
		}
	}

	if len(ast.GetStructLike()) > 0 {
		ns.Add("fmt", "fmt")
		ns.Add("thrift", DefaultThriftLib)
	}

	if len(ast.Services) > 0 {
		ns.Add("thrift", DefaultThriftLib)
		for _, svc := range ast.Services {
			if svc.Extends == "" || len(svc.Functions) > 0 {
				ns.Add("context", "context")
			}
			if len(svc.Functions) > 0 {
				ns.Add("fmt", "fmt")
			}
		}
	}

	structCount := len(ast.GetStructLike())
	ast.ForEachService(func(svc *parser.Service) bool {
		structCount += len(svc.Functions)
		return true
	})
	if structCount > 0 && cu.Features().KeepUnknownFields {
		ns.Add("unknown", DefaultUnknownLib)
	}

	if cu.Features().GenDeepEqual {
		ns.Add("strings", "strings")
		ns.Add("bytes", "bytes")
	} else if cu.Features().ValidateSet {
		ns.Add("reflect", "reflect")
	}
}

// UseStdLibrary claims to use a certain standard library.
// This function is designed to be called during template rendering to
// avoid tedious type checking for determine whether a library will be used.
func (cu *CodeUtils) UseStdLibrary(lib string) string {
	delete(cu.libNotUsed, lib)
	return ""
}
