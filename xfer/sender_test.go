package xfer_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/weaveworks/scope/xfer"
)

func TestHTTPSender(t *testing.T) {
	c := make(chan string)
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { c <- mustReadAll(r.Body) }))
	defer s.Close()

	sender := xfer.NewHTTPSender(s.URL, "some-auth-token", "some-probe-ID")
	defer sender.Stop()

	const str = "alright"
	sender.Send(ioutil.NopCloser(bytes.NewBufferString(str)))
	if want, have := str, <-c; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestHTTPSenderNonBlocking(t *testing.T) {
	log.SetOutput(ioutil.Discard) // logging the errors takes time and can cause test failure

	var (
		d = 5 * time.Millisecond
		c = make(chan string)
	)

	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { time.Sleep(d); c <- mustReadAll(r.Body) }))
	defer s.Close()

	sender := xfer.NewHTTPSender(s.URL, "some-auth-token", "some-probe-ID")
	defer sender.Stop()

	sender.Send(ioutil.NopCloser(bytes.NewBufferString("a")))
	sender.Send(ioutil.NopCloser(bytes.NewBufferString("dropped 1")))
	sender.Send(ioutil.NopCloser(bytes.NewBufferString("dropped 2")))
	sender.Send(ioutil.NopCloser(bytes.NewBufferString("dropped 3")))
	sender.Send(ioutil.NopCloser(bytes.NewBufferString("z")))

	if want, have := "a", <-c; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := "z", <-c; want != have {
		t.Errorf("want %q, have %q", want, have)
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
