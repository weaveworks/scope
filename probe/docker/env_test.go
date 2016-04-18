package docker_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	_ "github.com/weaveworks/scope/test"
)

func TestEnv(t *testing.T) {
	given := []string{
		"TERM=vt200",
		"SHELL=/bin/ksh",
		"FOO1=\"foo=bar\"",
		"FOO2",
	}
	nmd := report.MakeNode()

	nmd = docker.AddEnv(nmd, given)
	have := docker.ExtractEnv(nmd)

	if "vt200" != have["TERM"] {
		t.Errorf("Expected \"vt200\", got \"%s\"", have["TERM"])
	}

	if "/bin/ksh" != have["SHELL"] {
		t.Errorf("Expected \"/bin/ksh\", got \"%s\"", have["SHELL"])
	}

	if "\"foo=bar\"" != have["FOO1"] {
		t.Errorf("Expected \"\"foo=bar\"\", got \"%s\"", have["FOO1"])
	}

	if len(have) != 3 {
		t.Errorf("Expected only 3 items, got %d", len(have))
	}
}
