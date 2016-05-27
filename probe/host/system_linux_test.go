package host_test

import (
	"fmt"
	"syscall"
	"testing"

	"$GITHUB_URI/probe/host"
)

func TestUname(t *testing.T) {
	oldUname := host.Uname
	defer func() { host.Uname = oldUname }()

	const (
		release = "rls"
		version = "ver"
	)
	host.Uname = func(uts *syscall.Utsname) error {
		uts.Release = string2c(release)
		uts.Version = string2c(version)
		return nil
	}

	have, err := host.GetKernelVersion()
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("%s %s", release, version); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func string2c(s string) [65]int8 {
	var result [65]int8
	for i, c := range s {
		result[i] = int8(c)
	}
	return result
}
