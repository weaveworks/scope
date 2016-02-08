package detailed

import (
	"encoding/json"
	"strconv"
	"strings"

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
		Latest{ID: process.PID},
		Latest{ID: process.PPID},
		Latest{ID: process.Cmdline},
		Latest{ID: process.Threads},
	}
	containerNodeMetadata = []MetadataRowTemplate{
		Latest{ID: docker.ContainerID},
		Latest{ID: docker.ImageID},
		Latest{ID: docker.ContainerState},
		Latest{ID: docker.ContainerUptime},
		Latest{ID: docker.ContainerRestartCount},
		Set{ID: docker.ContainerIPs},
		Set{ID: docker.ContainerPorts},
		Latest{ID: docker.ContainerCreated},
		Latest{ID: docker.ContainerCommand},
		Latest{ID: overlay.WeaveMACAddress},
		Latest{ID: overlay.WeaveDNSHostname},
	}
	containerImageNodeMetadata = []MetadataRowTemplate{
		Latest{ID: docker.ImageID},
		Counter{ID: render.ContainersKey},
	}
	podNodeMetadata = []MetadataRowTemplate{
		Latest{ID: kubernetes.PodID},
		Latest{ID: kubernetes.Namespace},
		Latest{ID: kubernetes.PodCreated},
	}
	hostNodeMetadata = []MetadataRowTemplate{
		Latest{ID: host.HostName},
		Latest{ID: host.OS},
		Latest{ID: host.KernelVersion},
		Latest{ID: host.Uptime},
		Set{ID: host.LocalNetworks},
	}
)

// MetadataRowTemplate extracts some metadata rows from a node
type MetadataRowTemplate interface {
	MetadataRows(report.Node) []MetadataRow
}

// Latest extracts some metadata rows from a node's Latest
type Latest struct {
	ID string
}

// MetadataRows implements MetadataRowTemplate
func (l Latest) MetadataRows(n report.Node) []MetadataRow {
	if val, ok := n.Latest.Lookup(l.ID); ok {
		return []MetadataRow{{ID: l.ID, Value: val}}
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
	ID string
}

// MetadataRows implements MetadataRowTemplate
func (c Counter) MetadataRows(n report.Node) []MetadataRow {
	if val, ok := n.Counters.Lookup(c.ID); ok {
		return []MetadataRow{{ID: c.ID, Value: strconv.Itoa(val)}}
	}
	return nil
}

// MetadataRow is a row for the metadata table.
type MetadataRow struct {
	ID    string
	Value string
}

// Copy returns a value copy of a metadata row.
func (m MetadataRow) Copy() MetadataRow {
	return MetadataRow{
		ID:    m.ID,
		Value: m.Value,
	}
}

// MarshalJSON marshals this MetadataRow to json. It adds a label before
// rendering.
func (m MetadataRow) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID    string `json:"id"`
		Label string `json:"label"`
		Value string `json:"value"`
	}{
		ID:    m.ID,
		Label: Label(m.ID),
		Value: m.Value,
	})
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
