package host_test

import (
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestReporter(t *testing.T) {
	var (
		release     = "release"
		version     = "version"
		network     = "192.168.0.0/16"
		hostID      = "hostid"
		now         = "now"
		hostname    = "hostname"
		load        = "0.59 0.36 0.29"
		uptime      = "278h55m43s"
		kernel      = "release version"
		_, ipnet, _ = net.ParseCIDR(network)
		localNets   = report.Networks([]*net.IPNet{ipnet})
	)

	var (
		oldGetKernelVersion = host.GetKernelVersion
		oldGetLoad          = host.GetLoad
		oldGetUptime        = host.GetUptime
		oldNow              = host.Now
	)
	defer func() {
		host.GetKernelVersion = oldGetKernelVersion
		host.GetLoad = oldGetLoad
		host.GetUptime = oldGetUptime
		host.Now = oldNow
	}()
	host.GetKernelVersion = func() (string, error) { return release + " " + version, nil }
	host.GetLoad = func() string { return load }
	host.GetUptime = func() (time.Duration, error) { return time.ParseDuration(uptime) }
	host.Now = func() string { return now }

	want := report.MakeReport()
	want.Host.AddNode(report.MakeHostNodeID(hostID), report.MakeNodeWith(map[string]string{
		host.Timestamp:     now,
		host.HostName:      hostname,
		host.OS:            runtime.GOOS,
		host.Load:          load,
		host.Uptime:        uptime,
		host.KernelVersion: kernel,
	}).WithSets(report.Sets{
		host.LocalNetworks: report.MakeStringSet(network),
	}))
	have, _ := host.NewReporter(hostID, hostname, localNets).Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
