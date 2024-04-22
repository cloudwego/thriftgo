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

package fieldmask

// import (
// 	"testing"

// 	"github.com/cloudwego/thriftgo/internal/test_util"
// 	"github.com/cloudwego/thriftgo/plugin"
// )

// func TestGen(t *testing.T) {
// 	g, r := test_util.GenerateGolang("a.thrift", "gen-old/", nil, nil)
// 	if err := g.Persist(r); err != nil {
// 		panic(err)
// 	}
// 	g, r = test_util.GenerateGolang("a.thrift", "gen-new/", []plugin.Option{
// 		{"with_field_mask", ""},
// 		{"with_reflection", ""},
// 	}, nil)
// 	if err := g.Persist(r); err != nil {
// 		panic(err)
// 	}
// 	g, r = test_util.GenerateGolang("b.thrift", "gen-halfway/", []plugin.Option{
// 		{"with_field_mask", ""},
// 		{"field_mask_halfway", ""},
// 		{"with_reflection", ""},
// 	}, nil)
// 	if err := g.Persist(r); err != nil {
// 		panic(err)
// 	}
// 	g, r = test_util.GenerateGolang("b.thrift", "gen-zero/", []plugin.Option{
// 		{"with_field_mask", ""},
// 		{"field_mask_halfway", ""},
// 		{"with_reflection", ""},
// 		{"field_mask_zero_required", ""},
// 	}, nil)
// 	if err := g.Persist(r); err != nil {
// 		panic(err)
// 	}
// }
