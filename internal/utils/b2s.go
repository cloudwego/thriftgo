// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import "unsafe"

type sliceHeader struct {
	p unsafe.Pointer
	l int
	c int
}

type stringHeader struct {
	p unsafe.Pointer
	l int
}

func B2S(b []byte) (s string) {
	(*stringHeader)(unsafe.Pointer(&s)).p = (*sliceHeader)(unsafe.Pointer(&b)).p
	(*stringHeader)(unsafe.Pointer(&s)).l = (*sliceHeader)(unsafe.Pointer(&b)).l
	return
}

func S2B(s string) (b []byte) {
	(*sliceHeader)(unsafe.Pointer(&b)).p = (*stringHeader)(unsafe.Pointer(&s)).p
	(*sliceHeader)(unsafe.Pointer(&b)).l = (*stringHeader)(unsafe.Pointer(&s)).l
	(*sliceHeader)(unsafe.Pointer(&b)).c = (*stringHeader)(unsafe.Pointer(&s)).l
	return
}
