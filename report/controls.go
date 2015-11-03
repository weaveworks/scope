package report

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
