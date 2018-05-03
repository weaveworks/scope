package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// ConnectionStorageJoin returns connectionStorageJoin object
func ConnectionStorageJoin(toPV func(report.Node) []string, topology string) Renderer {
	return connectionStorageJoin{toPV: toPV, topology: topology}
}

// connectionStorageJoin holds the information about mapping of storage components
// along with TopologySelector
type connectionStorageJoin struct {
	toPV     func(report.Node) []string
	topology string
}

func (c connectionStorageJoin) Render(rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(rpt).Nodes

	var pvcNodes = map[string][]string{}
	for _, n := range inputNodes {
		pvName := c.toPV(n)
		for _, name := range pvName {
			pvcNodes[name] = append(pvcNodes[name], n.ID)
		}
	}

	return MapStorageEndpoints(
		func(m report.Node) []string {
			pvName, ok := m.Latest.Lookup(kubernetes.Name)
			if !ok {
				return []string{""}
			}
			id := pvcNodes[pvName]
			return id
		}, c.topology).Render(rpt)
}

// Map2PVName accepts PV Node and returns Volume name associated with PV Node.
func Map2PVName(m report.Node) []string {
	pvName, ok := m.Latest.Lookup(kubernetes.VolumeName)
	scName, ok1 := m.Latest.Lookup(kubernetes.StorageClassName)
	if !ok {
		pvName = ""
	}
	if !ok1 {
		scName = ""
	}
	return []string{pvName, scName}
}

// Map2PVCName returns pvc name
func Map2PVCName(m report.Node) []string {
	pvcName, ok := m.Latest.Lookup(kubernetes.VolumeClaim)
	if !ok {
		pvcName = ""
	}
	return []string{pvcName}
}

// Map2PVNode returns pv node ID
func Map2PVNode(n report.Node) []string {
	if pvNodeID, ok := n.Latest.Lookup(report.MakePersistentVolumeNodeID(n.ID)); ok {
		return []string{pvNodeID}
	}
	return []string{""}
}

type storageEndpointMapFunc func(report.Node) []string

// mapStorageEndpoints is the Renderer for rendering storage components together.
type mapStorageEndpoints struct {
	f        storageEndpointMapFunc
	topology string
}

// MapStorageEndpoints instantiates mapStorageEndpoints and returns same
func MapStorageEndpoints(f storageEndpointMapFunc, topology string) Renderer {
	return mapStorageEndpoints{f: f, topology: topology}
}

func (e mapStorageEndpoints) Render(rpt report.Report) Nodes {
	var endpoints Nodes
	if e.topology == report.PersistentVolumeClaim {
		endpoints = SelectPersistentVolume.Render(rpt)
		endpoints.Merge(SelectStorageClass.Render(rpt))
	}
	if e.topology == report.Pod {
		endpoints = SelectPersistentVolumeClaim.Render(rpt)
	}
	ret := newJoinResults(TopologySelector(e.topology).Render(rpt).Nodes)

	for _, n := range endpoints.Nodes {
		if id := e.f(n); len(id) > 0 {
			for _, nodeID := range id {
				if nodeID != "" {
					ret.addChild(n, nodeID, e.topology)
				}
			}
		}
	}
	if e.topology == report.PersistentVolumeClaim {
		ret.storageResult(endpoints)
		endpoints = SelectStorageClass.Render(rpt)
		for _, n := range endpoints.Nodes {
			if id := e.f(n); len(id) > 0 {
				for _, nodeID := range id {
					if nodeID != "" {
						ret.addChild(n, nodeID, e.topology)
					}
				}
			}
		}
		return ret.storageResult(endpoints)
	}
	return ret.storageResult(endpoints)
}
