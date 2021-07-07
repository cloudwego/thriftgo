// Copyright 2021 CloudWeGo Authors
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

package parser

import (
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestAnnotationsAppend(t *testing.T) {
	var as Annotations
	pairs := []struct {
		key   string
		value string
	}{
		{
			key:   "0",
			value: "zero",
		},
		{
			key:   "1",
			value: "one",
		},
	}
	for _, pair := range pairs {
		as.Append(pair.key, pair.value)
	}
	for i, anno := range as {
		test.Assert(t, pairs[i].key == anno.Key)
		test.Assert(t, pairs[i].value == anno.Values[0])
	}
}
