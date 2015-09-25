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
	url     string
	token   string
	probeID string
}

var fastClient = http.Client{
	Timeout: 5 * time.Second,
}

// NewHTTPPublisher returns an HTTPPublisher ready for use.
func NewHTTPPublisher(target, token, probeID string) (string, *HTTPPublisher, error) {
	targetAPI := sanitize.URL("http://", 0, "/api")(target)
	resp, err := fastClient.Get(targetAPI)
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
	return apiResponse.ID, &HTTPPublisher{
		url:     sanitize.URL("http://", 0, "/api/report")(target),
		token:   token,
		probeID: probeID,
	}, nil
}

func (p HTTPPublisher) String() string {
	return p.url
}

// Publish publishes the report to the URL.
func (p HTTPPublisher) Publish(r io.Reader) error {
	req, err := http.NewRequest("POST", p.url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", AuthorizationHeader(p.token))
	req.Header.Set(ScopeProbeIDHeader, p.probeID)
	req.Header.Set("Content-Encoding", "gzip")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	resp, err := http.DefaultClient.Do(req)
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

// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
// ID is currently set to the probe's hostname. It's designed to deduplicate
// reports from the same probe to the same receiver, in case the probe is
// configured to publish to multiple receivers that resolve to the same app.
const ScopeProbeIDHeader = "X-Scope-Probe-ID"
