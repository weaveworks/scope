package plugins

import (
	"net"
	"net/http"
	"time"
)

func makeUnixRoundTripper(address string, timeout time.Duration) (http.RoundTripper, error) {
	rt := &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return net.DialTimeout("unix", address, timeout)
		},
	}
	return rt, nil
}
