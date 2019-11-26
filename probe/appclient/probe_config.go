package appclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/certifi/gocertifi"
	"github.com/hashicorp/go-cleanhttp"

	"github.com/weaveworks/scope/common/xfer"
)

const (
	dialTimeout = 5 * time.Second
)

var certPool *x509.CertPool

func init() {
	var err error
	certPool, err = gocertifi.CACerts()
	if err != nil {
		panic(err)
	}
}

// ProbeConfig contains all the info needed for a probe to do HTTP requests
type ProbeConfig struct {
	BasicAuth    bool
	Token        string
	ProbeVersion string
	ProbeID      string
	Insecure     bool
}

func (pc ProbeConfig) authorizeHeaders(headers http.Header) {
	if pc.BasicAuth {
		headers.Set("Authorization", fmt.Sprintf("Basic %s", pc.Token))
	} else {
		headers.Set("Authorization", fmt.Sprintf("Scope-Probe token=%s", pc.Token))
	}
	headers.Set(xfer.ScopeProbeIDHeader, pc.ProbeID)
	headers.Set(xfer.ScopeProbeVersionHeader, pc.ProbeVersion)
	headers.Set("user-agent","Scope_Probe/"+pc.ProbeVersion )


}

func (pc ProbeConfig) authorizedRequest(method string, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err == nil {
		pc.authorizeHeaders(req.Header)
	}
	return req, err
}

func (pc ProbeConfig) getHTTPTransport(hostname string) *http.Transport {
	transport := cleanhttp.DefaultTransport()
	transport.DialContext = (&net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second,
	}).DialContext
	if pc.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		transport.TLSClientConfig = &tls.Config{
			RootCAs:    certPool,
			ServerName: hostname,
		}
	}
	return transport
}
