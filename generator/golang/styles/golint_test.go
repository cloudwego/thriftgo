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

package styles

import (
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestGoLint(t *testing.T) {
	g := new(GoLint)

	test.Assert(t, g.convertName("foo_bar") == "FooBar")
	test.Assert(t, g.convertName("foo_bar_baz") == "FooBarBaz")
	test.Assert(t, g.convertName("Foo_bar") == "FooBar")
	test.Assert(t, g.convertName("foo_WiFi") == "FooWiFi")
	test.Assert(t, g.convertName("id") == "ID")
	test.Assert(t, g.convertName("Id") == "ID")
	test.Assert(t, g.convertName("foo_id") == "FooID")
	test.Assert(t, g.convertName("fooId") == "FooID")
	test.Assert(t, g.convertName("fooUid") == "FooUID")
	test.Assert(t, g.convertName("idFoo") == "IDFoo")
	test.Assert(t, g.convertName("uidFoo") == "UIDFoo")
	test.Assert(t, g.convertName("midIdDle") == "MidIDDle")
	test.Assert(t, g.convertName("APIProxy") == "APIProxy")
	test.Assert(t, g.convertName("ApiProxy") == "APIProxy")
	test.Assert(t, g.convertName("apiProxy") == "APIProxy")
	test.Assert(t, g.convertName("_Leading") == "_Leading")
	test.Assert(t, g.convertName("___Leading") == "_Leading")
	test.Assert(t, g.convertName("trailing_") == "Trailing")
	test.Assert(t, g.convertName("trailing___") == "Trailing")
	test.Assert(t, g.convertName("a_b") == "AB")
	test.Assert(t, g.convertName("a__b") == "AB")
	test.Assert(t, g.convertName("a___b") == "AB")
	test.Assert(t, g.convertName("Rpc1150") == "RPC1150")
	test.Assert(t, g.convertName("case3_1") == "Case3_1")
	test.Assert(t, g.convertName("case3__1") == "Case3_1")
	test.Assert(t, g.convertName("IEEE802_16bit") == "IEEE802_16bit")
	test.Assert(t, g.convertName("IEEE802_16Bit") == "IEEE802_16Bit")
}
