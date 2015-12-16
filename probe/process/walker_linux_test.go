package process_test

import (
	"reflect"
	"testing"

	fs_hook "github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fs"
)

var mockFS = fs.Dir("",
	fs.Dir("proc",
		fs.Dir("3",
			fs.File{
				FName:     "comm",
				FContents: "curl\n",
			},
			fs.File{
				FName:     "cmdline",
				FContents: "curl\000google.com",
			},
			fs.File{
				FName:     "stat",
				FContents: "3 na R 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0",
			},
		),
		fs.Dir("2",
			fs.File{
				FName:     "comm",
				FContents: "bash\n",
			},
			fs.File{
				FName:     "cmdline",
				FContents: "",
			},
			fs.File{
				FName:     "stat",
				FContents: "2 na R 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0",
			},
		),
		fs.Dir("4",
			fs.File{
				FName:     "comm",
				FContents: "apache\n",
			},
			fs.File{
				FName:     "cmdline",
				FContents: "",
			},
			fs.File{
				FName:     "stat",
				FContents: "4 na R 3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0",
			},
		),
		fs.Dir("notapid"),
		fs.Dir("1",
			fs.File{
				FName:     "comm",
				FContents: "init\n",
			},
			fs.File{
				FName:     "cmdline",
				FContents: "",
			},
			fs.File{
				FName:     "stat",
				FContents: "1 na R 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0",
			},
		),
	),
)

func TestWalker(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()

	want := map[int]process.Process{
		3: {PID: 3, PPID: 2, Comm: "curl", Cmdline: "curl google.com", Threads: 1},
		2: {PID: 2, PPID: 1, Comm: "bash", Cmdline: "", Threads: 1},
		4: {PID: 4, PPID: 3, Comm: "apache", Cmdline: "", Threads: 1},
		1: {PID: 1, PPID: 0, Comm: "init", Cmdline: "", Threads: 1},
	}

	have := map[int]process.Process{}
	walker := process.NewWalker("/proc")
	err := walker.Walk(func(p, _ process.Process) {
		have[p.PID] = p
	})

	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
