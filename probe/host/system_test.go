package host_test

import (
	"strings"
	"testing"

	"github.com/weaveworks/scope/probe/host"
)

func TestGetKernelVersion(t *testing.T) {
	have, err := host.GetKernelVersion()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(have, "unknown") {
		t.Fatal(have)
	}
	t.Log(have)
}

func TestGetLoad(t *testing.T) {
	have := host.GetLoad()
	if have == nil {
		t.Fatal(have)
	}
	t.Log(have)
}

func TestGetUptime(t *testing.T) {
	have, err := host.GetUptime()
	if err != nil {
		t.Fatal(err)
	}
	if have == 0 {
		t.Fatal(have)
	}
	t.Log(have.String())
}
