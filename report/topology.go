package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Topology describes a specific view of a network. It consists of nodes and
// edges, and metadata about those nodes and edges, represented by
// EdgeMetadatas and Nodes respectively. Edges are directional, and embedded
// in the Node struct.
type Topology struct {
	Nodes // TODO(pb): remove Nodes intermediate type
	Controls
}

// MakeTopology gives you a Topology.
func MakeTopology() Topology {
	return Topology{
		Nodes:    map[string]Node{},
		Controls: Controls{},
	}
}

// AddNode adds node to the topology under key nodeID; if a
// node already exists for this key, nmd is merged with that node.
// The same topology is returned to enable chaining.
// This method is different from all the other similar methods
// in that it mutates the Topology, to solve issues of GC pressure.
func (t Topology) AddNode(nodeID string, nmd Node) Topology {
	if existing, ok := t.Nodes[nodeID]; ok {
		nmd = nmd.Merge(existing)
	}
	t.Nodes[nodeID] = nmd
	return t
}

// Copy returns a value copy of the Topology.
func (t Topology) Copy() Topology {
	return Topology{
		Nodes:    t.Nodes.Copy(),
		Controls: t.Controls.Copy(),
	}
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (t Topology) Merge(other Topology) Topology {
	return Topology{
		Nodes:    t.Nodes.Merge(other.Nodes),
		Controls: t.Controls.Merge(other.Controls),
	}
}

// Nodes is a collection of nodes in a topology. Keys are node IDs.
// TODO(pb): type Topology map[string]Node
type Nodes map[string]Node

// Copy returns a value copy of the Nodes.
func (n Nodes) Copy() Nodes {
	cp := make(Nodes, len(n))
	for k, v := range n {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (n Nodes) Merge(other Nodes) Nodes {
	cp := n.Copy()
	for k, v := range other {
		if n, ok := cp[k]; ok { // don't overwrite
			v = v.Merge(n)
		}
		cp[k] = v
	}
	return cp
}

// Node describes a superset of the metadata that probes can collect about a
// given node in a given topology, along with the edges emanating from the
// node and metadata about those edges.
type Node struct {
	Metadata  Metadata      `json:"metadata,omitempty"`
	Counters  Counters      `json:"counters,omitempty"`
	Sets      Sets          `json:"sets,omitempty"`
	Adjacency IDList        `json:"adjacency"`
	Edges     EdgeMetadatas `json:"edges,omitempty"`
	Controls  NodeControls  `json:"controls,omitempty"`
	Latest    LatestMap     `json:"latest,omitempty"`
	Metrics   Metrics       `json:"metrics,omitempty"`
}

// MakeNode creates a new Node with no initial metadata.
func MakeNode() Node {
	return Node{
		Metadata:  Metadata{},
		Counters:  Counters{},
		Sets:      Sets{},
		Adjacency: MakeIDList(),
		Edges:     EdgeMetadatas{},
		Controls:  MakeNodeControls(),
		Latest:    MakeLatestMap(),
		Metrics:   Metrics{},
	}
}

// MakeNodeWith creates a new Node with the supplied map.
func MakeNodeWith(m map[string]string) Node {
	return MakeNode().WithMetadata(m)
}

// WithMetadata returns a fresh copy of n, with Metadata m merged in.
func (n Node) WithMetadata(m map[string]string) Node {
	result := n.Copy()
	result.Metadata = result.Metadata.Merge(m)
	return result
}

// WithCounters returns a fresh copy of n, with Counters c merged in.
func (n Node) WithCounters(c map[string]int) Node {
	result := n.Copy()
	result.Counters = result.Counters.Merge(c)
	return result
}

// WithSet returns a fresh copy of n, with set merged in at key.
func (n Node) WithSet(key string, set StringSet) Node {
	result := n.Copy()
	existing := n.Sets[key]
	result.Sets[key] = existing.Merge(set)
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
	existing := n.Metrics[key]
	result.Metrics[key] = existing.Merge(metric)
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
	result.Edges[dst] = md
	return result
}

// WithControls returns a fresh copy of n, with cs added to Controls.
func (n Node) WithControls(cs ...string) Node {
	result := n.Copy()
	result.Controls = result.Controls.Add(cs...)
	return result
}

// WithLatest produces a new Node with k mapped to v in the Latest metadata.
func (n Node) WithLatest(k string, ts time.Time, v string) Node {
	result := n.Copy()
	result.Latest = result.Latest.Set(k, ts, v)
	return result
}

// Copy returns a value copy of the Node.
func (n Node) Copy() Node {
	cp := MakeNode()
	cp.Metadata = n.Metadata.Copy()
	cp.Counters = n.Counters.Copy()
	cp.Sets = n.Sets.Copy()
	cp.Adjacency = n.Adjacency.Copy()
	cp.Edges = n.Edges.Copy()
	cp.Controls = n.Controls.Copy()
	cp.Latest = n.Latest.Copy()
	cp.Metrics = n.Metrics.Copy()
	return cp
}

// Merge mergses the individual components of a node and returns a
// fresh node.
func (n Node) Merge(other Node) Node {
	cp := n.Copy()
	cp.Metadata = cp.Metadata.Merge(other.Metadata)
	cp.Counters = cp.Counters.Merge(other.Counters)
	cp.Sets = cp.Sets.Merge(other.Sets)
	cp.Adjacency = cp.Adjacency.Merge(other.Adjacency)
	cp.Edges = cp.Edges.Merge(other.Edges)
	cp.Controls = cp.Controls.Merge(other.Controls)
	cp.Latest = cp.Latest.Merge(other.Latest)
	cp.Metrics = cp.Metrics.Merge(other.Metrics)
	return cp
}

// Metadata is a string->string map.
type Metadata map[string]string

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge does not modify the receiver.
func (m Metadata) Merge(other Metadata) Metadata {
	result := m.Copy()
	for k, v := range other {
		result[k] = v // other takes precedence
	}
	return result
}

// Copy creates a deep copy of the Metadata.
func (m Metadata) Copy() Metadata {
	result := Metadata{}
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Counters is a string->int map.
type Counters map[string]int

// Merge merges two sets of counters into a fresh set of counters, summing
// values where appropriate.
func (c Counters) Merge(other Counters) Counters {
	result := c.Copy()
	for k, v := range other {
		result[k] = result[k] + v
	}
	return result
}

// Copy creates a deep copy of the Counters.
func (c Counters) Copy() Counters {
	result := Counters{}
	for k, v := range c {
		result[k] = v
	}
	return result
}

// Sets is a string->set-of-strings map.
type Sets map[string]StringSet

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (s Sets) Merge(other Sets) Sets {
	result := s.Copy()
	for k, v := range other {
		result[k] = result[k].Merge(v)
	}
	return result
}

// Copy returns a value copy of the sets map.
func (s Sets) Copy() Sets {
	result := Sets{}
	for k, v := range s {
		result[k] = v.Copy()
	}
	return result
}

// StringSet is a sorted set of unique strings. Clients must use the Add
// method to add strings.
type StringSet []string

// MakeStringSet makes a new StringSet with the given strings.
func MakeStringSet(strs ...string) StringSet {
	if len(strs) <= 0 {
		return nil
	}
	result := make([]string, len(strs))
	copy(result, strs)
	sort.Strings(result)
	for i := 1; i < len(result); { // shuffle down any duplicates
		if result[i-1] == result[i] {
			result = append(result[:i-1], result[i:]...)
			continue
		}
		i++
	}
	return StringSet(result)
}

// Add adds the strings to the StringSet. Add is the only valid way to grow a
// StringSet. Add returns the StringSet to enable chaining.
func (s StringSet) Add(strs ...string) StringSet {
	for _, str := range strs {
		i := sort.Search(len(s), func(i int) bool { return s[i] >= str })
		if i < len(s) && s[i] == str {
			// The list already has the element.
			continue
		}
		// It a new element, insert it in order.
		s = append(s, "")
		copy(s[i+1:], s[i:])
		s[i] = str
	}
	return s
}

// Merge combines the two StringSets and returns a new result.
func (s StringSet) Merge(other StringSet) StringSet {
	switch {
	case len(other) <= 0: // Optimise special case, to avoid allocating
		return s // (note unit test DeepEquals breaks if we don't do this)
	case len(s) <= 0:
		return other
	}
	result := make(StringSet, len(s)+len(other))
	for i, j, k := 0, 0, 0; ; k++ {
		switch {
		case i >= len(s):
			copy(result[k:], other[j:])
			return result[:k+len(other)-j]
		case j >= len(other):
			copy(result[k:], s[i:])
			return result[:k+len(s)-i]
		case s[i] < other[j]:
			result[k] = s[i]
			i++
		case s[i] > other[j]:
			result[k] = other[j]
			j++
		default: // equal
			result[k] = s[i]
			i++
			j++
		}
	}
}

// Copy returns a value copy of the StringSet.
func (s StringSet) Copy() StringSet {
	if s == nil {
		return s
	}
	result := make(StringSet, len(s))
	copy(result, s)
	return result
}

// Metrics is a string->metric map.
type Metrics map[string]Metric

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (m Metrics) Merge(other Metrics) Metrics {
	result := m.Copy()
	for k, v := range other {
		result[k] = result[k].Merge(v)
	}
	return result
}

// Copy returns a value copy of the sets map.
func (m Metrics) Copy() Metrics {
	result := Metrics{}
	for k, v := range m {
		result[k] = v.Copy()
	}
	return result
}

// Metric is a list of timeseries data. Clients must use the Add
// method to add values.
type Metric struct {
	Samples []Sample  `json:"samples"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	First   time.Time `json:"first"`
	Last    time.Time `json:"last"`
}

type Sample struct {
	Timestamp time.Time `json:"date"`
	Value     float64   `json:"value"`
}

// MakeMetric makes a new Metric.
func MakeMetric() Metric {
	return Metric{}
}

func (m Metric) WithFirst(t time.Time) Metric {
	m.First = t
	return m
}

// Add adds the sample to the Metric. Add is the only valid way to grow a
// Metric. Add returns the Metric to enable chaining.
func (m Metric) Add(t time.Time, v float64) Metric {
	i := sort.Search(len(m.Samples), func(i int) bool { return !t.After(m.Samples[i].Timestamp) })
	if i < len(m.Samples) && m.Samples[i].Timestamp.Equal(t) {
		// The list already has the element.
		return m
	}
	// It is a new element, insert it in order.
	m.Samples = append(m.Samples, Sample{})
	copy(m.Samples[i+1:], m.Samples[i:])
	m.Samples[i] = Sample{Timestamp: t, Value: v}
	if v > m.Max {
		m.Max = v
	}
	if v < m.Min {
		m.Min = v
	}
	if m.First.IsZero() || t.Before(m.First) {
		m.First = t
	}
	if m.Last.IsZero() || t.After(m.Last) {
		m.Last = t
	}
	return m
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {
	for _, sample := range other.Samples {
		m = m.Add(sample.Timestamp, sample.Value)
	}
	if !other.First.IsZero() && other.First.Before(m.First) {
		m.First = other.First
	}
	if !other.Last.IsZero() && other.Last.After(m.Last) {
		m.Last = other.Last
	}
	if other.Min < m.Min {
		m.Min = other.Min
	}
	if other.Max > m.Max {
		m.Max = other.Max
	}
	return m
}

// Copy returns a value copy of the Metric.
func (m Metric) Copy() Metric {
	var samples []Sample
	if m.Samples != nil {
		samples = make([]Sample, len(m.Samples))
		copy(samples, m.Samples)
	}
	return Metric{Samples: samples, Max: m.Max, Min: m.Min, First: m.First, Last: m.Last}
}

// Last returns the last sample in the metric.
// Returns nil if there are no samples.
func (m Metric) LastSample() *Sample {
	if len(m.Samples) == 0 {
		return nil
	}
	return &m.Samples[len(m.Samples)-1]
}

// EdgeMetadatas collect metadata about each edge in a topology. Keys are the
// remote node IDs, as in Adjacency.
type EdgeMetadatas map[string]EdgeMetadata

// Copy returns a value copy of the EdgeMetadatas.
func (e EdgeMetadatas) Copy() EdgeMetadatas {
	cp := make(EdgeMetadatas, len(e))
	for k, v := range e {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (e EdgeMetadatas) Merge(other EdgeMetadatas) EdgeMetadatas {
	cp := e.Copy()
	for k, v := range other {
		cp[k] = cp[k].Merge(v)
	}
	return cp
}

// Flatten flattens all the EdgeMetadatas in this set and returns the result.
// The original is not modified.
func (e EdgeMetadatas) Flatten() EdgeMetadata {
	result := EdgeMetadata{}
	for _, v := range e {
		result = result.Flatten(v)
	}
	return result
}

// EdgeMetadata describes a superset of the metadata that probes can possibly
// collect about a directed edge between two nodes in any topology.
type EdgeMetadata struct {
	EgressPacketCount  *uint64 `json:"egress_packet_count,omitempty"`
	IngressPacketCount *uint64 `json:"ingress_packet_count,omitempty"`
	EgressByteCount    *uint64 `json:"egress_byte_count,omitempty"`  // Transport layer
	IngressByteCount   *uint64 `json:"ingress_byte_count,omitempty"` // Transport layer
	MaxConnCountTCP    *uint64 `json:"max_conn_count_tcp,omitempty"`
}

// Copy returns a value copy of the EdgeMetadata.
func (e EdgeMetadata) Copy() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.EgressPacketCount),
		IngressPacketCount: cpu64ptr(e.IngressPacketCount),
		EgressByteCount:    cpu64ptr(e.EgressByteCount),
		IngressByteCount:   cpu64ptr(e.IngressByteCount),
		MaxConnCountTCP:    cpu64ptr(e.MaxConnCountTCP),
	}
}

// Reversed returns a value copy of the EdgeMetadata, with the direction reversed.
func (e EdgeMetadata) Reversed() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.IngressPacketCount),
		IngressPacketCount: cpu64ptr(e.EgressPacketCount),
		EgressByteCount:    cpu64ptr(e.IngressByteCount),
		IngressByteCount:   cpu64ptr(e.EgressByteCount),
		MaxConnCountTCP:    cpu64ptr(e.MaxConnCountTCP),
	}
}

func cpu64ptr(u *uint64) *uint64 {
	if u == nil {
		return nil
	}
	value := *u   // oh man
	return &value // this sucks
}

// Merge merges another EdgeMetadata into the receiver and returns the result.
// The receiver is not modified. The two edge metadatas should represent the
// same edge on different times.
func (e EdgeMetadata) Merge(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, max)
	return cp
}

// Flatten sums two EdgeMetadatas and returns the result. The receiver is not
// modified. The two edge metadata windows should be the same duration; they
// should represent different edges at the same time.
func (e EdgeMetadata) Flatten(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	// Note that summing of two maximums doesn't always give us the true
	// maximum. But it's a best effort.
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, sum)
	return cp
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	errs := []string{}

	// Check all node metadatas are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID, nmd := range t.Nodes {
		if nmd.Metadata == nil {
			errs = append(errs, fmt.Sprintf("node ID %q has nil metadata", nodeID))
		}
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}

		// Check all adjancency keys has entries in Node.
		for _, dstNodeID := range nmd.Adjacency {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node metadata missing from adjacency %q -> %q", nodeID, dstNodeID))
			}
		}

		// Check all the edge metadatas have entries in adjacencies
		for dstNodeID := range nmd.Edges {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node %s metadatas missing for edge %q", dstNodeID, nodeID))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}

	return nil
}

func merge(dst, src *uint64, op func(uint64, uint64) uint64) *uint64 {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = new(uint64)
	}
	(*dst) = op(*dst, *src)
	return dst
}

func sum(dst, src uint64) uint64 {
	return dst + src
}

func max(dst, src uint64) uint64 {
	if dst > src {
		return dst
	}
	return src
}
