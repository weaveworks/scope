package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
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
				{"Endpoint", clientIP, ""},
				{"Port", clientPort54001, ""},
			},
		},
		clientAddressNodeID: {
			Title:   "Origin Address",
			Numeric: false,
			Rows: []render.Row{
				{"Host name", clientHostName, ""},
			},
		},
		serverProcessNodeID: {
			Title:   "Origin Process",
			Numeric: false,
			Rows: []render.Row{
				{"Name (comm)", "apache", ""},
				{"PID", serverPID, ""},
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
