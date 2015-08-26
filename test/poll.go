package test

import (
	"reflect"
	"testing"
	"time"
)

// Poll repeatedly evaluates condition until we either timeout, or it suceeds.
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
		t.Fatal(Diff(want, h))
	}
}
