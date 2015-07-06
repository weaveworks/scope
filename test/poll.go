package test

import (
	"testing"
	"time"
)

// Poll repeatedly evaluates condition until we either timeout, or it suceeds.
func Poll(t *testing.T, d time.Duration, condition func() bool, msg string) {
	deadline := time.Now().Add(d)
	for {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		if condition() {
			return
		}
		time.Sleep(d / 10)
	}
}
