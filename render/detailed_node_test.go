package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func TestMakeDetailedNode(t *testing.T) {
	t.Skip("TODO")
}

func TestOriginTable(t *testing.T) {
	if _, ok := render.OriginTable(rpt, "not-found"); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]render.Table{
		client54001NodeID: {
			Title:   "Origin Endpoint",
			Numeric: false,
			Rows: []render.Row{
				{"Host name", clientHostName, ""},
				{"PID", "10001", ""},
				{"Process name", "curl", ""},
			},
		},
		clientAddressNodeID: {
			Title:   "Origin Address",
			Numeric: false,
			Rows: []render.Row{
				{"Host name", clientHostName, ""},
			},
		},
		report.MakeProcessNodeID(clientHostID, "4242"): {
			Title:   "Origin Process",
			Numeric: false,
			Rows: []render.Row{
				{"Name (comm)", "curl", ""},
				{"PID", "4242", ""},
			},
		},
		serverHostNodeID: {
			Title:   "Origin Host",
			Numeric: false,
			Rows: []render.Row{
				{"Host name", serverHostName, ""},
				{"Load", "0.01 0.01 0.01", ""},
				{"Operating system", "Linux", ""},
			},
		},
	} {
		have, ok := render.OriginTable(rpt, originID)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, diff(want, have))
		}
	}
}
