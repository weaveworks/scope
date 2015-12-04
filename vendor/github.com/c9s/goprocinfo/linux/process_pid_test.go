package linux

import (
	"reflect"
	"testing"
)

func TestMaxPID(t *testing.T) {

	max, err := ReadMaxPID("proc/sys_kernel_pid_max")

	if err != nil {
		t.Fatal("max pid read fail", err)
	}

	if max != 32768 {
		t.Error("unexpected value")
	}

	t.Logf("%+v", max)
}

func TestListPID(t *testing.T) {

	list, err := ListPID("proc", 32768)

	if err != nil {
		t.Fatal("list pid fail", err)
	}

	var expected = []uint64{884, 3323, 4854, 5811}

	if !reflect.DeepEqual(list, expected) {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", list)
}
