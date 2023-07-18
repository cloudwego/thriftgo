// Copyright 2023 CloudWeGo Authors
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

package utils

import (
	"testing"
)

func TestParseKV(t *testing.T) {
	// basic test
	input := "{k1:v1 k2:[{kk1:vv1 kkk1:vvv1},{kk2:vv2}] k3:v3 k4:v4 k5:{kkkk1:kvvvv1}}"
	kv, err := ParseKV(input)
	assert(t, err == nil && len(kv) == 5)
	assert(t, kv["k1"] == "v1")
	assert(t, kv["k2"] == "[{kk1:vv1 kkk1:vvv1},{kk2:vv2}]")
	assert(t, kv["k3"] == "v3")
	assert(t, kv["k4"] == "v4")
	assert(t, kv["k5"] == "{kkkk1:kvvvv1}")

	// space test
	input = "{k1:v1 \n  k2:[{kk1:vv1 kkk1:vvv1},{kk2:v  v2}] k3 : v3  \t k4: v4 k5:{  kkkk1 :kvvvv1}}"
	// input := "{k2:[{kk1:vv1 kkk1:vvv1},{kk2:v  v2}] k3 : v3}"
	kv, err = ParseKV(input)
	assert(t, err == nil && len(kv) == 5)
	assert(t, kv["k1"] == "v1")
	assert(t, kv["k2"] == "[{kk1:vv1 kkk1:vvv1},{kk2:v v2}]")
	assert(t, kv["k3"] == "v3")
	assert(t, kv["k4"] == "v4")
	assert(t, kv["k5"] == "{kkkk1:kvvvv1}")

	// illegal test
	input = "{k1:}"
	kv, err = ParseKV(input)
	assert(t, err != nil)
	input = "{k1:v1 k2:}"
	kv, err = ParseKV(input)
	assert(t, err != nil)
	input = "{k1:v1 k2}"
	kv, err = ParseKV(input)
	assert(t, err != nil)
	input = "{k1}"
	kv, err = ParseKV(input)
	assert(t, err != nil)

	// simple test
	input = "{k1:\"v2\"}"
	kv, err = ParseKV(input)
	assert(t, err == nil)

	input = "{k1:\"v2}"
	kv, err = ParseKV(input)
	assert(t, err != nil)

	input = "{k1:'v2\"}"
	kv, err = ParseKV(input)
	assert(t, err != nil)

	// bracket test
	input = "{k1:v2{}"
	kv, err = ParseKV(input)
	assert(t, err != nil)

	input = "{\n        valuemap:{k1:v1 k2:v2 k3:}\n        valuelist:[a,b,c,d]\n        valueset:[{email:e1},{email:e2}]\n        valuelistset:[[a,b,c],[d,e,f]]\n        valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]\n        valuemapStruct:[k1:{email:e1} k2:{email:e2}]\n    }"
	kv, err = ParseKV(input)
	assert(t, err == nil)
}

func TestParseArr(t *testing.T) {
	input := "[a,b,c]"
	arr, err := ParseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "b")
	assert(t, arr[2] == "c")

	input = "[a,\"b\",c]"
	arr, err = ParseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "\"b\"")
	assert(t, arr[2] == "c")

	input = "[a,'b,c]"
	arr, err = ParseArr(input)
	assert(t, err != nil)

	input = "[a,[b,c],c]"
	arr, err = ParseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,c]")
	assert(t, arr[2] == "c")

	input = "[a,[b,{c,d}],c,{e,f}]"
	arr, err = ParseArr(input)
	assert(t, err == nil && len(arr) == 4)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,{c,d}]")
	assert(t, arr[2] == "c")
	assert(t, arr[3] == "{e,f}")

	input = "[a ,[  b , {c, d} ],\nc,\t{e,f} ]"
	arr, err = ParseArr(input)
	assert(t, err == nil && len(arr) == 4)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,{c,d}]")
	assert(t, arr[2] == "c")
	assert(t, arr[3] == "{e,f}")
}

func assert(t *testing.T, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			val = append([]interface{}{"assertion failed:"}, val...)
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}
