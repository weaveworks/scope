package host

import (
	"runtime"
	"time"

	"github.com/weaveworks/scope/report"
)

// Keys for use in Node.Metadata.
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
)

// Exposed for testing.
const (
	ProcUptime = "/proc/uptime"
	ProcLoad   = "/proc/loadavg"
)

// Exposed for testing.
var (
	Now = func() string { return time.Now().UTC().Format(time.RFC3339Nano) }
)

// Reporter generates Reports containing the host topology.
type Reporter struct {
	hostID    string
	hostName  string
	localNets report.Networks
}

// NewReporter returns a Reporter which produces a report containing host
// topology for this host.
func NewReporter(hostID, hostName string, localNets report.Networks) *Reporter {
	return &Reporter{
		hostID:    hostID,
		hostName:  hostName,
		localNets: localNets,
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	var (
		rep        = report.MakeReport()
		localCIDRs []string
	)

	for _, localNet := range r.localNets {
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

	rep.Host.AddNode(report.MakeHostNodeID(r.hostID), report.MakeNodeWith(map[string]string{
		Timestamp:     Now(),
		HostName:      r.hostName,
		OS:            runtime.GOOS,
		KernelVersion: kernel,
		Uptime:        uptime.String(),
	}).WithSets(report.Sets{
		LocalNetworks: report.MakeStringSet(localCIDRs...),
	}).WithMetrics(GetLoad()))

	return rep, nil
}
