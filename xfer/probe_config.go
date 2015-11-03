package xfer

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/certifi/gocertifi"
)

// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
// ID is currently set to the a random string on probe startup.
const ScopeProbeIDHeader = "X-Scope-Probe-ID"

// ProbeConfig contains all the info needed for a probe to do HTTP requests
type ProbeConfig struct {
	Token    string
	ProbeID  string
	Insecure bool
}

func (pc ProbeConfig) authorizeHeaders(headers http.Header) {
	headers.Set("Authorization", fmt.Sprintf("Scope-Probe token=%s", pc.Token))
	headers.Set(ScopeProbeIDHeader, pc.ProbeID)
}

func (pc ProbeConfig) authorizedRequest(method string, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err == nil {
		pc.authorizeHeaders(req.Header)
	}
	return req, err
}

func (pc ProbeConfig) getHTTPTransport(hostname string) (*http.Transport, error) {
	if pc.Insecure {
		return &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}, nil
	}

	host, _, err := net.SplitHostPort(hostname)
	if err != nil {
		return nil, err
	}

	certPool, err := gocertifi.CACerts()
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    certPool,
			ServerName: host,
		},
	}, nil
}
