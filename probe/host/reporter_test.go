package host_test

import (
	"net"
	"runtime"
	"testing"
	"time"

	"$GITHUB_URI/common/mtime"
	"$GITHUB_URI/probe/host"
	"$GITHUB_URI/report"
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
			host.Load1:       report.MakeMetric().Add(timestamp, 1.0),
			host.CPUUsage:    report.MakeMetric().Add(timestamp, 30.0).WithMax(100.0),
			host.MemoryUsage: report.MakeMetric().Add(timestamp, 60.0).WithMax(100.0),
		}
		uptime      = "278h55m43s"
		kernel      = "release version"
		_, ipnet, _ = net.ParseCIDR(network)
	)

	mtime.NowForce(timestamp)
	defer mtime.NowReset()

	var (
		oldGetKernelVersion    = host.GetKernelVersion
		oldGetLoad             = host.GetLoad
		oldGetUptime           = host.GetUptime
		oldGetCPUUsagePercent  = host.GetCPUUsagePercent
		oldGetMemoryUsageBytes = host.GetMemoryUsageBytes
		oldGetLocalNetworks    = host.GetLocalNetworks
	)
	defer func() {
		host.GetKernelVersion = oldGetKernelVersion
		host.GetLoad = oldGetLoad
		host.GetUptime = oldGetUptime
		host.GetCPUUsagePercent = oldGetCPUUsagePercent
		host.GetMemoryUsageBytes = oldGetMemoryUsageBytes
		host.GetLocalNetworks = oldGetLocalNetworks
	}()
	host.GetKernelVersion = func() (string, error) { return release + " " + version, nil }
	host.GetLoad = func(time.Time) report.Metrics { return metrics }
	host.GetUptime = func() (time.Duration, error) { return time.ParseDuration(uptime) }
	host.GetCPUUsagePercent = func() (float64, float64) { return 30.0, 100.0 }
	host.GetMemoryUsageBytes = func() (float64, float64) { return 60.0, 100.0 }
	host.GetLocalNetworks = func() ([]*net.IPNet, error) { return []*net.IPNet{ipnet}, nil }

	rpt, err := host.NewReporter(hostID, hostname, "", "", nil).Report()
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
	for key, want := range metrics {
		wantSample := want.LastSample()
		if metric, ok := node.Metrics[key]; !ok {
			t.Errorf("Expected %s metric, but not found", key)
		} else if sample := metric.LastSample(); sample == nil {
			t.Errorf("Expected %s metric to have a sample, but there were none", key)
		} else if sample.Value != wantSample.Value {
			t.Errorf("Expected %s metric sample %f, got %f", key, wantSample, sample.Value)
		}
	}
}
