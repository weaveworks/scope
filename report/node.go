package report

import (
	"time"

	"github.com/weaveworks/scope/common/mtime"
)

// Node describes a superset of the metadata that probes can collect about a
// given node in a given topology, along with the edges emanating from the
// node and metadata about those edges.
type Node struct {
	ID             string                   `json:"id,omitempty"`
	Topology       string                   `json:"topology,omitempty"`
	Counters       Counters                 `json:"counters,omitempty"`
	Sets           Sets                     `json:"sets,omitempty"`
	Adjacency      IDList                   `json:"adjacency"`
	Edges          EdgeMetadatas            `json:"edges,omitempty"`
	Controls       NodeControls             `json:"controls,omitempty"`
	LatestControls NodeControlDataLatestMap `json:"latestControls,omitempty"`
	Latest         StringLatestMap          `json:"latest,omitempty"`
	Metrics        Metrics                  `json:"metrics,omitempty"`
	Parents        Sets                     `json:"parents,omitempty"`
	Children       NodeSet                  `json:"children,omitempty"`
}

// MakeNode creates a new Node with no initial metadata.
func MakeNode(id string) Node {
	return Node{
		ID:             id,
		Counters:       EmptyCounters,
		Sets:           EmptySets,
		Adjacency:      EmptyIDList,
		Edges:          EmptyEdgeMetadatas,
		Controls:       MakeNodeControls(),
		LatestControls: EmptyNodeControlDataLatestMap,
		Latest:         EmptyStringLatestMap,
		Metrics:        Metrics{},
		Parents:        EmptySets,
	}
}

// MakeNodeWith creates a new Node with the supplied map.
func MakeNodeWith(id string, m map[string]string) Node {
	return MakeNode(id).WithLatests(m)
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

// WithLatestActiveControls returns a fresh copy of n, with active controls cs added to LatestControls.
func (n Node) WithLatestActiveControls(cs ...string) Node {
	lcs := map[string]NodeControlData{}
	for _, control := range cs {
		lcs[control] = NodeControlData{}
	}
	return n.WithLatestControls(lcs)
}

// WithLatestControls returns a fresh copy of n, with lcs added to LatestControls.
func (n Node) WithLatestControls(lcs map[string]NodeControlData) Node {
	result := n.Copy()
	ts := mtime.Now()
	for k, v := range lcs {
		result.LatestControls = result.LatestControls.Set(k, ts, v)
	}
	return result
}

// WithLatestControl produces a new Node with control added to it
func (n Node) WithLatestControl(control string, ts time.Time, data NodeControlData) Node {
	result := n.Copy()
	result.LatestControls = result.LatestControls.Set(control, ts, data)
	return result
}

// WithParents returns a fresh copy of n, with sets merged in.
func (n Node) WithParents(parents Sets) Node {
	result := n.Copy()
	result.Parents = result.Parents.Merge(parents)
	return result
}

// PruneParents returns a fresh copy of n, without any parents.
func (n Node) PruneParents() Node {
	result := n.Copy()
	result.Parents = EmptySets
	return result
}

// WithChildren returns a fresh copy of n, with children merged in.
func (n Node) WithChildren(children NodeSet) Node {
	result := n.Copy()
	result.Children = result.Children.Merge(children)
	return result
}

// WithChild returns a fresh copy of n, with one child merged in.
func (n Node) WithChild(child Node) Node {
	result := n.Copy()
	result.Children = result.Children.Merge(MakeNodeSet(child))
	return result
}

// Copy returns a value copy of the Node.
func (n Node) Copy() Node {
	cp := MakeNode(n.ID)
	cp.Topology = n.Topology
	cp.Counters = n.Counters.Copy()
	cp.Sets = n.Sets.Copy()
	cp.Adjacency = n.Adjacency.Copy()
	cp.Edges = n.Edges.Copy()
	cp.Controls = n.Controls.Copy()
	cp.LatestControls = n.LatestControls.Copy()
	cp.Latest = n.Latest.Copy()
	cp.Metrics = n.Metrics.Copy()
	cp.Parents = n.Parents.Copy()
	cp.Children = n.Children.Copy()
	return cp
}

// Merge mergses the individual components of a node and returns a
// fresh node.
func (n Node) Merge(other Node) Node {
	id := n.ID
	if id == "" {
		id = other.ID
	}
	topology := n.Topology
	if topology == "" {
		topology = other.Topology
	} else if other.Topology != "" && topology != other.Topology {
		panic("Cannot merge nodes with different topology types: " + topology + " != " + other.Topology)
	}
	return Node{
		ID:             id,
		Topology:       topology,
		Counters:       n.Counters.Merge(other.Counters),
		Sets:           n.Sets.Merge(other.Sets),
		Adjacency:      n.Adjacency.Merge(other.Adjacency),
		Edges:          n.Edges.Merge(other.Edges),
		Controls:       n.Controls.Merge(other.Controls),
		LatestControls: n.LatestControls.Merge(other.LatestControls),
		Latest:         n.Latest.Merge(other.Latest),
		Metrics:        n.Metrics.Merge(other.Metrics),
		Parents:        n.Parents.Merge(other.Parents),
		Children:       n.Children.Merge(other.Children),
	}
}
