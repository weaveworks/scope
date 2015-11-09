package endpoint

import (
	"errors"
	"testing"
	"time"

	"github.com/weaveworks/scope/test"
)

func TestReverseResolver(t *testing.T) {
	tests := map[string][]string{
		"1.2.3.4": {"test.domain.name"},
		"4.3.2.1": {"im.a.little.tea.pot"},
	}

	revRes := newReverseResolver()
	defer revRes.stop()

	// Use a mocked resolver function.
	revRes.Resolver = func(addr string) (names []string, err error) {
		if names, ok := tests[addr]; ok {
			return names, nil
		}
		return []string{}, errors.New("invalid IP")
	}

	// Up the rate limit so the test runs faster.
	revRes.Throttle = time.Tick(time.Millisecond)

	for ip, names := range tests {
		test.Poll(t, 100*time.Millisecond, names, func() interface{} {
			ns, _ := revRes.get(ip)
			return ns
		})
	}
}
