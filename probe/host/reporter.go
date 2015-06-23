package host

import (
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/scope/probe/tag"
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
	ReadFile       = ioutil.ReadFile
)

type reporter struct {
	hostID   string
	hostName string
}

// NewReporter returns a Reporter which produces a report containing host
// topology for this host.
func NewReporter(hostID, hostName string) tag.Reporter {
	return &reporter{
		hostID:   hostID,
		hostName: hostName,
	}
}

func getUptime() (time.Duration, error) {
	var result time.Duration

	buf, err := ReadFile(ProcUptime)
	if err != nil {
		return result, err
	}

	fields := strings.Fields(string(buf))
	if len(fields) != 2 {
		return result, fmt.Errorf("invalid format: %s", string(buf))
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return result, err
	}

	return time.Duration(uptime) * time.Second, nil
}

func (r *reporter) Report() (report.Report, error) {
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

	uptime, err := getUptime()
	if err != nil {
		return rep, err
	}

	kernel, err := getKernelVersion()
	if err != nil {
		return rep, err
	}

	rep.Host.NodeMetadatas[report.MakeHostNodeID(r.hostID)] = report.NodeMetadata{
		Timestamp:     Now(),
		HostName:      r.hostName,
		LocalNetworks: strings.Join(localCIDRs, " "),
		OS:            runtime.GOOS,
		Load:          getLoad(),
		KernelVersion: kernel,
		Uptime:        uptime.String(),
	}

	return rep, nil
}
