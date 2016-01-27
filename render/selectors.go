package render

import (
	"github.com/weaveworks/scope/report"
)

// TopologySelector selects a single topology from a report.
// NB it is also a Renderer!
type TopologySelector func(r report.Report) RenderableNodes

// Render implements Renderer
func (t TopologySelector) Render(r report.Report) RenderableNodes {
	return t(r)
}

// Stats implements Renderer
func (t TopologySelector) Stats(r report.Report) Stats {
	return Stats{}
}

// Topology2RenderableNodes converts a topology to a set of RenderableNodes
func Topology2RenderableNodes(t report.Topology) RenderableNodes {
	result := EmptyRenderableNodes
	for id, nmd := range t.Nodes {
		result = result.Add(NewRenderableNode(id).WithNode(nmd))
	}

	// Push EdgeMetadata to both ends of the edges
	// We cannot (and should not) use ForEach here, as this loop depends on
	// adding and modifying entries of the set while iterating over it.
	// TODO: refactor this.
	keys := result.Keys()
	for _, srcNodeID := range keys {
		srcNode, _ := result.Lookup(srcNodeID)
		srcNode.Edges.ForEach(func(dstID string, emd report.EdgeMetadata) {
			srcNode.EdgeMetadata = srcNode.EdgeMetadata.Flatten(emd)

			dstNode, ok := result.Lookup(dstID)
			if !ok {
				dstNode = NewRenderableNode(dstID)
			}
			dstNode.EdgeMetadata = dstNode.EdgeMetadata.Flatten(emd.Reversed())
			result = result.Add(dstNode)
		})
		result = result.Add(srcNode)
	}
	return result
}

var (
	// SelectEndpoint selects the endpoint topology.
	SelectEndpoint = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Endpoint)
	})

	// SelectProcess selects the process topology.
	SelectProcess = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Process)
	})

	// SelectContainer selects the container topology.
	SelectContainer = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Container)
	})

	// SelectContainerImage selects the container image topology.
	SelectContainerImage = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.ContainerImage)
	})

	// SelectAddress selects the address topology.
	SelectAddress = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Address)
	})

	// SelectHost selects the address topology.
	SelectHost = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Host)
	})

	// SelectPod selects the pod topology.
	SelectPod = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Pod)
	})

	// SelectService selects the service topology.
	SelectService = TopologySelector(func(r report.Report) RenderableNodes {
		return Topology2RenderableNodes(r.Service)
	})
)
