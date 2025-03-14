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

import (
	"fmt"
	"strings"
	"testing"
)

func TestBitsetCodeGen(t *testing.T) {
	g := newBitsetCodeGen("bitset", "uint8")

	w := newCodewriter()

	// case: less or equal than 64 elements
	g.Add(1)
	g.Add(2)
	g.GenVar(w)
	srcEqual(t, w.String(), "var bitset uint8")
	w.Reset()

	g.GenSetbit(w, 2)
	srcEqual(t, w.String(), "bitset |= 0x2")
	w.Reset()

	g.GenIfNotSet(w, func(w *codewriter, id interface{}) {
		w.f("_ = %d", id)
	})
	srcEqual(t, w.String(), `if bitset & 0x1 == 0 { _ = 1 }
		if bitset & 0x2 == 0 { _ = 2 }`)
	w.Reset()

	g.Add(3)
	g.Add(4)
	g.Add(5)
	g.Add(6)
	g.Add(7)
	g.Add(8) // case: g.i > g.varbits/2
	g.GenIfNotSet(w, func(w *codewriter, id interface{}) {
		w.f("_ = %d", id)
	})
	srcEqual(t, w.String(), `if bitset != 0xff {
		if bitset & 0x1 == 0 { _ = 1 }
		if bitset & 0x2 == 0 { _ = 2 }
		if bitset & 0x4 == 0 { _ = 3 }
		if bitset & 0x8 == 0 { _ = 4 }
		if bitset & 0x10 == 0 { _ = 5 }
		if bitset & 0x20 == 0 { _ = 6 }
		if bitset & 0x40 == 0 { _ = 7 }
		if bitset & 0x80 == 0 { _ = 8 }
	}`)
	w.Reset()

	// case: more than varbits elements
	g = newBitsetCodeGen("bitset", "uint8")
	for i := 0; i < 17; i++ {
		g.Add(i + 100)
	}
	g.GenVar(w)
	srcEqual(t, w.String(), "var bitset [3]uint8")
	w.Reset()

	g.GenSetbit(w, 100)
	srcEqual(t, w.String(), "bitset[0] |= 0x1")
	w.Reset()

	g.GenSetbit(w, 115)
	srcEqual(t, w.String(), "bitset[1] |= 0x80")
	w.Reset()

	g.GenSetbit(w, 116)
	srcEqual(t, w.String(), "bitset[2] |= 0x1")
	w.Reset()

	g.GenIfNotSet(w, func(w *codewriter, id interface{}) {
		w.f("_ = %d", id)
	})
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "if bitset[0] != 0xff {")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(sb, "if bitset[0]&0x%x == 0 { _ = %d }\n", 1<<i, 100+i)
	}
	fmt.Fprintln(sb, "}")
	fmt.Fprintln(sb, "if bitset[1] != 0xff {")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(sb, "if bitset[1]&0x%x == 0 { _ = %d }\n", 1<<i, 108+i)
	}
	fmt.Fprintln(sb, "}")
	fmt.Fprintln(sb, "if bitset[2]&0x1 == 0 { _ = 116 }")
	srcEqual(t, w.String(), sb.String())
}
