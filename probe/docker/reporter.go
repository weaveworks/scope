package docker

import (
	"net"
	"strings"

	humanize "github.com/dustin/go-humanize"
	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

// Keys for use in Node
const (
	ImageID          = "docker_image_id"
	ImageName        = "docker_image_name"
	ImageSize        = "docker_image_size"
	ImageVirtualSize = "docker_image_virtual_size"
	ImageLabelPrefix = "docker_image_label_"
	IsInHostNetwork  = "docker_is_in_host_network"
	ImageTableID     = "image_table"
)

// Exposed for testing
var (
	ContainerMetadataTemplates = report.MetadataTemplates{
		ImageName:             {ID: ImageName, Label: "Image", From: report.FromLatest, Priority: 1},
		ContainerCommand:      {ID: ContainerCommand, Label: "Command", From: report.FromLatest, Priority: 2},
		ContainerStateHuman:   {ID: ContainerStateHuman, Label: "State", From: report.FromLatest, Priority: 3},
		ContainerUptime:       {ID: ContainerUptime, Label: "Uptime", From: report.FromLatest, Priority: 4},
		ContainerRestartCount: {ID: ContainerRestartCount, Label: "Restart #", From: report.FromLatest, Priority: 5},
		ContainerNetworks:     {ID: ContainerNetworks, Label: "Networks", From: report.FromSets, Priority: 6},
		ContainerIPs:          {ID: ContainerIPs, Label: "IPs", From: report.FromSets, Priority: 7},
		ContainerPorts:        {ID: ContainerPorts, Label: "Ports", From: report.FromSets, Priority: 8},
		ContainerCreated:      {ID: ContainerCreated, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 9},
		ContainerID:           {ID: ContainerID, Label: "ID", From: report.FromLatest, Truncate: 12, Priority: 10},
	}

	ContainerMetricTemplates = report.MetricTemplates{
		CPUTotalUsage: {ID: CPUTotalUsage, Label: "CPU", Format: report.PercentFormat, Priority: 1},
		MemoryUsage:   {ID: MemoryUsage, Label: "Memory", Format: report.FilesizeFormat, Priority: 2},
	}

	ContainerImageMetadataTemplates = report.MetadataTemplates{
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: "number", Priority: 2},
	}

	ContainerPropertyListTemplates = report.PropertyListTemplates{
		ImageTableID: {ID: ImageTableID, Label: "Image",
			FixedProperties: map[string]string{
				ImageID:          "ID",
				ImageName:        "Name",
				ImageSize:        "Size",
				ImageVirtualSize: "Virtual Size",
			},
		},
		LabelPrefix: {ID: LabelPrefix, Label: "Docker Labels", Prefix: LabelPrefix},
		EnvPrefix:   {ID: EnvPrefix, Label: "Environment Variables", Prefix: EnvPrefix},
	}

	ContainerImagePropertyListTemplates = report.PropertyListTemplates{
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
	result.Overlay = result.Overlay.Merge(r.overlayTopology())
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
		WithPropertyListTemplates(ContainerPropertyListTemplates)
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

		var networkInfo func(prefix string) (report.Sets, bool)
		networkInfo = func(prefix string) (ips report.Sets, isInHostNamespace bool) {
			container, ok := r.registry.GetContainerByPrefix(prefix)
			if !ok {
				return report.EmptySets, false
			}

			networkMode, ok := container.NetworkMode()
			if ok && strings.HasPrefix(networkMode, "container:") {
				return networkInfo(networkMode[10:])
			} else if ok && networkMode == NetworkModeHost {
				return hostNetworkInfo, true
			}

			return container.NetworkInfo(localAddrs), false
		}

		for _, node := range nodes {
			id, ok := report.ParseContainerNodeID(node.ID)
			if !ok {
				continue
			}
			networkInfo, isInHostNamespace := networkInfo(id)
			node = node.WithSets(networkInfo)
			// Indicate whether the container is in the host network
			// The container's NetworkMode is not enough due to
			// delegation (e.g. NetworkMode="container:foo" where
			// foo is a container in the host networking namespace)
			if isInHostNamespace {
				node = node.WithLatests(map[string]string{IsInHostNetwork: "true"})
			}
			result.AddNode(node)

		}
	}

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.MakeTopology().
		WithMetadataTemplates(ContainerImageMetadataTemplates).
		WithPropertyListTemplates(ContainerImagePropertyListTemplates)

	r.registry.WalkImages(func(image docker_client.APIImages) {
		imageID := trimImageID(image.ID)
		latests := map[string]string{
			ImageID:          imageID,
			ImageSize:        humanize.Bytes(uint64(image.Size)),
			ImageVirtualSize: humanize.Bytes(uint64(image.VirtualSize)),
		}
		if len(image.RepoTags) > 0 {
			latests[ImageName] = image.RepoTags[0]
		}
		nodeID := report.MakeContainerImageNodeID(imageID)
		node := report.MakeNodeWith(nodeID, latests)
		node = node.AddPrefixPropertyList(ImageLabelPrefix, image.Labels)
		result.AddNode(node)
	})

	return result
}

func (r *Reporter) overlayTopology() report.Topology {
	subnets := []string{}
	r.registry.WalkNetworks(func(network docker_client.Network) {
		for _, config := range network.IPAM.Config {
			subnets = append(subnets, config.Subnet)
		}

	})
	// Add both local and global networks to the LocalNetworks Set
	// since we treat container IPs as local
	node := report.MakeNode(report.MakeOverlayNodeID(report.DockerOverlayPeerPrefix, r.hostID)).WithSets(
		report.MakeSets().Add(host.LocalNetworks, report.MakeStringSet(subnets...)))
	return report.MakeTopology().AddNode(node)
}

// Docker sometimes prefixes ids with a "type" annotation, but it renders a bit
// ugly and isn't necessary, so we should strip it off
func trimImageID(id string) string {
	return strings.TrimPrefix(id, "sha256:")
}
