package host

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/weaveworks/common/mtime"
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
	CPUUsage      = "host_cpu_usage_percent"
	MemoryUsage   = "host_mem_usage_bytes"
	ScopeVersion  = "host_scope_version"
)

// Exposed for testing.
const (
	ProcUptime  = "/proc/uptime"
	ProcLoad    = "/proc/loadavg"
	ProcStat    = "/proc/stat"
	ProcMemInfo = "/proc/meminfo"
)

// Exposed for testing.
var (
	MetadataTemplates = report.MetadataTemplates{
		KernelVersion: {ID: KernelVersion, Label: "Kernel Version", From: report.FromLatest, Priority: 1},
		Uptime:        {ID: Uptime, Label: "Uptime", From: report.FromLatest, Priority: 2, Datatype: "duration"},
		HostName:      {ID: HostName, Label: "Hostname", From: report.FromLatest, Priority: 11},
		OS:            {ID: OS, Label: "OS", From: report.FromLatest, Priority: 12},
		LocalNetworks: {ID: LocalNetworks, Label: "Local Networks", From: report.FromSets, Priority: 13},
		ScopeVersion:  {ID: ScopeVersion, Label: "Scope Version", From: report.FromLatest, Priority: 14},
	}

	MetricTemplates = report.MetricTemplates{
		CPUUsage:    {ID: CPUUsage, Label: "CPU", Format: report.PercentFormat, Priority: 1},
		MemoryUsage: {ID: MemoryUsage, Label: "Memory", Format: report.FilesizeFormat, Priority: 2},
		Load1:       {ID: Load1, Label: "Load (1m)", Format: report.DefaultFormat, Group: "load", Priority: 11},
	}
)

// Reporter generates Reports containing the host topology.
type Reporter struct {
	sync.RWMutex
	hostID          string
	hostName        string
	probeID         string
	version         string
	pipes           controls.PipeClient
	hostShellCmd    []string
	handlerRegistry *controls.HandlerRegistry
	pipeIDToTTY     map[string]uintptr
}

// NewReporter returns a Reporter which produces a report containing host
// topology for this host.
func NewReporter(hostID, hostName, probeID, version string, pipes controls.PipeClient, handlerRegistry *controls.HandlerRegistry) *Reporter {
	r := &Reporter{
		hostID:          hostID,
		hostName:        hostName,
		probeID:         probeID,
		pipes:           pipes,
		version:         version,
		hostShellCmd:    getHostShellCmd(),
		handlerRegistry: handlerRegistry,
		pipeIDToTTY:     map[string]uintptr{},
	}
	r.registerControls()
	return r
}

// Name of this reporter, for metrics gathering
func (*Reporter) Name() string { return "Host" }

// GetLocalNetworks is exported for mocking
var GetLocalNetworks = report.GetLocalNetworks

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

	// TODO: Make sure uptime is accurate also when time travelling.
	uptime, err := GetUptime()
	if err != nil {
		return rep, err
	}

	kernelRelease, kernelVersion, err := GetKernelReleaseAndVersion()
	if err != nil {
		return rep, err
	}
	kernel := fmt.Sprintf("%s %s", kernelRelease, kernelVersion)

	rep.Host = rep.Host.WithMetadataTemplates(MetadataTemplates)
	rep.Host = rep.Host.WithMetricTemplates(MetricTemplates)

	now := mtime.Now()
	metrics := GetLoad(now)
	cpuUsage, max := GetCPUUsagePercent()
	metrics[CPUUsage] = report.MakeSingletonMetric(now, cpuUsage).WithMax(max)
	memoryUsage, max := GetMemoryUsageBytes()
	metrics[MemoryUsage] = report.MakeSingletonMetric(now, memoryUsage).WithMax(max)

	rep.Host.AddNode(
		report.MakeNodeWith(report.MakeHostNodeID(r.hostID), map[string]string{
			report.ControlProbeID: r.probeID,
			Timestamp:             mtime.Now().UTC().Format(time.RFC3339Nano),
			HostName:              r.hostName,
			OS:                    runtime.GOOS,
			KernelVersion:         kernel,
			Uptime:                strconv.Itoa(int(uptime / time.Millisecond)),
			ScopeVersion:          r.version,
		}).
			WithSets(report.MakeSets().
				Add(LocalNetworks, report.MakeStringSet(localCIDRs...)),
			).
			WithMetrics(metrics).
			WithLatestActiveControls(ExecHost),
	)

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
