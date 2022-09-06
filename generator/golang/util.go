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
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang/common"
	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	"github.com/cloudwego/thriftgo/generator/golang/styles"
	"github.com/cloudwego/thriftgo/generator/golang/templates"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
)

// Default libraries.
const (
	DefaultThriftLib  = "github.com/apache/thrift/lib/go/thrift"
	DefaultUnknownLib = "github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
	DefaultMetaLib    = "github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	defaultTemplate   = "default"
)

var escape = regexp.MustCompile(`\\.`)

// CodeUtils contains a set of utility functions.
type CodeUtils struct {
	backend.LogFunc
	packagePrefix string            // Package prefix for all generated codes.
	importReplace map[string]string // Customized imports, import path => replacement.
	features      Features          // Available features.
	namingStyle   styles.Naming     // Naming style.
	doInitialisms bool              // Make initialisms setting kept event naming style changes.

	rootScope   *Scope
	scopeCache  map[*parser.Thrift]*Scope
	useTemplate string
	alternative map[string][]string
}

// NewCodeUtils creates a new CodeUtils.
func NewCodeUtils(log backend.LogFunc) *CodeUtils {
	cu := &CodeUtils{
		LogFunc:       log,
		importReplace: make(map[string]string),
		features:      defaultFeatures,
		namingStyle:   styles.NewNamingStyle("thriftgo"),
		scopeCache:    make(map[*parser.Thrift]*Scope),
		useTemplate:   defaultTemplate,
		alternative:   templates.Alternative(),
	}
	return cu
}

// SetFeatures sets the feature set.
func (cu *CodeUtils) SetFeatures(fs Features) {
	cu.features = fs
}

// Features returns the current settings of generator features.
func (cu *CodeUtils) Features() Features {
	return cu.features
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
func (cu *CodeUtils) UsePackage(path, repl string) {
	cu.importReplace[path] = repl
}

// Template returns the current template name. Empty for the default.
func (cu *CodeUtils) Template() string {
	return cu.useTemplate
}

// UseTemplate specifies a different template to generate codes.
func (cu *CodeUtils) UseTemplate(value string) error {
	if value != defaultTemplate && cu.alternative[value] == nil {
		return fmt.Errorf("unknown template name: %q", value)
	}
	cu.useTemplate = value
	return nil
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

// Identify converts an raw name from IDL into an exported identifier in go.
func (cu *CodeUtils) Identify(name string) (s string, err error) {
	s = strings.TrimPrefix(name, prefix)
	s, err = cu.namingStyle.Identify(s)
	if err != nil {
		return "", err
	}
	return s, nil
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

// GetFilePath returns a path to the generated file for the given IDL.
// Note that the result is a path relative to the root output path.
func (cu *CodeUtils) GetFilePath(t *parser.Thrift) string {
	ref, _, pth := cu.ParseNamespace(t)
	full := filepath.Join(pth, ref+".go")
	if strings.HasSuffix(full, "_test.go") {
		full = strings.ReplaceAll(full, "_test.go", "_test_.go")
	}
	return full
}

// GetPackageName returns a go package name for the given thrift AST.
func (cu *CodeUtils) GetPackageName(ast *parser.Thrift) string {
	namespace := ast.GetNamespaceOrReferenceName("go")
	return cu.NamespaceToPackage(namespace)
}

// NamespaceToPackage converts a namespace to a package.
func (cu *CodeUtils) NamespaceToPackage(ns string) string {
	parts := strings.Split(ns, ".")
	return strings.ToLower(parts[len(parts)-1])
}

// NamespaceToImportPath returns an import path for the given namespace.
// Note that the result will not have the package prefix set with SetPackagePrefix.
func (cu *CodeUtils) NamespaceToImportPath(ns string) string {
	pkg := strings.ReplaceAll(ns, ".", "/")
	return pkg
}

// NamespaceToFullImportPath returns an import path for the given namespace.
// The result path will contain the package prefix if it is set.
func (cu *CodeUtils) NamespaceToFullImportPath(ns string) string {
	pth := cu.NamespaceToImportPath(ns)
	if cu.packagePrefix != "" {
		pth = filepath.Join(cu.packagePrefix, pth)
	}
	return pth
}

// Import returns the package name and the full import path for the given AST.
func (cu *CodeUtils) Import(t *parser.Thrift) (pkg, pth string) {
	ns := t.GetNamespaceOrReferenceName("go")
	pkg = cu.NamespaceToPackage(ns)
	pth = cu.NamespaceToFullImportPath(ns)
	return
}

// GenTags generates go tags for the given parser.Field.
func (cu *CodeUtils) GenTags(f *parser.Field, insertPoint string) (string, error) {
	return cu.genFieldTags(f, insertPoint, nil)
}

// GenFieldTags generates go tags for the given parser.Field.
func (cu *CodeUtils) GenFieldTags(f *Field, insertPoint string) (string, error) {
	var tags []string
	if cu.Features().FrugalTag {
		requiredness := strings.ToLower(f.Requiredness.String())
		tags = append(tags, fmt.Sprintf(`frugal:"%d,%s,%s"`, f.ID, requiredness, f.frugalTypeName))
	}
	return cu.genFieldTags(f.Field, insertPoint, tags)
}

func (cu *CodeUtils) genFieldTags(f *parser.Field, insertPoint string, extend []string) (string, error) {
	var tags []string
	switch f.Requiredness {
	case parser.FieldType_Required:
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d,required"`, f.Name, f.ID))
	case parser.FieldType_Optional:
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d,optional"`, f.Name, f.ID))
	default:
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d"`, f.Name, f.ID))
	}

	tags = append(tags, extend...)

	if gotags := f.Annotations.Get("go.tag"); len(gotags) > 0 {
		tag := gotags[0]
		if cu.Features().EscapeDoubleInTag {
			tag = escape.ReplaceAllStringFunc(tag, func(m string) string {
				if m[1] == '"' {
					return m[1:]
				}
				return m
			})
		}
		tags = append(tags, tag)
	} else {
		if cu.Features().GenDatabaseTag {
			tags = append(tags, fmt.Sprintf(`db:"%s"`, f.Name))
		}

		if cu.Features().GenerateJSONTag {
			id := f.Name
			if cu.Features().SnakeTyleJSONTag {
				id = snakify(id)
			}
			if f.Requiredness.IsOptional() && cu.Features().GenOmitEmptyTag {
				tags = append(tags, fmt.Sprintf(`json:"%s,omitempty"`, id))
			} else {
				tags = append(tags, fmt.Sprintf(`json:"%s"`, id))
			}
		}
	}
	str := fmt.Sprintf("`%s%s`", strings.Join(tags, " "), insertPoint)
	return str, nil
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

// MkRWCtx = MakeReadWriteContext.
// Check the documents of ReadWriteContext for more informations.
func (cu *CodeUtils) MkRWCtx(root *Scope, f *Field) (*ReadWriteContext, error) {
	r := NewResolver(root, cu)
	ctx, err := mkRWCtx(r, root, f.Type, nil)
	if err != nil {
		return nil, err
	}

	// adjust for fields
	ctx.Target = "p." + f.GoName().String()
	ctx.Source = "src"
	ctx.TypeName = f.GoTypeName()
	ctx.IsPointer = f.GoTypeName().IsPointer()
	return ctx, nil
}

// SetRootScope sets the root scope for rendering templates.
func (cu *CodeUtils) SetRootScope(s *Scope) {
	cu.rootScope = s
}

// RootScope returns the root scope previously set by SetRootScope.
func (cu *CodeUtils) RootScope() *Scope {
	return cu.rootScope
}

// BuildFuncMap builds a function map for templates.
func (cu *CodeUtils) BuildFuncMap() template.FuncMap {
	m := map[string]interface{}{
		"ToUpper":        strings.ToUpper,
		"ToLower":        strings.ToLower,
		"InsertionPoint": plugin.InsertionPoint,
		"Unexport":       common.Unexport,

		"Debug":          cu.Debug,
		"Features":       cu.Features,
		"GetPackageName": cu.GetPackageName,
		"GenTags":        cu.GenTags,
		"GenFieldTags":   cu.GenFieldTags,
		"MkRWCtx": func(f *Field) (*ReadWriteContext, error) {
			return cu.MkRWCtx(cu.rootScope, f)
		},

		"IsBaseType":        IsBaseType,
		"NeedRedirect":      NeedRedirect,
		"IsFixedLengthType": IsFixedLengthType,
		"SupportIsSet":      SupportIsSet,
		"GetTypeIDConstant": GetTypeIDConstant,
		"UseStdLibrary": func(libs ...string) string {
			cu.rootScope.imports.UseStdLibrary(libs...)
			return ""
		},
		"ServicePrefix": func(svc *Service) (string, error) {
			if svc == nil || svc.From().namespace == cu.rootScope.namespace {
				return "", nil
			}
			ast := svc.From().AST()
			inc := cu.rootScope.Includes().ByAST(ast)
			if inc == nil {
				return "", fmt.Errorf("unexpected service[%s] from scope[%s]", svc.Name, ast.Filename)
			}
			return inc.PackageName + ".", nil
		},
		"ServiceName": func(svc *Service) Name {
			if svc == nil {
				return Name("")
			}
			return svc.GoName()
		},
		"Marshal": func(s *StructLike) (res string) {
			bs, err := meta.Marshal(buildMeta(cu.rootScope.ast, s.StructLike))
			if err != nil {
				return fmt.Sprintf("<%s>", err.Error())
			}
			return prettifyBytesLiteral(fmt.Sprintf("%#v", bs))
		},
	}
	return m
}

var (
	snakeRE1 = regexp.MustCompile(`([^_])([A-Z][a-z]+)`)
	snakeRE2 = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func snakify(id string) string {
	id = snakeRE1.ReplaceAllString(id, `${1}_${2}`)
	id = snakeRE2.ReplaceAllString(id, `${1}_${2}`)
	return strings.ToLower(id)
}
