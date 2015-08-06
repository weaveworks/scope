package process_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/test"
)

func TestBasicWalk(t *testing.T) {
	var (
		procRoot = "/proc"
		procFunc = func(process.Process) {}
	)
	if err := process.NewWalker(procRoot).Walk(procFunc); err != nil {
		t.Fatal(err)
	}
}

func TestCache(t *testing.T) {
	processes := []process.Process{
		{PID: 1, PPID: 0, Comm: "init"},
		{PID: 2, PPID: 1, Comm: "bash"},
		{PID: 3, PPID: 1, Comm: "apache", Threads: 2},
		{PID: 4, PPID: 2, Comm: "ping", Cmdline: "ping foo.bar.local"},
	}
	walker := &mockWalker{
		processes: processes,
	}
	cachingWalker := process.NewCachingWalker(walker)
	err := cachingWalker.Update()
	if err != nil {
		t.Fatal(err)
	}

	have, err := all(cachingWalker)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	walker.processes = []process.Process{}
	have, err = all(cachingWalker)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	err = cachingWalker.Update()
	if err != nil {
		t.Fatal(err)
	}

	have, err = all(cachingWalker)
	want := []process.Process{}
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}

func all(w process.Walker) ([]process.Process, error) {
	all := []process.Process{}
	err := w.Walk(func(p process.Process) {
		all = append(all, p)
	})
	return all, err
}
