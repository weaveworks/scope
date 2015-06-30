package host

import (
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/weaveworks/scope/report"
)

// Keys for use in NodeMetadata
const (
	Timestamp     = "ts"
	HostName      = "host_name"
	LocalNetworks = "local_networks"
	OS            = "os"
	Load          = "load"
	KernelVersion = "kernel_version"
	Uptime        = "uptime"
)

// Exposed for testing
const (
	ProcUptime = "/proc/uptime"
	ProcLoad   = "/proc/loadavg"
)

// Exposed for testing
var (
	InterfaceAddrs = net.InterfaceAddrs
	Now            = func() string { return time.Now().UTC().Format(time.RFC3339Nano) }
)

// Reporter generates Reports containing the host topology.
type Reporter struct {
	hostID   string
	hostName string
}

// NewReporter returns a Reporter which produces a report containing host
// topology for this host.
func NewReporter(hostID, hostName string) *Reporter {
	return &Reporter{
		hostID:   hostID,
		hostName: hostName,
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	var (
		rep        = report.MakeReport()
		localCIDRs []string
	)

	localNets, err := InterfaceAddrs()
	if err != nil {
		return rep, err
	}
	for _, localNet := range localNets {
		// Not all networks are IP networks.
		if ipNet, ok := localNet.(*net.IPNet); ok {
			localCIDRs = append(localCIDRs, ipNet.String())
		}
	}

	uptime, err := GetUptime()
	if err != nil {
		return rep, err
	}

	kernel, err := GetKernelVersion()
	if err != nil {
		return rep, err
	}

	rep.Host.NodeMetadatas[report.MakeHostNodeID(r.hostID)] = report.NodeMetadata{
		Timestamp:     Now(),
		HostName:      r.hostName,
		LocalNetworks: strings.Join(localCIDRs, " "),
		OS:            runtime.GOOS,
		Load:          GetLoad(),
		KernelVersion: kernel,
		Uptime:        uptime.String(),
	}

	return rep, nil
}
