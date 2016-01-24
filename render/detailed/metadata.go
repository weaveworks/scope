package detailed

import (
	"fmt"
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
		ltst(process.PID, "PID"),
		ltst(process.PPID, "Parent PID"),
		ltst(process.Cmdline, "Command"),
		ltst(process.Threads, "# Threads"),
	)
	containerNodeMetadata = renderMetadata(
		ltst(docker.ContainerID, "ID"),
		ltst(docker.ImageID, "Image ID"),
		ltst(docker.ContainerState, "State"),
		ltst(docker.ContainerUptime, "Uptime"),
		ltst(docker.ContainerRestartCount, "Restart #"),
		sets(docker.ContainerIPs, "IPs"),
		sets(docker.ContainerPorts, "Ports"),
		ltst(docker.ContainerCreated, "Created"),
		ltst(docker.ContainerCommand, "Command"),
		ltst(overlay.WeaveMACAddress, "Weave MAC"),
		ltst(overlay.WeaveDNSHostname, "Weave DNS Hostname"),
		getDockerLabelRows,
	)
	containerImageNodeMetadata = renderMetadata(
		ltst(docker.ImageID, "Image ID"),
		getDockerLabelRows,
	)
	podNodeMetadata = renderMetadata(
		ltst(kubernetes.PodID, "ID"),
		ltst(kubernetes.Namespace, "Namespace"),
		ltst(kubernetes.PodCreated, "Created"),
	)
	hostNodeMetadata = renderMetadata(
		ltst(host.HostName, "Hostname"),
		ltst(host.OS, "Operating system"),
		ltst(host.KernelVersion, "Kernel version"),
		ltst(host.Uptime, "Uptime"),
		sets(host.LocalNetworks, "Local Networks"),
	)
)

// MetadataRow is a row for the metadata table.
type MetadataRow struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// Copy returns a value copy of a metadata row.
func (m MetadataRow) Copy() MetadataRow {
	return MetadataRow{
		ID:    m.ID,
		Label: m.Label,
		Value: m.Value,
	}
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

func sets(id, label string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Sets.Lookup(id); ok && len(val) > 0 {
			return []MetadataRow{{ID: id, Label: label, Value: strings.Join(val, ", ")}}
		}
		return nil
	}
}

func ltst(id, label string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Latest.Lookup(id); ok {
			return []MetadataRow{{ID: id, Label: label, Value: val}}
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
		rows = append(rows, MetadataRow{ID: "label_" + labelKey, Label: fmt.Sprintf("Label %q", labelKey), Value: labels[labelKey]})
	}
	return rows
}
