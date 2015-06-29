package tag

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

// WeaveTagger represents a single Weave router, presumably on the same host
// as the probe. It can produce an Overlay topology, and in theory can tag
// existing topologies with foreign keys to overlay -- though I'm not sure
// what that would look like in practice right now.
type WeaveTagger struct {
	url string
}

// NewWeaveTagger returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeaveTagger(weaveRouterAddress string) (*WeaveTagger, error) {
	s, err := sanitize("http://", 6784, "/status-json")(weaveRouterAddress)
	if err != nil {
		return nil, err
	}
	return &WeaveTagger{s}, nil
}

// Tag implements Tagger.
func (t WeaveTagger) Tag(r report.Report) (report.Report, error) {
	// The status-json endpoint doesn't return any link information, so
	// there's nothing to tag, yet.
	return r, nil
}

// OverlayTopology produces an overlay topology from the Weave router.
func (t WeaveTagger) OverlayTopology() report.Topology {
	topology := report.NewTopology()

	resp, err := http.Get(t.url)
	if err != nil {
		log.Printf("Weave Tagger: %v", err)
		return topology
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
		return topology
	}

	for _, peer := range status.Peers {
		topology.NodeMetadatas[report.MakeOverlayNodeID(peer.Name)] = report.NodeMetadata{
			WeavePeerName:     peer.Name,
			WeavePeerNickName: peer.NickName,
		}
	}

	return topology
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
