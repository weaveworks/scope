package plugins

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type unixRoundTripper struct {
	address string
	timeout time.Duration
}

func makeUnixRoundTripper(address string, timeout time.Duration) (http.RoundTripper, error) {
	return unixRoundTripper{address: address, timeout: timeout}, nil
}

func (t unixRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := net.DialTimeout("unix", t.address, t.timeout)
	if err != nil {
		return nil, err
	}
	client := httputil.NewClientConn(conn, nil)
	defer client.Close()
	return client.Do(req)
}
