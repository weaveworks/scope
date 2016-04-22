package render

import (
	"testing"
)

func TestDockerImageName(t *testing.T) {
	for _, input := range []struct{ in, name string }{
		{"foo/bar", "foo/bar"},
		{"foo/bar:baz", "foo/bar"},
		{"reg:123/foo/bar:baz", "foo/bar"},
		{"docker-registry.domain.name:5000/repo/image1:ver", "repo/image1"},
		{"foo", "foo"},
	} {
		name := ImageNameWithoutVersion(input.in)
		if name != input.name {
			t.Fatalf("%s: %s != %s", input.in, name, input.name)
		}
	}
}
