package linux

import (
	"testing"
)

func TestReadProcessCmdlineSimple(t *testing.T) {

	cmdline, err := ReadProcessCmdline("proc/3323/cmdline")

	if err != nil {
		t.Fatal("process cmdline read fail", err)
	}

	expected := "proftpd: (accepting connections)"

	if cmdline != expected {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", cmdline)
}

func TestReadProcessCmdlineComplex(t *testing.T) {

	cmdline, err := ReadProcessCmdline("proc/5811/cmdline")

	if err != nil {
		t.Fatal("process cmdline read fail", err)
	}

	expected := "/home/c9s/.config/sublime-text-2/Packages/User/GoSublime/linux-x64/bin/gosublime.margo_r14.12.06-1_go1.4.2.exe -oom 1000 -poll 30 -tag r14.12.06-1"

	if cmdline != expected {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", cmdline)
}
