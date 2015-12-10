package procspy

import (
	"bytes"
	"reflect"
	"syscall"
	"testing"

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
		),
	),
)

func TestWalkProcPid(t *testing.T) {
	oldReadDir, oldLstat, oldStat, oldOpen := readDir, lstat, stat, open
	defer func() { readDir, lstat, stat, open = oldReadDir, oldLstat, oldStat, oldOpen }()
	readDir, lstat, stat, open = mockFS.ReadDir, mockFS.Lstat, mockFS.Stat, mockFS.Open

	buf := bytes.Buffer{}
	have, err := walkProcPid(&buf)
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
