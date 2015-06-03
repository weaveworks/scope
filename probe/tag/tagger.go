package tag

import "github.com/weaveworks/scope/report"

// Tagger tags nodes with value-add node metadata.
type Tagger interface {
	Tag(r report.Report, ts report.TopologySelector, id string) report.NodeMetadata
	Stop()
}

// Apply tags all nodes in the report with all taggers.
func Apply(r report.Report, taggers []Tagger) report.Report {
	for _, tagger := range taggers {
		r.Endpoint = tagTopology(r, report.SelectEndpoint, r.Endpoint, tagger)
		r.Address = tagTopology(r, report.SelectAddress, r.Address, tagger)
		r.Process = tagTopology(r, report.SelectProcess, r.Process, tagger)
		r.Host = tagTopology(r, report.SelectHost, r.Host, tagger)
	}
	return r
}

func tagTopology(r report.Report, ts report.TopologySelector, t report.Topology, tagger Tagger) report.Topology {
	for nodeID := range t.NodeMetadatas {
		t.NodeMetadatas[nodeID] = t.NodeMetadatas[nodeID].Merge(tagger.Tag(r, ts, nodeID))
	}
	return t
}
