package docker

import (
	"net"
	"strings"

	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/report"
)

// Keys for use in Node
const (
	ImageID          = "docker_image_id"
	ImageName        = "docker_image_name"
	ImageLabelPrefix = "docker_image_label_"
)

// Exposed for testing
var (
	ContainerMetadataTemplates = report.MetadataTemplates{
		ContainerID:           {ID: ContainerID, Label: "ID", From: report.FromLatest, Truncate: 12, Priority: 1},
		ContainerStateHuman:   {ID: ContainerStateHuman, Label: "State", From: report.FromLatest, Priority: 2},
		ContainerCommand:      {ID: ContainerCommand, Label: "Command", From: report.FromLatest, Priority: 3},
		ImageID:               {ID: ImageID, Label: "Image ID", From: report.FromLatest, Truncate: 12, Priority: 11},
		ContainerUptime:       {ID: ContainerUptime, Label: "Uptime", From: report.FromLatest, Priority: 12},
		ContainerRestartCount: {ID: ContainerRestartCount, Label: "Restart #", From: report.FromLatest, Priority: 13},
		ContainerIPs:          {ID: ContainerIPs, Label: "IPs", From: report.FromSets, Priority: 14},
		ContainerPorts:        {ID: ContainerPorts, Label: "Ports", From: report.FromSets, Priority: 15},
		ContainerCreated:      {ID: ContainerCreated, Label: "Created", From: report.FromLatest, Priority: 16},
	}

	ContainerMetricTemplates = report.MetricTemplates{
		CPUTotalUsage: {ID: CPUTotalUsage, Label: "CPU", Format: report.PercentFormat, Priority: 1},
		MemoryUsage:   {ID: MemoryUsage, Label: "Memory", Format: report.FilesizeFormat, Priority: 2},
	}

	ContainerImageMetadataTemplates = report.MetadataTemplates{
		ImageID:          {ID: ImageID, Label: "Image ID", From: report.FromLatest, Truncate: 12, Priority: 1},
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: "number", Priority: 2},
	}

	ContainerTableTemplates = report.TableTemplates{
		LabelPrefix: {ID: LabelPrefix, Label: "Docker Labels", Prefix: LabelPrefix},
		EnvPrefix:   {ID: EnvPrefix, Label: "Environment Variables", Prefix: EnvPrefix},
	}

	ContainerImageTableTemplates = report.TableTemplates{
		ImageLabelPrefix: {ID: ImageLabelPrefix, Label: "Docker Labels", Prefix: ImageLabelPrefix},
	}

	ContainerControls = []report.Control{
		{
			ID:    AttachContainer,
			Human: "Attach",
			Icon:  "fa-desktop",
			Rank:  1,
		},
		{
			ID:    ExecContainer,
			Human: "Exec shell",
			Icon:  "fa-terminal",
			Rank:  2,
		},
		{
			ID:    StartContainer,
			Human: "Start",
			Icon:  "fa-play",
			Rank:  3,
		},
		{
			ID:    RestartContainer,
			Human: "Restart",
			Icon:  "fa-repeat",
			Rank:  4,
		},
		{
			ID:    PauseContainer,
			Human: "Pause",
			Icon:  "fa-pause",
			Rank:  5,
		},
		{
			ID:    UnpauseContainer,
			Human: "Unpause",
			Icon:  "fa-play",
			Rank:  6,
		},
		{
			ID:    StopContainer,
			Human: "Stop",
			Icon:  "fa-stop",
			Rank:  7,
		},
		{
			ID:    RemoveContainer,
			Human: "Remove",
			Icon:  "fa-trash-o",
			Rank:  8,
		},
		{
			ID:    AnalyzeTraffic,
			Human: "Analyze traffic",
			Icon:  "fa-plug",
			Rank:  9,
		},
	}
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	registry Registry
	hostID   string
	probeID  string
	probe    *probe.Probe
}

// NewReporter makes a new Reporter
func NewReporter(registry Registry, hostID string, probeID string, probe *probe.Probe) *Reporter {
	reporter := &Reporter{
		registry: registry,
		hostID:   hostID,
		probeID:  probeID,
		probe:    probe,
	}
	registry.WatchContainerUpdates(reporter.ContainerUpdated)
	return reporter
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Docker" }

// ContainerUpdated should be called whenever a container is updated.
func (r *Reporter) ContainerUpdated(n report.Node) {
	// Publish a 'short cut' report container just this container
	rpt := report.MakeReport()
	rpt.Shortcut = true
	rpt.Container.AddNode(n)
	r.probe.Publish(rpt)
}

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	localAddrs, err := report.LocalAddresses()
	if err != nil {
		return report.MakeReport(), nil
	}

	result := report.MakeReport()
	result.Container = result.Container.Merge(r.containerTopology(localAddrs))
	result.ContainerImage = result.ContainerImage.Merge(r.containerImageTopology())
	return result, nil
}

func getLocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	ips := []string{}
	for _, addr := range addrs {
		// Not all addrs are IPNets.
		if ipNet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipNet.IP.String())
		}
	}
	return ips, nil
}

func (r *Reporter) containerTopology(localAddrs []net.IP) report.Topology {
	result := report.MakeTopology().
		WithMetadataTemplates(ContainerMetadataTemplates).
		WithMetricTemplates(ContainerMetricTemplates).
		WithTableTemplates(ContainerTableTemplates)
	result.Controls.AddControls(ContainerControls)

	metadata := map[string]string{report.ControlProbeID: r.probeID}
	nodes := []report.Node{}
	r.registry.WalkContainers(func(c Container) {
		nodes = append(nodes, c.GetNode().WithLatests(metadata))
	})

	// Copy the IP addresses from other containers where they share network
	// namespaces & deal with containers in the host net namespace.  This
	// is recursive to deal with people who decide to be clever.
	{
		hostNetworkInfo := report.EmptySets
		if hostIPs, err := getLocalIPs(); err == nil {
			hostIPsWithScopes := addScopeToIPs(r.hostID, hostIPs)
			hostNetworkInfo = hostNetworkInfo.
				Add(ContainerIPs, report.MakeStringSet(hostIPs...)).
				Add(ContainerIPsWithScopes, report.MakeStringSet(hostIPsWithScopes...))
		}

		var networkInfo func(prefix string) report.Sets
		networkInfo = func(prefix string) report.Sets {
			container, ok := r.registry.GetContainerByPrefix(prefix)
			if !ok {
				return report.EmptySets
			}

			networkMode, ok := container.NetworkMode()
			if ok && strings.HasPrefix(networkMode, "container:") {
				return networkInfo(networkMode[10:])
			} else if ok && networkMode == NetworkModeHost {
				return hostNetworkInfo
			}

			return container.NetworkInfo(localAddrs)
		}

		for _, node := range nodes {
			id, ok := report.ParseContainerNodeID(node.ID)
			if !ok {
				continue
			}
			networkInfo := networkInfo(id)
			result.AddNode(node.WithSets(networkInfo))
		}
	}

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.MakeTopology().
		WithMetadataTemplates(ContainerImageMetadataTemplates).
		WithTableTemplates(ContainerImageTableTemplates)

	r.registry.WalkImages(func(image *docker_client.APIImages) {
		imageID := trimImageID(image.ID)
		nodeID := report.MakeContainerImageNodeID(imageID)
		node := report.MakeNodeWith(nodeID, map[string]string{
			ImageID: imageID,
		})
		node = node.AddTable(ImageLabelPrefix, image.Labels)

		if len(image.RepoTags) > 0 {
			node = node.WithLatests(map[string]string{ImageName: image.RepoTags[0]})
		}

		result.AddNode(node)
	})

	return result
}

// Docker sometimes prefixes ids with a "type" annotation, but it renders a bit
// ugly and isn't necessary, so we should strip it off
func trimImageID(id string) string {
	return strings.TrimPrefix(id, "sha256:")
}
