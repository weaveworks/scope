package process_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/process"
)

func TestTree(t *testing.T) {
	oldWalk := process.Walk
	defer func() { process.Walk = oldWalk }()

	process.Walk = func(_ string, f func(*process.Process)) error {
		for _, p := range []*process.Process{
			{PID: 1, PPID: 0},
			{PID: 2, PPID: 1},
			{PID: 3, PPID: 1},
			{PID: 4, PPID: 2},
		} {
			f(p)
		}
		return nil
	}

	tree, err := process.NewTree("foo")
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
