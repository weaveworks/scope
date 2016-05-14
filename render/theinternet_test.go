package render_test

import (
	"net"
	"reflect"
	"testing"

	"$GITHUB_URI/probe/host"
	"$GITHUB_URI/render"
	"$GITHUB_URI/report"
	"$GITHUB_URI/test"
)

func TestReportLocalNetworks(t *testing.T) {
	r := report.MakeReport().Merge(report.Report{
		Host: report.Topology{
			Nodes: report.Nodes{
				"nonets": report.MakeNode("nonets"),
				"foo": report.MakeNode("foo").WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet(
						"10.0.0.1/8", "192.168.1.1/24", "10.0.0.1/8", "badnet/33")),
				),
			},
		},
		Overlay: report.Topology{
			Nodes: report.Nodes{
				"router": report.MakeNode("router").WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("10.32.0.1/12")),
				),
			},
		},
	})
	want := report.Networks([]*net.IPNet{
		mustParseCIDR("10.0.0.1/8"),
		mustParseCIDR("192.168.1.1/24"),
		mustParseCIDR("10.32.0.1/12"),
	})
	have := render.LocalNetworks(r)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}
