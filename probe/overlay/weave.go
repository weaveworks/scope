package overlay

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/weaveworks/scope/report"
)

const (
	// WeavePeerName is the key for the peer name, typically a MAC address.
	WeavePeerName = "weave_peer_name"

	// WeavePeerNickName is the key for the peer nickname, typically a
	// hostname.
	WeavePeerNickName = "weave_peer_nick_name"
)

// Weave represents a single Weave router, presumably on the same host
// as the probe. It is both a Reporter and a Tagger: it produces an Overlay
// topology, and (in theory) can tag existing topologies with foreign keys to
// overlay -- though I'm not sure what that would look like in practice right
// now.
type Weave struct {
	url string
}

// NewWeave returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeave(weaveRouterAddress string) (*Weave, error) {
	s, err := sanitize("http://", 6784, "/status-json")(weaveRouterAddress)
	if err != nil {
		return nil, err
	}
	return &Weave{s}, nil
}

// Tag implements Tagger.
func (w Weave) Tag(r report.Report) (report.Report, error) {
	// The status-json endpoint doesn't return any link information, so
	// there's nothing to tag, yet.
	return r, nil
}

// Report implements Reporter.
func (w Weave) Report() (report.Report, error) {
	r := report.MakeReport()

	resp, err := http.Get(w.url)
	if err != nil {
		log.Printf("Weave Tagger: %v", err)
		return r, err
	}
	defer resp.Body.Close()

	var status struct {
		Peers []struct {
			Name     string `json:"Name"`
			NickName string `json:"NickName"`
		} `json:"Peers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		log.Printf("Weave Tagger: %v", err)
		return r, err
	}

	for _, peer := range status.Peers {
		r.Overlay.NodeMetadatas[report.MakeOverlayNodeID(peer.Name)] = report.NodeMetadata{
			WeavePeerName:     peer.Name,
			WeavePeerNickName: peer.NickName,
		}
	}
	return r, nil
}

func sanitize(scheme string, port int, path string) func(string) (string, error) {
	return func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("no host")
		}
		if !strings.HasPrefix(s, "http") {
			s = scheme + s
		}
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		if _, _, err = net.SplitHostPort(u.Host); err != nil {
			u.Host += fmt.Sprintf(":%d", port)
		}
		if u.Path != path {
			u.Path = path
		}
		return u.String(), nil
	}
}
