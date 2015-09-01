package proc_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/test"
)

type mockWalker struct {
	processes []proc.Process
}

func (m *mockWalker) Walk(f func(proc.Process)) error {
	for _, p := range m.processes {
		f(p)
	}
	return nil
}

func TestBasicWalk(t *testing.T) {
	var (
		procRoot = "/proc"
		procFunc = func(proc.Process) {}
	)
	if err := proc.NewWalker(procRoot).Walk(procFunc); err != nil {
		t.Fatal(err)
	}
}

func TestCache(t *testing.T) {
	processes := []proc.Process{
		{PID: 1, PPID: 0, Comm: "init"},
		{PID: 2, PPID: 1, Comm: "bash"},
		{PID: 3, PPID: 1, Comm: "apache", Threads: 2},
		{PID: 4, PPID: 2, Comm: "ping", Cmdline: "ping foo.bar.local"},
	}
	walker := &mockWalker{
		processes: processes,
	}
	cachingWalker := proc.NewCachingWalker(walker)
	err := cachingWalker.Tick()
	if err != nil {
		t.Fatal(err)
	}

	have, err := all(cachingWalker)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	walker.processes = []proc.Process{}
	have, err = all(cachingWalker)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	err = cachingWalker.Tick()
	if err != nil {
		t.Fatal(err)
	}

	have, err = all(cachingWalker)
	want := []proc.Process{}
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}

func all(w proc.Walker) ([]proc.Process, error) {
	all := []proc.Process{}
	err := w.Walk(func(p proc.Process) {
		all = append(all, p)
	})
	return all, err
}
