package render

import (
	"github.com/weaveworks/scope/report"
)

var PersistentVolumeRenderer = persistentVolumeRenderer{}

type persistentVolumeRenderer struct{}

func (p persistentVolumeRenderer) Render(rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for id, n := range rpt.PersistentVolumeClaim.Nodes {
		claims, _ := n.Parents.Lookup(report.PersistentVolumeClaim)
		for _, p := range claims {
			n = n.WithAdjacent(p)
		}
		nodes[id] = n
	}
	return Nodes{Nodes: nodes}
}
