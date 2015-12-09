package procspy

import (
	"bytes"
	"reflect"
	"syscall"
	"testing"

	fs_hook "github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/test/fs"
)

var mockFS = fs.Dir("",
	fs.Dir("proc",
		fs.Dir("1",
			fs.Dir("fd",
				fs.File{
					FName: "16",
					FStat: syscall.Stat_t{
						Ino:  45,
						Mode: syscall.S_IFSOCK,
					},
				},
			),
			fs.File{
				FName:     "comm",
				FContents: "foo\n",
			},
			fs.Dir("ns",
				fs.File{
					FName: "net",
					FStat: syscall.Stat_t{},
				},
			),
			fs.File{
				FName:     "stat",
				FContents: "1 na R 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1",
			},
		),
	),
)

func TestWalkProcPid(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()

	buf := bytes.Buffer{}
	have, err := walkProcPid(&buf, process.NewWalker(procRoot))
	if err != nil {
		t.Fatal(err)
	}
	want := map[uint64]Proc{
		45: {
			PID:  1,
			Name: "foo",
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("%+v", have)
	}
}
