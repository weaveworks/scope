package controls_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
)

func TestControls(t *testing.T) {
	registry := controls.NewDefaultHandlerRegistry()
	registry.Register("foo", func(req xfer.Request) xfer.Response {
		return xfer.Response{
			Value: "bar",
		}
	})
	defer registry.Rm("foo")

	want := xfer.Response{
		Value: "bar",
	}
	have := registry.HandleControlRequest(xfer.Request{
		Control: "foo",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}

func TestControlsNotFound(t *testing.T) {
	registry := controls.NewDefaultHandlerRegistry()
	want := xfer.Response{
		Error: "Control \"baz\" not recognised",
	}
	have := registry.HandleControlRequest(xfer.Request{
		Control: "baz",
	})
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}
