package render

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	mb                 = 1 << 20
	containerImageRank = 4
	containerRank      = 3
	processRank        = 2
	hostRank           = 1
	connectionsRank    = 0 // keep connections at the bottom until they are expandable in the UI
)

// DetailedNode is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type DetailedNode struct {
	ID         string  `json:"id"`
	LabelMajor string  `json:"label_major"`
	LabelMinor string  `json:"label_minor,omitempty"`
	Pseudo     bool    `json:"pseudo,omitempty"`
	Tables     []Table `json:"tables"`
}

// Table is a dataset associated with a node. It will be displayed in the
// detail panel when a user clicks on a node.
type Table struct {
	Title   string `json:"title"`   // e.g. Bandwidth
	Numeric bool   `json:"numeric"` // should the major column be right-aligned?
	Rank    int    `json:"-"`       // used to sort tables; not emitted.
	Rows    []Row  `json:"rows"`
}

// Row is a single entry in a Table dataset.
type Row struct {
	Key        string `json:"key"`                   // e.g. Ingress
	ValueMajor string `json:"value_major"`           // e.g. 25
	ValueMinor string `json:"value_minor,omitempty"` // e.g. KB/s
	Expandable bool   `json:"expandable,omitempty"`  // Whether it can be expanded (hidden by default)
}

type sortableRows []Row

func (r sortableRows) Len() int      { return len(r) }
func (r sortableRows) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r sortableRows) Less(i, j int) bool {
	switch {
	case r[i].Key != r[j].Key:
		return r[i].Key < r[j].Key

	case r[i].ValueMajor != r[j].ValueMajor:
		return r[i].ValueMajor < r[j].ValueMajor

	default:
		return r[i].ValueMinor < r[j].ValueMinor
	}
}

type sortableTables []Table

func (t sortableTables) Len() int           { return len(t) }
func (t sortableTables) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t sortableTables) Less(i, j int) bool { return t[i].Rank > t[j].Rank }

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r report.Report, n RenderableNode) DetailedNode {
	tables := sortableTables{}

	// Figure out if multiple hosts/containers are referenced by the renderableNode
	multiContainer, multiHost := getRenderingContext(r, n)

	// RenderableNode may be the result of merge operation(s), and so may have
	// multiple origins. The ultimate goal here is to generate tables to view
	// in the UI, so we skip the intermediate representations, but we could
	// add them later.
	connections := []Row{}
	for _, id := range n.Origins {
		if table, ok := OriginTable(r, id, multiHost, multiContainer); ok {
			tables = append(tables, table)
		} else if _, ok := r.Endpoint.NodeMetadatas[id]; ok {
			connections = append(connections, connectionDetailsRows(r.Endpoint, id)...)
		} else if _, ok := r.Address.NodeMetadatas[id]; ok {
			connections = append(connections, connectionDetailsRows(r.Address, id)...)
		}
	}

	if table, ok := connectionsTable(connections, r, n); ok {
		tables = append(tables, table)
	}

	// Sort tables by rank
	sort.Sort(tables)

	return DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
}

func getRenderingContext(r report.Report, n RenderableNode) (multiContainer bool, multiHost bool) {
	originHosts := make(map[string]struct{})
	originContainers := make(map[string]struct{})
	for _, id := range n.Origins {
		for _, topology := range r.Topologies() {
			if nmd, ok := topology.NodeMetadatas[id]; ok {
				originHosts[report.ExtractHostID(nmd)] = struct{}{}
				if id, ok := nmd.Metadata[docker.ContainerID]; ok {
					originContainers[id] = struct{}{}
				}
			}
			// Return early if possible
			multiHost = len(originHosts) > 1
			multiContainer = len(originContainers) > 1
			if multiHost && multiContainer {
				return
			}
		}
	}
	return
}

func connectionsTable(connections []Row, r report.Report, n RenderableNode) (Table, bool) {
	sec := r.Window.Seconds()
	rate := func(u *uint64) (float64, bool) {
		if u == nil {
			return 0.0, false
		}
		if sec <= 0 {
			return 0.0, true
		}
		return float64(*u) / sec, true
	}
	shortenByteRate := func(rate float64) (major, minor string) {
		switch {
		case rate > 1024*1024:
			return fmt.Sprintf("%.2f", rate/1024/1024), "MBps"
		case rate > 1024:
			return fmt.Sprintf("%.1f", rate/1024), "KBps"
		default:
			return fmt.Sprintf("%.0f", rate), "Bps"
		}
	}

	rows := []Row{}
	if n.EdgeMetadata.MaxConnCountTCP != nil {
		rows = append(rows, Row{"TCP connections", strconv.FormatUint(*n.EdgeMetadata.MaxConnCountTCP, 10), "", false})
	}
	if rate, ok := rate(n.EdgeMetadata.EgressPacketCount); ok {
		rows = append(rows, Row{"Egress packet rate", fmt.Sprintf("%.0f", rate), "packets/sec", false})
	}
	if rate, ok := rate(n.EdgeMetadata.IngressPacketCount); ok {
		rows = append(rows, Row{"Ingress packet rate", fmt.Sprintf("%.0f", rate), "packets/sec", false})
	}
	if rate, ok := rate(n.EdgeMetadata.EgressByteCount); ok {
		s, unit := shortenByteRate(rate)
		rows = append(rows, Row{"Egress byte rate", s, unit, false})
	}
	if rate, ok := rate(n.EdgeMetadata.IngressByteCount); ok {
		s, unit := shortenByteRate(rate)
		rows = append(rows, Row{"Ingress byte rate", s, unit, false})
	}
	if len(connections) > 0 {
		sort.Sort(sortableRows(connections))
		rows = append(rows, Row{Key: "Client", ValueMajor: "Server", Expandable: true})
		rows = append(rows, connections...)
	}
	if len(rows) > 0 {
		return Table{"Connections", true, connectionsRank, rows}, true
	}
	return Table{}, false
}

// OriginTable produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func OriginTable(r report.Report, originID string, addHostTags bool, addContainerTags bool) (Table, bool) {
	if nmd, ok := r.Process.NodeMetadatas[originID]; ok {
		return processOriginTable(nmd, addHostTags, addContainerTags)
	}
	if nmd, ok := r.Container.NodeMetadatas[originID]; ok {
		return containerOriginTable(nmd, addHostTags)
	}
	if nmd, ok := r.ContainerImage.NodeMetadatas[originID]; ok {
		return containerImageOriginTable(nmd)
	}
	if nmd, ok := r.Host.NodeMetadatas[originID]; ok {
		return hostOriginTable(nmd)
	}
	return Table{}, false
}

func connectionDetailsRows(topology report.Topology, originID string) []Row {
	rows := []Row{}
	labeler := func(nodeID string) (string, bool) {
		if _, addr, port, ok := report.ParseEndpointNodeID(nodeID); ok {
			return fmt.Sprintf("%s:%s", addr, port), true
		}
		if _, addr, ok := report.ParseAddressNodeID(nodeID); ok {
			return addr, true
		}
		return "", false
	}
	local, ok := labeler(originID)
	if !ok {
		return rows
	}
	// Firstly, collection outgoing connections from this node.
	originAdjID := report.MakeAdjacencyID(originID)
	for _, serverNodeID := range topology.Adjacency[originAdjID] {
		remote, ok := labeler(serverNodeID)
		if !ok {
			continue
		}
		rows = append(rows, Row{
			Key:        local,
			ValueMajor: remote,
			Expandable: true,
		})
	}
	// Next, scan the topology for incoming connections to this node.
	for clientAdjID, serverNodeIDs := range topology.Adjacency {
		if clientAdjID == originAdjID {
			continue
		}
		if !serverNodeIDs.Contains(originID) {
			continue
		}
		clientNodeID, ok := report.ParseAdjacencyID(clientAdjID)
		if !ok {
			continue
		}
		remote, ok := labeler(clientNodeID)
		if !ok {
			continue
		}
		rows = append(rows, Row{
			Key:        remote,
			ValueMajor: local,
			Expandable: true,
		})
	}
	return rows
}

func processOriginTable(nmd report.NodeMetadata, addHostTag bool, addContainerTag bool) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{process.PPID, "Parent PID"},
		{process.Cmdline, "Command"},
		{process.Threads, "# Threads"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	if containerID, ok := nmd.Metadata[docker.ContainerID]; ok && addContainerTag {
		rows = append([]Row{{Key: "Container ID", ValueMajor: containerID}}, rows...)
	}

	if addHostTag {
		rows = append([]Row{{Key: "Host", ValueMajor: report.ExtractHostID(nmd)}}, rows...)
	}

	var (
		title           = "Process"
		name, commFound = nmd.Metadata[process.Comm]
		pid, pidFound   = nmd.Metadata[process.PID]
	)
	if commFound {
		title += ` "` + name + `"`
	}
	if pidFound {
		title += " (" + pid + ")"
	}
	return Table{
		Title:   title,
		Numeric: false,
		Rows:    rows,
		Rank:    processRank,
	}, len(rows) > 0 || commFound || pidFound
}

func containerOriginTable(nmd report.NodeMetadata, addHostTag bool) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ContainerID, "ID"},
		{overlay.WeaveDNSHostname, "Weave DNS Hostname"},
		{docker.ImageID, "Image ID"},
		{docker.ContainerPorts, "Ports"},
		{docker.ContainerCreated, "Created"},
		{docker.ContainerCommand, "Command"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	if val, ok := nmd.Metadata[docker.MemoryUsage]; ok {
		memory, err := strconv.ParseFloat(val, 64)
		if err == nil {
			memoryStr := fmt.Sprintf("%0.2f", memory/float64(mb))
			rows = append(rows, Row{Key: "Memory Usage (MB):", ValueMajor: memoryStr, ValueMinor: ""})
		}
	}
	if addHostTag {
		rows = append([]Row{{Key: "Host", ValueMajor: report.ExtractHostID(nmd)}}, rows...)
	}

	var (
		title           = "Container"
		name, nameFound = nmd.Metadata[docker.ContainerName]
	)
	if nameFound {
		title += ` "` + name + `"`
	}

	return Table{
		Title:   title,
		Numeric: false,
		Rows:    rows,
		Rank:    containerRank,
	}, len(rows) > 0 || nameFound
}

func containerImageOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ImageID, "Image ID"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}
	title := "Container Image"
	var (
		nameFound bool
		name      string
	)
	if name, nameFound = nmd.Metadata[docker.ImageName]; nameFound {
		title += ` "` + name + `"`
	}
	return Table{
		Title:   title,
		Numeric: false,
		Rows:    rows,
		Rank:    containerImageRank,
	}, len(rows) > 0 || nameFound
}

func hostOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{host.Load, "Load"},
		{host.OS, "Operating system"},
		{host.KernelVersion, "Kernel version"},
		{host.Uptime, "Uptime"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	title := "Host"
	var (
		name      string
		foundName bool
	)
	if name, foundName = nmd.Metadata[host.HostName]; foundName {
		title += ` "` + name + `"`
	}
	return Table{
		Title:   title,
		Numeric: false,
		Rows:    rows,
		Rank:    hostRank,
	}, len(rows) > 0 || foundName
}
