package detailed

import (
	"strconv"
	"strings"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

var (
	processNodeMetadata = []MetadataRowTemplate{
		Latest{ID: process.PID, Prime: true},
		Latest{ID: process.Cmdline, Prime: true},
		Latest{ID: process.PPID, Prime: true},
		Latest{ID: process.Threads, Prime: true},
	}
	containerNodeMetadata = []MetadataRowTemplate{
		Latest{ID: docker.ContainerID, Truncate: 12, Prime: true},
		Latest{ID: docker.ContainerState, Prime: true},
		Latest{ID: docker.ContainerCommand, Prime: true},
		Latest{ID: docker.ImageID, Truncate: 12},
		Latest{ID: docker.ContainerUptime},
		Latest{ID: docker.ContainerRestartCount},
		Set{ID: docker.ContainerIPs},
		Set{ID: docker.ContainerPorts},
		Latest{ID: docker.ContainerCreated},
		Latest{ID: overlay.WeaveMACAddress},
		Latest{ID: overlay.WeaveDNSHostname},
	}
	containerImageNodeMetadata = []MetadataRowTemplate{
		Latest{ID: docker.ImageID, Truncate: 12, Prime: true},
		Counter{ID: render.ContainersKey, Prime: true},
	}
	podNodeMetadata = []MetadataRowTemplate{
		Latest{ID: kubernetes.PodID, Prime: true},
		Latest{ID: kubernetes.Namespace, Prime: true},
		Latest{ID: kubernetes.PodCreated, Prime: true},
	}
	hostNodeMetadata = []MetadataRowTemplate{
		Latest{ID: host.KernelVersion, Prime: true},
		Latest{ID: host.Uptime, Prime: true},
		Latest{ID: host.HostName},
		Latest{ID: host.OS},
		Set{ID: host.LocalNetworks},
	}
)

// MetadataRowTemplate extracts some metadata rows from a node
type MetadataRowTemplate interface {
	MetadataRows(report.Node) []MetadataRow
}

// Latest extracts some metadata rows from a node's Latest
type Latest struct {
	ID       string
	Truncate int  // If > 0, truncate the value to this length.
	Prime    bool // Whether the row should be shown by default
}

// MetadataRows implements MetadataRowTemplate
func (l Latest) MetadataRows(n report.Node) []MetadataRow {
	if val, ok := n.Latest.Lookup(l.ID); ok {
		if l.Truncate > 0 && len(val) > l.Truncate {
			val = val[:l.Truncate]
		}
		return []MetadataRow{{ID: l.ID, Value: val, Prime: l.Prime}}
	}
	return nil
}

// Set extracts some metadata rows from a node's Sets
type Set struct {
	ID string
}

// MetadataRows implements MetadataRowTemplate
func (s Set) MetadataRows(n report.Node) []MetadataRow {
	if val, ok := n.Sets.Lookup(s.ID); ok && len(val) > 0 {
		return []MetadataRow{{ID: s.ID, Value: strings.Join(val, ", ")}}
	}
	return nil
}

// Counter extracts some metadata rows from a node's Counters
type Counter struct {
	ID    string
	Prime bool
}

// MetadataRows implements MetadataRowTemplate
func (c Counter) MetadataRows(n report.Node) []MetadataRow {
	if val, ok := n.Counters.Lookup(c.ID); ok {
		return []MetadataRow{{
			ID:       c.ID,
			Value:    strconv.Itoa(val),
			Prime:    c.Prime,
			Datatype: number,
		}}
	}
	return nil
}

// MetadataRow is a row for the metadata table.
type MetadataRow struct {
	ID       string
	Value    string
	Prime    bool
	Datatype string
}

// Copy returns a value copy of a metadata row.
func (m MetadataRow) Copy() MetadataRow {
	return MetadataRow{
		ID:    m.ID,
		Value: m.Value,
	}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (MetadataRow) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*MetadataRow) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

type labelledMetadataRow struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Prime    bool   `json:"prime,omitempty"`
	Datatype string `json:"dataType,omitempty"`
}

// CodecEncodeSelf marshals this MetadataRow. It adds a label before
// rendering.
func (m *MetadataRow) CodecEncodeSelf(encoder *codec.Encoder) {
	in := labelledMetadataRow{
		ID:       m.ID,
		Label:    Label(m.ID),
		Value:    m.Value,
		Prime:    m.Prime,
		Datatype: m.Datatype,
	}
	encoder.Encode(in)
}

// CodecDecodeSelf implements codec.Selfer
func (m *MetadataRow) CodecDecodeSelf(decoder *codec.Decoder) {
	var in labelledMetadataRow
	decoder.Decode(&in)
	*m = MetadataRow{
		ID:       in.ID,
		Value:    in.Value,
		Prime:    in.Prime,
		Datatype: in.Datatype,
	}
}

// NodeMetadata produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetadata(n report.Node) []MetadataRow {
	renderers := map[string][]MetadataRowTemplate{
		report.Process:        processNodeMetadata,
		report.Container:      containerNodeMetadata,
		report.ContainerImage: containerImageNodeMetadata,
		report.Pod:            podNodeMetadata,
		report.Host:           hostNodeMetadata,
	}
	if templates, ok := renderers[n.Topology]; ok {
		rows := []MetadataRow{}
		for _, template := range templates {
			rows = append(rows, template.MetadataRows(n)...)
		}
		return rows
	}
	return nil
}
