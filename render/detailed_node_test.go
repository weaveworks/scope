package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/test"
)

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
				{"Address", clientIP, ""},
			},
		},
		serverProcessNodeID: {
			Title:   "Origin Process",
			Numeric: false,
			Rank:    2,
			Rows: []render.Row{
				{"Name (comm)", "apache", ""},
				{"PID", serverPID, ""},
			},
		},
		serverHostNodeID: {
			Title:   "Origin Host",
			Numeric: false,
			Rank:    1,
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
			t.Errorf("%q: %s", originID, test.Diff(want, have))
		}
	}
}

func TestMakeDetailedNode(t *testing.T) {
	renderableNode := render.ContainerRenderer.Render(rpt)[serverContainerID]
	have := render.MakeDetailedNode(rpt, renderableNode)
	want := render.DetailedNode{
		ID:         serverContainerID,
		LabelMajor: "server",
		LabelMinor: serverHostName,
		Pseudo:     false,
		Tables: []render.Table{
			{
				Title:   "Connections",
				Numeric: true,
				Rank:    100,
				Rows: []render.Row{
					{"Bytes ingress", "150", ""},
					{"Bytes egress", "1500", ""},
				},
			},
			{
				Title:   "Origin Container",
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{"ID", "5e4d3c2b1a", ""},
					{"Name", "server", ""},
					{"Image ID", "imageid456", ""},
				},
			},
			{
				Title:   "Origin Process",
				Numeric: false,
				Rank:    2,
				Rows: []render.Row{
					{"Name (comm)", "apache", ""},
					{"PID", "215", ""},
				},
			},
			{
				Title:   "Origin Host",
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{"Host name", "server.hostname.com", ""},
					{"Load", "0.01 0.01 0.01", ""},
					{"Operating system", "Linux", ""},
				},
			},
			{
				Title:   "Origin Endpoint",
				Numeric: false,
				Rows: []render.Row{
					{"Endpoint", "192.168.1.1", ""},
					{"Port", "80", ""},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
