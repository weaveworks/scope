package process_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/test"
)

type mockProcess struct {
	name, comm, cmdline string
}

func (p mockProcess) Name() string       { return p.name }
func (p mockProcess) Size() int64        { return 0 }
func (p mockProcess) Mode() os.FileMode  { return 0 }
func (p mockProcess) ModTime() time.Time { return time.Now() }
func (p mockProcess) IsDir() bool        { return true }
func (p mockProcess) Sys() interface{}   { return nil }

func TestWalker(t *testing.T) {
	oldReadDir, oldReadFile := process.ReadDir, process.ReadFile
	defer func() {
		process.ReadDir = oldReadDir
		process.ReadFile = oldReadFile
	}()

	processes := map[string]mockProcess{
		"3":       {name: "3", comm: "curl\n", cmdline: "curl\000google.com"},
		"2":       {name: "2", comm: "bash\n"},
		"4":       {name: "4", comm: "apache\n"},
		"notapid": {name: "notapid"},
		"1":       {name: "1", comm: "init\n"},
	}

	process.ReadDir = func(path string) ([]os.FileInfo, error) {
		result := []os.FileInfo{}
		for _, p := range processes {
			result = append(result, p)
		}
		return result, nil
	}

	process.ReadFile = func(path string) ([]byte, error) {
		splits := strings.Split(path, "/")

		pid := splits[len(splits)-2]
		process, ok := processes[pid]
		if !ok {
			return nil, fmt.Errorf("not found")
		}

		file := splits[len(splits)-1]
		switch file {
		case "comm":
			return []byte(process.comm), nil
		case "stat":
			pid, _ := strconv.Atoi(splits[len(splits)-2])
			parent := pid - 1
			return []byte(fmt.Sprintf("%d na R %d 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1", pid, parent)), nil
		case "cmdline":
			return []byte(process.cmdline), nil
		}

		return nil, fmt.Errorf("not found")
	}

	want := map[int]*process.Process{
		3: {PID: 3, PPID: 2, Comm: "curl", Cmdline: "curl google.com", Threads: 1},
		2: {PID: 2, PPID: 1, Comm: "bash", Cmdline: "", Threads: 1},
		4: {PID: 4, PPID: 3, Comm: "apache", Cmdline: "", Threads: 1},
		1: {PID: 1, PPID: 0, Comm: "init", Cmdline: "", Threads: 1},
	}

	have := map[int]*process.Process{}
	walker := process.NewWalker("unused")
	err := walker.Walk(func(p *process.Process) {
		have[p.PID] = p
	})

	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}

func TestCache(t *testing.T) {
	processes := []*process.Process{
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

	walker.processes = []*process.Process{}
	have, err = all(cachingWalker)
	if err != nil || !reflect.DeepEqual(processes, have) {
		t.Errorf("%v (%v)", test.Diff(processes, have), err)
	}

	err = cachingWalker.Update()
	if err != nil {
		t.Fatal(err)
	}

	have, err = all(cachingWalker)
	want := []*process.Process{}
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}

func all(w process.Walker) ([]*process.Process, error) {
	all := []*process.Process{}
	err := w.Walk(func(p *process.Process) {
		all = append(all, p)
	})
	return all, err
}
