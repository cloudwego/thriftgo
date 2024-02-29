// Copyright 2024 CloudWeGo Authors
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

package thrift_reflection

import (
	"reflect"
	"testing"
)

func Test_checkDuplicateAndRegister(t *testing.T) {
	testThrift := "test.thrift"
	testDesc := &FileDescriptor{
		Filepath: testThrift,
		Includes: map[string]string{
			"testKey": "testValue",
		},
	}
	type x struct{}
	pkgPath := reflect.TypeOf(x{}).PkgPath()
	checkDuplicateAndRegister(testDesc, pkgPath)

	// register the FileDescriptor with the same content and Filepath
	sameDesc := &FileDescriptor{
		Filepath: testThrift,
		Includes: map[string]string{
			"testKey": "testValue",
		},
	}
	checkDuplicateAndRegister(sameDesc, pkgPath)

	// register the FileDescriptor with the same Filepath and different content
	anotherDesc := &FileDescriptor{
		Filepath: testThrift,
		Includes: map[string]string{
			"anotherKey": "anotherValue",
		},
	}
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("Register FileDescriptor with the same Filepath and different content should panic")
		}
	}()
	checkDuplicateAndRegister(anotherDesc, pkgPath)
}
