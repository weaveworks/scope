package host_test

import (
	"net"
	"reflect"
	"runtime"
	"syscall"
	"testing"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

const (
	procLoad   = "0.59 0.36 0.29 1/200 12187"
	procUptime = "1004143.23 1263220.30"
	release    = "release"
	version    = "version"

	network  = "192.168.0.0/16"
	hostid   = "hostid"
	now      = "now"
	hostname = "hostname"
	load     = "0.59 0.36 0.29"
	uptime   = "278h55m43s"
	kernel   = "release version"
)

func string2c(s string) [65]int8 {
	var result [65]int8
	for i, c := range s {
		result[i] = int8(c)
	}
	return result
}

func TestReporter(t *testing.T) {
	oldInterfaceAddrs, oldNow, oldReadFile, oldUname := host.InterfaceAddrs, host.Now, host.ReadFile, host.Uname
	defer func() {
		host.InterfaceAddrs, host.Now, host.ReadFile, host.Uname = oldInterfaceAddrs, oldNow, oldReadFile, oldUname
	}()

	host.InterfaceAddrs = func() ([]net.Addr, error) {
		_, ipnet, _ := net.ParseCIDR(network)
		return []net.Addr{ipnet}, nil
	}

	host.Now = func() string { return now }

	host.ReadFile = func(filename string) ([]byte, error) {
		switch filename {
		case host.ProcUptime:
			return []byte(procUptime), nil
		case host.ProcLoad:
			return []byte(procLoad), nil
		default:
			panic(filename)
		}
	}

	host.Uname = func(uts *syscall.Utsname) error {
		uts.Release = string2c(release)
		uts.Version = string2c(version)
		return nil
	}

	r := host.NewReporter(hostid, hostname)
	have, _ := r.Report()
	want := report.MakeReport()
	want.Host.NodeMetadatas[report.MakeHostNodeID(hostid)] = report.NodeMetadata{
		host.Timestamp:     now,
		host.HostName:      hostname,
		host.LocalNetworks: network,
		host.OS:            runtime.GOOS,
		host.Load:          load,
		host.Uptime:        uptime,
		host.KernelVersion: kernel,
	}

	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
