package xfer

const (
	// AppPort is the default port that the app will use for its HTTP server.
	// The app publishes the API and user interface, and receives reports from
	// probes, on this port.
	AppPort = 4040

	// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
	// ID is currently set to the a random string on probe startup.
	ScopeProbeIDHeader = "X-Scope-Probe-ID"
)

// Details are some generic details that can be fetched from /api
type Details struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
}
