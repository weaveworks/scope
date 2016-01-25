package report

import (
	"time"

	"github.com/weaveworks/scope/common/mtime"
)

// Node describes a superset of the metadata that probes can collect about a
// given node in a given topology, along with the edges emanating from the
// node and metadata about those edges.
type Node struct {
	ID        string        `json:"id,omitempty"`
	Topology  string        `json:"topology,omitempty"`
	Counters  Counters      `json:"counters,omitempty"`
	Sets      Sets          `json:"sets,omitempty"`
	Adjacency IDList        `json:"adjacency"`
	Edges     EdgeMetadatas `json:"edges,omitempty"`
	Controls  NodeControls  `json:"controls,omitempty"`
	Latest    LatestMap     `json:"latest,omitempty"`
	Metrics   Metrics       `json:"metrics,omitempty"`
	Parents   Sets          `json:"parents,omitempty"`
}

// MakeNode creates a new Node with no initial metadata.
func MakeNode() Node {
	return Node{
		Counters:  EmptyCounters,
		Sets:      EmptySets,
		Adjacency: EmptyIDList,
		Edges:     EmptyEdgeMetadatas,
		Controls:  MakeNodeControls(),
		Latest:    EmptyLatestMap,
		Metrics:   Metrics{},
		Parents:   EmptySets,
	}
}

// MakeNodeWith creates a new Node with the supplied map.
func MakeNodeWith(m map[string]string) Node {
	return MakeNode().WithLatests(m)
}

// WithID returns a fresh copy of n, with ID changed.
func (n Node) WithID(id string) Node {
	result := n.Copy()
	result.ID = id
	return result
}

// WithTopology returns a fresh copy of n, with ID changed.
func (n Node) WithTopology(topology string) Node {
	result := n.Copy()
	result.Topology = topology
	return result
}

// Before is used for sorting nodes by topology and id
func (n Node) Before(other Node) bool {
	return n.Topology < other.Topology || (n.Topology == other.Topology && n.ID < other.ID)
}

// Equal is used for comparing nodes by topology and id
func (n Node) Equal(other Node) bool {
	return n.Topology == other.Topology && n.ID == other.ID
}

// After is used for sorting nodes by topology and id
func (n Node) After(other Node) bool {
	return other.Topology < n.Topology || (other.Topology == n.Topology && other.ID < n.ID)
}

// WithLatests returns a fresh copy of n, with Metadata m merged in.
func (n Node) WithLatests(m map[string]string) Node {
	result := n.Copy()
	ts := mtime.Now()
	for k, v := range m {
		result.Latest = result.Latest.Set(k, ts, v)
	}
	return result
}

// WithLatest produces a new Node with k mapped to v in the Latest metadata.
func (n Node) WithLatest(k string, ts time.Time, v string) Node {
	result := n.Copy()
	result.Latest = result.Latest.Set(k, ts, v)
	return result
}

// WithCounters returns a fresh copy of n, with Counters c merged in.
func (n Node) WithCounters(c map[string]int) Node {
	result := n.Copy()
	result.Counters = result.Counters.Merge(Counters{}.fromIntermediate(c))
	return result
}

// WithSet returns a fresh copy of n, with set merged in at key.
func (n Node) WithSet(key string, set StringSet) Node {
	result := n.Copy()
	result.Sets = result.Sets.Add(key, set)
	return result
}

// WithSets returns a fresh copy of n, with sets merged in.
func (n Node) WithSets(sets Sets) Node {
	result := n.Copy()
	result.Sets = result.Sets.Merge(sets)
	return result
}

// WithMetric returns a fresh copy of n, with metric merged in at key.
func (n Node) WithMetric(key string, metric Metric) Node {
	result := n.Copy()
	result.Metrics[key] = n.Metrics[key].Merge(metric)
	return result
}

// WithMetrics returns a fresh copy of n, with metrics merged in.
func (n Node) WithMetrics(metrics Metrics) Node {
	result := n.Copy()
	result.Metrics = result.Metrics.Merge(metrics)
	return result
}

// WithAdjacent returns a fresh copy of n, with 'a' added to Adjacency
func (n Node) WithAdjacent(a ...string) Node {
	result := n.Copy()
	result.Adjacency = result.Adjacency.Add(a...)
	return result
}

// WithEdge returns a fresh copy of n, with 'dst' added to Adjacency and md
// added to EdgeMetadata.
func (n Node) WithEdge(dst string, md EdgeMetadata) Node {
	result := n.Copy()
	result.Adjacency = result.Adjacency.Add(dst)
	result.Edges = result.Edges.Add(dst, md)
	return result
}

// WithControls returns a fresh copy of n, with cs added to Controls.
func (n Node) WithControls(cs ...string) Node {
	result := n.Copy()
	result.Controls = result.Controls.Add(cs...)
	return result
}

// WithParents returns a fresh copy of n, with sets merged in.
func (n Node) WithParents(parents Sets) Node {
	result := n.Copy()
	result.Parents = result.Parents.Merge(parents)
	return result
}

// Copy returns a value copy of the Node.
func (n Node) Copy() Node {
	cp := MakeNode()
	cp.ID = n.ID
	cp.Topology = n.Topology
	cp.Counters = n.Counters.Copy()
	cp.Sets = n.Sets.Copy()
	cp.Adjacency = n.Adjacency.Copy()
	cp.Edges = n.Edges.Copy()
	cp.Controls = n.Controls.Copy()
	cp.Latest = n.Latest.Copy()
	cp.Metrics = n.Metrics.Copy()
	cp.Parents = n.Parents.Copy()
	return cp
}

// Merge mergses the individual components of a node and returns a
// fresh node.
func (n Node) Merge(other Node) Node {
	cp := n.Copy()
	if cp.ID == "" {
		cp.ID = other.ID
	}
	if cp.Topology == "" {
		cp.Topology = other.Topology
	}
	cp.Counters = cp.Counters.Merge(other.Counters)
	cp.Sets = cp.Sets.Merge(other.Sets)
	cp.Adjacency = cp.Adjacency.Merge(other.Adjacency)
	cp.Edges = cp.Edges.Merge(other.Edges)
	cp.Controls = cp.Controls.Merge(other.Controls)
	cp.Latest = cp.Latest.Merge(other.Latest)
	cp.Metrics = cp.Metrics.Merge(other.Metrics)
	cp.Parents = cp.Parents.Merge(other.Parents)
	return cp
}
