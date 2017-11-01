package host_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/probe/host"

	"golang.org/x/sys/unix"
)

func TestUname(t *testing.T) {
	oldUname := host.Uname
	defer func() { host.Uname = oldUname }()

	const (
		release = "rls"
		version = "ver"
	)
	host.Uname = func(uts *unix.Utsname) error {
		copy(uts.Release[:], []byte(release))
		copy(uts.Version[:], []byte(version))
		return nil
	}

	haveRelease, haveVersion, err := host.GetKernelReleaseAndVersion()
	if err != nil {
		t.Fatal(err)
	}
	have := fmt.Sprintf("%s %s", haveRelease, haveVersion)
	if want := fmt.Sprintf("%s %s", release, version); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
