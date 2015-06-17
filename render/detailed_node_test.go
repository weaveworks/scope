package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
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
				Rows: []render.Row{
					{"Bytes ingress", "150", ""},
					{"Bytes egress", "1500", ""},
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
			{
				Title:   "Origin Process",
				Numeric: false,
				Rows: []render.Row{
					{"Name (comm)", "apache", ""},
					{"PID", "215", ""},
				},
			},
			{
				Title:   "Origin Container",
				Numeric: false,
				Rows: []render.Row{
					{"Container ID", "5e4d3c2b1a", ""},
					{"Container name", "server", ""},
				},
			},
			{
				Title:   "Origin Host",
				Numeric: false,
				Rows: []render.Row{
					{"Host name", "server.hostname.com", ""},
					{"Load", "0.01 0.01 0.01", ""},
					{"Operating system", "Linux", ""},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", diff(want, have))
	}
}
