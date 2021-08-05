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
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// StructDropContext wraps a thrift.TStruct into a TStructWithoutContext.
func StructDropContext(s thrift.TStruct) TStructWithoutContext {
	var x interface{} = s // to prevent 'impossible type assertion' compile error
	switch v := x.(type) {
	case TStructWithoutContext:
		return v
	case TStructWithContext:
		return &structWithoutContext{TStructWithContext: v}
	default:
		panic(fmt.Errorf("unexpected type %T", v))
	}
}

// StructWithContext wraps a thrift.TStruct into a TStructWithContext.
func StructWithContext(s thrift.TStruct) TStructWithContext {
	var x interface{} = s // to prevent 'impossible type assertion' compile error
	switch v := x.(type) {
	case TStructWithoutContext:
		return &structWithContext{TStructWithoutContext: v}
	case TStructWithContext:
		return v
	default:
		panic(fmt.Errorf("unexpected type %T", v))
	}
}

// TStructWithContext the new-style thrift.TStruct.
type TStructWithContext interface {
	Write(ctx context.Context, p thrift.TProtocol) error
	Read(ctx context.Context, p thrift.TProtocol) error
}

// TStructWithoutContext the old-style thrift.TStruct.
type TStructWithoutContext interface {
	Write(p thrift.TProtocol) error
	Read(p thrift.TProtocol) error
}

type structWithContext struct {
	TStructWithoutContext
}

func (swc *structWithContext) Write(ctx context.Context, p thrift.TProtocol) error {
	return swc.TStructWithoutContext.Write(p)
}
func (swc *structWithContext) Read(ctx context.Context, p thrift.TProtocol) error {
	return swc.TStructWithoutContext.Read(p)
}

type structWithoutContext struct {
	TStructWithContext
}

func (swc *structWithoutContext) Write(p thrift.TProtocol) error {
	return swc.TStructWithContext.Write(ctx, p)
}
func (swc *structWithoutContext) Read(p thrift.TProtocol) error {
	return swc.TStructWithContext.Read(ctx, p)
}
