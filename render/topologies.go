package render

import (
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
		Pseudo:   PanicPseudoNode,
	},
)

// ContainerImageRenderer is a Renderer which produces a renderable container
// image graph by merging the container graph and the container image topology.
var ContainerImageRenderer = MakeReduce(
	Map{
		MapFunc:  MapContainer2ContainerImage,
		Renderer: ContainerRenderer,
	},
	LeafMap{
		Selector: report.SelectContainerImage,
		Mapper:   MapContainerImageIdentity,
		Pseudo:   PanicPseudoNode,
	},
)

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
