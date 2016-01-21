package host_test

import (
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestReporter(t *testing.T) {
	var (
		release   = "release"
		version   = "version"
		network   = "192.168.0.0/16"
		hostID    = "hostid"
		hostname  = "hostname"
		timestamp = time.Now()
		load      = report.Metrics{
			host.Load1:       report.MakeMetric().Add(timestamp, 1.0),
			host.Load5:       report.MakeMetric().Add(timestamp, 5.0),
			host.Load15:      report.MakeMetric().Add(timestamp, 15.0),
			host.CPUUsage:    report.MakeMetric().Add(timestamp, 30.0).WithMax(100.0),
			host.MemoryUsage: report.MakeMetric().Add(timestamp, 60.0).WithMax(100.0),
		}
		uptime      = "278h55m43s"
		kernel      = "release version"
		_, ipnet, _ = net.ParseCIDR(network)
		localNets   = report.Networks([]*net.IPNet{ipnet})
	)

	mtime.NowForce(timestamp)
	defer mtime.NowReset()

	var (
		oldGetKernelVersion    = host.GetKernelVersion
		oldGetLoad             = host.GetLoad
		oldGetUptime           = host.GetUptime
		oldGetCPUUsagePercent  = host.GetCPUUsagePercent
		oldGetMemoryUsageBytes = host.GetMemoryUsageBytes
	)
	defer func() {
		host.GetKernelVersion = oldGetKernelVersion
		host.GetLoad = oldGetLoad
		host.GetUptime = oldGetUptime
		host.GetCPUUsagePercent = oldGetCPUUsagePercent
		host.GetMemoryUsageBytes = oldGetMemoryUsageBytes
	}()
	host.GetKernelVersion = func() (string, error) { return release + " " + version, nil }
	host.GetLoad = func(time.Time) report.Metrics { return load }
	host.GetUptime = func() (time.Duration, error) { return time.ParseDuration(uptime) }
	host.GetCPUUsagePercent = func() (float64, float64) { return 30.0, 100.0 }
	host.GetMemoryUsageBytes = func() (float64, float64) { return 60.0, 100.0 }

	want := report.MakeReport()
	want.Host.AddNode(report.MakeHostNodeID(hostID), report.MakeNodeWith(map[string]string{
		host.Timestamp:     timestamp.UTC().Format(time.RFC3339Nano),
		host.HostName:      hostname,
		host.OS:            runtime.GOOS,
		host.Uptime:        uptime,
		host.KernelVersion: kernel,
	}).WithSets(report.Sets{
		host.LocalNetworks: report.MakeStringSet(network),
	}).WithMetrics(load))
	have, _ := host.NewReporter(hostID, hostname, localNets).Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
