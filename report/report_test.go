package report_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestReportLocalNetworks(t *testing.T) {
	r := report.MakeReport()
	r.Merge(report.Report{Host: report.Topology{NodeMetadatas: report.NodeMetadatas{
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
			report.MakeAdjacencyID(client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(client54002EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(server80EndpointNodeID):    report.MakeIDList(client54001EndpointNodeID, client54002EndpointNodeID, report.TheInternet),
		}
		have := reportFixture.Squash().Endpoint.Adjacency
		if !reflect.DeepEqual(want, have) {
			t.Error(diff(want, have))
		}
	}
	{
		want := report.Adjacency{
			report.MakeAdjacencyID(clientAddressNodeID): report.MakeIDList(serverAddressNodeID),
			report.MakeAdjacencyID(serverAddressNodeID): report.MakeIDList(clientAddressNodeID, report.TheInternet),
		}
		have := reportFixture.Squash().Address.Adjacency
		if !reflect.DeepEqual(want, have) {
			t.Error(diff(want, have))
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
