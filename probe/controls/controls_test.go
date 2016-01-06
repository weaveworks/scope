package controls_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/test"
)

func TestControls(t *testing.T) {
	controls.Register("foo", func(req xfer.Request) xfer.Response {
		return xfer.Response{
			Value: "bar",
		}
	})
	defer controls.Rm("foo")

	want := xfer.Response{
		Value: "bar",
	}
	have := controls.HandleControlRequest(xfer.Request{
		Control: "foo",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}

func TestControlsNotFound(t *testing.T) {
	want := xfer.Response{
		Error: "Control \"baz\" not recognised",
	}
	have := controls.HandleControlRequest(xfer.Request{
		Control: "baz",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}
