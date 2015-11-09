package host

import (
	"github.com/weaveworks/scope/report"
)

// Tagger tags each node in each topology of a report with the origin host
// node ID of this (probe) host. Effectively, a foreign key linking every node
// in every topology to an origin host node in the host topology.
type Tagger struct {
	hostNodeID string
	probeID    string
}

// NewTagger tags each node with a foreign key linking it to its origin host
// in the host topology.
func NewTagger(hostID, probeID string) Tagger {
	return Tagger{
		hostNodeID: report.MakeHostNodeID(hostID),
		probeID:    probeID,
	}
}

// Tag implements Tagger.
func (t Tagger) Tag(r report.Report) (report.Report, error) {
	metadata := map[string]string{
		report.HostNodeID: t.hostNodeID,
		report.ProbeID:    t.probeID,
	}

	// Explicity don't tag Endpoints and Addresses - These topologies include pseudo nodes,
	// and as such do their own host tagging
	for _, topology := range []report.Topology{r.Process, r.Container, r.ContainerImage, r.Host, r.Overlay} {
		for id, node := range topology.Nodes {
			topology.AddNode(id, node.WithMetadata(metadata))
		}
	}
	return r, nil
}
