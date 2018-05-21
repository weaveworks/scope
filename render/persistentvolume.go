package render

import (
	"github.com/weaveworks/scope/report"
)

var PersistentVolumeRenderer = persistentVolumeRenderer{}

type persistentVolumeRenderer struct{}

func (p persistentVolumeRenderer) Render(rpt report.Report) Nodes {
	return Nodes{Nodes: rpt.PersistentVolumeClaim.Nodes}
}
