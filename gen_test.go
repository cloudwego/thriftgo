// Copyright 2022 CloudWeGo Authors
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
	"fmt"
	"io"
	"os"
	"strings"
)

func init() {
	idl = os.Getenv("IDL")
	if idl != "" {
		f, err := os.OpenFile(idl, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			panic(fmt.Errorf("open file %q: %w", idl, err))
		}
		ds := Build()
		writeIDL(f, ds)
	}
}

type Declaration interface {
	Type() string
	Name() string
}

type Definition interface {
	Print(io.Writer)
}

type Typedef struct {
	TypeName string
	Original Declaration
}

func (t *Typedef) Type() string { return t.TypeName }
func (t *Typedef) Name() string { return t.TypeName }
func (t *Typedef) Print(out io.Writer) {
	fmt.Fprintf(out, "typedef %s %s\n", t.Original.Type(), t.Name())
}

type BaseType struct {
	TypeName string
}

func (bt *BaseType) Type() string { return bt.TypeName }
func (bt *BaseType) Name() string { return strings.Title(bt.TypeName) }

type Enum struct {
	TypeName string
	Values   []int
}

func (e *Enum) Type() string { return e.TypeName }
func (e *Enum) Name() string { return strings.Title(e.TypeName) }
func (e *Enum) Print(out io.Writer) {
	fmt.Fprintf(out, "enum %s {\n", e.TypeName)
	last := -1
	for idx, v := range e.Values {
		i := idx + 1
		if v != last+1 {
			fmt.Fprintf(out, "    %s%d = %d,\n", e.Name(), i, v)
		} else {
			fmt.Fprintf(out, "    %s%d,\n", e.Name(), i)
		}
		last = v
	}
	fmt.Fprintf(out, "}\n")
}

type Container struct {
	TypeName string // map set list
	KeyType  Declaration
	ValType  Declaration
}

func (c *Container) Type() string {
	var ss []string
	if c.TypeName == "map" {
		ss = append(ss, c.KeyType.Type())
	}
	ss = append(ss, c.ValType.Type())
	return fmt.Sprintf("%s<%s>", c.TypeName, strings.Join(ss, ","))
}

func (c *Container) Name() string {
	ss := []string{c.TypeName}
	if c.TypeName == "map" {
		ss = append(ss, c.KeyType.Name())
	}
	ss = append(ss, c.ValType.Name())
	for i := range ss {
		ss[i] = strings.Title(ss[i])
	}
	return strings.Join(ss, "")
}

type StructLike struct {
	Category string // struct, union, exception
	TypeName string
	Prefix   string
	Fields   []Declaration
}

func (s *StructLike) Type() string { return s.TypeName }
func (s *StructLike) Name() string { return s.TypeName }
func (s *StructLike) Print(out io.Writer) {
	fmt.Fprintf(out, "%s %s {\n", s.Category, s.TypeName)
	for i, f := range s.Fields {
		t, n := f.Type(), s.Prefix+f.Name()
		if s.Category == "union" {
			fmt.Fprintf(out, "    %d: %s %s\n", i+1, t, n)
		} else {
			fmt.Fprintf(out, "    %d: %s %sDef\n", i*3+1, t, n)
			fmt.Fprintf(out, "    %d: required %s %sReq\n", i*3+2, t, n)
			fmt.Fprintf(out, "    %d: optional %s %sOpt\n", i*3+3, t, n)
		}
	}
	fmt.Fprintf(out, "}\n")
}

type Function struct {
	Name     string
	Response Declaration
	Requests []Declaration
	Throws   []Declaration
	Oneway   bool
}

func (f *Function) Print(out io.Writer) {
	fmt.Fprint(out, "    ") // indent
	var res string
	if f.Response == nil {
		if f.Oneway {
			res = "oneway void"
		} else {
			res = "void"
		}
	} else {
		res = f.Response.Type()
	}
	var rs []string
	for i, d := range f.Requests {
		rs = append(rs,
			fmt.Sprintf("%d: %s r%d", i+1, d.Type(), i+1))
	}

	fmt.Fprintf(out, "%s %s(%s)", res, f.Name, strings.Join(rs, ", "))
	var ts []string
	for i, t := range f.Throws {
		ts = append(ts,
			fmt.Sprintf("%d: %s e%d", i+1, t.Type(), i+1))
	}
	if !f.Oneway && len(ts) > 0 {
		fmt.Fprintf(out, "throws (%s)", strings.Join(ts, ", "))
	}
	fmt.Fprint(out, "\n")
}

type Service struct {
	Base      *Service
	Name      string
	Functions []*Function
}

func (s *Service) Print(out io.Writer) {
	if s.Base != nil {
		fmt.Fprintf(out, "service %s extends %s {\n", s.Name, s.Base.Name)
	} else {
		fmt.Fprintf(out, "service %s {\n", s.Name)
	}
	for _, f := range s.Functions {
		f.Print(out)
	}
	fmt.Fprintf(out, "}\n")
}

var baseTypes = (func() []Declaration {
	// names := "bool byte i8 i16 i32 i64 double string binary"
	names := "bool byte i16 i32 i64 double string binary"
	var bts []Declaration
	for _, t := range strings.Split(names, " ") {
		bts = append(bts, &BaseType{t})
	}
	return bts
})()

// construct container types from baseTypes and the given declarations.
func permutation(elem []Declaration) (res []Declaration) {
	res = elem
	for i := range elem {
		res = append(res, &Container{TypeName: "list", ValType: elem[i]})
	}
	for i := range elem {
		res = append(res, &Container{TypeName: "set", ValType: elem[i]})
	}
	for i := range elem {
		t := elem[i]
		if _, ok := t.(*Container); ok {
			t = baseTypes[0]
		}
		res = append(res, &Container{TypeName: "map", KeyType: t, ValType: elem[(i+1)%len(elem)]})
	}
	return
}

// permutation twice.
func permutation2(elem []Declaration) (res []Declaration) {
	res = permutation(elem)
	elem = res[len(elem):]
	elem = permutation(elem)[len(elem):]
	res = append(res, elem...)
	return
}

func typedefs(types []Declaration) (res []Declaration) {
	for _, t := range types {
		res = append(res, &Typedef{
			TypeName: "Alias" + t.Name(),
			Original: t,
		})
	}
	return
}

func buildFunctions(prefix string, ds []Declaration) (fs, gs []*Function) {
	var es, ns []Declaration
	for _, d := range ds {
		if s, ok := d.(*StructLike); ok && s.Category == "exception" {
			es = append(es, s)
		} else {
			ns = append(ns, d)
		}
	}

	var cnt int
	next := func() *Function {
		f := &Function{
			Name: fmt.Sprintf("%s%d", prefix, cnt),
		}
		cnt++
		return f
	}

	clone := func(f *Function) *Function {
		g := *f
		g.Name = fmt.Sprintf("%s%d", prefix, cnt)
		cnt++
		return &g
	}

	for i := 0; i < len(ns); i++ {
		f1 := next()
		f1.Requests = ns[i : i+1]
		if i > 0 {
			f1.Response = ns[i-1]
		}
		fs = append(fs, f1)
		if i+1 < len(ns) {
			f2 := clone(f1)
			f2.Requests = ns[i:]
			if i == 0 {
				f2.Oneway = true
			}
			fs = append(fs, f2)
		}
	}

	for i := 0; i < len(es); i++ {
		for _, f := range fs {
			g1 := clone(f)
			g1.Throws = es[i : i+1]
			gs = append(gs, g1)
			if i+1 < len(es) {
				g2 := clone(f)
				g2.Throws = es[i:]
				gs = append(gs, g2)
			}
		}
	}
	return fs, gs
}

func buildServices(ds []Declaration) (ss []Definition) {
	sa := &Service{Name: "SA"}
	sb := &Service{Name: "SB", Base: sa}
	sc := &Service{Name: "SC", Base: sb}

	sb.Functions, sc.Functions = buildFunctions("f", ds)
	return append(ss, sa, sb, sc)
}

// Build .
func Build() (ds []Definition) {
	v := &Enum{
		TypeName: "Enum",
		Values:   []int{1, 2, 3},
	}

	var bs, fs, ts []Declaration

	bs = append(baseTypes, v)
	fs = permutation(bs)
	u := &StructLike{
		Category: "union",
		TypeName: "Union",
		Prefix:   "u",
		Fields:   fs,
	}
	s := &StructLike{
		Category: "struct",
		TypeName: "Struct",
		Prefix:   "s",
		Fields:   fs,
	}
	e := &StructLike{
		Category: "exception",
		TypeName: "Exception",
		Prefix:   "e",
		Fields:   fs,
	}
	bs = append(bs, u, s, e)
	ts = typedefs(bs)
	fs = permutation2(append(bs, ts...))
	x := &StructLike{
		Category: "struct",
		TypeName: "Complex",
		Prefix:   "x",
		Fields:   fs,
	}
	for _, t := range ts {
		ds = append(ds, t.(Definition))
	}
	xx := []Declaration{
		bs[0], bs[1], bs[4], bs[6], bs[7], bs[8],
		v, u, s, e, x,
	}
	ds = append(ds, v, u, s, e, x)
	ds = append(ds, buildServices(xx)...)
	return ds
}

func writeIDL(out io.Writer, ds []Definition) {
	fmt.Fprintf(out, "// Generated by gen_test.go in github.com/cloudwego/thriftgo\n")
	fmt.Fprintf(out, "namespace * tests\n\n")
	for _, d := range ds {
		d.Print(out)
		fmt.Fprint(out, "\n")
	}
}
