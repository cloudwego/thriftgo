// Copyright 2021 CloudWeGo
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
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang/common"
	"github.com/cloudwego/thriftgo/generator/golang/styles"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
)

// Default libraries.
const (
	DefaultThriftLib  = "github.com/apache/thrift/lib/go/thrift"
	DefaultUnknownLib = "github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
)

// Errors
var (
	//ErrNotImplemented  = errors.New("Not implemented")
	ErrRootScopeNotSet = errors.New("root scope is not set")
)

// Features controls the behavior of CodeUtils.
type Features struct {
	MarshalEnumToText  bool `json_enum_as_text:"Generate MarshalText for enum values"`
	GenerateSetter     bool `gen_setter:"Generate Set* methods for fields"`
	GenDatabaseTag     bool `gen_db_tag:"Generate 'db:$field' tag"`
	GenOmitEmptyTag    bool `omitempty_for_optional:"Generate 'omitempty' tags for optional fields. Enabled by default."`
	TypedefAsTypeAlias bool `use_type_alias:"Generate type alias for typedef instead of type define. Enabled by default."`
	ValidateSet        bool `validate_set:"Generate codes to validate the uniqueness of set elements. Enabled by default."`
	ValueTypeForSIC    bool `value_type_in_container:"Genenerate value type for struct-like in container instead of pointer type."`
	ScanValueForEnum   bool `scan_value_for_enum:"Generate Scan and Value methods for enums to implement interfaces in std sql library."`
	ReorderFields      bool `reorder_fields:"Reorder fields of structs to improve memory usage."`
	TypedEnumString    bool `typed_enum_string:"Add type prefix to the string representation of enum values."`
	KeepUnknownFields  bool `keep_unknown_fields:"Genenerate codes to store unrecognized fields in structs."`
	GenDeepEqual       bool `gen_deep_equal:"Generate DeepEqual function for struct/union/exception."`
}

var defaultFeatures = Features{
	MarshalEnumToText:  false,
	GenerateSetter:     false,
	GenDatabaseTag:     false,
	GenOmitEmptyTag:    true,
	TypedefAsTypeAlias: true,
	ValidateSet:        true,
	ValueTypeForSIC:    false,
	ScanValueForEnum:   true,
	ReorderFields:      false,
	TypedEnumString:    false,
	KeepUnknownFields:  false,
	GenDeepEqual:       false,
}

var libs = map[string]string{
	"thrift":  DefaultThriftLib,
	"driver":  "database/sql/driver",
	"sql":     "database/sql",
	"unknown": DefaultUnknownLib,
}

// CodeUtils contains a set of functions used in the template execution.
type CodeUtils struct {
	backend.LogFunc
	ids           map[string]int           // Prefix => local variable index
	packagePrefix string                   // Package prefix for all generated codes.
	customized    map[string]string        // Customized imports, package => import path.
	features      Features                 // Available features.
	namingStyle   styles.Naming            // Naming style.
	synthesized   map[string]bool          // Names of synthesized types.
	fieldNames    map[*parser.Field]string // All resolved field names.
	setterNames   map[*parser.Field]string // All resolved field setter names.
	rootScope     *Scope
	scopeStack    []*Scope
	doInitialisms bool // Make initialisms setting kept event naming style changes.
}

// NewCodeUtils creates a new CodeUtils.
func NewCodeUtils(log backend.LogFunc) *CodeUtils {
	cu := &CodeUtils{
		LogFunc:     log,
		ids:         make(map[string]int),
		customized:  make(map[string]string),
		features:    defaultFeatures,
		namingStyle: styles.NewNamingStyle("thriftgo"),
		synthesized: make(map[string]bool),
	}
	return cu
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

// SetRootScope sets the root scope.
func (cu *CodeUtils) SetRootScope(s *Scope) {
	cu.rootScope = s
	cu.synthesized = make(map[string]bool)

	// install synthesized Args and Result type names
	for _, s := range cu.rootScope.ast.Services {
		for _, f := range s.Functions {
			at := cu.GetArgTypeName(s.Name, f)
			cu.synthesized[at] = true
			if f.Oneway {
				continue
			}
			rt := cu.GetResTypeName(s.Name, f)
			cu.synthesized[rt] = true
		}
	}

	funcs := append([]string{"Read", "Write", "String"})
	cu.fieldNames = make(map[*parser.Field]string)
	cu.setterNames = make(map[*parser.Field]string)
	for _, sl := range s.ast.GetStructLike() {
		used := funcs
		if cu.Features().GenerateSetter {
			for _, f := range sl.Fields {
				field, _ := cu.Unexport(f.Name)
				n, err := cu.Identify("set_"+field, used...)
				if err != nil {
					n = fmt.Sprintf("%s<%s>", f.Name, err.Error())
				}
				cu.setterNames[f] = n
				used = append(used, n)
			}
		}
		for _, f := range sl.Fields {
			n, err := cu.Identify(f.Name, used...)
			if err != nil {
				n = fmt.Sprintf("%s<%s>", f.Name, err.Error())
			}
			cu.fieldNames[f] = n
			used = append(used, n)
		}
	}

	if cu.Features().ReorderFields {
		for _, x := range s.ast.GetStructLike() {
			diff, err := cu.reorderFields(x)
			if err != nil {
				panic(err)
			}
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
	return m
}

// Features returns the current settings of generator features.
func (cu *CodeUtils) Features() Features {
	return cu.features
}

func (cu *CodeUtils) parseInclude(inc *parser.Include) (ref, pkg, pth string, err error) {
	if inc.Reference == nil {
		err = fmt.Errorf("include not parsed: '%s'", inc.Path)
		return
	}
	return cu.ParseNamespace(inc.Reference)
}

// ParseNamespace retrieves informations from the given AST and returns a
// reference name in the IDL, a package name for generated codes and an import path.
func (cu *CodeUtils) ParseNamespace(ast *parser.Thrift) (ref, pkg, pth string, err error) {
	ref = filepath.Base(ast.Filename)
	ref = strings.TrimSuffix(ref, filepath.Ext(ref))
	ns := ast.GetNamespaceOrReferenceName("go")
	pkg = cu.NamespaceToPackage(ns)
	pth = cu.NamespaceToImportPath(ns)
	return
}

// BuildScope creates a scope of the AST with its includes processed recursively.
func (cu *CodeUtils) BuildScope(ast *parser.Thrift) (*Scope, error) {
	s := NewScope(ast)
	err := s.init(cu)
	if err != nil {
		return nil, fmt.Errorf("process '%s' failed: %w", ast.Filename, err)
	}
	return s, nil
}

// ResolveSymbol resolves a type from AST to a TypeSymbol within the given scope.
func (cu *CodeUtils) ResolveSymbol(t *parser.Type, scope *Scope) (*TypeSymbol, error) {
	if sym, ok := baseTypes[t.Name]; ok {
		return sym, nil
	}

	sym := &TypeSymbol{}
	switch t.Name {
	case "map":
		key, err := cu.ResolveSymbol(t.KeyType, scope)
		if err != nil {
			return nil, fmt.Errorf("resolve key type of '%s' failed: %w", t, err)
		}
		if key.TypeID == typeids.Binary { // 'binary => string' for key type in map
			key = key.Clone()
			key.GoType = "string"
		}
		if key.TypeID == typeids.Struct {
			key = key.Clone()
			key.GoType = "*" + key.GoType
		}
		sym.KeyType = key
		fallthrough
	case "set", "list":
		val, err := cu.ResolveSymbol(t.ValueType, scope)
		if err != nil {
			return nil, fmt.Errorf("resolve value type of '%s' failed: %w", t, err)
		}

		if val.TypeID == typeids.Struct && !cu.Features().ValueTypeForSIC {
			val = val.Clone()
			val.GoType = "*" + val.GoType
		}

		sym.ValType = val

		if t.Name == "map" {
			sym.GoType = fmt.Sprintf("map[%s]%s", sym.KeyType.GoType, sym.ValType.GoType)
		} else {
			sym.GoType = fmt.Sprintf("[]%s", sym.ValType.GoType)
		}
		sym.TypeID = common.UpperFirstRune(t.Name)
		return sym, nil
	}

	parts := splitType(t.Name)
	switch len(parts) {
	case 1:
		if sym, ok := scope.name2type[t.Name]; ok {
			return sym, nil
		}
	case 2:
		ref, name := parts[0], parts[1]
		top, ok := scope.ref2scope[ref]
		if !ok {
			break
		}
		sym, ok := top.name2type[name]
		if !ok {
			break
		}

		res := &TypeSymbol{
			GoType: sym.GoType,
			TypeID: sym.TypeID,
			Enum:   sym.Enum,
		}
		if pkg := scope.ref2pkg[ref]; pkg != "" {
			res.GoType = pkg + "." + res.GoType
		}
		return res, nil
	default:
		return nil, fmt.Errorf("invalid type name: '%s'", t.Name)
	}
	return nil, fmt.Errorf("undefined type: '%s'", t)
}

// ResolveSymbolInRootScope resolves a type from AST to a TypeSymbol within the root scope.
func (cu *CodeUtils) ResolveSymbolInRootScope(t *parser.Type) (*TypeSymbol, error) {
	if cu.rootScope == nil {
		return nil, ErrRootScopeNotSet
	}
	return cu.ResolveSymbol(t, cu.rootScope)
}

// ResolveImports returns a map of import path to alias built from the include list
// of the IDL. An alias may be an empty string to indicate no alias is need for the
// import path.
func (cu *CodeUtils) ResolveImports() (map[string]string, error) {
	if cu.rootScope == nil {
		return nil, ErrRootScopeNotSet
	}

	imports := make(map[string]string)
	for pkg, path := range cu.rootScope.imports {
		imports[pkg] = path
	}

	for pkg, path := range cu.customized {
		imports[pkg] = path
	}

	res := make(map[string]string, len(imports))
	for alias, path := range imports {
		if alias == path || strings.HasSuffix(path, "/"+alias) {
			res[path] = ""
		} else {
			res[path] = alias
		}
	}
	return res, nil
}

// IsExported determines whether a name is exported.
func (cu *CodeUtils) IsExported(name string) bool {
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}

// GetFilePath returns a path to the generated file for the given IDL.
// Note that the result is a path relative to the root output path.
func (cu *CodeUtils) GetFilePath(t *parser.Thrift) (string, error) {
	ref, _, pth, err := cu.ParseNamespace(t)
	if err != nil {
		return "", err
	}
	full := filepath.Join(pth, ref+".go")
	if strings.HasSuffix(full, "_test.go") {
		full = strings.Replace(full, "_test.go", "_test_.go", -1)
	}
	return full, nil
}

// NamespaceToPackage converts a namespace to a package.
func (cu *CodeUtils) NamespaceToPackage(ns string) string {
	parts := strings.Split(ns, ".")
	return strings.ToLower(parts[len(parts)-1])
}

// NamespaceToImportPath returns an import path for the given namespace.
// Note that the result will not have the package prefix set with SetPackagePrefix.
func (cu *CodeUtils) NamespaceToImportPath(ns string) string {
	pkg := strings.Replace(ns, ".", "/", -1)
	return pkg
}

// GetDefaultValueTypeName returns a type name suitable for the default value of the given field.
func (cu *CodeUtils) GetDefaultValueTypeName(f *parser.Field) (string, error) {
	t, err := cu.ResolveFieldTypeName(f)
	if err != nil {
		return "", err
	}
	if yes, _ := cu.IsBaseType(f.Type); yes {
		t = cu.Deref(t)
	}
	return t, nil
}

// ID returns the ID of a field. If the ID is a minus number, the slash will be replaced by an underscore.
func (cu *CodeUtils) ID(v interface{}) string {
	if c, ok := v.(*ReadWriteContext); ok {
		return c.ID
	}
	if f, ok := v.(*parser.Field); ok {
		id := fmt.Sprint(f.ID)
		return strings.Replace(id, "-", "_", -1)
	}
	return fmt.Sprintf("<invalid type %T>", v)
}

// GetFieldInit returns the initialization code for a field.
// The given field must have a default value.
func (cu *CodeUtils) GetFieldInit(f *parser.Field) (string, error) {
	return cu.GetConstInit(f.Name, f.Type, f.Default)
}

// SearchValue searches the given id in the current scope.
func (cu *CodeUtils) SearchValue(id string) (string, error) {
	return cu.rootScope.SearchValue(id, cu)
}

// RefInfo contains information for querying references of a type.
type RefInfo struct {
	RefName string
	RawName string
	Package string
	Import  string
}

// SearchReference returns a RefInfo of the given identifier which must be a type.
// If the type is defined in current scope or not found, the return value will be nil.
func (cu *CodeUtils) SearchReference(id string) *RefInfo {
	if parts := splitType(id); len(parts) == 2 {
		pkg, ok := cu.rootScope.ref2pkg[parts[0]]
		if ok {
			pth, _ := cu.rootScope.pkg2path[pkg]
			return &RefInfo{
				RefName: parts[0],
				RawName: parts[1],
				Package: pkg,
				Import:  pth,
			}
		}
	}
	return nil
}

// GetConstInit returns the initialization code for a constant.
func (cu *CodeUtils) GetConstInit(name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	if cu.rootScope == nil {
		return "", ErrRootScopeNotSet
	}

	sym, err := cu.ResolveSymbolInRootScope(t)
	if err != nil {
		return "", err
	}
	switch sym.TypeID {
	case typeids.Bool:
		switch v.Type {
		case parser.ConstType_ConstInt:
			val := v.TypedValue.GetInt()
			return fmt.Sprint(val > 0), nil
		case parser.ConstType_ConstIdentifier:
			s := v.TypedValue.GetIdentifier()
			if s == "true" || s == "false" {
				return s, nil
			}
			val, err := cu.SearchValue(s)
			if err != nil {
				return "", err
			}
			return val, nil
		}

	case typeids.Byte, typeids.I8, typeids.I16, typeids.I32, typeids.I64:
		switch v.Type {
		case parser.ConstType_ConstInt:
			val := v.TypedValue.GetInt()
			return fmt.Sprint(val), nil
		case parser.ConstType_ConstIdentifier:
			s := v.TypedValue.GetIdentifier()
			if s == "true" {
				return "1", nil
			}
			if s == "false" {
				return "0", nil
			}
			val, err := cu.SearchValue(s)
			if err != nil {
				return "", err
			}
			// enum types require a explicit type conversion
			return fmt.Sprintf("%s(%s)", sym.GoType, val), nil
		}

	case typeids.Double:
		switch v.Type {
		case parser.ConstType_ConstInt:
			val := v.TypedValue.GetInt()
			return fmt.Sprint(val), nil
		case parser.ConstType_ConstDouble:
			val := v.TypedValue.GetDouble()
			return fmt.Sprint(val), nil
		case parser.ConstType_ConstIdentifier:
			s := v.TypedValue.GetIdentifier()
			if s == "true" {
				return "1", nil
			}
			if s == "false" {
				return "0", nil
			}
			val, err := cu.SearchValue(s)
			if err != nil {
				return "", err
			}
			return val, nil
		}

	case typeids.String, typeids.Binary:
		var str string
		switch v.Type {
		case parser.ConstType_ConstLiteral:
			str = fmt.Sprintf("\"%s\"", v.TypedValue.GetLiteral())
		case parser.ConstType_ConstIdentifier:
			s := v.TypedValue.GetIdentifier()
			if s == "true" || s == "false" {
				break
			}
			val, err := cu.SearchValue(s)
			if err != nil {
				return "", err
			}
			return val, nil
		}
		if sym.TypeID == typeids.Binary {
			str = fmt.Sprintf("[]byte(%s)", str)
		}
		return str, nil

	case typeids.Set, typeids.List:
		var ss []string
		switch v.Type {
		case parser.ConstType_ConstList:
			elemName := "element of " + name
			for _, elem := range v.TypedValue.GetList() {
				str, err := cu.GetConstInit(elemName, t.ValueType, elem)
				if err != nil {
					return "", err
				}
				ss = append(ss, str+",")
			}
			if len(ss) == 0 {
				return sym.GoType + "{}", nil
			}
			return fmt.Sprintf("%s{\n%s\n}", sym.GoType, strings.Join(ss, "\n")), nil
		case parser.ConstType_ConstInt, parser.ConstType_ConstDouble,
			parser.ConstType_ConstLiteral, parser.ConstType_ConstMap:
			return sym.GoType + "{}", nil
		}

	case typeids.Map:
		var kvs []string
		switch v.Type {
		case parser.ConstType_ConstMap:
			for _, mcv := range v.TypedValue.Map {
				keyName := "key of " + name
				key, err := cu.GetConstInit(keyName, t.KeyType, mcv.Key)
				if err != nil {
					return "", err
				}
				valName := "value of " + name
				val, err := cu.GetConstInit(valName, t.ValueType, mcv.Value)
				if err != nil {
					return "", err
				}
				kvs = append(kvs, fmt.Sprintf("%s: %s,", key, val))
			}
			if len(kvs) == 0 {
				return sym.GoType + "{}", nil
			}
			return fmt.Sprintf("%s{\n%s\n}", sym.GoType, strings.Join(kvs, "\n")), nil

		case parser.ConstType_ConstInt, parser.ConstType_ConstDouble,
			parser.ConstType_ConstLiteral, parser.ConstType_ConstList:
			return sym.GoType + "{}", nil
		}

	case typeids.Struct:
		if v.Type == parser.ConstType_ConstMap {
			var kvs []string
			for _, mcv := range v.TypedValue.Map {
				if mcv.Key.Type != parser.ConstType_ConstLiteral {
					return "", fmt.Errorf("expect literals as keys in default value of struct type '%s', got '%s'", name, mcv.Key.Type)
				}
				n := mcv.Key.TypedValue.GetLiteral()
				f, err := cu.getStructField(t, n)
				if err != nil {
					return "", err
				}
				key, err := cu.Identify(n)
				if err != nil {
					return "", err
				}

				valName := name + "." + f.Name
				val, err := cu.GetConstInit(valName, f.Type, mcv.Value)
				if err != nil {
					return "", err
				}
				if yes, _ := cu.NeedRedirect(f); yes {
					if yes, _ = cu.IsBaseType(f.Type); yes {
						// a trick to create pointers without temporary variables
						typ, _ := cu.ResolveFieldTypeName(f)
						val = fmt.Sprintf("(&struct{x %s}{%s}).x", cu.Deref(typ), val)
					}
					if !strings.HasPrefix(val, "&") {
						val = "&" + val
					}
				}
				kvs = append(kvs, fmt.Sprintf("%s: %s,", key, val))
			}
			if len(kvs) == 0 {
				return "&" + sym.GoType + "{}", nil
			}
			return fmt.Sprintf("&%s{\n%s\n}", cu.Deref(sym.GoType), strings.Join(kvs, "\n")), nil
		}
	}
	return "", fmt.Errorf("type error: '%s' was declared as type %s", name, t)
}

func (cu *CodeUtils) getStructField(t *parser.Type, field string) (*parser.Field, error) {
	var name, scope, ref = t.Name, cu.rootScope, ""

	// TODO: the below algorithm is not correct when resolving a type defined in a
	// indirect included IDL. Such a type may requires an extra packge import when
	// packge name confliction resolving which can not be done before reaching this
	// type. Thus the result import statements may be not correct and will cause a
	// 'undefined' error.
RESOLVE:
	parts := splitType(name)
	switch len(parts) {
	case 2:
		ref = parts[0]
		if tmp, ok := scope.ref2scope[ref]; !ok {
			break
		} else {
			scope, name = tmp, parts[1]
		}
		fallthrough
	case 1:
		if td, ok := scope.ast.GetTypedef(name); ok {
			name = td.Type.Name
			goto RESOLVE
		} else {
			for _, s := range scope.ast.GetStructLike() {
				if s.Name == name {
					for _, f := range s.GetFields() {
						if f.Name == field {
							if ref != "" {
								var ff parser.Field = *f
								ff.Type = cu.indirectType(ref, f.Type)
								return &ff, nil
							}
							return f, nil
						}
					}
				}
			}
		}
		return nil, fmt.Errorf("field '%s' not found in type '%s'", field, t.Name)
	}
	return nil, fmt.Errorf("undefined type: '%s' (scope %s)", t, scope.ast.Filename)
}

func (cu *CodeUtils) indirectType(ref string, t *parser.Type) *parser.Type {
	if t == nil {
		return nil
	}
	if baseTypes[t.Name] != nil || isContainerTypes[t.Name] {
		return t
	}
	var tt parser.Type
	if !strings.Contains(t.Name, ".") {
		tt.Name = ref + "." + t.Name
	}
	tt.KeyType = cu.indirectType(ref, t.KeyType)
	tt.ValueType = cu.indirectType(ref, t.ValueType)
	return &tt
}

func (cu *CodeUtils) getGlobals() (cs, ncs []*parser.Constant, err error) {
	if cu.rootScope == nil {
		return nil, nil, ErrRootScopeNotSet
	}
	for _, c := range cu.rootScope.ast.Constants {
		isConst, err := cu.matchTypeID(c.Type,
			typeids.Bool,
			typeids.Byte,
			typeids.I16,
			typeids.I32,
			typeids.I64,
			typeids.Double,
			typeids.String)
		if err != nil {
			return nil, nil, err
		}
		if isConst {
			cs = append(cs, c)
		} else {
			ncs = append(ncs, c)
		}
	}
	return
}

// GetConstGlobals returns all global variables defined in the root scope that result in constant value.
func (cu *CodeUtils) GetConstGlobals() (cs []*parser.Constant, err error) {
	cs, _, err = cu.getGlobals()
	return
}

// GetNonConstGlobals returns all global variables defined in the root scope that result in non-constant values.
func (cu *CodeUtils) GetNonConstGlobals() (ncs []*parser.Constant, err error) {
	_, ncs, err = cu.getGlobals()
	return
}

// Identify converts an raw name from IDL into an exported identifier in go.
// If the identifier potentially conflicts with some synthesized name or matches
// any used name, an "_" will be appended to it.
func (cu *CodeUtils) Identify(name string, used ...string) (string, error) {
	s, err := cu.namingStyle.Identify(name)
	if err != nil {
		return "", err
	}
	if cu.synthesized[name] {
		return s, nil
	}
	if strings.HasPrefix(s, "New") || strings.HasSuffix(s, "Args") || strings.HasSuffix(s, "Result") {
		s += "_"
		return s, nil
	}
	for i := range used {
		if used[i] == s {
			s += "_"
			break
		}
	}
	return s, nil
}

// Identify0 is a help function for Identify.
func (cu *CodeUtils) Identify0(name string, used []string) (string, error) {
	return cu.Identify(name, used...)
}

// ParamName adds a suffix to the name if it is a keyword of golang.
func (cu *CodeUtils) ParamName(name string) string {
	if isKeywords[name] {
		return name + "_a1"
	}
	return name
}

// ResolveFieldName returns the resolved name of the given field.
func (cu *CodeUtils) ResolveFieldName(f *parser.Field) (string, error) {
	name, ok := cu.fieldNames[f]
	if !ok {
		return cu.Identify(f.Name)
	}
	return name, nil
}

// GetFieldSetterName returns a name of the setter function for the given field.
func (cu *CodeUtils) GetFieldSetterName(f *parser.Field) (string, error) {
	name, ok := cu.setterNames[f]
	if !ok {
		field, _ := cu.Unexport(f.Name)
		return cu.Identify("set_" + field)
	}
	return name, nil
}

// ToStructLike get StructLike from synthesized or StructLike
func (cu *CodeUtils) ToStructLike(v interface{}) *parser.StructLike {
	if t, ok := v.(*parser.StructLike); ok {
		return t
	}
	if s, ok := v.(*Synthesized); ok {
		return s.StructLike
	}
	return nil
}

// FieldNamesx returns all identifiers of the fields of a struct-like.
func (cu *CodeUtils) FieldNamesx(v interface{}) (names []string, err error) {
	var fs []*parser.Field
	if s := cu.ToStructLike(v); s != nil {
		fs = s.Fields
	}
	for _, f := range fs {
		name, err := cu.ResolveFieldName(f)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

// Unexport returns an unexported form of the given identifier.
func (cu *CodeUtils) Unexport(name string) (string, error) {
	return common.LowerFirstRune(name), nil
}

// MakeEnumValueName returns a legal identifier for a enum value.
func (cu *CodeUtils) MakeEnumValueName(typeName string, valueName string) string {
	return typeName + "_" + valueName
}

// ResolveFieldTypeName returns a legal type name in go for the given field.
func (cu *CodeUtils) ResolveFieldTypeName(f *parser.Field) (string, error) {
	sym, err := cu.ResolveSymbolInRootScope(f.Type)
	if err != nil {
		return "", err
	}

	if yes, err := cu.NeedRedirect(f); err == nil {
		if yes {
			return "*" + cu.Deref(sym.GoType), nil
		}
	} else {
		return "", err
	}
	return sym.GoType, nil
}

func (cu *CodeUtils) reorderFields(s *parser.StructLike) (*sizeDiff, error) {
	if len(s.Fields) == 0 {
		return nil, nil
	}

	fs := make([]*parser.Field, len(s.Fields))
	copy(fs, s.Fields)

	var diff sizeDiff
	var a1, a2 align
	sizes := make(map[*parser.Field]int, len(fs))
	for _, f := range fs {
		sym, err := cu.ResolveSymbolInRootScope(f.Type)
		if err != nil {
			return nil, err
		}
		yes, err := cu.NeedRedirect(f)
		if err != nil {
			return nil, err
		}
		if yes {
			sizes[f] = pointerSize
		} else {
			if sym.IsEnumType() {
				sizes[f] = sizeof[typeids.I64]
			} else {
				sizes[f] = sizeof[sym.TypeID]
			}
		}
		a1.add(sizes[f])
	}

	sort.Slice(fs, func(i, j int) bool {
		return sizes[fs[i]] >= sizes[fs[j]]
	})

	for _, f := range fs {
		a2.add(sizes[f])
	}

	diff.original = a1.padded()
	diff.arranged = a2.padded()
	if diff.original != diff.arranged {
		s.Fields = fs
	}
	return &diff, nil
}

// ResolveTypeName returns a legal type name in go for the given AST type.
func (cu *CodeUtils) ResolveTypeName(t *parser.Type) (string, error) {
	sym, err := cu.ResolveSymbolInRootScope(t)
	if err != nil {
		return "", err
	}
	if sym.TypeID == typeids.Struct {
		return "*" + sym.GoType, nil
	}
	return sym.GoType, nil
}

func (cu *CodeUtils) matchTypeID(t *parser.Type, typeIDs ...string) (bool, error) {
	sym, err := cu.ResolveSymbolInRootScope(t)
	if err != nil {
		return false, err
	}
	for _, tid := range typeIDs {
		if tid == sym.TypeID {
			return true, nil
		}
	}
	return false, nil
}

// IsBaseType determines whether the given type is a base type.
func (cu *CodeUtils) IsBaseType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t,
		typeids.Bool,
		typeids.Byte,
		typeids.I16,
		typeids.I32,
		typeids.I64,
		typeids.Double,
		typeids.String,
		typeids.Binary)
}

// IsFixedLengthType determines whether the given type is a fixed length type.
func (cu *CodeUtils) IsFixedLengthType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t,
		typeids.Bool,
		typeids.Byte,
		typeids.I16,
		typeids.I32,
		typeids.I64,
		typeids.Double)
}

// IsSetType determines whether the given type is a set type.
func (cu *CodeUtils) IsSetType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t, typeids.Set)
}

// IsContainerType determines whether the given type is a container type.
func (cu *CodeUtils) IsContainerType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t, typeids.Map, typeids.Set, typeids.List)
}

// IsStructLike determines whether the given type is a struct-like type.
func (cu *CodeUtils) IsStructLike(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t, typeids.Struct)
}

// IsStringType reports whether the given type resolves into a string type.
func (cu *CodeUtils) IsStringType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t, typeids.String)
}

// IsBinaryType reports whether the given type resolves into a binary type.
func (cu *CodeUtils) IsBinaryType(t *parser.Type) (bool, error) {
	return cu.matchTypeID(t, typeids.Binary)
}

// IsEnumType reports whether the given type resolves into an enum type.
func (cu *CodeUtils) IsEnumType(t *parser.Type) (bool, error) {
	sym, err := cu.ResolveSymbolInRootScope(t)
	if err != nil {
		return false, err
	}
	return sym.IsEnumType(), nil
}

// GenTags generates go tags for the given parser.Field.
func (cu *CodeUtils) GenTags(f *parser.Field, insertPoint string) (string, error) {
	var tags []string
	if f.Requiredness == parser.FieldType_Required {
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d,required"`, f.Name, f.ID))
	} else {
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d"`, f.Name, f.ID))
	}

	if gotags := f.Annotations["go.tag"]; gotags != "" {
		tags = append(tags, gotags)
	} else {
		if cu.Features().GenDatabaseTag {
			tags = append(tags, fmt.Sprintf(`db:"%s"`, f.Name))
		}

		if cu.IsOptional(f) && cu.Features().GenOmitEmptyTag {
			tags = append(tags, fmt.Sprintf(`json:"%s,omitempty"`, f.Name))
		} else {
			tags = append(tags, fmt.Sprintf(`json:"%s"`, f.Name))
		}
	}
	str := fmt.Sprintf("`%s%s`", strings.Join(tags, " "), insertPoint)
	return str, nil
}

// IsSetterOfResponseType determines whether the field in th given type is the Success field of a response wrapper type.
func (cu *CodeUtils) IsSetterOfResponseType(typeName, fieldName string) (bool, error) {
	yes := strings.HasSuffix(typeName, "Result") && fieldName == "Success"
	return yes, nil
}

// IsRequired reports whether the given field is required.
func (cu *CodeUtils) IsRequired(f *parser.Field) bool {
	return f.Requiredness == parser.FieldType_Required
}

// IsOptional reports whether the given field is optional.
func (cu *CodeUtils) IsOptional(f *parser.Field) bool {
	return f.Requiredness == parser.FieldType_Optional
}

// HasDefaultValue reports whether the given field is specified with a default value.
func (cu *CodeUtils) HasDefaultValue(f *parser.Field) bool {
	return f.IsSetDefault()
}

// SupportIsSet determines whether a field supports IsSet query.
func (cu *CodeUtils) SupportIsSet(f *parser.Field) (yes bool, err error) {
	if yes, err = cu.IsStructLike(f.Type); err != nil || yes {
		return
	}
	if cu.IsOptional(f) {
		return true, nil
	}
	return false, nil
}

// NeedRedirect deterimines whether the given field should result in a pointer type.
// Condition: struct-like || (optional non-binary base type without default vlaue)
func (cu *CodeUtils) NeedRedirect(f *parser.Field) (yes bool, err error) {
	if yes, err = cu.IsStructLike(f.Type); err != nil || yes {
		return
	}

	if cu.IsOptional(f) && !cu.HasDefaultValue(f) {
		isBinary, err := cu.IsBinaryType(f.Type)
		if err != nil || isBinary {
			// binary types produce slice types
			return false, err
		}

		isBaseType, err := cu.IsBaseType(f.Type)
		if err != nil {
			return false, err
		}
		return isBaseType, nil
	}

	return false, nil
}

// IsPointerTypeName reports whether the type name has a prefix "*".
func (cu *CodeUtils) IsPointerTypeName(typeName string) bool {
	return strings.HasPrefix(typeName, "*")
}

// GetNewFunc returns the construction function of the given type.
func (cu *CodeUtils) GetNewFunc(typeName string) string {
	if idx := strings.Index(typeName, "."); idx >= 0 {
		idx++
		return typeName[:idx] + "New" + typeName[idx:]
	}
	return "New" + typeName
}

// Deref removes the "&" and "*" prefix of the given code.
func (cu *CodeUtils) Deref(code string) string {
	return strings.TrimLeft(code, "&*")
}

// GetTypeIDConstant returns a constnat appropriate for
func (cu *CodeUtils) GetTypeIDConstant(t *parser.Type) (string, error) {
	tid, err := cu.GetTypeID(t)
	if err != nil {
		return "", err
	}
	tid = strings.ToUpper(tid)
	if tid == "BINARY" {
		tid = "STRING"
	}
	return tid, nil
}

// GetTypeID returns the thrift type ID literal for the given type which is suitable
// to concate with "Read" or "Write" to produce a valid method name in the TProtocol
// interface. Note that enum types results in I32.
func (cu *CodeUtils) GetTypeID(t *parser.Type) (string, error) {
	// Bool|Byte|I16|I32|I64|Double|String|Binary|Set|List|Map|Struct
	sym, err := cu.ResolveSymbolInRootScope(t)
	if err != nil {
		return "", err
	}
	return sym.TypeID, nil
}

// ResetIDGenerator resets the varID to zero.
func (cu *CodeUtils) ResetIDGenerator() (none string) {
	cu.ids = make(map[string]int)
	return
}

// GenID returns a local variable with the given name as prefix.
func (cu *CodeUtils) GenID(prefix string) (name string) {
	name = prefix
	if id := cu.ids[prefix]; id > 0 {
		name += fmt.Sprint(id)
	}
	cu.ids[prefix]++
	return
}

// BaseServicePrefix returns the package prefix of the base of the given service.
// If the given service has no base service or the base service is in current IDL,
// then the package prefix will be "".
func (cu *CodeUtils) BaseServicePrefix(t *parser.Service) (string, error) {
	if t.Extends != "" {
		if cu.rootScope == nil {
			return "", ErrRootScopeNotSet
		}
		ref, pkg, svc := "", "", t.Extends
		scope := cu.rootScope
		parts := splitType(t.Extends)
		switch len(parts) {
		case 2:
			ref, svc = parts[0], parts[1]
			scope = cu.rootScope.ref2scope[ref]
			if scope == nil {
				return "", fmt.Errorf("undefined service: '%s'", t.Extends)
			}
			pkg = cu.rootScope.ref2pkg[ref]
			if pkg != "" {
				pkg = pkg + "."
			}
			fallthrough
		case 1:
			for _, s := range scope.ast.Services {
				if s.Name == svc {
					return pkg, nil
				}
			}
			return "", fmt.Errorf("undefined service: '%s'", t.Extends)
		default:
			return "", fmt.Errorf("invalid base service: '%s'", t.Extends)
		}
	}
	return "", nil
}

// GetServiceIdentifier returns the identifier without a package prefix.
func (cu *CodeUtils) GetServiceIdentifier(id string) (string, error) {
	if id == "" {
		return "", nil
	}
	parts := splitType(id)
	switch len(parts) {
	case 1:
		return cu.Identify(parts[0])
	case 2:
		return cu.Identify(parts[1])
	default:
		return "", fmt.Errorf("invalid service name: '%s'", id)
	}
}

// GetArgTypeName returns the type name of the parameter wrapper type for the given function.
func (cu *CodeUtils) GetArgTypeName(svc string, f *parser.Function) string {
	return svc + "_" + common.LowerFirstRune(f.Name) + "_args"
}

// GetResTypeName returns the type name of the result wrapper type for the given function.
func (cu *CodeUtils) GetResTypeName(svc string, f *parser.Function) string {
	return svc + "_" + common.LowerFirstRune(f.Name) + "_result"
}

// Synthesized wraps the synthesized types with the service name it belongs to.
type Synthesized struct {
	*parser.StructLike
	Service string
}

// GetStructName returns the name of the given struct-like object for the
// thrift.TProtocol.WriteFieldBegin function.
func (cu *CodeUtils) GetStructName(obj interface{}) (string, error) {
	switch obj.(type) {
	case *parser.StructLike:
		return obj.(*parser.StructLike).Name, nil
	case *Synthesized:
		s := obj.(*Synthesized)
		return strings.TrimPrefix(s.Name, s.Service+"_"), nil
	}
	return "", fmt.Errorf("unsupported type: %T", obj)
}

// BuildArgsType creates a parser.StructLike for a given function's parameters.
func (cu *CodeUtils) BuildArgsType(svc string, f *parser.Function) (*Synthesized, error) {
	name := cu.GetArgTypeName(svc, f)
	args := &Synthesized{
		StructLike: &parser.StructLike{
			Category: "struct",
			Name:     name,
			Fields:   f.Arguments,
		},
		Service: svc,
	}
	return args, nil
}

// BuildResType creates a parser.StructLike for a given function's result type.
func (cu *CodeUtils) BuildResType(svc string, f *parser.Function) (*Synthesized, error) {
	name := cu.GetResTypeName(svc, f)
	res := &Synthesized{
		StructLike: &parser.StructLike{
			Category: "struct",
			Name:     name,
		},
		Service: svc,
	}
	if !f.Void {
		field := &parser.Field{
			ID:           0,
			Name:         "success",
			Requiredness: parser.FieldType_Optional,
			Type:         f.FunctionType,
		}
		res.Fields = append(res.Fields, field)
	}
	res.Fields = append(res.Fields, f.Throws...)
	return res, nil
}

// ReadWriteContext contains information for generating codes in ReadField* and
// WriteField* functions.
type ReadWriteContext struct {
	ID        string       // A field ID or ""
	Target    string       // A field name or a temporary variable
	Source    string       // A field name or a temporary variable
	FieldName string       // A field name from IDL or ""
	TypeName  string       // A type name.
	TypeID    string       // A type ID or ""
	Type      *parser.Type // The target type
	IsPointer bool         // Whether the target type is a pointer type
	IsMapKey  bool         // Whether the target type is a map key type
	NeedDecl  bool         // Whether a declaration of target is needed
}

// MkRWCtx = MakeReadWriteContext.
func (cu *CodeUtils) MkRWCtx(obj interface{}, tgtName string, needDecl bool, isMapKey bool) (*ReadWriteContext, error) {
	return cu.MkRWCtx2(obj, tgtName, "", needDecl, isMapKey)
}

// MkRWCtx2 = MakeReadWriteContext2.
func (cu *CodeUtils) MkRWCtx2(obj interface{}, tgtName, srcName string, needDecl bool, isMapKey bool) (*ReadWriteContext, error) {
	switch obj.(type) {
	case *parser.Field:
		f := obj.(*parser.Field)

		t, err := cu.ResolveFieldName(f)
		if err != nil {
			return nil, fmt.Errorf("MkRWCtx failed to Identify %+v: %w", f, err)
		}

		tn, err := cu.ResolveFieldTypeName(f)
		if err != nil {
			return nil, fmt.Errorf("MkRWCtx failed to resolve field type: %w", err)
		}

		ti, err := cu.GetTypeID(f.Type)
		if err != nil {
			return nil, fmt.Errorf("MkRWCtx failed to get type ID: %w", err)
		}

		ctx := &ReadWriteContext{
			ID:        cu.ID(f),
			Target:    "p." + t,
			Source:    "src",
			FieldName: f.Name,
			TypeName:  tn,
			TypeID:    ti,
			Type:      f.Type,
			IsPointer: cu.IsPointerTypeName(tn),
			IsMapKey:  isMapKey,
			NeedDecl:  needDecl,
		}
		return ctx, nil
	case *parser.Type:
		t := obj.(*parser.Type)

		tn, err := cu.ResolveTypeName(t)
		if err != nil {
			return nil, fmt.Errorf("MkRWCtx failed to resolve field type: %w", err)
		}

		ti, err := cu.GetTypeID(t)
		if err != nil {
			return nil, fmt.Errorf("MkRWCtx failed to get type ID: %w", err)
		}

		ctx := &ReadWriteContext{
			Target:    tgtName,
			Source:    srcName,
			FieldName: "",
			TypeName:  tn,
			TypeID:    ti,
			Type:      t,
			IsPointer: cu.IsPointerTypeName(tn),
			IsMapKey:  isMapKey,
			NeedDecl:  needDecl,
		}

		if ctx.IsMapKey {
			switch ctx.TypeID {
			case typeids.Struct:
				ctx.TypeName = cu.Deref(ctx.TypeName)
				ctx.IsPointer = false
			case typeids.Binary:
				ctx.TypeName = "string"
			}
		}

		return ctx, nil
	default:
		return nil, fmt.Errorf("MkRWCtx: unsupported type %T", obj)
	}
}

// GetKeyType returns the key type of the given type. T must be a map type.
func (cu *CodeUtils) GetKeyType(t *parser.Type) (*parser.Type, error) {
	yes, err := cu.matchTypeID(t, typeids.Map)
	if err != nil {
		return nil, err
	}
	if !yes {
		return nil, fmt.Errorf("expect map type, got: '%s'", t)
	}
	if tt := cu.rootScope.Dealias(t.Name); tt != nil {
		t = tt
	}
	return t.KeyType, nil
}

// GetValType returns the value type of the given type. T must be a container type.
func (cu *CodeUtils) GetValType(t *parser.Type) (*parser.Type, error) {
	yes, err := cu.IsContainerType(t)
	if err != nil {
		return nil, err
	}
	if !yes {
		return nil, fmt.Errorf("expect container type, got: '%s'", t)
	}
	if tt := cu.rootScope.Dealias(t.Name); tt != nil {
		t = tt
	}
	return t.ValueType, nil
}

// Debug .
func (cu *CodeUtils) Debug(vs ...interface{}) string {
	var ss []string
	for _, v := range vs {
		ss = append(ss, fmt.Sprintf("%T(%+v)", v, v))
	}
	println("[DEBUG]", strings.Join(ss, " "))
	return ""
}

func splitType(id string) []string {
	if id == "" {
		return []string{}
	}
	idx := strings.LastIndex(id, ".")
	if idx == -1 {
		return []string{id}
	}
	return []string{id[:idx], id[idx+1:]}
}

func splitValue(id string) (sss [][]string) {
	if id == "" {
		return
	}
	idx := strings.LastIndex(id, ".")
	if idx == -1 {
		sss = append(sss, []string{id})
		return
	}

	a, b := id, ""
	for idx != -1 {
		a, b = id[:idx], id[idx+1:]
		sss = append(sss, []string{a, b})
		idx = strings.LastIndex(a, ".")
	}
	return
}
