package xfer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/weaveworks/scope/common/sanitize"
)

// HTTPPublisher publishes buffers by POST to a fixed endpoint.
type HTTPPublisher struct {
	ProbeConfig

	url    string
	client *http.Client
}

// NewHTTPPublisher returns an HTTPPublisher ready for use.
func NewHTTPPublisher(hostname, target, token, probeID string, insecure bool) (string, *HTTPPublisher, error) {
	pc := ProbeConfig{
		Token:    token,
		ProbeID:  probeID,
		Insecure: insecure,
	}

	httpTransport, err := pc.getHTTPTransport(hostname)
	if err != nil {
		return "", nil, err
	}

	p := &HTTPPublisher{
		ProbeConfig: pc,
		url:         sanitize.URL("", 0, "/api/report")(target),
		client: &http.Client{
			Transport: httpTransport,
		},
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: httpTransport,
	}
	req, err := pc.authorizedRequest("GET", sanitize.URL("", 0, "/api")(target), nil)
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

func (p HTTPPublisher) String() string {
	return p.url
}

// Publish publishes the report to the URL.
func (p HTTPPublisher) Publish(r io.Reader) error {
	req, err := p.ProbeConfig.authorizedRequest("POST", p.url, r)
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
func (p *HTTPPublisher) Stop() {
	// We replace the HTTPPublishers pretty regularly, so we need to ensure the
	// underlying connections get closed, or we end up with lots of idle
	// goroutines on the server (see #604)
	p.client.Transport.(*http.Transport).CloseIdleConnections()
}
