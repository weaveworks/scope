package xfer_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/weaveworks/scope/xfer"
)

func TestMultiPublisher(t *testing.T) {
	var (
		a1 = &mockPublisher{} // target a, endpoint 1
		a2 = &mockPublisher{} // target a, endpoint 2 (duplicate)
		b2 = &mockPublisher{} // target b, endpoint 2 (duplicate)
		b3 = &mockPublisher{} // target b, endpoint 3
	)

	sum := func() int { return a1.count + a2.count + b2.count + b3.count }

	mp := xfer.NewMultiPublisher(func(endpoint string) (string, xfer.Publisher, error) {
		switch endpoint {
		case "a1":
			return "1", a1, nil
		case "a2":
			return "2", a2, nil
		case "b2":
			return "2", b2, nil
		case "b3":
			return "3", b3, nil
		default:
			return "", nil, fmt.Errorf("invalid endpoint %s", endpoint)
		}
	})
	defer mp.Stop()

	mp.Set("a", []string{"a1", "a2"})
	mp.Set("b", []string{"b2", "b3"})

	for i := 1; i < 10; i++ {
		if err := mp.Publish(&bytes.Buffer{}); err != nil {
			t.Error(err)
		}
		if want, have := 3*i, sum(); want != have {
			t.Errorf("want %d, have %d", want, have)
		}
	}
}

type mockPublisher struct{ count int }

func (p *mockPublisher) Publish(io.Reader) error { p.count++; return nil }
func (p *mockPublisher) Stop()                   {}
