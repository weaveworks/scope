package host

import (
	"github.com/weaveworks/scope/report"
)

// Tagger tags each node in each topology of a report with the origin host
// node ID of this (probe) host. Effectively, a foreign key linking every node
// in every topology to an origin host node in the host topology.
type Tagger struct {
	hostNodeID string
}

// NewTagger tags each node with a foreign key linking it to its origin host
// in the host topology.
func NewTagger(hostID string) Tagger {
	return Tagger{
		hostNodeID: report.MakeHostNodeID(hostID),
	}
}

// Name of this tagger, for metrics gathering
func (Tagger) Name() string { return "Host" }

// Tag implements Tagger.
func (t Tagger) Tag(r report.Report) (report.Report, error) {
	var (
		metadata = map[string]string{report.HostNodeID: t.hostNodeID}
	)

	// Explicitly don't tag Endpoints, Addresses and Overlay nodes - These topologies include pseudo nodes,
	// and as such do their own host tagging.
	// Don't tag Pods so they can be reported centrally.
	for _, topology := range []report.Topology{r.Process, r.Container, r.ContainerImage, r.Host} {
		for _, node := range topology.Nodes {
			topology.ReplaceNode(node.WithLatests(metadata).WithParent(report.Host, t.hostNodeID))
		}
	}
	return r, nil
}
