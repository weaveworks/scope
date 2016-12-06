package procspy

import (
	"net"
	"reflect"
	"testing"
	"time"

	fs_hook "github.com/weaveworks/common/fs"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/test"
)

func TestLinuxConnections(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()
	scanner := NewConnectionScanner(process.NewWalker("/proc"))
	defer scanner.Stop()

	// let the background scanner finish its first pass
	time.Sleep(1 * time.Second)

	iter, err := scanner.Connections(true)
	if err != nil {
		t.Fatal(err)
	}
	have := iter.Next()
	want := &Connection{
		LocalAddress:  net.ParseIP("0.0.0.0").To4(),
		LocalPort:     42688,
		RemoteAddress: net.ParseIP("0.0.0.0").To4(),
		RemotePort:    0,
		inode:         5107,
		Proc: Proc{
			PID:  1,
			Name: "foo",
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}

	if have := iter.Next(); have != nil {
		t.Fatal(have)
	}

}
