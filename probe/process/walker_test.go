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
	name    string
	cmdline string
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
		"3":       {name: "3", cmdline: "curl\000google.com"},
		"2":       {name: "2"},
		"4":       {name: "4"},
		"notapid": {name: "notapid"},
		"1":       {name: "1"},
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
		3: {PID: 3, PPID: 2, Comm: "(unknown)", Cmdline: "curl google.com", Threads: 1},
		2: {PID: 2, PPID: 1, Comm: "(unknown)", Cmdline: "", Threads: 1},
		4: {PID: 4, PPID: 3, Comm: "(unknown)", Cmdline: "", Threads: 1},
		1: {PID: 1, PPID: 0, Comm: "(unknown)", Cmdline: "", Threads: 1},
	}

	have := map[int]*process.Process{}
	err := process.Walk("unused", func(p *process.Process) {
		have[p.PID] = p
	})

	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
