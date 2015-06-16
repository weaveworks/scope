package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func TestUngroupedMapping(t *testing.T) {
	for i, c := range []struct {
		f                                      render.LeafMapFunc
		id                                     string
		meta                                   report.NodeMetadata
		wantOK                                 bool
		wantID, wantMajor, wantMinor, wantRank string
	}{
		{
			f:  render.NetworkHostname,
			id: report.MakeAddressNodeID("", "1.2.3.4"),
			meta: report.NodeMetadata{
				"name": "my.host",
			},
			wantOK:    true,
			wantID:    "host:my.host",
			wantMajor: "my",
			wantMinor: "host",
			wantRank:  "my",
		},
		{
			f:  render.NetworkHostname,
			id: report.MakeAddressNodeID("", "1.2.3.4"),
			meta: report.NodeMetadata{
				"name": "localhost",
			},
			wantOK:    true,
			wantID:    "host:localhost",
			wantMajor: "localhost",
			wantMinor: "",
			wantRank:  "localhost",
		},
		{
			f:  render.ProcessPID,
			id: "not-used-beta",
			meta: report.NodeMetadata{
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
			f:  render.MapEndpoint2Container,
			id: "foo-id",
			meta: report.NodeMetadata{
				"pid":    "42",
				"name":   "curl",
				"domain": "hosta",
			},
			wantOK:    true,
			wantID:    "uncontained",
			wantMajor: "Uncontained",
			wantMinor: "",
			wantRank:  "uncontained",
		},
		{
			f:  render.MapEndpoint2Container,
			id: "bar-id",
			meta: report.NodeMetadata{
				"pid":                   "42",
				"name":                  "curl",
				"domain":                "hosta",
				"docker_container_id":   "d321fe0",
				"docker_container_name": "walking_sparrow",
				"docker_image_id":       "1101fff",
				"docker_image_name":     "org/app:latest",
			},
			wantOK:    true,
			wantID:    "d321fe0",
			wantMajor: "",
			wantMinor: "hosta",
			wantRank:  "",
		},
	} {
		identity := fmt.Sprintf("(%d %s %v)", i, c.id, c.meta)

		m, haveOK := c.f(c.meta)
		if want, have := c.wantOK, haveOK; want != have {
			t.Errorf("%s: map OK error: want %v, have %v", identity, want, have)
		}
		if want, have := c.wantID, m.ID; want != have {
			t.Errorf("%s: map ID error: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantMajor, m.LabelMajor; want != have {
			t.Errorf("%s: map major label: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantMinor, m.LabelMinor; want != have {
			t.Errorf("%s: map minor label: want %#v, have %#v", identity, want, have)
		}
		if want, have := c.wantRank, m.Rank; want != have {
			t.Errorf("%s: map rank: want %#v, have %#v", identity, want, have)
		}
	}
}

func TestGroupedMapping(t *testing.T) {
	t.Skipf("not yet implemented") // TODO
}
