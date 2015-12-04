package linux

import (
	"reflect"
	"testing"
)

func TestReadProcessIO(t *testing.T) {

	io, err := ReadProcessIO("proc/3323/io")

	if err != nil {
		t.Fatal("process io read fail", err)
	}

	expected := &ProcessIO{
		RChar:               3865585,
		WChar:               183294,
		Syscr:               6697,
		Syscw:               997,
		ReadBytes:           90112,
		WriteBytes:          45056,
		CancelledWriteBytes: 0,
	}

	if !reflect.DeepEqual(io, expected) {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", io)
}
