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

type mockedProcess struct {
	ID, Comm, Cmdline string
}

func (p mockedProcess) Name() string       { return p.ID }
func (p mockedProcess) Size() int64        { return 0 }
func (p mockedProcess) Mode() os.FileMode  { return 0 }
func (p mockedProcess) ModTime() time.Time { return time.Now() }
func (p mockedProcess) IsDir() bool        { return true }
func (p mockedProcess) Sys() interface{}   { return nil }

func TestProcReaderProcesses(t *testing.T) {
	processes := map[string]mockedProcess{
		"3":       {ID: "3", Comm: "curl\n", Cmdline: "curl\000google.com"},
		"2":       {ID: "2", Comm: "bash\n"},
		"4":       {ID: "4", Comm: "apache\n"},
		"notapid": {ID: "notapid"},
		"1":       {ID: "1", Comm: "init\n"},
	}

	want := map[int]process.Process{
		3: {PID: 3, PPID: 2, Comm: "curl", Cmdline: "curl google.com", Threads: 1, Inodes: []uint64{}},
		2: {PID: 2, PPID: 1, Comm: "bash", Cmdline: "", Threads: 1, Inodes: []uint64{}},
		4: {PID: 4, PPID: 3, Comm: "apache", Cmdline: "", Threads: 1, Inodes: []uint64{}},
		1: {PID: 1, PPID: 0, Comm: "init", Cmdline: "", Threads: 1, Inodes: []uint64{}},
	}

	// use a mocked /proc that reads from our mocked processes
	procDir := mockedDir{
		ReadDirNamesFunc: func(path string) ([]string, error) {
			result := []string{}
			for k := range processes {
				result = append(result, k)
			}
			return result, nil
		},
		OpenFunc: func(filename string) (process.File, error) {
			splits := strings.Split(filename, "/")
			pid := splits[len(splits)-2]
			process, ok := processes[pid]
			if !ok {
				return nil, fmt.Errorf("not found")
			}

			file := splits[len(splits)-1]
			var content []byte
			switch file {
			case "comm":
				content = []byte(process.Comm)
			case "stat":
				pid, _ := strconv.Atoi(splits[len(splits)-2])
				parent := pid - 1
				content = []byte(fmt.Sprintf("%d na R %d 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1", pid, parent))
			case "cmdline":
				content = []byte(process.Cmdline)
			default:
				return nil, fmt.Errorf("not found")
			}

			return mockedFileWithBytes{content}, nil
		},
	}

	procReader := process.NewReader(procDir, false)
	procReader.Read()

	have := map[int]process.Process{}
	err := procReader.Processes(func(p process.Process) {
		have[p.PID] = p
	})
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
