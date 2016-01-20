package detailed

import (
	"encoding/json"
	"sort"
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
		meta(process.PID),
		meta(process.PPID),
		meta(process.Cmdline),
		meta(process.Threads),
	)
	containerNodeMetadata = renderMetadata(
		meta(docker.ContainerID),
		meta(docker.ImageID),
		ltst(docker.ContainerState),
		sets(docker.ContainerIPs),
		sets(docker.ContainerPorts),
		meta(docker.ContainerCreated),
		meta(docker.ContainerCommand),
		meta(overlay.WeaveMACAddress),
		meta(overlay.WeaveDNSHostname),
		getDockerLabelRows,
	)
	containerImageNodeMetadata = renderMetadata(
		meta(docker.ImageID),
		getDockerLabelRows,
	)
	podNodeMetadata = renderMetadata(
		meta(kubernetes.PodID),
		meta(kubernetes.Namespace),
		meta(kubernetes.PodCreated),
	)
	hostNodeMetadata = renderMetadata(
		meta(host.HostName),
		meta(host.OS),
		meta(host.KernelVersion),
		meta(host.Uptime),
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
		"group":               groupNodeMetadata,
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

func meta(id string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Metadata[id]; ok {
			return []MetadataRow{{ID: id, Value: val}}
		}
		return nil
	}
}

func sets(id string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Sets[id]; ok && len(val) > 0 {
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

func getDockerLabelRows(nmd report.Node) []MetadataRow {
	rows := []MetadataRow{}
	// Add labels in alphabetical order
	labels := docker.ExtractLabels(nmd)
	labelKeys := make([]string, 0, len(labels))
	for k := range labels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	for _, labelKey := range labelKeys {
		rows = append(rows, MetadataRow{ID: "label_" + labelKey, Value: labels[labelKey]})
	}
	return rows
}

func groupNodeMetadata(n report.Node) []MetadataRow {
	rows := []MetadataRow{}

	for _, template := range []struct {
		keys   []string
		lookup func(string, report.Node) (string, bool)
	}{
		{
			n.Metadata.Keys(),
			func(key string, n report.Node) (string, bool) {
				val, ok := n.Metadata[key]
				return val, ok
			},
		},
		{
			n.Latest.Keys(),
			func(key string, n report.Node) (string, bool) {
				return n.Latest.Lookup(key)
			},
		},
		{
			n.Sets.Keys(),
			func(key string, n report.Node) (string, bool) {
				if val, ok := n.Sets[key]; ok && len(val) > 0 {
					return strings.Join(val, ", "), true
				}
				return "", false
			},
		},
	} {
		sort.Strings(template.keys)
		for _, key := range template.keys {
			if strings.HasPrefix(key, "_") {
				continue
			}
			val, ok := template.lookup(key, n)
			if !ok {
				continue
			}

			rows = append(rows, MetadataRow{
				ID:    key,
				Value: val,
			})
		}
	}
	return rows
}
