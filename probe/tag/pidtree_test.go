package tag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fileinfo struct {
	name string
}

func (f fileinfo) Name() string       { return f.name }
func (f fileinfo) Size() int64        { return 0 }
func (f fileinfo) Mode() os.FileMode  { return 0 }
func (f fileinfo) ModTime() time.Time { return time.Now() }
func (f fileinfo) IsDir() bool        { return true }
func (f fileinfo) Sys() interface{}   { return nil }

func TestPIDTree(t *testing.T) {
	oldReadDir, oldReadFile := readDir, readFile
	defer func() {
		readDir = oldReadDir
		readFile = oldReadFile
	}()

	readDir = func(path string) ([]os.FileInfo, error) {
		return []os.FileInfo{
			fileinfo{"3"}, fileinfo{"2"}, fileinfo{"4"},
			fileinfo{"notapid"}, fileinfo{"1"},
		}, nil
	}

	readFile = func(path string) ([]byte, error) {
		splits := strings.Split(path, "/")
		if splits[len(splits)-1] != "stat" {
			return nil, fmt.Errorf("not stat")
		}
		pid, err := strconv.Atoi(splits[len(splits)-2])
		if err != nil {
			return nil, err
		}
		parent := pid - 1
		return []byte(fmt.Sprintf("%d na R %d", pid, parent)), nil
	}

	pidtree, err := NewPIDTree("/proc")
	if err != nil {
		t.Fatalf("newPIDTree error: %v", err)
	}

	for pid, want := range map[int]int{
		2: 1,
		3: 2,
	} {
		have, err := pidtree.GetParent(pid)
		if err != nil || !reflect.DeepEqual(want, have) {
			t.Errorf("%d: want %#v, have %#v (%v)", pid, want, have, err)
		}
	}
}
