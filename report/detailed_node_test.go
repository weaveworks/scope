package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestMakeDetailedNode(t *testing.T) {
	rendered := report.Render(reportFixture, report.SelectEndpoint, report.ProcessPID, report.BasicPseudoNode)
	renderableNodeID := report.MakeProcessNodeID(serverHostID, "215") // what ProcessPID does
	renderableNode, ok := rendered[renderableNodeID]
	if !ok {
		t.Fatalf("couldn't find %q", renderableNodeID)
	}
	have := report.MakeDetailedNode(reportFixture, renderableNode)
	want := report.DetailedNode{
		ID:         renderableNodeID,
		LabelMajor: "apache",
		LabelMinor: "(unknown) (215)", // unknown because we don't put a host_name in the process node metadata
		Pseudo:     false,
		Tables: []report.Table{
			{
				Title:   "Connections",
				Numeric: true,
				Rows: []report.Row{
					//{"TCP connections", "0", ""},
					{"Bytes ingress", "310", ""},
					{"Bytes egress", "3100", ""},
				},
			},
			{
				Title:   "Origin Host",
				Numeric: false,
				Rows: []report.Row{
					{"Host name", "server.host.com", ""},
					{"Load", "0.01 0.01 0.01", ""},
					{"Operating system", "Linux", ""},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}
