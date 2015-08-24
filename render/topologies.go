package render

import (
	"fmt"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = LeafMap{
	Selector: report.SelectEndpoint,
	Mapper:   MapEndpointIdentity,
	Pseudo:   GenericPseudoNode(report.EndpointIDAddresser),
}

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = MakeReduce(
	Map{
		MapFunc:  MapEndpoint2Process,
		Renderer: EndpointRenderer,
	},
	LeafMap{
		Selector: report.SelectProcess,
		Mapper:   MapProcessIdentity,
		Pseudo:   PanicPseudoNode,
	},
)

// ProcessWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type ProcessWithContainerNameRenderer struct{}

// Render produces a process graph where the minor labels contain the
// container name, if found.
func (r ProcessWithContainerNameRenderer) Render(rpt report.Report) RenderableNodes {
	processes := ProcessRenderer.Render(rpt)
	containers := LeafMap{
		Selector: report.SelectContainer,
		Mapper:   MapContainerIdentity,
		Pseudo:   PanicPseudoNode,
	}.Render(rpt)

	for id, p := range processes {
		pid, ok := p.NodeMetadata.Metadata[process.PID]
		if !ok {
			continue
		}
		containerID, ok := p.NodeMetadata.Metadata[docker.ContainerID]
		if !ok {
			continue
		}
		container, ok := containers[containerID]
		if !ok {
			continue
		}
		p.LabelMinor = fmt.Sprintf("%s (%s:%s)", report.ExtractHostID(p.NodeMetadata), container.LabelMajor, pid)
		processes[id] = p
	}

	return processes
}

// EdgeMetadata produces an EdgeMetadata for a given edge.
func (r ProcessWithContainerNameRenderer) EdgeMetadata(rpt report.Report, localID, remoteID string) report.EdgeMetadata {
	return ProcessRenderer.EdgeMetadata(rpt, localID, remoteID)
}

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
			f: func(n RenderableNode) bool {
				_, inContainer := n.NodeMetadata.Metadata[docker.ContainerID]
				_, isConnected := n.NodeMetadata.Metadata[IsConnected]
				return inContainer || isConnected
			},
			Renderer: CustomRenderer{
				RenderFunc: ColorConnected,
				Renderer:   ProcessRenderer,
			},
		},
	},

	LeafMap{
		Selector: report.SelectContainer,
		Mapper:   MapContainerIdentity,
		Pseudo:   PanicPseudoNode,
	},

	// This mapper brings in short lived connections by joining with container IPs.
	// We need to be careful to ensure we only include each edge once.  Edges brought in
	// by the above renders will have a pid, so its enough to filter out any nodes with
	// pids.
	Map{
		MapFunc: MapIP2Container,
		Renderer: FilterUnconnected(
			MakeReduce(
				LeafMap{
					Selector: report.SelectContainer,
					Mapper:   MapContainer2IP,
					Pseudo:   PanicPseudoNode,
				},
				LeafMap{
					Selector: report.SelectEndpoint,
					Mapper:   MapEndpoint2IP,
					Pseudo:   IPPseudoNode,
				},
			),
		),
	},
)

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
			LeafMap{
				Selector: report.SelectContainerImage,
				Mapper:   MapContainerImageIdentity,
				Pseudo:   PanicPseudoNode,
			},
		),
	},
}

// AddressRenderer is a Renderer which produces a renderable address
// graph from the address topology.
var AddressRenderer = LeafMap{
	Selector: report.SelectAddress,
	Mapper:   MapAddressIdentity,
	Pseudo:   GenericPseudoNode(report.AddressIDAddresser),
}

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology and address graph.
var HostRenderer = MakeReduce(
	Map{
		MapFunc:  MapAddress2Host,
		Renderer: AddressRenderer,
	},
	LeafMap{
		Selector: report.SelectHost,
		Mapper:   MapHostIdentity,
		Pseudo:   PanicPseudoNode,
	},
)
