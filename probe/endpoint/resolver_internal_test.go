package endpoint

import (
	"errors"
	"testing"
)

func TestReverseResolver(t *testing.T) {
	tests := map[string]string{
		"8.8.8.8": "google-public-dns-a.google.com",
		"8.8.4.4": "google-public-dns-b.google.com",
	}

	revRes := newReverseResolver()

	// use a mocked resolver function
	revRes.resolver = func(addr string) (names []string, err error) {
		if name, ok := tests[addr]; ok {
			return []string{name}, nil
		}
		return []string{}, errors.New("invalid IP")
	}

	// first time: no names are returned for our reverse resolutions
	for ip := range tests {
		if have, err := revRes.Get(ip, true); have != "" || err == nil {
			t.Errorf("we didn't get an error, or the cache was not empty, when trying to resolve '%q'", ip)
		}
	}

	// so, if we check again these IPs, we should have the names now
	for ip, want := range tests {
		have, err := revRes.Get(ip, true)
		if err != nil {
			t.Errorf("%s: %v", ip, err)
		}
		if want != have {
			t.Errorf("%s: want %q, have %q", ip, want, have)
		}
	}
}
