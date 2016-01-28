package overlay

import (
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/scope/common/backoff"
	"github.com/weaveworks/scope/common/weave"
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

// Weave represents a single Weave router, presumably on the same host
// as the probe. It is both a Reporter and a Tagger: it produces an Overlay
// topology, and (in theory) can tag existing topologies with foreign keys to
// overlay -- though I'm not sure what that would look like in practice right
// now.
type Weave struct {
	client weave.Client
	hostID string

	mtx         sync.RWMutex
	statusCache weave.Status
	psCache     map[string]weave.PSEntry

	backoff backoff.Interface
}

// NewWeave returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeave(hostID string, client weave.Client) *Weave {
	w := &Weave{
		client:  client,
		hostID:  hostID,
		psCache: map[string]weave.PSEntry{},
	}

	w.backoff = backoff.New(w.collect, "collecting weave info")
	w.backoff.SetInitialBackoff(5 * time.Second)
	go w.backoff.Start()
	return w
}

// Name of this reporter/tagger/ticker, for metrics gathering
func (*Weave) Name() string { return "Weave" }

// Stop gathering weave ps output.
func (w *Weave) Stop() {
	w.backoff.Stop()
}

func (w *Weave) collect() (done bool, err error) {
	// If we fail to get info from weave
	// we should wipe away stale data
	defer func() {
		if err != nil {
			w.mtx.Lock()
			defer w.mtx.Unlock()
			w.statusCache = weave.Status{}
			w.psCache = map[string]weave.PSEntry{}
		}
	}()

	if err = w.ps(); err != nil {
		return
	}
	if err = w.status(); err != nil {
		return
	}

	return
}

func (w *Weave) ps() error {
	psEntriesByPrefix, err := w.client.PS()
	if err != nil {
		return err
	}

	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.psCache = psEntriesByPrefix
	return nil
}

func (w *Weave) status() error {
	status, err := w.client.Status()
	if err != nil {
		return err
	}

	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.statusCache = status
	return nil
}

// Tag implements Tagger.
func (w *Weave) Tag(r report.Report) (report.Report, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	// Put information from weaveDNS on the container nodes
	for _, entry := range w.statusCache.DNS.Entries {
		if entry.Tombstone > 0 {
			continue
		}
		nodeID := report.MakeContainerNodeID(entry.ContainerID)
		node, ok := r.Container.Nodes[nodeID]
		if !ok {
			continue
		}
		w, _ := node.Latest.Lookup(WeaveDNSHostname)
		hostnames := report.IDList(strings.Fields(w))
		hostnames = hostnames.Add(strings.TrimSuffix(entry.Hostname, "."))
		r.Container.Nodes[nodeID] = node.WithLatests(map[string]string{WeaveDNSHostname: strings.Join(hostnames, " ")})
	}

	// Put information from weave ps on the container nodes
	for id, node := range r.Container.Nodes {
		prefix, _ := node.Latest.Lookup(docker.ContainerID)
		prefix = prefix[:12]
		entry, ok := w.psCache[prefix]
		if !ok {
			continue
		}

		ipsWithScope := report.MakeStringSet()
		for _, ip := range entry.IPs {
			ipsWithScope = ipsWithScope.Add(report.MakeAddressNodeID("", ip))
		}
		node = node.WithSet(docker.ContainerIPs, report.MakeStringSet(entry.IPs...))
		node = node.WithSet(docker.ContainerIPsWithScopes, ipsWithScope)
		node = node.WithLatests(map[string]string{
			WeaveMACAddress: entry.MACAddress,
		})
		r.Container.Nodes[id] = node
	}
	return r, nil
}

// Report implements Reporter.
func (w *Weave) Report() (report.Report, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	r := report.MakeReport()
	for _, peer := range w.statusCache.Router.Peers {
		r.Overlay.AddNode(report.MakeOverlayNodeID(peer.Name), report.MakeNodeWith(map[string]string{
			WeavePeerName:     peer.Name,
			WeavePeerNickName: peer.NickName,
		}))
	}
	return r, nil
}
