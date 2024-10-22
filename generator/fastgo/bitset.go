/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fastgo

// bitsetCodeGen ...
// it's used by generate required fields bitset
type bitsetCodeGen struct {
	varname  string
	typename string
	varbits  uint

	i uint
	m map[interface{}]uint
}

// newBitsetCodeGen ...
// varname - definition name of a bitset
// typename - it generates `var $varname [N]$typename`
func newBitsetCodeGen(varname, typename string) *bitsetCodeGen {
	ret := &bitsetCodeGen{
		varname:  varname,
		typename: typename,
		m:        map[interface{}]uint{},
	}
	switch typename {
	case "byte", "uint8":
		ret.varbits = 8
	case "uint16":
		ret.varbits = 16
	case "uint32":
		ret.varbits = 32
	case "uint64":
		ret.varbits = 64
	default:
		panic(typename)
	}
	return ret
}

// Add adds a `v` to bitsetCodeGen. `v` must be uniq.
// it will be used by `GenSetbit` and `GenIfNotSet`
func (g *bitsetCodeGen) Add(v interface{}) {
	_, ok := g.m[v]
	if ok {
		panic("duplicated")
	}
	g.m[v] = g.i
	g.i++
}

// Len ...
func (g *bitsetCodeGen) Len() int {
	return len(g.m)
}

// GenVar generates the definition of a bitset
// if generates nothing if Add not called
func (g *bitsetCodeGen) GenVar(w *codewriter) {
	if g.i == 0 {
		return
	}
	bits := g.varbits
	if g.i <= bits {
		w.f("var %s %s", g.varname, g.typename)
		return
	}
	w.f("var %s [%d]%s", g.varname, (g.i+bits-1)/bits, g.typename)
}

func (g *bitsetCodeGen) bitvalue(i uint) uint64 {
	i = i % g.varbits
	return 1 << uint64(i)
}

func (g *bitsetCodeGen) bitsvalue(n uint) uint64 {
	if n > g.varbits {
		panic(n)
	}
	ret := uint64(0)
	for i := uint(0); i < n; i++ {
		ret |= 1 << uint64(i)
	}
	return ret
}

// GenSetbit generates setbit code for v, vmust be added to bitsetCodeGen
func (g *bitsetCodeGen) GenSetbit(w *codewriter, v interface{}) {
	i, ok := g.m[v]
	if !ok {
		panic("[BUG] unknown v?")
	}
	if g.i <= g.varbits {
		w.f("%s |= 0x%x", g.varname, g.bitvalue(i))
	} else {
		w.f("%s[%d] |= 0x%x", g.varname, i/g.varbits, g.bitvalue(i))
	}
}

// GenIfNotSet generates `if` code for each v
func (g *bitsetCodeGen) GenIfNotSet(w *codewriter, f func(w *codewriter, v interface{})) {
	if len(g.m) == 0 {
		return
	}
	m := make(map[uint]interface{})
	for k, v := range g.m {
		m[v] = k
	}
	if g.i <= g.varbits {
		if g.i > g.varbits/2 {
			w.f("if %s != 0x%x {", g.varname, g.bitsvalue(g.i))
			defer w.f("}")
		}
		for i := uint(0); i < g.i; i++ {
			w.f("if %s & 0x%x == 0 {", g.varname, g.bitvalue(i))
			f(w, m[i])
			w.f("}")
		}
		return
	}
	i := uint(0)
	for i+g.varbits < g.i {
		w.f("if %s[%d] !=  0x%x {", g.varname, i/g.varbits, g.bitsvalue(g.varbits))
		end := i + g.varbits
		for ; i < end; i++ {
			w.f("if %s[%d] & 0x%x == 0 {", g.varname, i/g.varbits, g.bitvalue(i))
			f(w, m[i])
			w.f("}")
		}
		w.f("}")
	}
	if i < g.i {
		if g.i%g.varbits > g.varbits/2 {
			w.f("if %s[%d] != 0x%x {", g.varname, i/g.varbits, g.bitsvalue(g.i%g.varbits))
			defer w.f("}")
		}
		for ; i < g.i; i++ {
			w.f("if %s[%d] & 0x%x == 0 {", g.varname, i/g.varbits, g.bitvalue(i))
			f(w, m[i])
			w.f("}")
		}
	}
}
