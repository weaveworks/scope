package xfer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/weaveworks/scope/common/sanitize"
)

// HTTPPublisher publishes buffers by POST to a fixed endpoint.
type HTTPPublisher struct {
	url     string
	token   string
	probeID string
	client  *http.Client
}

var fastClient = &http.Client{
	Timeout: 5 * time.Second,
}

// NewHTTPPublisher returns an HTTPPublisher ready for use.
func NewHTTPPublisher(target, token, probeID string, insecure bool) (string, *HTTPPublisher, error) {
	p := &HTTPPublisher{
		url:     sanitize.URL("", 0, "/api/report")(target),
		token:   token,
		probeID: probeID,
		client:  http.DefaultClient,
	}
	client := fastClient
	if insecure {
		allowInsecure(fastClient)
		allowInsecure(p.client)
	}
	req, err := p.authorizedRequest("GET", sanitize.URL("", 0, "/api")(target), nil)
	if err != nil {
		return "", nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	var apiResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", nil, err
	}
	return apiResponse.ID, p, nil
}

func (p HTTPPublisher) authorizedRequest(method string, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err == nil {
		req.Header.Set("Authorization", AuthorizationHeader(p.token))
		req.Header.Set(ScopeProbeIDHeader, p.probeID)
	}
	return req, err
}

func (p HTTPPublisher) String() string {
	return p.url
}

// Publish publishes the report to the URL.
func (p HTTPPublisher) Publish(r io.Reader) error {
	req, err := p.authorizedRequest("POST", p.url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "gzip")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

// Stop implements Publisher
func (p HTTPPublisher) Stop() {}

// AuthorizationHeader returns a value suitable for an HTTP Authorization
// header, based on the passed token string.
func AuthorizationHeader(token string) string {
	return fmt.Sprintf("Scope-Probe token=%s", token)
}

func allowInsecure(c *http.Client) {
	c.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
// ID is currently set to the probe's hostname. It's designed to deduplicate
// reports from the same probe to the same receiver, in case the probe is
// configured to publish to multiple receivers that resolve to the same app.
const ScopeProbeIDHeader = "X-Scope-Probe-ID"
