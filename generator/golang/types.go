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
	"fmt"
	"path/filepath"
	"unsafe"

	"github.com/cloudwego/thriftgo/parser"
)

// ConstSymbol associates the constant node in AST with resolved type information.
type ConstSymbol struct {
	AST  *parser.Constant
	Type *TypeSymbol
}

// TypeSymbol stands for a resolved type symbol in the AST.
type TypeSymbol struct {
	GoType  string // The go type name
	TypeID  string // Bool|Byte|I16|I32|I64|Double|String|Binary|Set|List|Map|Struct
	KeyType *TypeSymbol
	ValType *TypeSymbol
	Enum    *parser.Enum
}

// IsEnumType reports whether the symbol is an enum type.
func (ts *TypeSymbol) IsEnumType() bool {
	return ts.Enum != nil
}

// Clone returns a copy of the current TypeSymbol.
func (ts *TypeSymbol) Clone() *TypeSymbol {
	var res = *ts
	return &res
}

// TypeAlias returns a new TypeSymbol with its GoType renamed to the given name.
func (ts *TypeSymbol) TypeAlias(name string) *TypeSymbol {
	res := ts.Clone()
	res.GoType = name
	return res
}

// Scope contains the type symbols defined in a thrift IDL and references of its pkg2paths.
type Scope struct {
	ast        *parser.Thrift
	pkg        string                  // package of this IDL
	ref2scope  map[string]*Scope       // refname => scope
	ref2pkg    map[string]string       // refname => package
	pkg2path   map[string]string       // package => import path
	imports    map[string]string       // package aquired by generated code
	name2type  map[string]*TypeSymbol  // types defined in this IDL
	name2const map[string]*ConstSymbol // constants defined in this IDL
}

// NewScope creates an uninitialized scope from the given IDL.
func NewScope(ast *parser.Thrift) *Scope {
	return &Scope{
		ast:        ast,
		ref2scope:  make(map[string]*Scope),
		ref2pkg:    make(map[string]string),
		pkg2path:   make(map[string]string),
		imports:    make(map[string]string),
		name2type:  make(map[string]*TypeSymbol),
		name2const: make(map[string]*ConstSymbol),
	}
}

// Dealias returns the aliased type of a typedef.  If the type matched the id can not be found
// or is not a typedef, the return value will be nil.
func (s *Scope) Dealias(id string) *parser.Type {
	ast, name := s.ast, id
	parts := splitType(id)
	switch len(parts) {
	case 2:
		top, ok := s.ref2scope[parts[0]]
		if !ok {
			break
		}
		ast, name = top.ast, parts[1]
		fallthrough
	case 1:
		for _, t := range ast.GetTypedefs() {
			if t.Alias == name {
				return t.Type
			}
		}
	}
	return nil
}

// SearchValue searches a value with the given identifier in the current scope and its pkg2paths.
func (s *Scope) SearchValue(id string, cu *CodeUtils) (string, error) {
	if val, ok := builtinConstant[id]; ok {
		return val, nil
	}

	var results []string
	for _, parts := range splitValue(id) {
		switch len(parts) {
		case 1: // should be a const value
			if _, ok := s.name2const[id]; ok {
				res, err := cu.Identify(id)
				if err == nil {
					results = append(results, res)
				}
			}
		case 2: // should be an enum value or an external const value
			enum, value := parts[0], parts[1]
			if sym, ok := s.name2type[enum]; ok && sym.IsEnumType() {
				for _, v := range sym.Enum.Values {
					if v.Name == value {
						results = append(results, cu.MakeEnumValueName(sym.GoType, value))
					}
				}
			}
			ref, value := parts[0], parts[1]
			if top, ok := s.ref2scope[ref]; ok {
				res, err := top.SearchValue(value, cu)
				if err == nil {
					results = append(results, s.ref2pkg[ref]+"."+res)
				}
			}
		case 3: // should be an external enum value
			ref, enum, value := parts[0], parts[1], parts[2]
			top, ok := s.ref2scope[ref]
			if ok {
				res, err := top.SearchValue(enum+"."+value, cu)
				if err != nil {
					return res, nil
				}
				results = append(results, s.ref2pkg[ref]+"."+res)
			}
		}
	}
	switch len(results) {
	case 0:
		return "", fmt.Errorf(`undefined value: "%s"`, id)
	case 1:
		return results[0], nil
	default:
		return "", fmt.Errorf(`ambiguous value: "%s"`, id)
	}
}

// AddLibrary is used by the templates to add dependencies that are not known
// before the generating stage.
func (s *Scope) AddLibrary(pkg string) {
	if pkg != "" {
		if inc, ok := s.pkg2path[pkg]; ok {
			s.imports[pkg] = inc
			return
		}
		if lib, ok := libs[pkg]; ok {
			s.imports[pkg] = lib
			return
		}
		s.imports[pkg] = pkg
	}
}

func (s *Scope) addStandardLibraries(cu *CodeUtils) {
	if len(s.ast.Enums) > 0 {
		s.AddLibrary("fmt")
		if cu.Features().ScanValueForEnum {
			s.AddLibrary("driver")
			s.AddLibrary("sql")
		}
	}
	if len(s.ast.GetStructLike()) > 0 {
		s.AddLibrary("fmt")
		s.AddLibrary("thrift")
		if cu.Features().KeepUnknownFields {
			s.AddLibrary("unknown")
		}
	}
	if len(s.ast.Services) > 0 {
		s.AddLibrary("context")
		s.AddLibrary("thrift")
		for _, svc := range s.ast.Services {
			if len(svc.Functions) > 0 {
				s.AddLibrary("fmt")
			}
		}
	}
}

func (s *Scope) addImports(cu *CodeUtils) {
	process := func(parts []string) {
		if len(parts) > 1 {
			ref := parts[0]
			if s.ref2scope[ref] != nil {
				s.AddLibrary(s.ref2pkg[ref])
			}
		}
	}

	s.walkTypes(func(t *parser.Type) bool {
		// TODO: should not add "reflect" when there are set typedefs but not used in current scope
		if t.Name == "set" && cu.Features().ValidateSet && !cu.Features().GenDeepEqual {
			s.AddLibrary("reflect")
		}
		process(splitType(t.Name))
		return true
	})

	for _, svc := range s.ast.Services {
		process(splitType(svc.Extends))
	}

	s.walkValues(func(v *parser.ConstValue) bool {
		if v.Type == parser.ConstType_ConstIdentifier {
			id := v.TypedValue.GetIdentifier()
			for _, ss := range splitValue(id) {
				process(ss)
			}
		}
		return true
	})
	if cu.Features().GenDeepEqual {
		cu.SetRootScope(s)
		s.addDeepEqualImports(cu)
	}
}

func (s *Scope) addDeepEqualImports(cu *CodeUtils) {
	var addSymbolImports func(*TypeSymbol, bool)
	addTypeImports := func(t *parser.Type, isKey bool) {
		sym, err := cu.ResolveSymbolInRootScope(t)
		if err != nil {
			cu.LogFunc.Warn(fmt.Sprintf("resolve symbol failed %+v in %s", t, s.ast.Filename))
			return
		}
		addSymbolImports(sym, isKey)
	}
	addSymbolImports = func(t *TypeSymbol, isKey bool) {
		if t.KeyType != nil {
			addSymbolImports(t.KeyType, true)
		}
		if t.ValType != nil {
			addSymbolImports(t.ValType, false)
		}
		if t.TypeID == typeids.Binary && !isKey {
			s.AddLibrary("bytes")
		}
		if t.TypeID == typeids.String && !isKey {
			s.AddLibrary("strings")
		}
	}
	// add deepequal imports for StructLikes field
	var processSet func(t *parser.Type)
	processSet = func(t *parser.Type) {
		if ok, _ := cu.IsSetType(t); ok {
			addTypeImports(t.ValueType, false)
		} else {
			if t.KeyType != nil {
				processSet(t.KeyType)
			}
			if t.ValueType != nil {
				processSet(t.ValueType)
			}
		}
	}
	for _, st := range s.ast.Structs {
		for _, f := range st.Fields {
			processSet(f.Type)
		}
	}
	for _, st := range s.ast.Unions {
		for _, f := range st.Fields {
			processSet(f.Type)
		}
	}
	for _, st := range s.ast.Exceptions {
		for _, f := range st.Fields {
			processSet(f.Type)
		}
	}
	for _, svc := range s.ast.Services {
		for _, f := range svc.Functions {
			for _, arg := range f.Arguments {
				processSet(arg.Type)
			}
			processSet(f.FunctionType)
		}
	}
	// add deepequal imports for StructLikes have DeepEqual function
	for _, st := range s.ast.Structs {
		for _, f := range st.Fields {
			addTypeImports(f.Type, false)
		}
	}
	for _, st := range s.ast.Unions {
		for _, f := range st.Fields {
			addTypeImports(f.Type, false)
		}
	}
	for _, st := range s.ast.Exceptions {
		for _, f := range st.Fields {
			addTypeImports(f.Type, false)
		}
	}
	for _, svc := range s.ast.Services {
		for _, f := range svc.Functions {
			for _, arg := range f.Arguments {
				addTypeImports(arg.Type, false)
			}
			addTypeImports(f.FunctionType, false)
		}
	}
}

func (s *Scope) init(cu *CodeUtils) error {
	s.addStandardLibraries(cu)
	for _, inc := range s.ast.Includes {
		ref, pkg, pth, err := cu.parseInclude(inc)
		if err != nil {
			return err
		}

		if _, ok := s.ref2pkg[ref]; ok {
			return fmt.Errorf("includes conflict: %s", inc.Path)
		}

		if inc.Reference.GetNamespaceOrReferenceName("go") == s.ast.GetNamespaceOrReferenceName("go") {
			// same go package as the current AST
			s.ref2pkg[ref] = ""
		} else {
			if cu.packagePrefix != "" {
				pth = filepath.Join(cu.packagePrefix, pth)
			}

			tmp, idx := pkg, 0
			for s.imports[tmp] != "" || (s.pkg2path[tmp] != "" && s.pkg2path[tmp] != pth) {
				tmp = fmt.Sprintf("%s%d", pkg, idx)
				idx++
			}
			s.ref2pkg[ref] = tmp
			s.pkg2path[tmp] = pth
		}

		scope, err := cu.BuildScope(inc.Reference)
		if err != nil {
			return err
		}
		s.ref2scope[ref] = scope
	}

	if err := s.resolveTypes(cu); err != nil {
		return err
	}
	if err := s.resolveValues(cu); err != nil {
		return err
	}
	s.addImports(cu)
	return nil
}

// resolveTypes resolves all types defined in the IDL associated with the current scope.
func (s *Scope) resolveTypes(cu *CodeUtils) (ex error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				ex = e
				return
			}
			ex = fmt.Errorf("%+v", x)
		}
	}() // catch exception
	check := func(name string) {
		if _, ok := s.name2type[name]; ok {
			panic(fmt.Errorf("type names conflict: '%s'", name))
		}
	}
	must := func(s string, e error) string {
		if e != nil {
			panic(e)
		}
		return s
	}

	for _, t := range s.ast.Enums {
		check(t.Name)
		id := must(cu.Identify(t.Name))
		s.name2type[t.Name] = &TypeSymbol{
			GoType: id,
			TypeID: "I32", // thrift transport enum as int32 but in golang it is generated as int64
			Enum:   t,
		}
	}

	for _, t := range s.ast.GetStructLike() {
		check(t.Name)

		if t.Category == "union" {
			for _, f := range t.Fields {
				f.Requiredness = parser.FieldType_Optional
			}
		}

		id := must(cu.Identify(t.Name))
		s.name2type[t.Name] = &TypeSymbol{
			GoType: id,
			TypeID: "Struct",
		}
	}

	// typedefs are resolved at last because they may refer to other types in the IDL
	for _, t := range s.ast.Typedefs {
		check(t.Alias)
		sym, err := cu.ResolveSymbol(t.Type, s)
		if err != nil {
			panic(err)
		}
		id := must(cu.Identify(t.Alias))
		s.name2type[t.Alias] = sym.TypeAlias(id)
	}
	return nil
}

// resolveValues resolves all constant defined in the IDL associated with the current scope.
func (s *Scope) resolveValues(cu *CodeUtils) error {
	for _, c := range s.ast.Constants {
		sym, err := cu.ResolveSymbol(c.Type, s)
		if err != nil {
			return err
		}
		s.name2const[c.Name] = &ConstSymbol{c, sym}
	}
	return nil
}

// walkTypes walks throught the AST and visit each parser.Type node.
// The result of visit function decides whether to continue the iteration.
func (s *Scope) walkTypes(visit func(*parser.Type) bool) {
	var at, bt func(t *parser.Type)
	ch := make(chan *parser.Type)

	bt = func(t *parser.Type) { at(t) } // make at recursive
	at = func(t *parser.Type) {
		cont := visit(t)
		if !cont {
			close(ch)
		}
		switch t.Name {
		case "map":
			bt(t.KeyType)
			fallthrough
		case "set", "list":
			bt(t.ValueType)
			return
		}
	}

	go func() {
		defer func() { recover() }()
		for _, t := range s.ast.Typedefs {
			ch <- t.Type
		}
		for _, t := range s.ast.Constants {
			ch <- t.Type
		}
		for _, t := range s.ast.GetStructLike() {
			for _, f := range t.Fields {
				ch <- f.Type
			}
		}
		for _, t := range s.ast.Services {
			for _, f := range t.Functions {
				ch <- f.FunctionType
				for _, a := range f.Arguments {
					ch <- a.Type
				}
				for _, x := range f.Throws {
					ch <- x.Type
				}
			}
		}
		close(ch)
	}()

	for t := range ch {
		at(t)
	}
}

// walkValues walks throught the AST and visit each *parser.ConstValue node.
// The result of visit function decides whether to continue the iteration.
func (s *Scope) walkValues(visit func(*parser.ConstValue) bool) {
	var at, bt func(t *parser.ConstValue)
	ch := make(chan *parser.ConstValue)

	bt = func(v *parser.ConstValue) { at(v) } // make at recursive
	at = func(v *parser.ConstValue) {
		cont := visit(v)
		if !cont {
			close(ch)
		}
		switch v.Type {
		case parser.ConstType_ConstMap:
			for _, kv := range v.TypedValue.Map {
				bt(kv.Key)
				bt(kv.Value)
			}
		case parser.ConstType_ConstList:
			for _, e := range v.TypedValue.List {
				bt(e)
			}
		}
	}

	go func() {
		defer func() { recover() }()
		for _, v := range s.ast.Constants {
			if v.Value != nil {
				ch <- v.Value
			}
		}
		for _, v := range s.ast.GetStructLike() {
			for _, f := range v.Fields {
				if f.Default != nil {
					ch <- f.Default
				}
			}
		}
		close(ch)
	}()

	for v := range ch {
		at(v)
	}
}

var builtinConstant = map[string]string{
	"true": "true", "false": "false",
}

var typeids = struct {
	Bool   string
	Byte   string
	I8     string
	I16    string
	I32    string
	I64    string
	Double string
	String string
	Binary string
	Set    string
	List   string
	Map    string
	Struct string
}{
	Bool:   "Bool",
	Byte:   "Byte",
	I8:     "Byte", // i8 is byte
	I16:    "I16",
	I32:    "I32",
	I64:    "I64",
	Double: "Double",
	String: "String",
	Binary: "Binary",
	Set:    "Set",
	List:   "List",
	Map:    "Map",
	Struct: "Struct",
}

var baseTypes = map[string]*TypeSymbol{
	"bool":   {GoType: "bool", TypeID: "Bool"},
	"byte":   {GoType: "int8", TypeID: "Byte"},
	"i8":     {GoType: "int8", TypeID: "Byte"},
	"i16":    {GoType: "int16", TypeID: "I16"},
	"i32":    {GoType: "int32", TypeID: "I32"},
	"i64":    {GoType: "int64", TypeID: "I64"},
	"double": {GoType: "float64", TypeID: "Double"},
	"string": {GoType: "string", TypeID: "String"},
	"binary": {GoType: "[]byte", TypeID: "Binary"},
}

var isContainerTypes = map[string]bool{"map": true, "set": true, "list": true}

var isKeywords = map[string]bool{
	"break":       true,
	"default":     true,
	"func":        true,
	"interface":   true,
	"select":      true,
	"case":        true,
	"defer":       true,
	"go":          true,
	"map":         true,
	"struct":      true,
	"chan":        true,
	"else":        true,
	"goto":        true,
	"package":     true,
	"switch":      true,
	"const":       true,
	"fallthrough": true,
	"if":          true,
	"range":       true,
	"type":        true,
	"continue":    true,
	"for":         true,
	"import":      true,
	"return":      true,
	"var":         true,
}

// assuming the generated codes run on the same architecture as thriftgo.
const pointerSize = int(unsafe.Sizeof((*int)(nil)))

var sizeof = (func() map[string]int {
	return map[string]int{
		typeids.Bool:   1,
		typeids.Byte:   1,
		typeids.I8:     1,
		typeids.I16:    2,
		typeids.I32:    4,
		typeids.I64:    8,
		typeids.Double: 8, // float64
		typeids.String: pointerSize * 2,
		typeids.Binary: pointerSize * 3,
		typeids.Set:    pointerSize * 3,
		typeids.List:   pointerSize * 3,
		typeids.Map:    pointerSize,
		typeids.Struct: pointerSize, // as pointer
	}
})()

// align implements the structure padding algorithm of golang.
type align struct {
	unit int
	size int
}

func (a *align) add(size int) {
	unit := size
	if unit >= pointerSize {
		unit = pointerSize
	}
	if unit > a.unit {
		a.unit = unit
	}

	if pad := a.padded(); pad-a.size < size {
		a.size = pad
	}
	a.size += size
}

func (a *align) padded() int {
	return (a.size + a.unit - 1) / a.unit * a.unit
}

type sizeDiff struct {
	original int
	arranged int
}

func (d *sizeDiff) percent() float64 {
	return float64(d.arranged-d.original) / float64(d.original) * 100
}
