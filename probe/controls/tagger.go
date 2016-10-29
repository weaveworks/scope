package controls

import (
	"time"

	"github.com/weaveworks/scope/report"
)

// Tagger adds controls to nodes
type Tagger struct {
}

// NewTagger tags each node with generic controls
func NewTagger() Tagger {
	return Tagger{}
}

// Name of this tagger, for metrics gathering
func (Tagger) Name() string { return "Controls" }

// Tag implements Tagger.
func (t Tagger) Tag(r report.Report) (report.Report, error) {
	for _, topology := range []report.Topology{r.Process, r.Container, r.ContainerImage, r.Host, r.Pod} {
		// Each topology has most likely not many controls that need
		// to be propagated. Gather the list in propagatingControls.
		var propagatingControls []report.Control
		for _, control := range topology.Controls {
			if control.AlwaysPropagated {
				propagatingControls = append(propagatingControls, control)
			}
		}

		for _, node := range topology.Nodes {
			for _, control := range propagatingControls {
				metadata := map[string]string{"thisisatest": control.ID}
				topology.AddNode(node.WithLatests(metadata))
				topology.AddNode(node.WithLatestControl(control.ID, time.Now(), report.NodeControlData{Dead: false}))
			}
		}
	}
	return r, nil
}
