package report_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestReportCopy(t *testing.T) {
	one := report.MakeReport()
	two := one.Copy()
	two.Merge(reportFixture)
	if want, have := report.MakeReport(), one; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestReportLocalNetworks(t *testing.T) {
	r := report.MakeReport()
	r = r.Merge(report.Report{Host: report.Topology{NodeMetadatas: report.NodeMetadatas{
		"nonets": {},
		"foo":    {"local_networks": "10.0.0.1/8 192.168.1.1/24 10.0.0.1/8 badnet/33"},
	}}})
	if want, have := []*net.IPNet{
		mustParseCIDR("10.0.0.1/8"),
		mustParseCIDR("192.168.1.1/24"),
	}, r.LocalNetworks(); !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestReportSquash(t *testing.T) {
	{
		want := report.Adjacency{
			report.MakeAdjacencyID(clientHostID, client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(clientHostID, client54002EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(serverHostID, server80EndpointNodeID):    report.MakeIDList(client54001EndpointNodeID, client54002EndpointNodeID, report.TheInternet),
		}
		have := reportFixture.Squash().Endpoint.Adjacency
		if !reflect.DeepEqual(want, have) {
			t.Error(diff(want, have))
		}
	}
	{
		want := report.Adjacency{
			report.MakeAdjacencyID(clientHostID, clientAddressNodeID): report.MakeIDList(serverAddressNodeID),
			report.MakeAdjacencyID(serverHostID, serverAddressNodeID): report.MakeIDList(clientAddressNodeID, report.TheInternet),
		}
		have := reportFixture.Squash().Address.Adjacency
		if !reflect.DeepEqual(want, have) {
			t.Error(diff(want, have))
		}
	}
}

func TestReportEdgeMetadata(t *testing.T) {
	have := reportFixture.EdgeMetadata(report.SelectEndpoint, report.ProcessName, "apache", "curl")
	want := report.EdgeMetadata{
		WithBytes:    true,
		BytesEgress:  100,
		BytesIngress: 10,
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestReportOriginTable(t *testing.T) {
	if _, ok := reportFixture.OriginTable("not-found"); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]report.Table{
		report.MakeProcessNodeID(clientHostID, "4242"): {
			Title:   "Origin Process",
			Numeric: false,
			Rows: []report.Row{
				{"Host name", "client.host.com", ""},
			},
		},
		clientAddressNodeID: {
			Title:   "Origin Address",
			Numeric: false,
			Rows: []report.Row{
				{"Host name", "client.host.com", ""},
			},
		},
		report.MakeProcessNodeID(clientHostID, "4242"): {
			Title:   "Origin Process",
			Numeric: false,
			Rows: []report.Row{
				{"Process name", "curl", ""},
				{"PID", "4242", ""},
				{"Docker container ID", "a1b2c3d4e5", ""},
				{"Docker container name", "fixture-container", ""},
				{"Docker image ID", "0000000000", ""},
				{"Docker image name", "fixture/container:latest", ""},
			},
		},
		serverHostNodeID: {
			Title:   "Origin Host",
			Numeric: false,
			Rows: []report.Row{
				{"Host name", "server.host.com", ""},
				{"Load", "0.01 0.01 0.01", ""},
				{"Operating system", "Linux", ""},
			},
		},
	} {
		have, ok := reportFixture.OriginTable(originID)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, diff(want, have))
		}
	}
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}
