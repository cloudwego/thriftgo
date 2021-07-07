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

package semantic_test

import (
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
	"github.com/cloudwego/thriftgo/semantic"
)

func TestSplitType(t *testing.T) {
	ss := semantic.SplitType("")
	test.Assert(t, len(ss) == 0)

	ss = semantic.SplitType("a")
	test.Assert(t, len(ss) == 1 && ss[0] == "a")

	ss = semantic.SplitType("a.b")
	test.Assert(t, len(ss) == 2 && ss[0] == "a" && ss[1] == "b", ss)

	ss = semantic.SplitType("a.b.c")
	test.Assert(t, len(ss) == 2 && ss[0] == "a.b" && ss[1] == "c")

	ss = semantic.SplitType("a.b.c.d")
	test.Assert(t, len(ss) == 2 && ss[0] == "a.b.c" && ss[1] == "d")
}

func TestSplitValue(t *testing.T) {
	sss := semantic.SplitValue("")
	test.Assert(t, len(sss) == 0)

	sss = semantic.SplitValue("a")
	test.Assert(t, len(sss) == 1 && len(sss[0]) == 1 && sss[0][0] == "a")

	sss = semantic.SplitValue("a.b")
	test.Assert(t, len(sss) == 1 && len(sss[0]) == 2 && sss[0][0] == "a" && sss[0][1] == "b")

	sss = semantic.SplitValue("a.b.c") // IDL<a.b>.const<c> or IDL<a>.enum<b>.value<c>
	test.Assert(t, len(sss) == 2 && len(sss[0]) == 2 && len(sss[1]) == 3)
	test.Assert(t, sss[0][0] == "a.b" && sss[0][1] == "c")
	test.Assert(t, sss[1][0] == "a" && sss[1][1] == "b" && sss[1][2] == "c")
}
