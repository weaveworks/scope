package host

import (
	"net"
	"runtime"
	"time"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Keys for use in Node.Latest.
const (
	Timestamp     = "ts"
	HostName      = "host_name"
	LocalNetworks = "local_networks"
	OS            = "os"
	KernelVersion = "kernel_version"
	Uptime        = "uptime"
	Load1         = "load1"
	Load5         = "load5"
	Load15        = "load15"
	CPUUsage      = "host_cpu_usage_percent"
	MemoryUsage   = "host_mem_usage_bytes"
)

// Exposed for testing.
const (
	ProcUptime  = "/proc/uptime"
	ProcLoad    = "/proc/loadavg"
	ProcStat    = "/proc/stat"
	ProcMemInfo = "/proc/meminfo"
)

// Reporter generates Reports containing the host topology.
type Reporter struct {
	hostID   string
	hostName string
	probeID  string
	pipes    controls.PipeClient
}

// NewReporter returns a Reporter which produces a report containing host
// topology for this host.
func NewReporter(hostID, hostName, probeID string, pipes controls.PipeClient) *Reporter {
	r := &Reporter{
		hostID:   hostID,
		hostName: hostName,
		probeID:  probeID,
		pipes:    pipes,
	}
	r.registerControls()
	return r
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Host" }

// GetLocalNetworks is exported for mocking
var GetLocalNetworks = func() ([]*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	localNets := report.Networks{}
	for _, addr := range addrs {
		// Not all addrs are IPNets.
		if ipNet, ok := addr.(*net.IPNet); ok {
			localNets = append(localNets, ipNet)
		}
	}
	return localNets, nil
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	var (
		rep        = report.MakeReport()
		localCIDRs []string
	)

	localNets, err := GetLocalNetworks()
	if err != nil {
		return rep, nil
	}
	for _, localNet := range localNets {
		localCIDRs = append(localCIDRs, localNet.String())
	}

	uptime, err := GetUptime()
	if err != nil {
		return rep, err
	}

	kernel, err := GetKernelVersion()
	if err != nil {
		return rep, err
	}

	now := mtime.Now()
	metrics := GetLoad(now)
	cpuUsage, max := GetCPUUsagePercent()
	metrics[CPUUsage] = report.MakeMetric().Add(now, cpuUsage).WithMax(max)
	memoryUsage, max := GetMemoryUsageBytes()
	metrics[MemoryUsage] = report.MakeMetric().Add(now, memoryUsage).WithMax(max)

	metadata := map[string]string{report.ControlProbeID: r.probeID}
	rep.Host.AddNode(report.MakeHostNodeID(r.hostID), report.MakeNodeWith(map[string]string{
		Timestamp:     mtime.Now().UTC().Format(time.RFC3339Nano),
		HostName:      r.hostName,
		OS:            runtime.GOOS,
		KernelVersion: kernel,
		Uptime:        uptime.String(),
	}).WithSets(report.EmptySets.
		Add(LocalNetworks, report.MakeStringSet(localCIDRs...)),
	).WithMetrics(metrics).WithControls(ExecHost).WithLatests(metadata))

	rep.Host.Controls.AddControl(report.Control{
		ID:    ExecHost,
		Human: "Exec shell",
		Icon:  "fa-terminal",
	})

	return rep, nil
}

// Stop stops the reporter.
func (r *Reporter) Stop() {
	r.deregisterControls()
}
