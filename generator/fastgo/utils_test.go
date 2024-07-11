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
	"go/format"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func srcEqual(t *testing.T, a, b string) {
	_, file, line, _ := runtime.Caller(1)
	location := fmt.Sprintf("@ %s:%d", filepath.Base(file), line)
	b0, err := format.Source([]byte(a))
	if err != nil {
		t.Log("syntax err", err, a, location)
		t.FailNow()
	}
	b1, err := format.Source([]byte(b))
	if err != nil {
		t.Log("syntax err", err, a, location)
		t.FailNow()
	}
	s0 := strings.TrimSpace(string(b0))
	s1 := strings.TrimSpace(string(b1))

	test.Assert(t, s0 == s1, fmt.Sprintf("\n%s\n != \n%s\n %s", s0, s1, location))
}
