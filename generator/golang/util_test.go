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

package golang

import "testing"

func TestSnakify(t *testing.T) {
	cases := []struct{ original, expected string }{
		{"a", "a"},
		{"A", "a"},
		{"AB", "ab"},
		{"HTTPRequest", "http_request"},
		{"HTTP1Method", "http1_method"},
		{"GetUserIP", "get_user_ip"},
	}
	for _, c := range cases {
		res := snakify(c.original)
		if res != c.expected {
			t.Logf("snakify(%q) => %q. Expected: %q", c.original, res, c.expected)
			t.Fail()
		}
	}
}
