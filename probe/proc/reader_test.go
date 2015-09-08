package proc_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/test"
)

func TestProcReaderBasic(t *testing.T) {
	procFunc := func(proc.Process) {}
	if err := proc.NewProcReader(proc.EmptyProcDir).Processes(procFunc); err != nil {
		t.Fatal(err)
	}
}

func TestCachingProcReader(t *testing.T) {
	all := func(w proc.ProcReader) ([]proc.Process, error) {
		all := []proc.Process{}
		err := w.Processes(func(p proc.Process) {
			all = append(all, p)
		})
		return all, err
	}

	processes := []proc.Process{
		{PID: 1, PPID: 0, Comm: "init"},
		{PID: 2, PPID: 1, Comm: "bash"},
		{PID: 3, PPID: 1, Comm: "apache", Threads: 2},
		{PID: 4, PPID: 2, Comm: "ping", Cmdline: "ping foo.bar.local"},
	}
	procReader := &proc.MockedProcReader{
		Procs: processes,
	}
	cachingProcReader := proc.NewCachingProcReader(procReader, true)
	err := cachingProcReader.Tick()
	if err != nil {
		t.Fatal(err)
	}

	have, err := all(cachingProcReader)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	procReader.Procs = []proc.Process{}
	have, err = all(cachingProcReader)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	err = cachingProcReader.Tick()
	if err != nil {
		t.Fatal(err)
	}

	have, err = all(cachingProcReader)
	want := []proc.Process{}
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
