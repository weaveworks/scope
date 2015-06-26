package process_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/process"
)

func TestTree(t *testing.T) {
	walker := &mockWalker{
		processes: []*process.Process{
			{PID: 1, PPID: 0, Comm: "init"},
			{PID: 2, PPID: 1, Comm: "bash"},
			{PID: 3, PPID: 1, Comm: "apache", Threads: 2},
			{PID: 4, PPID: 2, Comm: "ping", Cmdline: "ping foo.bar.local"},
		},
	}

	tree, err := process.NewTree(walker)
	if err != nil {
		t.Fatalf("newProcessTree error: %v", err)
	}

	for pid, want := range map[int]int{2: 1, 3: 1, 4: 2} {
		have, err := tree.GetParent(pid)
		if err != nil || !reflect.DeepEqual(want, have) {
			t.Errorf("%d: want %#v, have %#v (%v)", pid, want, have, err)
		}
	}
}
