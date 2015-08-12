package render_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestReportLocalNetworks(t *testing.T) {
	r := report.MakeReport().Merge(report.Report{
		Host: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				"nonets": report.MakeNodeMetadata(),
				"foo": report.MakeNodeMetadataWith(map[string]string{
					host.LocalNetworks: "10.0.0.1/8 192.168.1.1/24 10.0.0.1/8 badnet/33",
				}),
			},
		},
	})
	want := report.Networks([]*net.IPNet{
		mustParseCIDR("10.0.0.1/8"),
		mustParseCIDR("192.168.1.1/24"),
	})
	have := render.LocalNetworks(r)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestParseNetworks(t *testing.T) {
	var (
		hugenetStr  = "1.0.0.0/8"
		bignetStr   = "10.1.0.1/16"
		smallnetStr = "5.6.7.8/32"
		hugenet     = mustParseCIDR(hugenetStr)
		bignet      = mustParseCIDR(bignetStr)
		smallnet    = mustParseCIDR(smallnetStr)
	)
	for _, tc := range []struct {
		input string
		want  report.Networks
	}{
		{"", report.Networks{}},
		{fmt.Sprintf("%s", bignetStr), report.Networks([]*net.IPNet{bignet})},
		{fmt.Sprintf("%s %s", bignetStr, bignetStr), report.Networks([]*net.IPNet{bignet})},
		{fmt.Sprintf("%s foo %s oops  %s", hugenetStr, smallnetStr, hugenetStr), report.Networks([]*net.IPNet{hugenet, smallnet})},
	} {
		if want, have := tc.want, render.ParseNetworks(tc.input); !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
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
