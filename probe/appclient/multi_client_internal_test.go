package appclient

import (
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	n := 3
	s := newSemaphore(n)

	// First n should be fine
	for i := 0; i < n; i++ {
		ok := make(chan struct{})
		go func() { s.acquire(); close(ok) }()
		select {
		case <-ok:
		case <-time.After(10 * time.Millisecond):
			t.Errorf("p (%d) failed", i+1)
		}
	}

	// This should block
	ok := make(chan struct{})
	go func() { s.acquire(); close(ok) }()
	select {
	case <-ok:
		t.Errorf("%dth p OK, but should block", n+1)
	case <-time.After(10 * time.Millisecond):
		//t.Logf("%dth p blocks, as expected", n+1)
	}

	s.release()

	select {
	case <-ok:
	case <-time.After(10 * time.Millisecond):
		t.Errorf("%dth p didn't resolve in time", n+1)
	}
}
