package main

import (
	"log"

	"github.com/weaveworks/scope/report"
)

// Tagger tags nodes with value-add node metadata.
type Tagger interface {
	Tag(r report.Report) (report.Report, error)
}

// Reporter generates Reports.
type Reporter interface {
	Report() (report.Report, error)
}

// Ticker is something which will be invoked every spyDuration.
// It's useful for things that should be updated on that interval.
// For example, cached shared state between Taggers and Reporters.
type Ticker interface {
	Tick() error
}

// Apply tags the report with all the taggers.
func Apply(r report.Report, taggers []Tagger) report.Report {
	var err error
	for _, tagger := range taggers {
		r, err = tagger.Tag(r)
		if err != nil {
			log.Printf("error applying tagger: %v", err)
		}
	}
	return r
}

// Topology is the Node key for the origin topology.
const Topology = "topology"

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from. It's
// kind of a proof-of-concept tagger, useful primarily for debugging.
func newTopologyTagger() Tagger {
	return &topologyTagger{}
}

// Tag implements Tagger
func (topologyTagger) Tag(r report.Report) (report.Report, error) {
	for val, topology := range map[string]*report.Topology{
		"endpoint":        &(r.Endpoint),
		"address":         &(r.Address),
		"process":         &(r.Process),
		"container":       &(r.Container),
		"container_image": &(r.ContainerImage),
		"host":            &(r.Host),
		"overlay":         &(r.Overlay),
	} {
		other := report.MakeNodeWith(map[string]string{Topology: val})
		for id := range topology.Nodes {
			topology.AddNode(id, other)
		}
	}
	return r, nil
}
