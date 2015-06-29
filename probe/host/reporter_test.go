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

const (
	release  = "release"
	version  = "version"
	network  = "192.168.0.0/16"
	hostID   = "hostid"
	now      = "now"
	hostname = "hostname"
	load     = "0.59 0.36 0.29"
	uptime   = "278h55m43s"
	kernel   = "release version"
)

func TestReporter(t *testing.T) {
	var (
		oldGetKernelVersion = host.GetKernelVersion
		oldGetLoad          = host.GetLoad
		oldGetUptime        = host.GetUptime
		oldInterfaceAddrs   = host.InterfaceAddrs
		oldNow              = host.Now
	)
	defer func() {
		host.GetKernelVersion = oldGetKernelVersion
		host.GetLoad = oldGetLoad
		host.GetUptime = oldGetUptime
		host.InterfaceAddrs = oldInterfaceAddrs
		host.Now = oldNow
	}()
	host.GetKernelVersion = func() (string, error) { return release + " " + version, nil }
	host.GetLoad = func() string { return load }
	host.GetUptime = func() (time.Duration, error) { return time.ParseDuration(uptime) }
	host.Now = func() string { return now }
	host.InterfaceAddrs = func() ([]net.Addr, error) { _, ipnet, _ := net.ParseCIDR(network); return []net.Addr{ipnet}, nil }

	want := report.MakeReport()
	want.Host.NodeMetadatas[report.MakeHostNodeID(hostID)] = report.NodeMetadata{
		host.Timestamp:     now,
		host.HostName:      hostname,
		host.LocalNetworks: network,
		host.OS:            runtime.GOOS,
		host.Load:          load,
		host.Uptime:        uptime,
		host.KernelVersion: kernel,
	}
	r := host.NewReporter(hostID, hostname)
	have, _ := r.Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
