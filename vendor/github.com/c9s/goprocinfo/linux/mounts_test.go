package linux

import (
	"testing"
)

func TestMounts(t *testing.T) {
	mounts, err := ReadMounts("proc/mounts")
	if err != nil {
		t.Fatal("mounts read fail")
	}
	t.Logf("%+v", mounts)
	if mounts.Mounts[0].Device != "rootfs" {
		t.Fatal("unexpected value")
	}
	if mounts.Mounts[1].FSType != "proc" {
		t.Fatal("unexpected value")
	}
}
