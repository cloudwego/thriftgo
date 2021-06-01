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

package test

import "reflect"

// testingTB is a subset of common methods between *testing.T and *testing.B.
type testingTB interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// Assert .
func Assert(t testingTB, cond bool, val ...interface{}) {
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

// Assertf .
func Assertf(t testingTB, cond bool, format string, val ...interface{}) {
	t.Helper()
	if !cond {
		t.Fatalf(format, val...)
	}
}

// DeepEqual .
func DeepEqual(t testingTB, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatal("assertion failed")
	}
}

// Panic .
func Panic(t testingTB, fn func()) {
	t.Helper()
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("assertion failed: did not panic")
		}
	}()
	fn()
}
