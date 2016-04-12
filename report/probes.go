package report

import (
	"time"
)

// Probes contains details of the probe(s) which generated a report.
type Probes map[string]Probe

// Copy produces a copy of the Probes
func (ps Probes) Copy() Probes {
	result := Probes{}
	for id, probe := range ps {
		result[id] = probe.Copy()
	}
	return result
}

// Merge two sets of Probes, keeping the records with the latest LastSeen
func (ps Probes) Merge(other Probes) Probes {
	result := ps.Copy()
	for id, probe := range other {
		o, ok := result[id]
		if !ok || o.LastSeen.Before(probe.LastSeen) {
			result[id] = probe
		}
	}
	return result
}

// Probe is the details for a single probe that generated a report.
type Probe struct {
	ID       string    `json:"id"`
	LastSeen time.Time `json:"lastSeen"`
}

// Copy produces a copy of the Probe
func (p Probe) Copy() Probe {
	return Probe{
		ID:       p.ID,
		LastSeen: p.LastSeen,
	}
}
