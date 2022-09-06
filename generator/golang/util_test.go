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
