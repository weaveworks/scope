package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestMakeDetailedNode(t *testing.T) {
	t.Skip("TODO")
}

func TestOriginTable(t *testing.T) {
	if _, ok := report.OriginTable(reportFixture, "not-found"); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]report.Table{
		client54001EndpointNodeID: {
			Title:   "Origin Endpoint",
			Numeric: false,
			Rows:    []report.Row{{"Host name", clientHostName, ""}},
		},
		clientAddressNodeID: {
			Title:   "Origin Address",
			Numeric: false,
			Rows: []report.Row{
				{"Host name", clientHostName, ""},
			},
		},
		report.MakeProcessNodeID(clientHostID, "4242"): {
			Title:   "Origin Process",
			Numeric: false,
			Rows: []report.Row{
				{"Name (comm)", "curl", ""},
				{"PID", "4242", ""},
			},
		},
		serverHostNodeID: {
			Title:   "Origin Host",
			Numeric: false,
			Rows: []report.Row{
				{"Host name", serverHostName, ""},
				{"Load", "0.01 0.01 0.01", ""},
				{"Operating system", "Linux", ""},
			},
		},
	} {
		have, ok := report.OriginTable(reportFixture, originID)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, diff(want, have))
		}
	}
}
