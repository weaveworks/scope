package process_test

import (
	"reflect"
	"testing"

	fs_hook "github.com/weaveworks/common/fs"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/common/test/fs"
	"github.com/weaveworks/scope/probe/process"
)

var mockFS = fs.Dir("",
	fs.Dir("proc",
		fs.Dir("3",
			fs.File{
				FName:     "cmdline",
				FContents: "curl\000google.com",
			},
			fs.File{
				FName:     "stat",
				FContents: "3 na R 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 2 2048",
			},
			fs.File{
				FName:     "limits",
				FContents: "Limit Soft-Limit Hard-Limit Units\nMax open files 32768 65536 files",
			},
			fs.Dir("fd", fs.File{FName: "0"}, fs.File{FName: "1"}, fs.File{FName: "2"}),
		),
		fs.Dir("2",
			fs.File{
				FName:     "cmdline",
				FContents: "bash",
			},
			fs.File{
				FName:     "stat",
				FContents: "2 na R 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0",
			},
			fs.File{
				FName:     "limits",
				FContents: ``,
			},
			fs.Dir("fd", fs.File{FName: "1"}, fs.File{FName: "2"}),
		),
		fs.Dir("4",
			fs.File{
				FName:     "cmdline",
				FContents: "apache",
			},
			fs.File{
				FName:     "stat",
				FContents: "4 na R 3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0",
			},
			fs.File{
				FName:     "limits",
				FContents: ``,
			},
			fs.Dir("fd", fs.File{FName: "0"}),
		),
		fs.Dir("notapid"),
		fs.Dir("1",
			fs.File{
				FName:     "cmdline",
				FContents: "init",
			},
			fs.File{
				FName:     "stat",
				FContents: "1 na R 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0",
			},
			fs.File{
				FName:     "limits",
				FContents: ``,
			},
			fs.Dir("fd"),
		),
	),
)

func TestWalker(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()

	want := map[int]process.Process{
		3: {PID: 3, PPID: 2, Name: "curl", Cmdline: "curl google.com", Threads: 1, RSSBytes: 8192, RSSBytesLimit: 2048, OpenFilesCount: 3, OpenFilesLimit: 32768},
		2: {PID: 2, PPID: 1, Name: "bash", Cmdline: "bash", Threads: 1, OpenFilesCount: 2},
		4: {PID: 4, PPID: 3, Name: "apache", Cmdline: "apache", Threads: 1, OpenFilesCount: 1},
		1: {PID: 1, PPID: 0, Name: "init", Cmdline: "init", Threads: 1, OpenFilesCount: 0},
	}

	have := map[int]process.Process{}
	walker := process.NewWalker("/proc", false)
	err := walker.Walk(func(p, _ process.Process) {
		have[p.PID] = p
	})

	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%v (%v)", test.Diff(want, have), err)
	}
}
