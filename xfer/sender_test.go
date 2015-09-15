package xfer_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	"github.com/weaveworks/scope/xfer"

	"testing"
)

func TestHTTPSender(t *testing.T) {
	c := make(chan string)
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { c <- mustReadAll(r.Body) }))
	defer s.Close()

	const authToken, probeID = "123", "abc"
	sender := xfer.NewHTTPSender(s.URL, authToken, probeID)
	defer sender.Stop()

	const str = "alright"
	sender.Send(ioutil.NopCloser(bytes.NewBufferString(str)))
	if want, have := str, <-c; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestHTTPSenderNonBlocking(t *testing.T) {
	log.SetOutput(ioutil.Discard) // logging the errors takes time and can cause test failure

	delay := 5 * time.Millisecond
	s := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { time.Sleep(delay) }))
	defer s.Close()

	const authToken, probeID = "123", "abc"
	sender := xfer.NewHTTPSender(s.URL, authToken, probeID)
	defer sender.Stop()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			sender.Send(ioutil.NopCloser(bytes.NewBufferString(strconv.Itoa(i))))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(delay / 2):
		t.Errorf("HTTPSender appears to be blocking")
	}
}

func TestMultiSender(t *testing.T) {
	sender := &mockSender{}
	multi := xfer.NewMultiSender(func(string) (xfer.Sender, error) { return sender, nil })

	multi.Add("one copy")
	const first = "hello"
	multi.Send(ioutil.NopCloser(bytes.NewBufferString(first)))
	if want, have := first, sender.buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	sender.buf.Reset()
	multi.Add("two copies")
	const second = "world"
	multi.Send(ioutil.NopCloser(bytes.NewBufferString(second)))
	if want, have := second+second, sender.buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func mustReadAll(r io.Reader) string {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

type mockSender struct {
	mtx sync.Mutex
	buf bytes.Buffer
}

func (s *mockSender) Send(rc io.ReadCloser) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	io.Copy(&s.buf, rc)
	rc.Close()
}

func (s *mockSender) Stop() {}
