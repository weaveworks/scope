package report

import (
	"fmt"
	"strings"
)

// Topology describes a specific view of a network. It consists of
// nodes with metadata, and edges. Edges are directional, and embedded
// in the Node struct.
type Topology struct {
	Shape             string            `json:"shape,omitempty"`
	Tag               string            `json:"tag,omitempty"`
	Label             string            `json:"label,omitempty"`
	LabelPlural       string            `json:"label_plural,omitempty"`
	Nodes             Nodes             `json:"nodes,omitempty" deepequal:"nil==empty"`
	Controls          Controls          `json:"controls,omitempty" deepequal:"nil==empty"`
	MetadataTemplates MetadataTemplates `json:"metadata_templates,omitempty"`
	MetricTemplates   MetricTemplates   `json:"metric_templates,omitempty"`
	TableTemplates    TableTemplates    `json:"table_templates,omitempty"`
}

// MakeTopology gives you a Topology.
func MakeTopology() Topology {
	return Topology{
		Nodes:    map[string]Node{},
		Controls: Controls{},
	}
}

// WithMetadataTemplates merges some metadata templates into this topology,
// returning a new topology.
func (t Topology) WithMetadataTemplates(other MetadataTemplates) Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               t.Tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Merge(other),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// WithMetricTemplates merges some metadata templates into this topology,
// returning a new topology.
func (t Topology) WithMetricTemplates(other MetricTemplates) Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               t.Tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Merge(other),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// WithTableTemplates merges some table templates into this topology,
// returning a new topology.
func (t Topology) WithTableTemplates(other TableTemplates) Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               t.Tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Merge(other),
	}
}

// WithShape sets the shape of nodes from this topology, returning a new topology.
func (t Topology) WithShape(shape string) Topology {
	return Topology{
		Shape:             shape,
		Tag:               t.Tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// WithTag sets the tag of nodes from this topology, returning a new topology.
func (t Topology) WithTag(tag string) Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// WithLabel sets the label terminology of this topology, returning a new topology.
func (t Topology) WithLabel(label, labelPlural string) Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               t.Tag,
		Label:             label,
		LabelPlural:       labelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// AddNode adds node to the topology under key nodeID; if a
// node already exists for this key, nmd is merged with that node.
// This method is different from all the other similar methods
// in that it mutates the Topology, to solve issues of GC pressure.
func (t Topology) AddNode(node Node) {
	if existing, ok := t.Nodes[node.ID]; ok {
		node = node.Merge(existing)
	}
	t.Nodes[node.ID] = node
}

// ReplaceNode adds node to the topology under key nodeID; if a
// node already exists for this key, node replaces that node.
// Like AddNode, it mutates the Topology
func (t Topology) ReplaceNode(node Node) {
	t.Nodes[node.ID] = node
}

// GetShape returns the current topology shape, or the default if there isn't one.
func (t Topology) GetShape() string {
	if t.Shape == "" {
		return Circle
	}
	return t.Shape
}

// Copy returns a value copy of the Topology.
func (t Topology) Copy() Topology {
	return Topology{
		Shape:             t.Shape,
		Tag:               t.Tag,
		Label:             t.Label,
		LabelPlural:       t.LabelPlural,
		Nodes:             t.Nodes.Copy(),
		Controls:          t.Controls.Copy(),
		MetadataTemplates: t.MetadataTemplates.Copy(),
		MetricTemplates:   t.MetricTemplates.Copy(),
		TableTemplates:    t.TableTemplates.Copy(),
	}
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (t Topology) Merge(other Topology) Topology {
	shape := t.Shape
	if shape == "" {
		shape = other.Shape
	}
	label, labelPlural := t.Label, t.LabelPlural
	if label == "" {
		label, labelPlural = other.Label, other.LabelPlural
	}
	tag := t.Tag
	if tag == "" {
		tag = other.Tag
	}
	return Topology{
		Shape:             shape,
		Tag:               tag,
		Label:             label,
		LabelPlural:       labelPlural,
		Nodes:             t.Nodes.Merge(other.Nodes),
		Controls:          t.Controls.Merge(other.Controls),
		MetadataTemplates: t.MetadataTemplates.Merge(other.MetadataTemplates),
		MetricTemplates:   t.MetricTemplates.Merge(other.MetricTemplates),
		TableTemplates:    t.TableTemplates.Merge(other.TableTemplates),
	}
}

// UnsafeMerge merges the other object into this one, modifying the original.
func (t *Topology) UnsafeMerge(other Topology) {
	if t.Shape == "" {
		t.Shape = other.Shape
	}
	if t.Label == "" {
		t.Label, t.LabelPlural = other.Label, other.LabelPlural
	}
	if t.Tag == "" {
		t.Tag = other.Tag
	}
	t.Nodes.UnsafeMerge(other.Nodes)
	t.Controls = t.Controls.Merge(other.Controls)
	t.MetadataTemplates = t.MetadataTemplates.Merge(other.MetadataTemplates)
	t.MetricTemplates = t.MetricTemplates.Merge(other.MetricTemplates)
	t.TableTemplates = t.TableTemplates.Merge(other.TableTemplates)
}

// Nodes is a collection of nodes in a topology. Keys are node IDs.
// TODO(pb): type Topology map[string]Node
type Nodes map[string]Node

// Copy returns a value copy of the Nodes.
func (n Nodes) Copy() Nodes {
	if n == nil {
		return nil
	}
	cp := make(Nodes, len(n))
	for k, v := range n {
		cp[k] = v
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (n Nodes) Merge(other Nodes) Nodes {
	if len(other) > len(n) {
		n, other = other, n
	}
	if len(other) == 0 {
		return n
	}
	cp := n.Copy()
	cp.UnsafeMerge(other)
	return cp
}

// UnsafeMerge merges the other object into this one, modifying the original.
func (n *Nodes) UnsafeMerge(other Nodes) {
	for k, v := range other {
		if existing, ok := (*n)[k]; ok { // don't overwrite
			(*n)[k] = v.Merge(existing)
		} else {
			(*n)[k] = v
		}
	}
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	errs := []string{}

	// Check all nodes are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID, nmd := range t.Nodes {
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}

		// Check all adjancency keys has entries in Node.
		for _, dstNodeID := range nmd.Adjacency {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node missing from adjacency %q -> %q", nodeID, dstNodeID))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}

	return nil
}
