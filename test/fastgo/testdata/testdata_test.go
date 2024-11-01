//go:build testfastgo

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

// add `go:build testfastgo` for preventing build failure without generated code
// run manually: go test -v -tags testfastgo

package testdata

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func P[T any](v T) *T { return &v }

func TestTestTypes(t *testing.T) {
	p0 := &TestTypes{
		B2:      P(true),
		Byte2:   P(int8(11)),
		I802:    P(int8(12)),
		I162:    P(int16(13)),
		I322:    P(int32(14)),
		Dbl2:    P(float64(15)),
		Str2:    P("16"),
		Bin0:    []byte{},
		Bin1:    []byte{},
		Bin2:    []byte("17"),
		Num2:    P(Numberz(18)),
		UID2:    P(UserID(19)),
		Msg0:    &Msg{},
		Msg1:    &Msg{},
		Map111:  map[int32]string{},
		Map112:  map[int32]string{111: "111", 112: "112"},
		Map121:  map[int32]int32{},
		Map122:  map[int32]int32{1: 121, 2: 122},
		Map131:  map[string]*Msg{},
		Map132:  map[string]*Msg{"131": &Msg{Message: "132"}},
		List141: []int32{},
		List142: []int32{142},
		List151: []string{},
		List152: []string{"142"},
		List161: []*Msg{},
		List162: []*Msg{&Msg{Message: "162"}},
		Set171:  []int32{},
		Set172:  []int32{172},
		Set181:  []string{},
		Set182:  []string{"172"},
		Mix191:  []map[int32]int32{},
		Mix192:  []map[int32]int32{{1: 2}, {3: 4}},
		Mix201:  map[int32][]int32{},
		Mix202:  map[int32][]int32{201: []int32{202}},
	}
	sz := p0.BLength()
	b := p0.FastAppend(nil)
	require.Equal(t, sz, len(b))
	p1 := &TestTypes{}
	off, err := p1.FastRead(b)
	require.NoError(t, err)
	require.Equal(t, sz, off)
	require.Equal(t, p0, p1)
}
