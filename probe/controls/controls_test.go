package controls_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/xfer"
)

func TestControls(t *testing.T) {
	controls.Register("foo", func(req xfer.Request) xfer.Response {
		return xfer.Response{
			Value: "bar",
		}
	})
	defer controls.Rm("foo")

	want := xfer.Response{
		ID:    1234,
		Value: "bar",
	}
	have := controls.HandleControlRequest(xfer.Request{
		ID:      1234,
		Control: "foo",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}

func TestControlsNotFound(t *testing.T) {
	want := xfer.Response{
		ID:    3456,
		Error: "Control 'baz' not recognised",
	}
	have := controls.HandleControlRequest(xfer.Request{
		ID:      3456,
		Control: "baz",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}
