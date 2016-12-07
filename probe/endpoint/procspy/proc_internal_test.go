package procspy

import (
	"bytes"
	"reflect"
	"syscall"
	"testing"
	"time"

	fs_hook "github.com/weaveworks/common/fs"
	"github.com/weaveworks/common/test/fs"
	"github.com/weaveworks/scope/probe/process"
)

var mockFS = fs.Dir("",
	fs.Dir("proc",
		fs.Dir("1",
			fs.Dir("fd",
				fs.File{
					FName: "16",
					FStat: syscall.Stat_t{
						Ino:  5107,
						Mode: syscall.S_IFSOCK,
					},
				},
			),
			fs.File{
				FName:     "cmdline",
				FContents: "foo",
			},
			fs.Dir("ns",
				fs.File{
					FName: "net",
					FStat: syscall.Stat_t{},
				},
			),
			fs.Dir("net",
				fs.File{
					FName: "tcp",
					FContents: `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:A6C0 00000000:0000 01 00000000:00000000 00:00000000 00000000   105        0 5107 1 ffff8800a6aaf040 100 0 0 10 2d
`,
				},
				fs.File{
					FName: "tcp6",
				},
			),
			fs.File{
				FName:     "stat",
				FContents: "1 na R 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0",
			},
			fs.File{
				FName:     "limits",
				FContents: "",
			},
		),
	),
)

func TestWalkProcPid(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()

	buf := bytes.Buffer{}
	walker := process.NewWalker(procRoot)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	pWalker := newPidWalker(walker, ticker.C, 1)
	have, err := pWalker.walk(&buf)
	if err != nil {
		t.Fatal(err)
	}
	want := map[uint64]*Proc{
		5107: {
			PID:  1,
			Name: "foo",
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("%+v", have)
	}
}
