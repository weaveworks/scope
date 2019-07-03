package host_test

import (
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

func TestReporter(t *testing.T) {
	var (
		release   = "release"
		version   = "version"
		network   = "192.168.0.0/16"
		hostID    = "hostid"
		hostname  = "hostname"
		timestamp = time.Now()
		metrics   = report.Metrics{
			host.Load1:       report.MakeSingletonMetric(timestamp, 1.0),
			host.CPUUsage:    report.MakeSingletonMetric(timestamp, 30.0).WithMax(100.0),
			host.MemoryUsage: report.MakeSingletonMetric(timestamp, 60.0).WithMax(100.0),
		}
		uptime      = "3600" // one hour
		kernel      = "release version"
		_, ipnet, _ = net.ParseCIDR(network)
	)

	mtime.NowForce(timestamp)
	defer mtime.NowReset()

	var (
		oldGetKernelReleaseAndVersion = host.GetKernelReleaseAndVersion
		oldGetLoad                    = host.GetLoad
		oldGetUptime                  = host.GetUptime
		oldGetCPUUsagePercent         = host.GetCPUUsagePercent
		oldGetMemoryUsageBytes        = host.GetMemoryUsageBytes
		oldGetLocalNetworks           = host.GetLocalNetworks
	)
	defer func() {
		host.GetKernelReleaseAndVersion = oldGetKernelReleaseAndVersion
		host.GetLoad = oldGetLoad
		host.GetUptime = oldGetUptime
		host.GetCPUUsagePercent = oldGetCPUUsagePercent
		host.GetMemoryUsageBytes = oldGetMemoryUsageBytes
		host.GetLocalNetworks = oldGetLocalNetworks
	}()
	host.GetKernelReleaseAndVersion = func() (string, string, error) { return release, version, nil }
	host.GetLoad = func(time.Time) report.Metrics { return metrics }
	host.GetUptime = func() (time.Duration, error) { return time.Hour, nil }
	host.GetCPUUsagePercent = func() (float64, float64) { return 30.0, 100.0 }
	host.GetMemoryUsageBytes = func() (float64, float64) { return 60.0, 100.0 }
	host.GetLocalNetworks = func() ([]*net.IPNet, error) { return []*net.IPNet{ipnet}, nil }

	hr := controls.NewDefaultHandlerRegistry()
	rpt, err := host.NewReporter(hostID, hostname, "probe-id", "", nil, hr).Report()
	if err != nil {
		t.Fatal(err)
	}

	nodeID := report.MakeHostNodeID(hostID)
	node, ok := rpt.Host.Nodes[nodeID]
	if !ok {
		t.Errorf("Expected host node %q, but not found", nodeID)
	}

	// Should have a bunch of expected latest keys
	for _, tuple := range []struct {
		key, want string
	}{
		{host.Timestamp, timestamp.UTC().Format(time.RFC3339Nano)},
		{host.HostName, hostname},
		{host.OS, runtime.GOOS},
		{host.Uptime, uptime},
		{host.KernelVersion, kernel},
		{report.ControlProbeID, "probe-id"},
	} {
		if have, ok := node.Latest.Lookup(tuple.key); !ok || have != tuple.want {
			t.Errorf("Expected %s %q, got %q", tuple.key, tuple.want, have)
		}
	}

	// Should have the local network
	if have, ok := node.Sets.Lookup(host.LocalNetworks); !ok || !have.Contains(network) {
		t.Errorf("Expected host.LocalNetworks to include %q, got %q", network, have)
	}

	// Should have metrics
	//for key, want := range metrics {
	//	wantSample, _ := want.LastSample()
	//	if metric, ok := node.Metrics[key]; !ok {
	//		t.Errorf("Expected %s metric, but not found", key)
	//	} else if sample, ok := metric.LastSample(); !ok {
	//		t.Errorf("Expected %s metric to have a sample, but there were none", key)
	//	} else if sample.Value != wantSample.Value {
	//		t.Errorf("Expected %s metric sample %f, got %f", key, wantSample.Value, sample.Value)
	//	}
	//}
}
