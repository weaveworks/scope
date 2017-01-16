package overlay

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/common/backoff"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"

	docker_client "github.com/fsouza/go-dockerclient"
)

// Keys for use in Node
const (
	WeavePeerName              = "weave_peer_name"
	WeavePeerNickName          = "weave_peer_nick_name"
	WeaveDNSHostname           = "weave_dns_hostname"
	WeaveMACAddress            = "weave_mac_address"
	WeaveVersion               = "weave_version"
	WeaveEncryption            = "weave_encryption"
	WeaveProtocol              = "weave_protocol"
	WeavePeerDiscovery         = "weave_peer_discovery"
	WeaveTargetCount           = "weave_target_count"
	WeaveConnectionCount       = "weave_connection_count"
	WeavePeerCount             = "weave_peer_count"
	WeaveTrustedSubnets        = "weave_trusted_subnet_count"
	WeaveIPAMTableID           = "weave_ipam_table"
	WeaveIPAMStatus            = "weave_ipam_status"
	WeaveIPAMRange             = "weave_ipam_range"
	WeaveIPAMDefaultSubnet     = "weave_ipam_default_subnet"
	WeaveDNSTableID            = "weave_dns_table"
	WeaveDNSDomain             = "weave_dns_domain"
	WeaveDNSUpstream           = "weave_dns_upstream"
	WeaveDNSTTL                = "weave_dns_ttl"
	WeaveDNSEntryCount         = "weave_dns_entry_count"
	WeaveProxyTableID          = "weave_proxy_table"
	WeaveProxyStatus           = "weave_proxy_status"
	WeaveProxyAddress          = "weave_proxy_address"
	WeavePluginTableID         = "weave_plugin_table"
	WeavePluginStatus          = "weave_plugin_status"
	WeavePluginDriver          = "weave_plugin_driver"
	WeaveConnectionsListPrefix = "weave_connections_list_"
)

var (
	// NewDockerClientStub is used for testing
	NewDockerClientStub = newDockerClient

	containerNotRunningRE = regexp.MustCompile(`Container .* is not running\n`)

	containerMetadata = report.MetadataTemplates{
		WeaveMACAddress:  {ID: WeaveMACAddress, Label: "Weave MAC", From: report.FromLatest, Priority: 17},
		WeaveDNSHostname: {ID: WeaveDNSHostname, Label: "Weave DNS Name", From: report.FromLatest, Priority: 18},
	}

	weaveMetadata = report.MetadataTemplates{
		WeaveVersion:         {ID: WeaveVersion, Label: "Version", From: report.FromLatest, Priority: 1},
		WeaveProtocol:        {ID: WeaveProtocol, Label: "Protocol", From: report.FromLatest, Priority: 2},
		WeavePeerName:        {ID: WeavePeerName, Label: "Name", From: report.FromLatest, Priority: 3},
		WeaveEncryption:      {ID: WeaveEncryption, Label: "Encryption", From: report.FromLatest, Priority: 4},
		WeavePeerDiscovery:   {ID: WeavePeerDiscovery, Label: "Peer Discovery", From: report.FromLatest, Priority: 5},
		WeaveTargetCount:     {ID: WeaveTargetCount, Label: "Targets", From: report.FromLatest, Priority: 6},
		WeaveConnectionCount: {ID: WeaveConnectionCount, Label: "Connections", From: report.FromLatest, Priority: 8},
		WeavePeerCount:       {ID: WeavePeerCount, Label: "Peers", From: report.FromLatest, Priority: 7},
		WeaveTrustedSubnets:  {ID: WeaveTrustedSubnets, Label: "Trusted Subnets", From: report.FromSets, Priority: 9},
	}

	weavePropertyListTemplates = report.PropertyListTemplates{
		WeaveIPAMTableID: {ID: WeaveIPAMTableID, Label: "IPAM",
			FixedProperties: map[string]string{
				WeaveIPAMStatus:        "Status",
				WeaveIPAMRange:         "Range",
				WeaveIPAMDefaultSubnet: "Default Subnet",
			},
		},
		WeaveDNSTableID: {ID: WeaveDNSTableID, Label: "DNS",
			FixedProperties: map[string]string{
				WeaveDNSDomain:     "Domain",
				WeaveDNSUpstream:   "Upstream",
				WeaveDNSTTL:        "TTL",
				WeaveDNSEntryCount: "Entries",
			},
		},
		WeaveProxyTableID: {ID: WeaveProxyTableID, Label: "Proxy",
			FixedProperties: map[string]string{
				WeaveProxyStatus:  "Status",
				WeaveProxyAddress: "Address",
			},
		},
		WeavePluginTableID: {ID: WeavePluginTableID, Label: "Plugin",
			FixedProperties: map[string]string{
				WeavePluginStatus: "Status",
				WeavePluginDriver: "Driver Name",
			},
		},
		WeaveConnectionsListPrefix: {
			ID:     WeaveConnectionsListPrefix,
			Label:  "Connections",
			Prefix: WeaveConnectionsListPrefix,
		},
	}
)

// DockerClient is used for testing
type DockerClient interface {
	CreateExec(docker_client.CreateExecOptions) (*docker_client.Exec, error)
	StartExec(string, docker_client.StartExecOptions) error
	InspectContainer(id string) (*docker_client.Container, error)
}

func newDockerClient(endpoint string) (DockerClient, error) {
	return docker_client.NewClient(endpoint)
}

// Weave represents a single Weave router, presumably on the same host
// as the probe. It is both a Reporter and a Tagger: it produces an Overlay
// topology, and (in theory) can tag existing topologies with foreign keys to
// overlay -- though I'm not sure what that would look like in practice right
// now.
type Weave struct {
	client       weave.Client
	dockerClient DockerClient
	hostID       string

	mtx                sync.RWMutex
	statusCache        weave.Status
	psCache            map[string]weave.PSEntry
	proxyRunningCache  bool
	proxyAddressCache  string
	pluginRunningCache bool

	backoff       backoff.Interface
	psBackoff     backoff.Interface
	proxyBackoff  backoff.Interface
	pluginBackoff backoff.Interface
}

// NewWeave returns a new Weave tagger based on the Weave router at
// address. The address should be an IP or FQDN, no port.
func NewWeave(hostID string, client weave.Client, dockerEndpoint string) (*Weave, error) {
	dockerClient, err := NewDockerClientStub(dockerEndpoint)
	if err != nil {
		return nil, err
	}
	w := &Weave{
		client:       client,
		dockerClient: dockerClient,
		hostID:       hostID,
		psCache:      map[string]weave.PSEntry{},
	}

	w.backoff = backoff.New(w.status, "collecting weave status")
	w.backoff.SetInitialBackoff(5 * time.Second)
	go w.backoff.Start()

	w.psBackoff = backoff.New(w.ps, "collecting weave ps")
	w.psBackoff.SetInitialBackoff(10 * time.Second)
	go w.psBackoff.Start()

	w.proxyBackoff = backoff.New(w.proxyStatus, "collecting weave proxy status")
	w.proxyBackoff.SetInitialBackoff(10 * time.Second)
	go w.proxyBackoff.Start()

	w.pluginBackoff = backoff.New(w.pluginStatus, "collecting weave plugin status")
	w.pluginBackoff.SetInitialBackoff(10 * time.Second)
	go w.pluginBackoff.Start()

	return w, nil
}

// Name of this reporter/tagger/ticker, for metrics gathering
func (*Weave) Name() string { return "Weave" }

// Stop gathering weave ps output.
func (w *Weave) Stop() {
	w.backoff.Stop()
	w.psBackoff.Stop()
	w.proxyBackoff.Stop()
	w.pluginBackoff.Stop()
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

func filterContainerNotFound(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *docker_client.Error:
		// This is really ugly, but this error comes from the client in some cases
		// and there is no other way to distinguish it :(
		dockerError := err.(*docker_client.Error)
		if containerNotRunningRE.MatchString(dockerError.Message) {
			return nil
		}
	case *docker_client.ContainerNotRunning:
		return nil
	case *docker_client.NoSuchContainer:
		return nil
	}

	return err
}

func (w *Weave) proxyStatus() (bool, error) {
	update := func(running bool, address string) {
		w.mtx.Lock()
		defer w.mtx.Unlock()
		w.proxyRunningCache = running
		w.proxyAddressCache = address
	}

	exec, err := w.dockerClient.CreateExec(docker_client.CreateExecOptions{
		AttachStdout: true,
		Cmd:          []string{"curl", "-s", "--unix-socket", "status.sock", "http:/status"},
		Container:    "weaveproxy",
	})
	if err != nil {
		update(false, "")
		return false, filterContainerNotFound(err)
	}
	out := bytes.NewBuffer(nil)
	err = w.dockerClient.StartExec(exec.ID, docker_client.StartExecOptions{
		OutputStream: out,
	})
	if err != nil {
		update(true, "")
		return false, filterContainerNotFound(err)
	}

	update(true, out.String())

	return false, nil
}

func (w *Weave) pluginStatus() (bool, error) {
	update := func(running bool) {
		w.mtx.Lock()
		defer w.mtx.Unlock()

		w.pluginRunningCache = running
	}

	c, err := w.dockerClient.InspectContainer("weaveplugin")
	if err != nil {
		update(false)
		return false, filterContainerNotFound(err)
	}

	update(c.State.Running)

	return false, nil
}

// Tag implements Tagger.
func (w *Weave) Tag(r report.Report) (report.Report, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	// Put information from weaveDNS on the container nodes
	if w.statusCache.DNS != nil {
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
	r.Container = r.Container.WithMetadataTemplates(containerMetadata)
	r.Overlay = r.Overlay.WithMetadataTemplates(weaveMetadata).WithPropertyListTemplates(weavePropertyListTemplates)

	// We report nodes for all peers (not just the current node) to highlight peers not monitored by Scope
	// (i.e. without a running probe)
	// Note: this will cause redundant information (n^2) if all peers have a running probe
	for _, peer := range w.statusCache.Router.Peers {
		node := w.getPeerNode(peer)
		r.Overlay.AddNode(node)
	}
	if w.statusCache.IPAM != nil {
		r.Overlay.AddNode(
			report.MakeNode(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, w.statusCache.Router.Name)).
				WithSet(host.LocalNetworks, report.MakeStringSet(w.statusCache.IPAM.DefaultSubnet)),
		)
	}
	return r, nil
}

// getPeerNode obtains an Overlay topology node for representing a peer in the Weave network
func (w *Weave) getPeerNode(peer weave.Peer) report.Node {
	node := report.MakeNode(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, peer.Name))
	latests := map[string]string{
		WeavePeerName:     peer.Name,
		WeavePeerNickName: peer.NickName,
	}

	// Peer corresponding to current host
	if peer.Name == w.statusCache.Router.Name {
		latests, node = w.addCurrentPeerInfo(latests, node)
	}

	for _, conn := range peer.Connections {
		if conn.Outbound {
			node = node.WithAdjacent(report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, conn.Name))
		}
	}

	return node.WithLatests(latests)
}

// addCurrentPeerInfo adds information exclusive to the Overlay topology node representing current Weave Net peer
// (i.e. in the same host as the reporting Scope probe)
func (w *Weave) addCurrentPeerInfo(latests map[string]string, node report.Node) (map[string]string, report.Node) {
	latests[report.HostNodeID] = w.hostID
	latests[WeaveVersion] = w.statusCache.Version
	latests[WeaveEncryption] = "disabled"
	if w.statusCache.Router.Encryption {
		latests[WeaveEncryption] = "enabled"
	}
	latests[WeavePeerDiscovery] = "disabled"
	if w.statusCache.Router.PeerDiscovery {
		latests[WeavePeerDiscovery] = "enabled"
	}
	if w.statusCache.Router.ProtocolMinVersion == w.statusCache.Router.ProtocolMaxVersion {
		latests[WeaveProtocol] = fmt.Sprintf("%d", w.statusCache.Router.ProtocolMinVersion)
	} else {
		latests[WeaveProtocol] = fmt.Sprintf("%d..%d", w.statusCache.Router.ProtocolMinVersion, w.statusCache.Router.ProtocolMaxVersion)
	}
	latests[WeaveTargetCount] = fmt.Sprintf("%d", len(w.statusCache.Router.Targets))
	latests[WeaveConnectionCount] = fmt.Sprintf("%d", len(w.statusCache.Router.Connections))
	latests[WeavePeerCount] = fmt.Sprintf("%d", len(w.statusCache.Router.Peers))
	node = node.WithSet(WeaveTrustedSubnets, report.MakeStringSet(w.statusCache.Router.TrustedSubnets...))
	if w.statusCache.IPAM != nil {
		latests[WeaveIPAMStatus] = getIPAMStatus(*w.statusCache.IPAM)
		latests[WeaveIPAMRange] = w.statusCache.IPAM.Range
		latests[WeaveIPAMDefaultSubnet] = w.statusCache.IPAM.DefaultSubnet
	}
	if w.statusCache.DNS != nil {
		latests[WeaveDNSDomain] = w.statusCache.DNS.Domain
		latests[WeaveDNSUpstream] = strings.Join(w.statusCache.DNS.Upstream, ", ")
		latests[WeaveDNSTTL] = fmt.Sprintf("%d", w.statusCache.DNS.TTL)
		dnsEntryCount := 0
		for _, entry := range w.statusCache.DNS.Entries {
			if entry.Tombstone == 0 {
				dnsEntryCount++
			}
		}
		latests[WeaveDNSEntryCount] = fmt.Sprintf("%d", dnsEntryCount)
	}
	latests[WeaveProxyStatus] = "not running"
	if w.proxyRunningCache {
		latests[WeaveProxyStatus] = "running"
		latests[WeaveProxyAddress] = w.proxyAddressCache
	}
	latests[WeavePluginStatus] = "not running"
	if w.pluginRunningCache {
		latests[WeavePluginStatus] = "running"
		latests[WeavePluginDriver] = "weave"
	}
	node = node.AddPrefixPropertyList(WeaveConnectionsListPrefix, getConnectionsList(w.statusCache.Router))
	node = node.WithParents(report.EmptySets.Add(report.Host, report.MakeStringSet(w.hostID)))

	return latests, node
}

func getConnectionsList(router weave.Router) map[string]string {
	const (
		outboundArrow = "->"
		inboundArrow  = "<-"
	)
	table := make(map[string]string, len(router.Connections))
	for _, conn := range router.Connections {
		arrow := inboundArrow
		if conn.Outbound {
			arrow = outboundArrow
		}
		// TODO: we should probably use a multicolumn table for this
		//       but there is no mechanism to support it yet.
		key := fmt.Sprintf("%s %s", arrow, conn.Address)
		value := fmt.Sprintf("%s, %s", conn.State, conn.Info)
		table[key] = value
	}
	return table
}

func getIPAMStatus(ipam weave.IPAM) string {
	allIPAMOwnersUnreachable := func(ipam weave.IPAM) bool {
		for _, entry := range ipam.Entries {
			if entry.Size > 0 && entry.IsKnownPeer {
				return false
			}
		}
		return true
	}

	if len(ipam.Entries) > 0 {
		if allIPAMOwnersUnreachable(ipam) {
			return "all ranges owned by unreachable peers"
		} else if len(ipam.PendingAllocates) > 0 {
			return "waiting for grant"

		} else {
			return "ready"
		}
	}

	if ipam.Paxos != nil {
		if ipam.Paxos.Elector {
			return fmt.Sprintf(
				"awaiting consensus (quorum: %d, known: %d)",
				ipam.Paxos.Quorum,
				ipam.Paxos.KnownNodes,
			)
		}
		return "priming"
	}

	return "idle"
}
