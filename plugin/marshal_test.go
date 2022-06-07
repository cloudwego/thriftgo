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

package plugin

import (
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestMarshalUnmarshal(t *testing.T) {
	{
		isUnion := map[reflect.Type]bool{
			reflect.TypeOf((*parser.ConstTypedValue)(nil)).Elem(): true,
		}
		req1 := NewRequest()
		test.ThriftRandomFill(req1, isUnion)
		bs1, err := MarshalRequest(req1)
		test.Assert(t, err == nil && len(bs1) > 0, err, len(bs1))

		req2, err := UnmarshalRequest(bs1)
		test.Assert(t, err == nil, err)

		bs2, err := MarshalRequest(req2)
		test.Assert(t, err == nil)
		test.DeepEqual(t, bs1, bs2)
	}

	{
		res1 := NewResponse()
		test.ThriftRandomFill(res1, nil)
		bs1, err := MarshalResponse(res1)
		test.Assert(t, err == nil && len(bs1) > 0, err, len(bs1))

		res2, err := UnmarshalResponse(bs1)
		test.Assert(t, err == nil, err)

		bs2, err := MarshalResponse(res2)
		test.Assert(t, err == nil)
		test.DeepEqual(t, bs1, bs2)
	}
}
