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

package compatible

import (
	"context"
	"reflect"

	"github.com/apache/thrift/lib/go/thrift"
)

// Call wraps an invocation to the thrift.TClient's Call to discard its
// ResponseMeta return value.
func Call(ctx context.Context, cli interface{}, method string, args, result interface{}) error {
	return call(ctx, cli, method, args, result)
}

// TCaller is the old-style TClient interface.
type TCaller interface {
	Call(ctx context.Context, method string, args, result thrift.TStruct) error
}

var call func(ctx context.Context, cli interface{}, method string, args, result interface{}) error

func init() {
	var v interface{} = new(thrift.TStandardClient)
	if _, ok := v.(TCaller); ok {
		call = func(ctx context.Context, cli interface{}, method string, args, result interface{}) error {
			return cli.(TCaller).Call(ctx, method, args.(thrift.TStruct), result.(thrift.TStruct))
		}
	} else {
		// To ensure the generated code runs well with both v0.13.0 and v0.14.0
		// of the apache/thrift go library, we must not use any new type or API.
		// So to discard the ResponseMeta without knowning its type definition,
		// the only possible approach is reflection.
		call = func(ctx context.Context, cli interface{}, method string, args, result interface{}) error {
			mv := reflect.ValueOf(cli).MethodByName("Call")
			rs := mv.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(method),
				reflect.ValueOf(StructWithContext(args.(thrift.TStruct))),
				reflect.ValueOf(StructWithContext(result.(thrift.TStruct))),
			})
			if err := rs[len(rs)-1].Interface(); err != nil {
				return err.(error)
			}
			return nil
		}
	}
}
