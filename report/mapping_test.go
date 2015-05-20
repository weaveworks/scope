package report

import (
	"fmt"
	"testing"
)

func TestMapping(t *testing.T) {
	for i, c := range []struct {
		f                                      MapFunc
		id                                     string
		meta                                   NodeMetadata
		wantOK                                 bool
		wantID, wantMajor, wantMinor, wantRank string
	}{
		{
			f:  NetworkHostname,
			id: ScopeDelim + "1.2.3.4",
			meta: NodeMetadata{
				"name": "my.host",
			},
			wantOK:    true,
			wantID:    "host:my.host",
			wantMajor: "my",
			wantMinor: "host",
			wantRank:  "my",
		},
		{
			f:  NetworkHostname,
			id: ScopeDelim + "1.2.3.4",
			meta: NodeMetadata{
				"name": "localhost",
			},
			wantOK:    true,
			wantID:    "host:localhost",
			wantMajor: "localhost",
			wantMinor: "",
			wantRank:  "localhost",
		},
		{
			f:  NetworkIP,
			id: ScopeDelim + "1.2.3.4",
			meta: NodeMetadata{
				"name": "my.host",
			},
			wantOK:    true,
			wantID:    "addr:" + ScopeDelim + "1.2.3.4",
			wantMajor: "1.2.3.4",
			wantMinor: "my.host",
			wantRank:  "1.2.3.4",
		},
		{
			f:  ProcessName,
			id: "not-used-alpha",
			meta: NodeMetadata{
				"pid":    "42",
				"name":   "curl",
				"domain": "hosta",
			},
			wantOK:    true,
			wantID:    "proc:hosta:curl",
			wantMajor: "curl",
			wantMinor: "hosta",
			wantRank:  "curl",
		},
		{
			f:  ProcessPID,
			id: "not-used-beta",
			meta: NodeMetadata{
				"pid":    "42",
				"name":   "curl",
				"domain": "hosta",
			},
			wantOK:    true,
			wantID:    "pid:hosta:42",
			wantMajor: "curl",
			wantMinor: "hosta (42)",
			wantRank:  "42",
		},
		{
			f:  ProcessCgroup,
			id: "not-used-delta",
			meta: NodeMetadata{
				"pid":    "42",
				"name":   "curl",
				"domain": "hosta",
				"cgroup": "systemd",
			},
			wantOK:    true,
			wantID:    "cgroup:hosta:systemd",
			wantMajor: "systemd",
			wantMinor: "hosta",
			wantRank:  "systemd",
		},
		{
			f:  ProcessCgroup,
			id: "not-used-kappa",
			meta: NodeMetadata{
				"pid":    "42536",
				"domain": "hosta",
				"cgroup": "", // missing cgroup, and
				"name":   "", // missing name
			},
			wantOK:    false,
			wantID:    "cgroup:hosta:",
			wantMajor: "",
			wantMinor: "hosta",
			wantRank:  "",
		},
		{
			f:  ProcessCgroup,
			id: "not-used-gamma",
			meta: NodeMetadata{
				"pid":    "42536",
				"domain": "hosta",
				"cgroup": "",              // missing cgroup, but
				"name":   "elasticsearch", // having name
			},
			wantOK:    true,
			wantID:    "cgroup:hosta:elasticsearch",
			wantMajor: "elasticsearch",
			wantMinor: "hosta",
			wantRank:  "elasticsearch",
		},
		{
			f:  ProcessName,
			id: "not-used-iota",
			meta: NodeMetadata{
				"pid":    "42",
				"domain": "hosta",
				"name":   "", // missing name
			},
			wantOK:    false,
			wantID:    "proc:hosta:",
			wantMajor: "",
			wantMinor: "hosta",
			wantRank:  "",
		},
	} {
		identity := fmt.Sprintf("(%d %s %v)", i, c.id, c.meta)

		m, haveOK := c.f(c.id, c.meta, false)
		if want, have := c.wantOK, haveOK; want != have {
			t.Errorf("%s: map OK error: want %v, have %v", identity, want, have)
		}
		if want, have := c.wantID, m.ID; want != have {
			t.Errorf("%s: map ID error: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantMajor, m.Major; want != have {
			t.Errorf("%s: map major label: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantMinor, m.Minor; want != have {
			t.Errorf("%s: map minor label: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantRank, m.Rank; want != have {
			t.Errorf("%s: map rank: want %#v, have %#v", identity, want, have)
		}
	}
}
