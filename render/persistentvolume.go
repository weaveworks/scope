package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// PersistentVolumeRenderer is the common renderer for all the storage components.
var PersistentVolumeRenderer = Memoise(
	MakeReduce(
		ConnectionStorageJoin(
			Map2PVName,
			report.PersistentVolumeClaim,
		),
		MakeFilter(
			func(n report.Node) bool {
				claimName, ok := n.Latest.Lookup(kubernetes.VolumeClaim)
				if claimName == "" {
					return !ok
				}
				return ok
			},
			MakeReduce(
				PropagateSingleMetrics(report.Container,
					MakeMap(
						Map3Parent([]string{report.Pod}),
						MakeFilter(
							ComposeFilterFuncs(
								IsRunning,
								Complement(isPauseContainer),
							),
							ContainerWithImageNameRenderer,
						),
					),
				),
				ConnectionStorageJoin(
					Map2PVCName,
					report.Pod,
				),
			),
		),
		MapStorageEndpoints(
			Map2PVNode,
			report.PersistentVolume,
		),
	),
)

// Map3Parent returns a MapFunc which maps Nodes to some parent grouping.
func Map3Parent(
	// The topology IDs to look for parents in
	topologies []string,
) MapFunc {
	return func(n report.Node) report.Nodes {
		result := report.Nodes{}
		for _, topology := range topologies {
			if groupIDs, ok := n.Parents.Lookup(topology); ok {
				for _, id := range groupIDs {
					node := NewDerivedNode(id, n).WithTopology(topology)
					node.Counters = node.Counters.Add(n.Topology, 1)
					result[id] = node
				}
			}
		}
		return result
	}
}

// ConnectionStorageJoin returns connectionStorageJoin object
func ConnectionStorageJoin(toPV func(report.Node) string, topology string) Renderer {
	return connectionStorageJoin{toPV: toPV, topology: topology}
}

// connectionStorageJoin holds the information about mapping of storage components
// along with TopologySelector
type connectionStorageJoin struct {
	toPV     func(report.Node) string
	topology string
}

func (c connectionStorageJoin) Render(rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(rpt).Nodes

	var pvcNodes = map[string]string{}
	for _, n := range inputNodes {
		pvName := c.toPV(n)
		pvcNodes[pvName] = n.ID
	}

	return MapStorageEndpoints(
		func(m report.Node) string {
			pvName, ok := m.Latest.Lookup(kubernetes.Name)
			if !ok {
				return ""
			}
			id := pvcNodes[pvName]
			return id
		}, c.topology).Render(rpt)
}

// Map2PVName accepts PV Node and returns Volume name associated with PV Node.
func Map2PVName(m report.Node) string {
	pvName, ok := m.Latest.Lookup(kubernetes.VolumeName)
	if !ok {
		pvName = ""
	}
	return pvName
}

// Map2PVCName returns pvc name
func Map2PVCName(m report.Node) string {
	pvcName, ok := m.Latest.Lookup(kubernetes.VolumeClaim)
	if !ok {
		pvcName = ""
	}
	return pvcName
}

// Map2PVNode returns pv node ID
func Map2PVNode(n report.Node) string {
	if pvNodeID, ok := n.Latest.Lookup(report.MakePersistentVolumeNodeID(n.ID)); ok {
		return pvNodeID
	}
	return ""
}

// mapStorageEndpoints is the Renderer for rendering storage components together.
type mapStorageEndpoints struct {
	f        endpointMapFunc
	topology string
}

// MapStorageEndpoints instantiates mapStorageEndpoints and returns same
func MapStorageEndpoints(f endpointMapFunc, topology string) Renderer {
	return mapStorageEndpoints{f: f, topology: topology}
}

func (e mapStorageEndpoints) Render(rpt report.Report) Nodes {

	var endpoints Nodes
	if e.topology == "persistent_volume_claim" {
		endpoints = SelectPersistentVolume.Render(rpt)
	}
	if e.topology == "pod" {
		endpoints = SelectPersistentVolumeClaim.Render(rpt)
	}
	ret := newJoinResults(TopologySelector(e.topology).Render(rpt).Nodes)

	for _, n := range endpoints.Nodes {
		if id := e.f(n); id != "" {
			ret.addChild(n, id, e.topology)
		}
	}
	return ret.storageResult(endpoints)
}
