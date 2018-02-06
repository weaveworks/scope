package report

// DNSRecord contains names that an IP address maps to
type DNSRecord struct {
	Forward StringSet `json:"forward,omitempty"`
	Reverse StringSet `json:"reverse,omitempty"`
}

// DNSRecords contains all address->name mappings for a report
type DNSRecords map[string]DNSRecord

// Copy makes a copy of the DNSRecords
func (r DNSRecords) Copy() DNSRecords {
	cp := make(DNSRecords, len(r))
	for k, v := range r {
		cp[k] = v
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (r DNSRecords) Merge(other DNSRecords) DNSRecords {
	if len(other) > len(r) {
		r, other = other, r
	}
	cp := r.Copy()
	for k, v := range other {
		if v2, ok := cp[k]; ok {
			cp[k] = DNSRecord{
				Forward: v.Forward.Merge(v2.Forward),
				Reverse: v.Reverse.Merge(v2.Reverse),
			}
		} else {
			cp[k] = v
		}
	}
	return cp
}

// FirstMatch returns the first DNS name where match() returns true
func (r DNSRecords) FirstMatch(id string, match func(name string) bool) (string, bool) {
	_, addr, _, ok := ParseEndpointNodeID(id)
	if !ok {
		return "", false
	}
	// we rely on StringSets being sorted, to make selection deterministic
	// prioritize forward names
	for _, hostname := range r[addr].Forward {
		if match(hostname) {
			return hostname, true
		}
	}
	for _, hostname := range r[addr].Reverse {
		if match(hostname) {
			return hostname, true
		}
	}
	return "", false
}
