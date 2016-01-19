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
		meta(process.PID, "PID"),
		meta(process.PPID, "Parent PID"),
		meta(process.Cmdline, "Command"),
		meta(process.Threads, "# Threads"),
	)
	containerNodeMetadata = renderMetadata(
		meta(docker.ContainerID, "ID"),
		meta(docker.ImageID, "Image ID"),
		ltst(docker.ContainerState, "State"),
		sets(docker.ContainerIPs, "IPs"),
		sets(docker.ContainerPorts, "Ports"),
		meta(docker.ContainerCreated, "Created"),
		meta(docker.ContainerCommand, "Command"),
		meta(overlay.WeaveMACAddress, "Weave MAC"),
		meta(overlay.WeaveDNSHostname, "Weave DNS Hostname"),
		getDockerLabelRows,
	)
	containerImageNodeMetadata = renderMetadata(
		meta(docker.ImageID, "Image ID"),
		getDockerLabelRows,
	)
	podNodeMetadata = renderMetadata(
		meta(kubernetes.PodID, "ID"),
		meta(kubernetes.Namespace, "Namespace"),
		meta(kubernetes.PodCreated, "Created"),
	)
	hostNodeMetadata = renderMetadata(
		meta(host.HostName, "Hostname"),
		meta(host.OS, "Operating system"),
		meta(host.KernelVersion, "Kernel version"),
		meta(host.Uptime, "Uptime"),
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

func meta(id, label string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Metadata[id]; ok {
			return []MetadataRow{{ID: id, Label: label, Value: val}}
		}
		return nil
	}
}

func sets(id, label string) func(report.Node) []MetadataRow {
	return func(n report.Node) []MetadataRow {
		if val, ok := n.Sets[id]; ok && len(val) > 0 {
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
