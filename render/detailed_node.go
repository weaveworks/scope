package render

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	mb                 = 1 << 20
	connectionsRank    = 100
	containerImageRank = 4
	containerRank      = 3
	processRank        = 2
	hostRank           = 1
	endpointRank       = 0 // this is the least important table, so sort to bottom
	addressRank        = 0 // also least important; never merged with endpoints
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
}

type tables []Table

func (t tables) Len() int           { return len(t) }
func (t tables) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tables) Less(i, j int) bool { return t[i].Rank > t[j].Rank }

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r report.Report, n RenderableNode) DetailedNode {
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

	tables := tables{}
	{
		rows := []Row{}
		if n.EdgeMetadata.MaxConnCountTCP != nil {
			rows = append(rows, Row{"TCP connections", strconv.FormatUint(*n.EdgeMetadata.MaxConnCountTCP, 10), ""})
		}
		if rate, ok := rate(n.EdgeMetadata.EgressPacketCount); ok {
			rows = append(rows, Row{"Egress packet rate", fmt.Sprintf("%.0f", rate), "packets/sec"})
		}
		if rate, ok := rate(n.EdgeMetadata.IngressPacketCount); ok {
			rows = append(rows, Row{"Ingress packet rate", fmt.Sprintf("%.0f", rate), "packets/sec"})
		}
		if rate, ok := rate(n.EdgeMetadata.EgressByteCount); ok {
			s, unit := shortenByteRate(rate)
			rows = append(rows, Row{"Egress byte rate", s, unit})
		}
		if rate, ok := rate(n.EdgeMetadata.IngressByteCount); ok {
			s, unit := shortenByteRate(rate)
			rows = append(rows, Row{"Ingress byte rate", s, unit})
		}
		if len(rows) > 0 {
			tables = append(tables, Table{"Connections", true, connectionsRank, rows})
		}
	}

	// RenderableNode may be the result of merge operation(s), and so may have
	// multiple origins. The ultimate goal here is to generate tables to view
	// in the UI, so we skip the intermediate representations, but we could
	// add them later.
	connections := []Row{}
	for _, id := range n.Origins {
		if table, ok := OriginTable(r, id); ok {
			tables = append(tables, table)
		} else if nmd, ok := r.Endpoint.NodeMetadatas[id]; ok {
			connections = append(connections, connectionDetailsRows(r.Endpoint, id, nmd)...)
		}
	}
	if len(connections) > 0 {
		tables = append(tables, connectionDetailsTable(connections))
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

// OriginTable produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func OriginTable(r report.Report, originID string) (Table, bool) {
	if nmd, ok := r.Address.NodeMetadatas[originID]; ok {
		return addressOriginTable(nmd)
	}
	if nmd, ok := r.Process.NodeMetadatas[originID]; ok {
		return processOriginTable(nmd)
	}
	if nmd, ok := r.Container.NodeMetadatas[originID]; ok {
		return containerOriginTable(nmd)
	}
	if nmd, ok := r.ContainerImage.NodeMetadatas[originID]; ok {
		return containerImageOriginTable(nmd)
	}
	if nmd, ok := r.Host.NodeMetadatas[originID]; ok {
		return hostOriginTable(nmd)
	}
	return Table{}, false
}

func connectionDetailsRows(endpointTopology report.Topology, originID string, nmd report.NodeMetadata) []Row {
	rows := []Row{}
	local := fmt.Sprintf("%s:%s", nmd.Metadata[endpoint.Addr], nmd.Metadata[endpoint.Port])
	adjacencies := endpointTopology.Adjacency[report.MakeAdjacencyID(originID)]
	sort.Strings(adjacencies)
	for _, adj := range adjacencies {
		if _, address, port, ok := report.ParseEndpointNodeID(adj); ok {
			rows = append(rows, Row{
				Key:        local,
				ValueMajor: fmt.Sprintf("%s:%s", address, port),
			})
		}
	}
	return rows
}

func connectionDetailsTable(connectionRows []Row) Table {
	return Table{
		Title:   "Connection Details",
		Numeric: false,
		Rows:    append([]Row{{Key: "Local", ValueMajor: "Remote"}}, connectionRows...),
		Rank:    endpointRank,
	}
}

func addressOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd.Metadata[endpoint.Addr]; ok {
		rows = append(rows, Row{"Address", val, ""})
	}
	return Table{
		Title:   "Origin Address",
		Numeric: false,
		Rows:    rows,
		Rank:    addressRank,
	}, len(rows) > 0
}

func processOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{process.Comm, "Name (comm)"},
		{process.PID, "PID"},
		{process.PPID, "Parent PID"},
		{process.Cmdline, "Command"},
		{process.Threads, "# Threads"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	return Table{
		Title:   "Origin Process",
		Numeric: false,
		Rows:    rows,
		Rank:    processRank,
	}, len(rows) > 0
}

func containerOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ContainerID, "ID"},
		{docker.ContainerName, "Name"},
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

	return Table{
		Title:   "Origin Container",
		Numeric: false,
		Rows:    rows,
		Rank:    containerRank,
	}, len(rows) > 0
}

func containerImageOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ImageID, "Image ID"},
		{docker.ImageName, "Image name"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}
	return Table{
		Title:   "Origin Container Image",
		Numeric: false,
		Rows:    rows,
		Rank:    containerImageRank,
	}, len(rows) > 0
}

func hostOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{host.HostName, "Host name"},
		{host.Load, "Load"},
		{host.OS, "Operating system"},
		{host.KernelVersion, "Kernel version"},
		{host.Uptime, "Uptime"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	return Table{
		Title:   "Origin Host",
		Numeric: false,
		Rows:    rows,
		Rank:    hostRank,
	}, len(rows) > 0
}
