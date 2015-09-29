package xfer_test

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/xfer"
)

func TestBackgroundPublisher(t *testing.T) {
	mp := mockPublisher{}
	backgroundPublisher := xfer.NewBackgroundPublisher(&mp)
	defer backgroundPublisher.Stop()
	runtime.Gosched()

	for i := 1; i <= 10; i++ {
		err := backgroundPublisher.Publish(&bytes.Buffer{})
		if err != nil {
			t.Fatalf("%v", err)
		}

		test.Poll(t, 100*time.Millisecond, i, func() interface{} {
			return mp.count
		})
	}
}
