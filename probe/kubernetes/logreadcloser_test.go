package kubernetes_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/weaveworks/scope/probe/kubernetes"
)

func TestLogReadCloser(t *testing.T) {
	s0 := []byte("abcdefghijklmnopqrstuvwxyz")
	s1 := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s2 := []byte("0123456789012345")

	r0 := ioutil.NopCloser(bytes.NewReader(s0))
	r1 := ioutil.NopCloser(bytes.NewReader(s1))
	r2 := ioutil.NopCloser(bytes.NewReader(s2))

	l := kubernetes.NewLogReadCloser(r0, r1, r2)

	buf := make([]byte, 3000)
	count := 0
	for {
		n, err := l.Read(buf[count:])
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
		}
		count += n
	}

	total := len(s0) + len(s1) + len(s2)
	if count != total {
		t.Errorf("Must read %v characters, but got %v", total, count)
	}

	// check every byte
	byteCounter := map[byte]int{}
	byteCount(byteCounter, s0)
	byteCount(byteCounter, s1)
	byteCount(byteCounter, s2)

	for i := 0; i < count; i++ {
		b := buf[i]
		v, ok := byteCounter[b]
		if ok {
			v--
			byteCounter[b] = v
		}
	}

	for b, c := range byteCounter {
		if c != 0 {
			t.Errorf("%v should be 0 instead of %v", b, c)
		}
	}

	err := l.Close()
	if err != nil {
		t.Errorf("Close must not return an error: %v", err)
	}
}

func byteCount(accumulator map[byte]int, s []byte) {
	for _, b := range s {
		v, ok := accumulator[b]
		if !ok {
			v = 0
		}
		v++
		accumulator[b] = v
	}
}
