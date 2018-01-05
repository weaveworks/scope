package kubernetes_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/weaveworks/scope/probe/kubernetes"
)

func TestLogReadCloser(t *testing.T) {
	data0 := []byte("abcdefghijklmno\npqrstuvwxyz\n")
	data1 := []byte("ABCDEFGHI\nJKLMNOPQRSTUVWXYZ\n")
	data2 := []byte("012345678901\n2345\n\n678\n")

	label0 := "zero"
	label1 := "one"
	label2 := "two"
	longestlabelLength := len(label0)

	readClosersWithLabel := map[io.ReadCloser]string{}
	r0 := ioutil.NopCloser(bytes.NewReader(data0))
	readClosersWithLabel[r0] = label0
	r1 := ioutil.NopCloser(bytes.NewReader(data1))
	readClosersWithLabel[r1] = label1
	r2 := ioutil.NopCloser(bytes.NewReader(data2))
	readClosersWithLabel[r2] = label2

	l := kubernetes.NewLogReadCloser(readClosersWithLabel)

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

	// convert to string for easier comparison
	result := map[string]int{}
	lineCounter(result, longestlabelLength, label0, data0)
	lineCounter(result, longestlabelLength, label1, data1)
	lineCounter(result, longestlabelLength, label2, data2)

	str := string(buf[:count])
	for _, line := range strings.SplitAfter(str, "\n") {
		v, ok := result[line]
		if ok {
			result[line] = v - 1
		}
	}

	for line, v := range result {
		if v != 0 {
			t.Errorf("Line %v has not be read from reader", line)
		}
	}

	err := l.Close()
	if err != nil {
		t.Errorf("Close must not return an error: %v", err)
	}
}

func lineCounter(counter map[string]int, pad int, label string, data []byte) {
	for _, str := range strings.SplitAfter(string(data), "\n") {
		if len(str) == 0 {
			// SplitAfter ends with an empty string if the last character is '\n'
			continue
		}
		line := fmt.Sprintf("[%-*s] %v", pad, label, str)
		v, ok := counter[line]
		if !ok {
			v = 0
		}
		v++
		counter[line] = v
	}
}
