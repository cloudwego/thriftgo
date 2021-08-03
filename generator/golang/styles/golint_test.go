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
	test.Assert(t, g.convertName("id_2_app") == "ID2App")
	test.Assert(t, g.convertName("id_2_app_2_url") == "ID2App2Url")
	test.Assert(t, g.convertName("version_2_3_4_alpha_1") == "Version2_3_4Alpha1")
	test.Assert(t, g.convertName("Id_2_App") == "ID2App")
	test.Assert(t, g.convertName("Id_2_App_2_Url") == "ID2App2URL")

	g.UseInitialisms(false)

	test.Assert(t, g.convertName("foo_bar") == "FooBar")
	test.Assert(t, g.convertName("foo_bar_baz") == "FooBarBaz")
	test.Assert(t, g.convertName("Foo_bar") == "FooBar")
	test.Assert(t, g.convertName("foo_WiFi") == "Foo_WiFi")
	test.Assert(t, g.convertName("id") == "Id")
	test.Assert(t, g.convertName("Id") == "Id")
	test.Assert(t, g.convertName("foo_id") == "FooId")
	test.Assert(t, g.convertName("fooId") == "FooId")
	test.Assert(t, g.convertName("fooUid") == "FooUid")
	test.Assert(t, g.convertName("idFoo") == "IdFoo")
	test.Assert(t, g.convertName("uidFoo") == "UidFoo")
	test.Assert(t, g.convertName("midIdDle") == "MidIdDle")
	test.Assert(t, g.convertName("APIProxy") == "APIProxy")
	test.Assert(t, g.convertName("ApiProxy") == "ApiProxy")
	test.Assert(t, g.convertName("apiProxy") == "ApiProxy")
	test.Assert(t, g.convertName("_Leading") == "__Leading")
	test.Assert(t, g.convertName("___Leading") == "____Leading")
	test.Assert(t, g.convertName("trailing_") == "Trailing_")
	test.Assert(t, g.convertName("trailing___") == "Trailing___")
	test.Assert(t, g.convertName("a_b") == "AB")
	test.Assert(t, g.convertName("a__b") == "A_B")
	test.Assert(t, g.convertName("a___b") == "A__B")
	test.Assert(t, g.convertName("Rpc1150") == "Rpc1150")
	test.Assert(t, g.convertName("case3_1") == "Case3_1")
	test.Assert(t, g.convertName("case3__1") == "Case3__1")
	test.Assert(t, g.convertName("IEEE802_16bit") == "IEEE802_16bit")
	test.Assert(t, g.convertName("IEEE802_16Bit") == "IEEE802_16Bit")
	test.Assert(t, g.convertName("id_2_app") == "Id_2App")
	test.Assert(t, g.convertName("id_2_app_2_url") == "Id_2App_2Url")
	test.Assert(t, g.convertName("version_2_3_4_alpha_1") == "Version_2_3_4Alpha_1")
	test.Assert(t, g.convertName("Id_2_App") == "Id_2_App")
	test.Assert(t, g.convertName("Id_2_App_2_Url") == "Id_2_App_2_Url")
}
