package appclient_test

import (
	"bytes"
	"io"
	"net/url"
	"runtime"
	"testing"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
)

type mockClient struct {
	id      string
	count   int
	stopped int
	publish int
}

func (c *mockClient) Details() (xfer.Details, error) {
	return xfer.Details{ID: c.id}, nil
}

func (c *mockClient) ControlConnection() {
	c.count++
}

func (c *mockClient) Target() url.URL {
	return url.URL{}
}

func (c *mockClient) ReTarget(_ url.URL) {
}

func (c *mockClient) Stop() {
	c.stopped++
}

func (c *mockClient) Publish(io.Reader) error {
	c.publish++
	return nil
}

func (c *mockClient) PipeConnection(_ string, _ xfer.Pipe) {}
func (c *mockClient) PipeClose(_ string) error             { return nil }

var (
	a1      = &mockClient{id: "1"} // hostname a, app id 1
	a2      = &mockClient{id: "2"} // hostname a, app id 2
	b2      = &mockClient{id: "2"} // hostname b, app id 2 (duplicate)
	b3      = &mockClient{id: "3"} // hostname b, app id 3
	factory = func(hostname string, url url.URL) (appclient.AppClient, error) {
		switch url.Host {
		case "a1":
			return a1, nil
		case "a2":
			return a2, nil
		case "b2":
			return b2, nil
		case "b3":
			return b3, nil
		}
		panic(url.Host)
	}
)

func TestMultiClient(t *testing.T) {
	var (
		expect = func(i, j int) {
			if i != j {
				_, file, line, _ := runtime.Caller(1)
				t.Fatalf("%s:%d: %d != %d", file, line, i, j)
			}
		}
	)

	mp := appclient.NewMultiAppClient(factory, false)
	defer mp.Stop()

	// Add two hostnames with overlapping apps, check we don't add the same app twice
	mp.Set("a", []url.URL{{Host: "a1"}, {Host: "a2"}})
	mp.Set("b", []url.URL{{Host: "b2"}, {Host: "b3"}})
	expect(a1.count, 1)
	expect(a2.count+b2.count, 1)
	expect(b3.count, 1)

	// Now drop the overlap, check we don't remove the app
	mp.Set("b", []url.URL{{Host: "b3"}})
	expect(a1.count, 1)
	expect(a2.count+b2.count, 1)
	expect(b3.count, 1)

	// Now check we remove apps
	mp.Set("b", []url.URL{})
	expect(b3.stopped, 1)
}

func TestMultiClientPublish(t *testing.T) {
	mp := appclient.NewMultiAppClient(factory, false)
	defer mp.Stop()

	sum := func() int { return a1.publish + a2.publish + b2.publish + b3.publish }

	mp.Set("a", []url.URL{{Host: "a1"}, {Host: "a2"}})
	mp.Set("b", []url.URL{{Host: "b2"}, {Host: "b3"}})

	for i := 1; i < 10; i++ {
		if err := mp.Publish(&bytes.Buffer{}); err != nil {
			t.Error(err)
		}
		if want, have := 3*i, sum(); want != have {
			t.Errorf("want %d, have %d", want, have)
		}
	}
}
