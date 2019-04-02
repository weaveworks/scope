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
	ImageID          = report.DockerImageID
	ImageName        = report.DockerImageName
	ImageTag         = report.DockerImageTag
	ImageSize        = report.DockerImageSize
	ImageVirtualSize = report.DockerImageVirtualSize
	IsInHostNetwork  = report.DockerIsInHostNetwork
	ImageLabelPrefix = "docker_image_label_"
	ImageTableID     = "image_table"
	ServiceName      = report.DockerServiceName
	StackNamespace   = report.DockerStackNamespace
	DefaultNamespace = "No stack"
)

// Exposed for testing
var (
	ContainerMetadataTemplates = report.MetadataTemplates{
		ImageTag:              {ID: ImageTag, Label: "Image tag", From: report.FromLatest, Priority: 1},
		ImageName:             {ID: ImageName, Label: "Image name", From: report.FromLatest, Priority: 2},
		ContainerCommand:      {ID: ContainerCommand, Label: "Command", From: report.FromLatest, Priority: 3},
		ContainerStateHuman:   {ID: ContainerStateHuman, Label: "State", From: report.FromLatest, Priority: 4},
		ContainerUptime:       {ID: ContainerUptime, Label: "Uptime", From: report.FromLatest, Priority: 5, Datatype: report.Duration},
		ContainerRestartCount: {ID: ContainerRestartCount, Label: "Restart #", From: report.FromLatest, Priority: 6},
		ContainerNetworks:     {ID: ContainerNetworks, Label: "Networks", From: report.FromSets, Priority: 7},
		ContainerIPs:          {ID: ContainerIPsWithScopes, Label: "IPs", From: report.FromSetsIPScope, Priority: 8},
		ContainerPorts:        {ID: ContainerPorts, Label: "Ports", From: report.FromSets, Priority: 9},
		ContainerCreated:      {ID: ContainerCreated, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 10},
		ContainerID:           {ID: ContainerID, Label: "ID", From: report.FromLatest, Truncate: 12, Priority: 11},
	}

	ContainerMetricTemplates = report.MetricTemplates{
		CPUTotalUsage: {ID: CPUTotalUsage, Label: "CPU", Format: report.PercentFormat, Priority: 1},
		MemoryUsage:   {ID: MemoryUsage, Label: "Memory", Format: report.FilesizeFormat, Priority: 2},
	}

	ContainerImageMetadataTemplates = report.MetadataTemplates{
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: report.Number, Priority: 2},
	}

	ContainerTableTemplates = report.TableTemplates{
		ImageTableID: {
			ID:    ImageTableID,
			Label: "Image",
			Type:  report.PropertyListType,
			FixedRows: map[string]string{
				// Prepend spaces as a hack to keep at the top when sorted.
				ImageID:          " ID",
				ImageName:        " Name",
				ImageTag:         " Tag",
				ImageSize:        "Size",
				ImageVirtualSize: "Virtual size",
			},
		},
		LabelPrefix: {
			ID:     LabelPrefix,
			Label:  "Docker labels",
			Type:   report.PropertyListType,
			Prefix: LabelPrefix,
		},
		EnvPrefix: {
			ID:     EnvPrefix,
			Label:  "Environment variables",
			Type:   report.PropertyListType,
			Prefix: EnvPrefix,
		},
	}

	ContainerImageTableTemplates = report.TableTemplates{
		ImageLabelPrefix: {
			ID:     ImageLabelPrefix,
			Label:  "Docker labels",
			Type:   report.PropertyListType,
			Prefix: ImageLabelPrefix,
		},
	}

	ContainerControls = []report.Control{
		{
			ID:    AttachContainer,
			Human: "Attach",
			Icon:  "fa fa-desktop",
			Rank:  1,
		},
		{
			ID:    ExecContainer,
			Human: "Exec shell",
			Icon:  "fa fa-terminal",
			Rank:  2,
		},
		{
			ID:    StartContainer,
			Human: "Start",
			Icon:  "fa fa-play",
			Rank:  3,
		},
		{
			ID:    RestartContainer,
			Human: "Restart",
			Icon:  "fa fa-redo",
			Rank:  4,
		},
		{
			ID:    PauseContainer,
			Human: "Pause",
			Icon:  "fa fa-pause",
			Rank:  5,
		},
		{
			ID:    UnpauseContainer,
			Human: "Unpause",
			Icon:  "fa fa-play",
			Rank:  6,
		},
		{
			ID:    StopContainer,
			Human: "Stop",
			Icon:  "fa fa-stop",
			Rank:  7,
		},
		{
			ID:    RemoveContainer,
			Human: "Remove",
			Icon:  "far fa-trash-alt",
			Rank:  8,
		},
	}

	SwarmServiceMetadataTemplates = report.MetadataTemplates{
		ServiceName:    {ID: ServiceName, Label: "Service name", From: report.FromLatest, Priority: 0},
		StackNamespace: {ID: StackNamespace, Label: "Stack namespace", From: report.FromLatest, Priority: 1},
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
	result.SwarmService = result.SwarmService.Merge(r.swarmServiceTopology())
	return result, nil
}

func getLocalIPs() ([]string, error) {
	ipnets, err := report.GetLocalNetworks()
	if err != nil {
		return nil, err
	}
	ips := []string{}
	for _, ipnet := range ipnets {
		ips = append(ips, ipnet.IP.String())
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
		hostNetworkInfo := report.MakeSets()
		if hostIPs, err := getLocalIPs(); err == nil {
			hostIPsWithScopes := addScopeToIPs(r.hostID, hostIPs)
			hostNetworkInfo = hostNetworkInfo.
				Add(ContainerIPsWithScopes, report.MakeStringSet(hostIPsWithScopes...))
		}

		var networkInfo func(prefix string) (report.Sets, bool)
		networkInfo = func(prefix string) (ips report.Sets, isInHostNamespace bool) {
			container, ok := r.registry.GetContainerByPrefix(prefix)
			if !ok {
				return report.MakeSets(), false
			}

			networkMode, ok := container.NetworkMode()
			if ok && strings.HasPrefix(networkMode, "container:") {
				return networkInfo(networkMode[10:])
			} else if ok && networkMode == "host" {
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
		WithTableTemplates(ContainerImageTableTemplates)

	r.registry.WalkImages(func(image docker_client.APIImages) {
		imageID := trimImageID(image.ID)
		latests := map[string]string{
			ImageID:          imageID,
			ImageSize:        humanize.Bytes(uint64(image.Size)),
			ImageVirtualSize: humanize.Bytes(uint64(image.VirtualSize)),
		}
		if len(image.RepoTags) > 0 {
			imageFullName := image.RepoTags[0]
			latests[ImageName] = ImageNameWithoutTag(imageFullName)
			latests[ImageTag] = ImageNameTag(imageFullName)
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
	t := report.MakeTopology()
	t.AddNode(node)
	return t
}

func (r *Reporter) swarmServiceTopology() report.Topology {
	return report.MakeTopology().WithMetadataTemplates(SwarmServiceMetadataTemplates)
}

// Docker sometimes prefixes ids with a "type" annotation, but it renders a bit
// ugly and isn't necessary, so we should strip it off
func trimImageID(id string) string {
	return strings.TrimPrefix(id, "sha256:")
}
