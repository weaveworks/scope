package overlay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/common/exec"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

const (
	// WeavePeerName is the key for the peer name, typically a MAC address.
	WeavePeerName = "weave_peer_name"

	// WeavePeerNickName is the key for the peer nickname, typically a
	// hostname.
	WeavePeerNickName = "weave_peer_nick_name"

	// WeaveDNSHostname is the ket for the WeaveDNS hostname
	WeaveDNSHostname = "weave_dns_hostname"

	// WeaveMACAddress is the key for the mac address of the container on the
	// weave network, to be found in container node metadata
	WeaveMACAddress = "weave_mac_address"
)

var weavePsMatch = regexp.MustCompile(`^([0-9a-f]{12}) ((?:[0-9a-f][0-9a-f]\:){5}(?:[0-9a-f][0-9a-f]))(.*)$`)
var ipMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})(/[0-9]+)`)

// Weave represents a single Weave router, presumably on the same host
// as the probe. It is both a Reporter and a Tagger: it produces an Overlay
// topology, and (in theory) can tag existing topologies with foreign keys to
// overlay -- though I'm not sure what that would look like in practice right
// now.
type Weave struct {
	url    string
	hostID string
}

type weaveStatus struct {
	Router struct {
		Peers []struct {
			Name     string
			NickName string
		}
	}

	DNS struct {
		Entries []struct {
			Hostname    string
			ContainerID string
			Tombstone   int64
		}
	}
}

// NewWeave returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeave(hostID, weaveRouterAddress string) (*Weave, error) {
	s, err := sanitize("http://", 6784, "/report")(weaveRouterAddress)
	if err != nil {
		return nil, err
	}
	return &Weave{
		url:    s,
		hostID: hostID,
	}, nil
}

func (w Weave) update() (weaveStatus, error) {
	var result weaveStatus
	req, err := http.NewRequest("GET", w.url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("Weave Tagger: got %d", resp.StatusCode)
	}

	return result, json.NewDecoder(resp.Body).Decode(&result)
}

type psEntry struct {
	containerIDPrefix string
	macAddress        string
	ips               []string
}

func (w Weave) ps() ([]psEntry, error) {
	var result []psEntry
	cmd := exec.Command("weave", "--local", "ps")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}
	if err := cmd.Start(); err != nil {
		return result, err
	}
	defer func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Weave tagger, cmd failed: %v", err)
		}
	}()
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		groups := weavePsMatch.FindStringSubmatch(line)
		if len(groups) == 0 {
			continue
		}
		containerIDPrefix, macAddress, ips := groups[1], groups[2], []string{}
		for _, ipGroup := range ipMatch.FindAllStringSubmatch(groups[3], -1) {
			ips = append(ips, ipGroup[1])
		}
		result = append(result, psEntry{containerIDPrefix, macAddress, ips})
	}
	return result, scanner.Err()
}

func (w Weave) tagContainer(r report.Report, containerIDPrefix, macAddress string, ips []string) {
	for nodeid, nmd := range r.Container.NodeMetadatas {
		idPrefix := nmd.Metadata[docker.ContainerID][:12]
		if idPrefix != containerIDPrefix {
			continue
		}

		existingIPs := report.MakeIDList(docker.ExtractContainerIPs(nmd)...)
		existingIPs = existingIPs.Add(ips...)
		nmd.Metadata[docker.ContainerIPs] = strings.Join(existingIPs, " ")
		nmd.Metadata[WeaveMACAddress] = macAddress
		r.Container.NodeMetadatas[nodeid] = nmd
		break
	}
}

// Tag implements Tagger.
func (w Weave) Tag(r report.Report) (report.Report, error) {
	status, err := w.update()
	if err != nil {
		return r, nil
	}

	for _, entry := range status.DNS.Entries {
		if entry.Tombstone > 0 {
			continue
		}
		nodeID := report.MakeContainerNodeID(w.hostID, entry.ContainerID)
		node, ok := r.Container.NodeMetadatas[nodeID]
		if !ok {
			continue
		}
		hostnames := report.IDList(strings.Fields(node.Metadata[WeaveDNSHostname]))
		hostnames = hostnames.Add(strings.TrimSuffix(entry.Hostname, "."))
		node.Metadata[WeaveDNSHostname] = strings.Join(hostnames, " ")
		r.Container.NodeMetadatas[nodeID] = node
	}

	psEntries, err := w.ps()
	if err != nil {
		return r, nil
	}
	for _, e := range psEntries {
		w.tagContainer(r, e.containerIDPrefix, e.macAddress, e.ips)
	}
	return r, nil
}

// Report implements Reporter.
func (w Weave) Report() (report.Report, error) {
	r := report.MakeReport()
	status, err := w.update()
	if err != nil {
		return r, err
	}

	for _, peer := range status.Router.Peers {
		r.Overlay.NodeMetadatas[report.MakeOverlayNodeID(peer.Name)] = report.MakeNodeMetadataWith(map[string]string{
			WeavePeerName:     peer.Name,
			WeavePeerNickName: peer.NickName,
		})
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
