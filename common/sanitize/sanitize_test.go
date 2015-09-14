package sanitize_test

import (
	"testing"

	"github.com/weaveworks/scope/common/sanitize"
)

func TestSanitizeURL(t *testing.T) {
	for _, input := range []struct {
		scheme string
		port   int
		path   string
		input  string
		want   string
	}{
		{"", 0, "", "", ""},
		{"", 0, "", "foo", "http://foo"},
		{"", 80, "", "foo", "http://foo:80"},
		{"", 0, "some/path", "foo", "http://foo/some/path"},
		{"", 0, "/some/path", "foo", "http://foo/some/path"},
		{"https://", 0, "", "foo", "https://foo"},
		{"https://", 80, "", "foo", "https://foo:80"},
		{"https://", 0, "some/path", "foo", "https://foo/some/path"},
		{"https://", 0, "", "http://foo", "http://foo"}, // specified scheme beats default...
		{"", 9999, "", "foo:80", "http://foo:80"},       // specified port beats default...
		{"", 0, "/bar", "foo/baz", "http://foo/bar"},    // ...but default path beats specified!
	} {
		if want, have := input.want, sanitize.URL(input.scheme, input.port, input.path)(input.input); want != have {
			t.Errorf("sanitize.URL(%q, %d, %q)(%q): want %q, have %q", input.scheme, input.port, input.path, input.input, want, have)
			continue
		}
	}
}
