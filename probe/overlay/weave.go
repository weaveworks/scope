package overlay

import (
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/scope/common/backoff"
	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
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

	backoff   backoff.Interface
	psBackoff backoff.Interface
}

// NewWeave returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeave(hostID string, client weave.Client) *Weave {
	w := &Weave{
		client:  client,
		hostID:  hostID,
		psCache: map[string]weave.PSEntry{},
	}

	w.backoff = backoff.New(w.status, "collecting weave status")
	w.backoff.SetInitialBackoff(5 * time.Second)
	go w.backoff.Start()

	w.psBackoff = backoff.New(w.ps, "collecting weave ps")
	w.psBackoff.SetInitialBackoff(10 * time.Second)
	go w.psBackoff.Start()

	return w
}

// Name of this reporter/tagger/ticker, for metrics gathering
func (*Weave) Name() string { return "Weave" }

// Stop gathering weave ps output.
func (w *Weave) Stop() {
	w.backoff.Stop()
	w.psBackoff.Stop()
}

func (w *Weave) ps() (bool, error) {
	psEntriesByPrefix, err := w.client.PS()

	w.mtx.Lock()
	defer w.mtx.Unlock()

	if err != nil {
		w.psCache = map[string]weave.PSEntry{}
	} else {
		w.psCache = psEntriesByPrefix
	}
	return false, err
}

func (w *Weave) status() (bool, error) {
	status, err := w.client.Status()

	w.mtx.Lock()
	defer w.mtx.Unlock()

	if err != nil {
		w.statusCache = weave.Status{}
	} else {
		w.statusCache = status
	}
	return false, err
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
	const maxPrefixSize = 12
	for id, node := range r.Container.Nodes {
		prefix, ok := node.Latest.Lookup(docker.ContainerID)
		if !ok {
			continue
		}
		if len(prefix) > maxPrefixSize {
			prefix = prefix[:maxPrefixSize]
		}
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
	r.Container = r.Container.WithMetadataTemplates(report.MetadataTemplates{
		WeaveMACAddress:  {ID: WeaveMACAddress, Label: "Weave MAC", From: report.FromLatest, Priority: 17},
		WeaveDNSHostname: {ID: WeaveDNSHostname, Label: "Weave DNS Name", From: report.FromLatest, Priority: 18},
	})
	for _, peer := range w.statusCache.Router.Peers {
		node := report.MakeNodeWith(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, peer.Name),
			map[string]string{
				WeavePeerName:     peer.Name,
				WeavePeerNickName: peer.NickName,
			})
		if peer.Name == w.statusCache.Router.Name {
			node = node.WithLatest(report.HostNodeID, mtime.Now(), w.hostID)
			node = node.WithParents(report.EmptySets.Add(report.Host, report.MakeStringSet(w.hostID)))
		}
		for _, conn := range peer.Connections {
			if conn.Outbound {
				node = node.WithAdjacent(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, conn.Name))
			}
		}
		r.Overlay.AddNode(node)
	}
	if w.statusCache.IPAM.DefaultSubnet != "" {
		r.Overlay.AddNode(
			report.MakeNode(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, w.statusCache.Router.Name)).WithSets(
				report.MakeSets().Add(host.LocalNetworks, report.MakeStringSet(w.statusCache.IPAM.DefaultSubnet)),
			),
		)
	}
	return r, nil
}
