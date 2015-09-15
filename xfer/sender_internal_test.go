package xfer

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func TestPooledBuffer(t *testing.T) {
	pool := &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
	buf := pool.Get().(*bytes.Buffer)
	pb := pooledBuffer{buf, pool}

	const str = "Okay"
	if n, err := fmt.Fprintf(pb, str); err != nil {
		t.Error(err)
	} else if want, have := len(str), n; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if err := pb.Close(); err != nil {
		t.Error(err)
	}

	if want, have := errDoubleClose, pb.Close(); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
