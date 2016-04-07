package render

import (
	"net"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = FilterNonProcspied(SelectEndpoint)

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = MakeReduce(
	MakeMap(
		MapEndpoint2Process,
		EndpointRenderer,
	),
	SelectProcess,
)

// processWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type processWithContainerNameRenderer struct {
	Renderer
}

func (r processWithContainerNameRenderer) Render(rpt report.Report) report.Nodes {
	processes := r.Renderer.Render(rpt)
	containers := SelectContainer.Render(rpt)

	outputs := report.Nodes{}
	for id, p := range processes {
		outputs[id] = p
		containerID, timestamp, ok := p.Latest.LookupEntry(docker.ContainerID)
		if !ok {
			continue
		}
		container, ok := containers[report.MakeContainerNodeID(containerID)]
		if !ok {
			continue
		}
		output := p.Copy()
		output.Latest = output.Latest.Set(docker.ContainerID, timestamp, containerID)
		if containerName, timestamp, ok := container.Latest.LookupEntry(docker.ContainerName); ok {
			output.Latest = output.Latest.Set(docker.ContainerName, timestamp, containerName)
		}
		outputs[id] = output
	}
	return outputs
}

// ProcessWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
var ProcessWithContainerNameRenderer = processWithContainerNameRenderer{ProcessRenderer}

// ProcessNameRenderer is a Renderer which produces a renderable process
// name graph by munging the progess graph.
var ProcessNameRenderer = MakeMap(
	MapProcess2Name,
	ProcessRenderer,
)

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
// NB We only want processes in container _or_ processes with network connections
// but we need to be careful to ensure we only include each edge once, by only
// including the ProcessRenderer once.
var ContainerRenderer = MakeReduce(
	MakeSilentFilter(
		func(n report.Node) bool {
			_, isConnected := n.Latest.Lookup(IsConnected)
			return n.Topology != Pseudo || isConnected
		},
		MakeMap(
			MapProcess2Container,
			ColorConnected(ProcessRenderer),
		),
	),

	// This mapper brings in short lived connections by joining with container IPs.
	// We need to be careful to ensure we only include each edge once.  Edges brought in
	// by the above renders will have a pid, so its enough to filter out any nodes with
	// pids.
	SilentFilterUnconnected(MakeMap(
		MapIP2Container,
		MakeReduce(
			MakeMap(
				MapContainer2IP,
				SelectContainer,
			),
			MakeMap(
				MapEndpoint2IP,
				SelectEndpoint,
			),
		),
	)),

	SelectContainer,
)

type containerWithHostIPsRenderer struct {
	Renderer
}

// Render produces a process graph where the ips for host network mode are set
// to the host's IPs.
func (r containerWithHostIPsRenderer) Render(rpt report.Report) report.Nodes {
	containers := r.Renderer.Render(rpt)
	hosts := SelectHost.Render(rpt)

	outputs := report.Nodes{}
	for id, c := range containers {
		outputs[id] = c
		networkMode, ok := c.Latest.Lookup(docker.ContainerNetworkMode)
		if !ok || networkMode != docker.NetworkModeHost {
			continue
		}

		h, ok := hosts[report.MakeHostNodeID(report.ExtractHostID(c))]
		if !ok {
			continue
		}

		newIPs := report.MakeStringSet()
		hostNetworks, _ := h.Sets.Lookup(host.LocalNetworks)
		for _, cidr := range hostNetworks {
			if ip, _, err := net.ParseCIDR(cidr); err == nil {
				newIPs = newIPs.Add(ip.String())
			}
		}

		output := c.Copy()
		output.Sets = c.Sets.Add(docker.ContainerIPs, newIPs)
		outputs[id] = output
	}
	return outputs
}

// ContainerWithHostIPsRenderer is a Renderer which produces a container graph
// enriched with host IPs on containers where NetworkMode is Host
var ContainerWithHostIPsRenderer = containerWithHostIPsRenderer{ContainerRenderer}

type containerWithImageNameRenderer struct {
	Renderer
}

// Render produces a process graph where the minor labels contain the
// container name, if found.  It also merges the image node metadata into the
// container metadata.
func (r containerWithImageNameRenderer) Render(rpt report.Report) report.Nodes {
	containers := r.Renderer.Render(rpt)
	images := SelectContainerImage.Render(rpt)

	outputs := report.Nodes{}
	for id, c := range containers {
		outputs[id] = c
		imageID, ok := c.Latest.Lookup(docker.ImageID)
		if !ok {
			continue
		}
		image, ok := images[report.MakeContainerImageNodeID(imageID)]
		if !ok {
			continue
		}
		output := c.Copy()
		output.Latest = image.Latest.Merge(c.Latest)
		outputs[id] = output
	}
	return outputs
}

// ContainerWithImageNameRenderer is a Renderer which produces a container
// graph where the ranks are the image names, not their IDs
var ContainerWithImageNameRenderer = containerWithImageNameRenderer{ContainerWithHostIPsRenderer}

// ContainerImageRenderer is a Renderer which produces a renderable container
// image graph by merging the container graph and the container image topology.
var ContainerImageRenderer = MakeReduce(
	MakeMap(
		MapContainer2ContainerImage,
		ContainerRenderer,
	),
	SelectContainerImage,
)

// ContainerHostnameRenderer is a Renderer which produces a renderable container
// by hostname graph..
var ContainerHostnameRenderer = MakeMap(
	MapContainer2Hostname,
	ContainerRenderer,
)

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology.
var HostRenderer = MakeReduce(
	MakeMap(
		MapEndpoint2Host,
		EndpointRenderer,
	),
	MakeMap(
		MapX2Host,
		ColorConnected(ProcessRenderer),
	),
	MakeMap(
		MapX2Host,
		ContainerRenderer,
	),
	MakeMap(
		MapX2Host,
		ContainerImageRenderer,
	),
	// Pods don't have a host id - #1142
	// MakeMap(
	// 	MapX2Host,
	// 		SelectPod,
	// ),
	SelectHost,
)

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = MakeReduce(
	MakeSilentFilter(
		func(n report.Node) bool {
			_, isConnected := n.Latest.Lookup(IsConnected)
			return n.Topology != Pseudo || isConnected
		},
		ColorConnected(MakeMap(
			MapContainer2Pod,
			ContainerRenderer,
		)),
	),
	SelectPod,
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = MakeReduce(
	MakeMap(
		MapPod2Service,
		PodRenderer,
	),
	SelectService,
)
