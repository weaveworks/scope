package render

import (
	"fmt"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = Map{
	MapFunc:  MapEndpointIdentity,
	Renderer: SelectEndpoint,
}

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = MakeReduce(
	Map{
		MapFunc:  MapEndpoint2Process,
		Renderer: EndpointRenderer,
	},
	Map{
		MapFunc:  MapProcessIdentity,
		Renderer: SelectProcess,
	},
)

// processWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type processWithContainerNameRenderer struct {
	Renderer
}

func (r processWithContainerNameRenderer) Render(rpt report.Report) RenderableNodes {
	processes := r.Renderer.Render(rpt)
	containers := Map{
		MapFunc:  MapContainerIdentity,
		Renderer: SelectContainer,
	}.Render(rpt)

	for id, p := range processes {
		pid, ok := p.Node.Metadata[process.PID]
		if !ok {
			continue
		}
		containerID, ok := p.Node.Metadata[docker.ContainerID]
		if !ok {
			continue
		}
		container, ok := containers[containerID]
		if !ok {
			continue
		}
		p.LabelMinor = fmt.Sprintf("%s (%s:%s)", report.ExtractHostID(p.Node), container.LabelMajor, pid)
		processes[id] = p
	}

	return processes
}

// ProcessWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
var ProcessWithContainerNameRenderer = processWithContainerNameRenderer{ProcessRenderer}

// ProcessRenderer is a Renderer which produces a renderable process
// name graph by munging the progess graph.
var ProcessNameRenderer = Map{
	MapFunc: MapCountProcessName,
	Renderer: Map{
		MapFunc:  MapProcess2Name,
		Renderer: ProcessRenderer,
	},
}

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
var ContainerRenderer = MakeReduce(
	Map{
		MapFunc: MapProcess2Container,

		// We only want processes in container _or_ processes with network connections
		// but we need to be careful to ensure we only include each edge once, by only
		// including the ProcessRenderer once.
		Renderer: Filter{
			FilterFunc: func(n RenderableNode) bool {
				_, inContainer := n.Node.Metadata[docker.ContainerID]
				_, isConnected := n.Node.Metadata[IsConnected]
				return inContainer || isConnected
			},
			Renderer: ColorConnected(ProcessRenderer),
		},
	},

	Map{
		MapFunc:  MapContainerIdentity,
		Renderer: SelectContainer,
	},

	// This mapper brings in short lived connections by joining with container IPs.
	// We need to be careful to ensure we only include each edge once.  Edges brought in
	// by the above renders will have a pid, so its enough to filter out any nodes with
	// pids.
	Map{
		MapFunc: MapIP2Container,
		Renderer: FilterUnconnected(
			MakeReduce(
				Map{
					MapFunc:  MapContainer2IP,
					Renderer: SelectContainer,
				},
				Map{
					MapFunc:  MapEndpoint2IP,
					Renderer: SelectEndpoint,
				},
			),
		),
	},
)

type containerWithImageNameRenderer struct {
	Renderer
}

// Render produces a process graph where the minor labels contain the
// container name, if found.
func (r containerWithImageNameRenderer) Render(rpt report.Report) RenderableNodes {
	containers := r.Renderer.Render(rpt)
	images := Map{
		MapFunc:  MapContainerImageIdentity,
		Renderer: SelectContainerImage,
	}.Render(rpt)

	for id, c := range containers {
		imageID, ok := c.Node.Metadata[docker.ImageID]
		if !ok {
			continue
		}
		image, ok := images[imageID]
		if !ok {
			continue
		}
		c.Rank = imageNameWithoutVersion(image.LabelMajor)
		containers[id] = c
	}

	return containers
}

// ContainerWithImageNameRenderer is a Renderer which produces a container
// graph where the ranks are the image names, not their IDs
var ContainerWithImageNameRenderer = containerWithImageNameRenderer{ContainerRenderer}

// ContainerImageRenderer is a Renderer which produces a renderable container
// image graph by merging the container graph and the container image topology.
var ContainerImageRenderer = Map{
	MapFunc: MapCountContainers,
	Renderer: Map{
		MapFunc: MapContainerImage2Name,
		Renderer: MakeReduce(
			Map{
				MapFunc:  MapContainer2ContainerImage,
				Renderer: ContainerRenderer,
			},
			Map{
				MapFunc:  MapContainerImageIdentity,
				Renderer: SelectContainerImage,
			},
		),
	},
}

// AddressRenderer is a Renderer which produces a renderable address
// graph from the address topology.
var AddressRenderer = Map{
	MapFunc:  MapAddressIdentity,
	Renderer: SelectAddress,
}

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology and address graph.
var HostRenderer = MakeReduce(
	Map{
		MapFunc:  MapAddress2Host,
		Renderer: AddressRenderer,
	},
	Map{
		MapFunc:  MapHostIdentity,
		Renderer: SelectHost,
	},
)

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = Map{
	MapFunc: MapCountContainers,
	Renderer: MakeReduce(
		Map{
			MapFunc:  MapPodIdentity,
			Renderer: SelectPod,
		},
		Map{
			MapFunc:  MapContainer2Pod,
			Renderer: ContainerRenderer,
		},
	),
}

// PodsServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = Map{
	MapFunc: MapCountPods,
	Renderer: MakeReduce(
		Map{
			MapFunc:  MapPod2Service,
			Renderer: PodRenderer,
		},
		Map{
			MapFunc:  MapServiceIdentity,
			Renderer: SelectService,
		},
	),
}
