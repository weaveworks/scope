package xfer_test

import (
	"runtime"
	"testing"

	"github.com/weaveworks/scope/xfer"
)

type mockClient struct {
	id      string
	count   int
	stopped int
}

func (c *mockClient) Details() (xfer.Details, error) {
	return xfer.Details{ID: c.id}, nil
}

func (c *mockClient) ControlConnection(handler xfer.ControlHandler) {
	c.count++
}

func (c *mockClient) Stop() {
	c.stopped++
}

func TestMultiClient(t *testing.T) {
	var (
		a1      = &mockClient{id: "1"} // hostname a, app id 1
		a2      = &mockClient{id: "2"} // hostname a, app id 2
		b2      = &mockClient{id: "2"} // hostname b, app id 2 (duplicate)
		b3      = &mockClient{id: "3"} // hostname b, app id 3
		factory = func(_ xfer.ProbeConfig, hostname, target string) (xfer.AppClient, error) {
			switch target {
			case "a1":
				return a1, nil
			case "a2":
				return a2, nil
			case "b2":
				return b2, nil
			case "b3":
				return b3, nil
			}
			t.Fatal(target)
			return a1, nil
		}
		controlHandler = xfer.ControlHandlerFunc(func(_ xfer.Request) xfer.Response {
			return xfer.Response{}
		})
		expect = func(i, j int) {
			if i != j {
				_, file, line, _ := runtime.Caller(1)
				t.Fatalf("%s:%d: %d != %d", file, line, i, j)
			}
		}
	)

	mp := xfer.NewMultiAppClient(xfer.ProbeConfig{}, controlHandler, factory)
	defer mp.Stop()

	// Add two hostnames with overlapping apps, check we don't add the same app twice
	mp.Set("a", []string{"a1", "a2"})
	mp.Set("b", []string{"b2", "b3"})
	expect(a1.count, 1)
	expect(a2.count+b2.count, 1)
	expect(b3.count, 1)

	// Now drop the overlap, check we don't remove the app
	mp.Set("b", []string{"b3"})
	expect(a1.count, 1)
	expect(a2.count+b2.count, 1)
	expect(b3.count, 1)

	// Now check we remove apps
	mp.Set("b", []string{})
	expect(b3.stopped, 1)
}
