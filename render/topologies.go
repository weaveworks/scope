package render

import (
	"fmt"
	"net"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = MakeMap(
	MapEndpointIdentity,
	SelectEndpoint,
)

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = MakeReduce(
	MakeMap(
		MapEndpoint2Process,
		EndpointRenderer,
	),
	MakeMap(
		MapProcessIdentity,
		SelectProcess,
	),
)

// processWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type processWithContainerNameRenderer struct {
	Renderer
}

func (r processWithContainerNameRenderer) Render(rpt report.Report) RenderableNodes {
	processes := r.Renderer.Render(rpt)
	containers := MakeMap(
		MapContainerIdentity,
		SelectContainer,
	).Render(rpt)

	outputs := RenderableNodes{}
	for id, p := range processes {
		outputs[id] = p
		pid, ok := p.Node.Latest.Lookup(process.PID)
		if !ok {
			continue
		}
		containerID, ok := p.Node.Latest.Lookup(docker.ContainerID)
		if !ok {
			continue
		}
		container, ok := containers[MakeContainerID(containerID)]
		if !ok {
			continue
		}
		output := p.Copy()
		output.LabelMinor = fmt.Sprintf("%s (%s:%s)", report.ExtractHostID(p.Node), container.LabelMajor, pid)
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
	MapCountProcessName,
	MakeMap(
		MapProcess2Name,
		ProcessRenderer,
	),
)

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
// NB We only want processes in container _or_ processes with network connections
// but we need to be careful to ensure we only include each edge once, by only
// including the ProcessRenderer once.
var ContainerRenderer = MakeReduce(
	MakeFilter(
		func(n RenderableNode) bool {
			_, inContainer := n.Node.Latest.Lookup(docker.ContainerID)
			_, isConnected := n.Node.Latest.Lookup(IsConnected)
			return inContainer || isConnected
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
	FilterUnconnected(MakeMap(
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

	MakeMap(
		MapContainerIdentity,
		SelectContainer,
	),
)

type containerWithHostIPsRenderer struct {
	Renderer
}

// Render produces a process graph where the ips for host network mode are set
// to the host's IPs.
func (r containerWithHostIPsRenderer) Render(rpt report.Report) RenderableNodes {
	containers := r.Renderer.Render(rpt)
	hosts := MakeMap(
		MapHostIdentity,
		SelectHost,
	).Render(rpt)

	outputs := RenderableNodes{}
	for id, c := range containers {
		outputs[id] = c
		networkMode, ok := c.Node.Latest.Lookup(docker.ContainerNetworkMode)
		if !ok || networkMode != docker.NetworkModeHost {
			continue
		}

		h, ok := hosts[MakeHostID(report.ExtractHostID(c.Node))]
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
func (r containerWithImageNameRenderer) Render(rpt report.Report) RenderableNodes {
	containers := r.Renderer.Render(rpt)
	images := MakeMap(
		MapContainerImageIdentity,
		SelectContainerImage,
	).Render(rpt)

	outputs := RenderableNodes{}
	for id, c := range containers {
		outputs[id] = c
		imageID, ok := c.Node.Latest.Lookup(docker.ImageID)
		if !ok {
			continue
		}
		image, ok := images[MakeContainerImageID(imageID)]
		if !ok {
			continue
		}
		output := c.Copy()
		output.Rank = ImageNameWithoutVersion(image.LabelMajor)
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
var ContainerImageRenderer = MakeMap(
	MapCountContainers,
	MakeMap(
		MapContainerImage2Name,
		MakeReduce(
			MakeMap(
				MapContainer2ContainerImage,
				ContainerRenderer,
			),
			MakeMap(
				MapContainerImageIdentity,
				SelectContainerImage,
			),
		),
	),
)

// ContainerHostnameRenderer is a Renderer which produces a renderable container
// by hostname graph..
var ContainerHostnameRenderer = MakeMap(
	MapCountContainers,
	MakeMap(
		MapContainer2Hostname,
		ContainerRenderer,
	),
)

// AddressRenderer is a Renderer which produces a renderable address
// graph from the address topology.
var AddressRenderer = MakeMap(
	MapAddressIdentity,
	SelectAddress,
)

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology and address graph.
var HostRenderer = MakeReduce(
	MakeMap(
		MapX2Host,
		FilterPseudo(ContainerImageRenderer),
	),
	MakeMap(
		MapX2Host,
		MakeMap(
			MapPodIdentity,
			SelectPod,
		),
	),
	MakeMap(
		MapX2Host,
		AddressRenderer,
	),
	MakeMap(
		MapHostIdentity,
		SelectHost,
	),
)

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = MakeMap(
	MapCountContainers,
	MakeReduce(
		MakeMap(
			MapContainer2Pod,
			ContainerRenderer,
		),
		MakeMap(
			MapPodIdentity,
			SelectPod,
		),
	),
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = MakeMap(
	MapCountPods,
	MakeReduce(
		MakeMap(
			MapPod2Service,
			PodRenderer,
		),
		MakeMap(
			MapServiceIdentity,
			SelectService,
		),
	),
)
