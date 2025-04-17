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

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
)

func errTypeMissMatch(name string, ft *parser.Type, v *parser.ConstValue) error {
	return fmt.Errorf("type error: '%s' was declared as type %s, but got default value of type %s", name, ft, v)
}

// Resolver resolves names for types, names and initialization value for thrift AST
// nodes in a scope (the root scope).
type Resolver struct {
	root *Scope
	util *CodeUtils
}

// NewResolver creates a new Resolver with the given scope.
func NewResolver(root *Scope, cu *CodeUtils) *Resolver {
	return &Resolver{
		root: root,
		util: cu,
	}
}

// GetDefaultValueTypeName returns a type name suitable for the default value of the given field.
func (r *Resolver) GetDefaultValueTypeName(f *parser.Field) (TypeName, error) {
	t, err := r.ResolveFieldTypeName(f)
	if err != nil {
		return "", err
	}
	if IsBaseType(f.Type) {
		t = t.Deref()
	}
	return t, nil
}

// GetFieldInit returns the initialization code for a field.
// The given field must have a default value.
func (r *Resolver) GetFieldInit(f *parser.Field) (Code, error) {
	return r.GetConstInit(f.Name, f.Type, f.Default)
}

// GetConstInit returns the initialization code for a constant.
func (r *Resolver) GetConstInit(name string, t *parser.Type, v *parser.ConstValue) (Code, error) {
	return r.ResolveConst(r.root, name, t, v)
}

// ResolveFieldTypeName returns a legal type name in go for the given field.
func (r *Resolver) ResolveFieldTypeName(f *parser.Field) (TypeName, error) {
	tn, err := r.GetTypeName(r.root, f.Type)
	if err != nil {
		return "", err
	}

	if NeedRedirect(f) && !checkRefInterfaceType(r.util, r.root, f.Type) {
		return "*" + tn.Deref(), nil
	}
	return tn, nil
}

// ResolveTypeName returns a legal type name in go for the given AST type.
func (r *Resolver) ResolveTypeName(t *parser.Type) (TypeName, error) {
	tn, err := r.GetTypeName(r.root, t)
	if err != nil {
		return "", err
	}
	if t.Category.IsStructLike() && !checkRefInterfaceType(r.util, r.root, t) {
		return "*" + tn, nil
	}
	return tn, nil
}

// GetTypeName returns a an type name (with selector if necessary) of the
// given type to be used in the root file.
// The type t must be a parser.Type associated with g.
func (r *Resolver) GetTypeName(g *Scope, t *parser.Type) (name TypeName, err error) {
	str, err := r.getTypeName(g, t)
	return TypeName(str), err
}

func (r *Resolver) getTypeName(g *Scope, t *parser.Type) (name string, err error) {
	if ref := t.GetReference(); ref != nil {
		g = g.includes[ref.Index].Scope
		name = g.globals.Get(ref.Name)
	} else {
		if s := baseTypes[t.Name]; s != "" {
			return s, nil
		}
		if isContainerTypes[t.Name] {
			return r.getContainerTypeName(g, t)
		}
		name = g.globals.Get(t.Name)
	}

	if name == "" {
		return "", fmt.Errorf("getTypeName failed: type[%v] file[%s]", t, g.ast.Filename)
	}

	if g.namespace != r.root.namespace {
		pkg := r.root.includeIDL(r.util, g.ast)
		name = pkg + "." + name
	}
	return
}

func (r *Resolver) getContainerTypeName(g *Scope, t *parser.Type) (name string, err error) {
	if t.Name == "map" {
		var k string
		if t.KeyType.Category == parser.Category_Binary {
			k = "string" // 'binary => string' for key type in map
		} else {
			k, err = r.getTypeName(g, t.KeyType)
			if err != nil {
				return "", fmt.Errorf("resolve key type of '%s' failed: %w", t, err)
			}
			if t.KeyType.Category.IsStructLike() && !checkRefInterfaceType(r.util, g, t.KeyType) {
				// when a struct-like is used as key of a map, it must
				// generte a pointer type instead of the struct itself
				k = "*" + k
			}
		}
		name = fmt.Sprintf("map[%s]", k)
	} else {
		name = "[]" // sets and lists compile into slices
	}

	v, err := r.getTypeName(g, t.ValueType)
	if err != nil {
		return "", fmt.Errorf("resolve value type of '%s' failed: %w", t, err)
	}

	if t.ValueType.Category.IsStructLike() && !r.util.Features().ValueTypeForSIC && !checkRefInterfaceType(r.util, g, t.ValueType) {
		v = "*" + v // generate pointer type for struct-like by default
	}
	return name + v, nil // map[k]v or []v
}

// getIDValue returns the literal representation of a const value.
// The extra must be associated with g and from a const value that has
// type parser.ConstType_ConstIdentifier.
func (r *Resolver) getIDValue(g *Scope, extra *parser.ConstValueExtra) (v string, t *parser.Type, ok bool) {
	if extra.Index == -1 {
		if extra.IsEnum {
			enum, ok := g.ast.GetEnum(extra.Sel)
			if !ok {
				return "", t, false
			}
			if en := g.Enum(enum.Name); en != nil {
				if ev := en.Value(extra.Name); ev != nil {
					v = ev.GoName().String()
					t = &parser.Type{
						Name:     enum.Name,
						Category: parser.Category_Enum,
					}
				}
			}
		} else {
			v = g.globals.Get(extra.Name)
			con, ok := g.ast.GetConstant(extra.Name)
			if !ok {
				return "", t, false
			}
			t = con.Type
		}
	} else {
		g = g.includes[extra.Index].Scope
		extra = &parser.ConstValueExtra{
			Index:  -1,
			IsEnum: extra.IsEnum,
			Name:   extra.Name,
			Sel:    extra.Sel,
		}
		return r.getIDValue(g, extra)
	}
	_, rootPkg := r.util.Import(r.root.ast)
	_, constPkg := r.util.Import(g.ast)
	if v != "" && rootPkg != constPkg {
		pkg := r.root.includeIDL(r.util, g.ast)
		v = pkg + "." + v
	}
	return v, t, v != ""
}

// ResolveConst returns the initialization code for a constant or a default value.
// The type t must be a parser.Type associated with g.
func (r *Resolver) ResolveConst(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (Code, error) {
	str, err := r.resolveConst(g, name, t, v)
	return Code(str), err
}

func (r *Resolver) resolveConst(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	switch t.Category {
	case parser.Category_Bool:
		return r.onBool(g, name, t, v)

	case parser.Category_Byte, parser.Category_I16, parser.Category_I32, parser.Category_I64:
		return r.onInt(g, name, t, v)

	case parser.Category_Double:
		return r.onDouble(g, name, t, v)

	case parser.Category_String, parser.Category_Binary:
		return r.onStrBin(g, name, t, v)

	case parser.Category_Enum:
		return r.onEnum(g, name, t, v)

	case parser.Category_Set, parser.Category_List:
		return r.onSetOrList(g, name, t, v)

	case parser.Category_Map:
		return r.onMap(g, name, t, v)

	case parser.Category_Struct, parser.Category_Union, parser.Category_Exception:
		return r.onStructLike(g, name, t, v)
	}
	return "", fmt.Errorf("type error: '%s' was declared as type %s but got value[%v] of category[%s]", name, t, v, t.Category)
}

func (r *Resolver) onBool(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	switch v.Type {
	case parser.ConstType_ConstInt:
		val := v.TypedValue.GetInt()
		return fmt.Sprint(val > 0), nil
	case parser.ConstType_ConstDouble:
		val := v.TypedValue.GetDouble()
		return fmt.Sprint(val > 0), nil
	case parser.ConstType_ConstIdentifier:
		s := v.TypedValue.GetIdentifier()
		if s == "true" || s == "false" {
			return s, nil
		}

		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", s)
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		return val, nil
	}
	return "", errTypeMissMatch(name, t, v)
}

func (r *Resolver) onInt(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
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
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", s)
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		goType, _ := r.getTypeName(g, t)
		val = fmt.Sprintf("%s(%s)", goType, val)
		return val, nil
	}
	return "", errTypeMissMatch(name, t, v)
}

func (r *Resolver) onDouble(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	switch v.Type {
	case parser.ConstType_ConstInt:
		val := v.TypedValue.GetInt()
		return fmt.Sprint(val) + ".0", nil
	case parser.ConstType_ConstDouble:
		val := v.TypedValue.GetDouble()
		return fmt.Sprint(val), nil
	case parser.ConstType_ConstIdentifier:
		s := v.TypedValue.GetIdentifier()
		if s == "true" {
			return "1.0", nil
		}
		if s == "false" {
			return "0.0", nil
		}
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", s)
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		return val, nil
	}
	return "", errTypeMissMatch(name, t, v)
}

func (r *Resolver) onStrBin(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (res string, err error) {
	defer func() {
		if err == nil && t.Category == parser.Category_Binary {
			res = "[]byte(" + res + ")"
		}
	}()
	switch v.Type {
	case parser.ConstType_ConstLiteral:
		raw := strings.ReplaceAll(v.TypedValue.GetLiteral(), "\"", "\\\"")
		return fmt.Sprintf(`"%s"`, raw), nil
	case parser.ConstType_ConstIdentifier:
		s := v.TypedValue.GetIdentifier()
		if s == "true" || s == "false" {
			break
		}

		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", s)
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		return val, nil
	default:
	}
	return "", errTypeMissMatch(name, t, v)
}

func (r *Resolver) onEnum(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	switch v.Type {
	case parser.ConstType_ConstInt:
		return fmt.Sprintf("%d", v.TypedValue.GetInt()), nil
	case parser.ConstType_ConstIdentifier:
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", v.TypedValue.GetIdentifier())
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		return val, nil
	}
	return "", fmt.Errorf("expect const value for %q is a int or enum, got %+v", name, v)
}

func (r *Resolver) onSetOrList(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	goType, err := r.getTypeName(g, t)
	if err != nil {
		return "", err
	}
	var ss []string
	switch v.Type {
	case parser.ConstType_ConstList:
		elemName := "element of " + name
		for _, elem := range v.TypedValue.GetList() {
			str, err := r.resolveConst(g, elemName, t.ValueType, elem)
			if err != nil {
				return "", err
			}
			ss = append(ss, str+",")
		}
		if len(ss) == 0 {
			return goType + "{}", nil
		}
		return fmt.Sprintf("%s{\n%s\n}", goType, strings.Join(ss, "\n")), nil

	case parser.ConstType_ConstIdentifier:
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", v.TypedValue.GetIdentifier())
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		if val != "true" && val != "false" {
			return val, nil
		}

	}
	// fault tolerance
	return goType + "{}", nil
}

func (r *Resolver) onMap(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	goType, err := r.getTypeName(g, t)
	if err != nil {
		return "", err
	}
	var kvs []string
	switch v.Type {
	case parser.ConstType_ConstMap:
		for _, mcv := range v.TypedValue.Map {
			keyName := "key of " + name
			key, err := r.resolveConst(g, keyName, r.bin2str(t.KeyType), mcv.Key)
			if err != nil {
				return "", err
			}
			valName := "value of " + name
			val, err := r.resolveConst(g, valName, t.ValueType, mcv.Value)
			if err != nil {
				return "", err
			}
			kvs = append(kvs, fmt.Sprintf("%s: %s,", key, val))
		}
		if len(kvs) == 0 {
			return goType + "{}", nil
		}
		return fmt.Sprintf("%s{\n%s\n}", goType, strings.Join(kvs, "\n")), nil

	case parser.ConstType_ConstIdentifier:
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", v.TypedValue.GetIdentifier())
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		if val != "true" && val != "false" {
			return val, nil
		}
	}
	// fault tolerance
	return goType + "{}", nil
}

func (r *Resolver) onStructLike(g *Scope, name string, t *parser.Type, v *parser.ConstValue) (string, error) {
	goType, err := r.getTypeName(g, t)
	if err != nil {
		return "", err
	}
	if v.Type == parser.ConstType_ConstIdentifier {
		val, cate, ok := r.getIDValue(g, v.Extra)
		if !ok {
			return "", fmt.Errorf("undefined value: %q", v.TypedValue.GetIdentifier())
		}
		if err := r.typeMatch(t, cate, name); err != nil {
			return "", err
		}
		if val != "true" && val != "false" {
			return val, nil
		}
	}

	if v.Type != parser.ConstType_ConstMap {
		// constant value of a struct-like must be a map literal
		return "", errTypeMissMatch(name, t, v)
	}

	// get the target struct-like with typedef dereferenced
	file, st, err := r.getStructLike(g, t)
	if err != nil {
		return "", err
	}

	var kvs []string
	for _, mcv := range v.TypedValue.Map {
		if mcv.Key.Type != parser.ConstType_ConstLiteral {
			return "", fmt.Errorf("expect literals as keys in default value of struct type '%s', got '%s'", name, mcv.Key.Type)
		}
		n := mcv.Key.TypedValue.GetLiteral()

		f, ok := st.GetField(n)
		if !ok {
			return "", fmt.Errorf("field %q not found in %q (%q): %v",
				n, st.Name, file.ast.Filename, v,
			)
		}
		typ, err := r.getTypeName(file, f.Type)
		if err != nil {
			return "", fmt.Errorf("get type name of %q in %q (%q): %w",
				n, st.Name, file.ast.Filename, err,
			)
		}

		key := file.StructLike(st.Name).Field(f.Name).GoName().String()
		val, err := r.resolveConst(file, st.Name+"."+f.Name, f.Type, mcv.Value)
		if err != nil {
			return "", err
		}

		if NeedRedirect(f) {
			if IsBaseType(f.Type) {
				// a trick to create pointers without temporary variables
				val = fmt.Sprintf("(&struct{x %s}{%s}).x", typ, val)
			}
			if !strings.HasPrefix(val, "&") {
				val = "&" + val
			}
		}
		kvs = append(kvs, fmt.Sprintf("%s: %s,", key, val))
	}
	if len(kvs) == 0 {
		return "&" + goType + "{}", nil
	}
	return fmt.Sprintf("&%s{\n%s\n}", goType, strings.Join(kvs, "\n")), nil
}

func (r *Resolver) getStructLike(g *Scope, t *parser.Type) (f *Scope, s *parser.StructLike, err error) {
	ast, x, err := semantic.Deref(g.ast, t)
	if err != nil {
		err = fmt.Errorf("expect %q a typedef or struct-like in %q: %w",
			t.Name, g.ast.Filename, err)
		return nil, nil, err
	}
	if ast == g.ast {
		f = g
	} else {
		if f = r.util.scopeCache[ast]; f == nil {
			panic(fmt.Errorf("%q not build", ast.Filename))
		}
	}
	for _, y := range ast.GetStructLikes() {
		if x.Name == y.Name {
			s = y
		}
	}
	if s == nil {
		err = fmt.Errorf("expect %q a struct-like in %q: not found: %v",
			x.Name, ast.Filename, x == t)
		return nil, nil, err
	}
	return
}

func (r *Resolver) typeMatch(field *parser.Type, value *parser.Type, name string) error {
	if field.Category.IsBool() {
		if !value.Category.IsBool() {
			return fmt.Errorf("type of %s is not bool type", name)
		}
		return nil
	}
	if field.Category.IsInteger() {
		if !value.Category.IsDigital() {
			return fmt.Errorf("type of %s is not digital type", name)
		}
		return nil
	}
	if field.Category.IsDouble() {
		if !value.Category.IsDouble() {
			return fmt.Errorf("type of %s is not double type", name)
		}
		return nil
	}
	if field.Category.IsString() {
		if !value.Category.IsString() {
			return fmt.Errorf("type of %s is not string type", name)
		}
		return nil
	}
	if field.Category.IsBinary() {
		if !value.Category.IsString() && !value.Category.IsBinary() {
			return fmt.Errorf("type of %s is not string or binary type", name)
		}
		return nil
	}
	if field.Category.IsEnum() {
		if !value.Category.IsEnum() {
			return fmt.Errorf("type of %s is not enum type", name)
		}
		if field.NameWithReference() != value.NameWithReference() {
			return fmt.Errorf("enum type of %s is not %s, %s", name, field.NameWithReference(), value.NameWithReference())
		}
		return nil
	}
	if field.Category.IsSet() {
		if !value.Category.IsSet() {
			return fmt.Errorf("type of %s is not set type", name)
		}
		return r.typeMatch(field.ValueType, value.ValueType, name)
	}
	if field.Category.IsList() {
		if !value.Category.IsList() && !value.Category.IsSet() {
			return fmt.Errorf("type of %s is not set or list type", name)
		}
		return r.typeMatch(field.ValueType, value.ValueType, name)
	}
	if field.Category.IsMap() {
		if !value.Category.IsMap() {
			return fmt.Errorf("type of %s is not map type", name)
		}
		if err := r.typeMatch(field.KeyType, value.KeyType, name); err != nil {
			return err
		}
		return r.typeMatch(field.ValueType, value.ValueType, name)
	}
	if field.Category.IsStruct() {
		if !value.Category.IsStruct() {
			return fmt.Errorf("type of %s is not struct type", name)
		}
		if field.NameWithReference() != value.NameWithReference() {
			return fmt.Errorf("type of %s is not %s", name, field.NameWithReference())
		}
		return nil
	}
	if field.Category.IsUnion() {
		if !value.Category.IsUnion() {
			return fmt.Errorf("type of %s is not union type", name)
		}
		if field.NameWithReference() != value.NameWithReference() {
			return fmt.Errorf("type of %s is not %s", name, field.NameWithReference())
		}
		return nil
	}
	if field.Category.IsException() {
		if !value.Category.IsException() {
			return fmt.Errorf("type of %s is not exception type", name)
		}
		if field.NameWithReference() != value.NameWithReference() {
			return fmt.Errorf("type of %s is not %s", name, field.NameWithReference())
		}
		return nil
	}
	return fmt.Errorf("type of %s not matched %s", name, field.NameWithReference())
}

func (r *Resolver) bin2str(t *parser.Type) *parser.Type {
	if t.Category == parser.Category_Binary {
		r := *t
		r.Category = parser.Category_String
		return &r
	}
	return t
}

// getRefTypeAnnotation returns annotations of t's referring type
func getRefTypeAnnotation(g *Scope, t *parser.Type) parser.Annotations {
	ref := t.GetReference()
	if ref == nil {
		return nil
	}
	refScope := g.includes[ref.Index].Scope
	switch t.Category {
	case parser.Category_Constant:
		if c := refScope.Constant(ref.Name); c != nil {
			return c.Annotations
		}
	case parser.Category_Enum:
		if e := refScope.Enum(ref.Name); e != nil {
			return e.Annotations
		}
	case parser.Category_Struct:
		fallthrough
	case parser.Category_Union:
		fallthrough
	case parser.Category_Exception:
		if st := refScope.StructLike(ref.Name); st != nil {
			return st.Annotations
		}
	case parser.Category_Typedef:
		if def := refScope.Typedef(ref.Name); def != nil {
			return def.Annotations
		}
	}
	return nil
}

// isRefInterfaceType verifies whether t refers Interface type
func isRefInterfaceType(g *Scope, t *parser.Type) bool {
	annos := getRefTypeAnnotation(g, t)
	return annotationContainsTrue(annos, interfaceAnnotation)
}

// checkRefInterfaceType checks whether EnableRefInterface feature has been set and t refers Interface type
func checkRefInterfaceType(cu *CodeUtils, g *Scope, t *parser.Type) bool {
	return cu.Features().EnableRefInterface && isRefInterfaceType(g, t)
}
