package render

import (
	"github.com/weaveworks/scope/report"
)

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = LeafMap{
	Selector: report.SelectEndpoint,
	Mapper:   MapEndpointIdentity,
	Pseudo:   GenericPseudoNode,
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
		Pseudo:   GenericPseudoNode,
	},
)

// ProcessRenderer is a Renderer which produces a renderable process
// name graph by munging the progess graph.
var ProcessNameRenderer = Map{
	MapFunc:  MapProcess2Name,
	Renderer: ProcessRenderer,
}

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
var ContainerRenderer = MakeReduce(
	Map{
		MapFunc:  MapProcess2Container,
		Renderer: ProcessRenderer,
	},
	LeafMap{
		Selector: report.SelectContainer,
		Mapper:   MapContainerIdentity,
		Pseudo:   GenericPseudoNode,
	},
)
