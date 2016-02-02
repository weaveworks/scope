package detailed

import (
	"encoding/json"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

var (
	processNodeMetadata = renderMetadata(
		ltst(process.PID),
		ltst(process.PPID),
		ltst(process.Cmdline),
		ltst(process.Threads),
	)
	containerNodeMetadata = renderMetadata(
		ltst(docker.ContainerID),
		ltst(docker.ImageID),
		ltst(docker.ContainerState),
		ltst(docker.ContainerUptime),
		ltst(docker.ContainerRestartCount),
		sets(docker.ContainerIPs),
		sets(docker.ContainerPorts),
		ltst(docker.ContainerCreated),
		ltst(docker.ContainerCommand),
		ltst(overlay.WeaveMACAddress),
		ltst(overlay.WeaveDNSHostname),
	)
	containerImageNodeMetadata = renderMetadata(
		ltst(docker.ImageID),
	)
	podNodeMetadata = renderMetadata(
		ltst(kubernetes.PodID),
		ltst(kubernetes.Namespace),
		ltst(kubernetes.PodCreated),
	)
	hostNodeMetadata = renderMetadata(
		ltst(host.HostName),
		ltst(host.OS),
		ltst(host.KernelVersion),
		ltst(host.Uptime),
		sets(host.LocalNetworks),
	)
)

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
	renderers := map[string]func(report.Node) []MetadataRow{
		report.Process:        processNodeMetadata,
		report.Container:      containerNodeMetadata,
		report.ContainerImage: containerImageNodeMetadata,
		report.Pod:            podNodeMetadata,
		report.Host:           hostNodeMetadata,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		return renderer(n)
	}
	return nil
}

func renderMetadata(templates ...func(report.Node) []MetadataRow) func(report.Node) []MetadataRow {
	return func(nmd report.Node) []MetadataRow {
		rows := []MetadataRow{}
		for _, template := range templates {
			rows = append(rows, template(nmd)...)
		}
		return rows
	}
}

func sets(id string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Sets.Lookup(id); ok && len(val) > 0 {
			return []MetadataRow{{ID: id, Value: strings.Join(val, ", ")}}
		}
		return nil
	}
}

func ltst(id string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Latest.Lookup(id); ok {
			return []MetadataRow{{ID: id, Value: val}}
		}
		return nil
	}
}
