package test

import (
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/test/reflect"
)

// Poll repeatedly evaluates condition until we either timeout, or it succeeds.
func Poll(t *testing.T, d time.Duration, want interface{}, have func() interface{}) {
	deadline := time.Now().Add(d)
	for {
		if time.Now().After(deadline) {
			break
		}
		if reflect.DeepEqual(want, have()) {
			return
		}
		time.Sleep(d / 10)
	}
	h := have()
	if !reflect.DeepEqual(want, h) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: %s", file, line, test.Diff(want, h))
	}
}
