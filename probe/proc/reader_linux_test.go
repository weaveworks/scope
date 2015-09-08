package proc_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/test"
)

func TestProcReaderProcesses(t *testing.T) {
	processes := map[string]proc.MockedProcess{
		"3":       {Id: "3", Comm: "curl\n", Cmdline: "curl\000google.com"},
		"2":       {Id: "2", Comm: "bash\n"},
		"4":       {Id: "4", Comm: "apache\n"},
		"notapid": {Id: "notapid"},
		"1":       {Id: "1", Comm: "init\n"},
	}

	want := map[int]proc.Process{
		3: {PID: 3, PPID: 2, Comm: "curl", Cmdline: "curl google.com", Threads: 1, Inodes: []uint64{}},
		2: {PID: 2, PPID: 1, Comm: "bash", Cmdline: "", Threads: 1, Inodes: []uint64{}},
		4: {PID: 4, PPID: 3, Comm: "apache", Cmdline: "", Threads: 1, Inodes: []uint64{}},
		1: {PID: 1, PPID: 0, Comm: "init", Cmdline: "", Threads: 1, Inodes: []uint64{}},
	}

	// use a mocked /proc that reads from our mocked processes
	procDir := proc.MockedProcDir{
		ReadDirFunc: func(path string) ([]os.FileInfo, error) {
			result := []os.FileInfo{}
			for _, p := range processes {
				result = append(result, p)
			}
			return result, nil
		},

		ReadFileFunc: func(path string) ([]byte, error) {
			splits := strings.Split(path, "/")
			pid := splits[len(splits)-2]
			process, ok := processes[pid]
			if !ok {
				return nil, fmt.Errorf("not found")
			}

			file := splits[len(splits)-1]
			switch file {
			case "comm":
				return []byte(process.Comm), nil
			case "stat":
				pid, _ := strconv.Atoi(splits[len(splits)-2])
				parent := pid - 1
				return []byte(fmt.Sprintf("%d na R %d 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1", pid, parent)), nil
			case "cmdline":
				return []byte(process.Cmdline), nil
			}

			return nil, fmt.Errorf("not found")
		},
	}

	procReader := proc.NewProcReader(procDir)
	have := map[int]proc.Process{}
	err := procReader.Processes(func(p proc.Process) {
		have[p.PID] = p
	})
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
