package docker_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestLabels(t *testing.T) {
	given = []string{
		"TERM=vt200",
		"SHELL=/bin/ksh",
		"FOO1=\"foo=bar\"",
		"FOO2",
	}
	want := map[string]string{
		"TERM":  "vt200",
		"SHELL": "/bin/ksh",
	}
	nmd := report.MakeNode()

	nmd = docker.AddEnv(nmd, want)
	have := docker.ExtractEnv(nmd)

	if !reflect.Equal("vt200", have["TERM"]) {
		t.Error(test.Diff("vt200", have["TERM"]))
	}

	if !reflect.Equal("/bin/ksh", have["SHELL"]) {
		t.Error(test.Diff("/bin/ksh", have["SHELL"]))
	}

	if !reflect.Equal("\"foo=bar\"", have["FOO1"]) {
		t.Error(test.Diff("\"foo=bar\"", have["FOO1"]))
	}

	if !reflect.Equal(nil, have["FOO2"]) {
		t.Error(test.Diff(nil, have["FOO2"]))
	}
}
