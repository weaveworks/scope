package report

// Controls describe the control tags within the Nodes
type Controls map[string]Control

// A Control basically describes an RPC
type Control struct {
	ID           string `json:"id"`
	Human        string `json:"human"`
	Icon         string `json:"icon"` // from https://fortawesome.github.io/Font-Awesome/cheatsheet/ please
	Confirmation string `json:"confirmation,omitempty"`
	Rank         int    `json:"rank"`
	ProbeID      string `json:"probeId,omitempty"`
}

// Merge merges other with cs, returning a fresh Controls.
func (cs Controls) Merge(other Controls) Controls {
	if len(other) > len(cs) {
		cs, other = other, cs
	}
	if len(other) == 0 {
		return cs
	}
	result := cs.Copy()
	for k, v := range other {
		result[k] = v
	}
	return result
}

// Copy produces a copy of cs.
func (cs Controls) Copy() Controls {
	if cs == nil {
		return nil
	}
	result := Controls{}
	for k, v := range cs {
		result[k] = v
	}
	return result
}

// AddControl adds c added to cs.
func (cs Controls) AddControl(c Control) {
	cs[c.ID] = c
}

// AddControls adds a collection of controls to cs.
func (cs Controls) AddControls(controls []Control) {
	for _, c := range controls {
		cs[c.ID] = c
	}
}
