package report

import (
	"time"
)

// Controls describe the control tags within the Nodes
type Controls map[string]Control

// A Control basically describes an RPC
type Control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"` // from https://fortawesome.github.io/Font-Awesome/cheatsheet/ please
}

// Merge merges other with cs, returning a fresh Controls.
func (cs Controls) Merge(other Controls) Controls {
	result := cs.Copy()
	for k, v := range other {
		result[k] = v
	}
	return result
}

// Copy produces a copy of cs.
func (cs Controls) Copy() Controls {
	result := Controls{}
	for k, v := range cs {
		result[k] = v
	}
	return result
}

// AddControl returns a fresh Controls, c added to cs.
func (cs Controls) AddControl(c Control) {
	cs[c.ID] = c
}

// NodeControls represent the individual controls that are valid
// for a given node at a given point in time.  Its is immutable.
type NodeControls struct {
	Timestamp int64  `json:"timestamp"`
	Controls  IDList `json:"controls"`
}

// MakeNodeControls makes a new NodeControls
func MakeNodeControls() NodeControls {
	return NodeControls{
		Timestamp: time.Now().Unix(),
		Controls:  MakeIDList(),
	}
}

// Copy is a noop, as NodeControls is immutable
func (nc NodeControls) Copy() NodeControls {
	return nc
}

// Merge returns the newest of the two NodeControls; it does not take the union
// of the valid Controls.
func (nc NodeControls) Merge(other NodeControls) NodeControls {
	if other.Timestamp > nc.Timestamp {
		return other
	}
	return nc
}

// Add the new control IDs to this NodeControls, producing a fresh NodeControls.
func (nc NodeControls) Add(ids ...string) NodeControls {
	return NodeControls{
		Timestamp: time.Now().Unix(),
		Controls:  nc.Controls.Add(ids...),
	}
}
