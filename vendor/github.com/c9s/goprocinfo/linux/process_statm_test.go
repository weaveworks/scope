package linux

import (
	"reflect"
	"testing"
)

func TestReadProcessStatm(t *testing.T) {

	statm, err := ReadProcessStatm("proc/3323/statm")

	if err != nil {
		t.Fatal("process statm read fail", err)
	}

	expected := &ProcessStatm{
		Size:     4053,
		Resident: 522,
		Share:    174,
		Text:     174,
		Lib:      0,
		Data:     286,
		Dirty:    0,
	}

	if !reflect.DeepEqual(statm, expected) {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", statm)
}
