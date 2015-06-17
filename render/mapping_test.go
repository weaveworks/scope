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
