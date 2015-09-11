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
	"sync"

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

	mtx    sync.RWMutex
	status weaveStatus
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

// Tick implements Ticker
func (w *Weave) Tick() error {
	req, err := http.NewRequest("GET", w.url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Weave Tagger: got %d", resp.StatusCode)
	}

	var result weaveStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.status = result
	return nil
}

type psEntry struct {
	containerIDPrefix string
	macAddress        string
	ips               []string
}

func (w *Weave) ps() ([]psEntry, error) {
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

// Tag implements Tagger.
func (w *Weave) Tag(r report.Report) (report.Report, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	// Put information from weaveDNS on the container nodes
	for _, entry := range w.status.DNS.Entries {
		if entry.Tombstone > 0 {
			continue
		}
		nodeID := report.MakeContainerNodeID(w.hostID, entry.ContainerID)
		node, ok := r.Container.Nodes[nodeID]
		if !ok {
			continue
		}
		hostnames := report.IDList(strings.Fields(node.Metadata[WeaveDNSHostname]))
		hostnames = hostnames.Add(strings.TrimSuffix(entry.Hostname, "."))
		node.Metadata[WeaveDNSHostname] = strings.Join(hostnames, " ")
	}

	// Put information from weave ps on the container nodes
	psEntries, err := w.ps()
	if err != nil {
		return r, nil
	}
	containersByPrefix := map[string]report.Node{}
	for _, node := range r.Container.Nodes {
		prefix := node.Metadata[docker.ContainerID][:12]
		containersByPrefix[prefix] = node
	}
	for _, e := range psEntries {
		node, ok := containersByPrefix[e.containerIDPrefix]
		if !ok {
			continue
		}

		existingIPs := report.MakeIDList(docker.ExtractContainerIPs(node)...)
		existingIPs = existingIPs.Add(e.ips...)
		node.Metadata[docker.ContainerIPs] = strings.Join(existingIPs, " ")
		node.Metadata[WeaveMACAddress] = e.macAddress
	}
	return r, nil
}

// Report implements Reporter.
func (w *Weave) Report() (report.Report, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	r := report.MakeReport()
	for _, peer := range w.status.Router.Peers {
		r.Overlay.Nodes[report.MakeOverlayNodeID(peer.Name)] = report.MakeNodeWith(map[string]string{
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
