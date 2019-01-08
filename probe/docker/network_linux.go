// withNetNS function requires a fix that first appeared in Go version 1.10
// +build go1.10

package docker

import (
	"fmt"
	"net"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

// Code adapted from github.com/weaveworks/weave/net/netdev.go

// Return any non-local IP addresses for processID if in a non-root namespace
func namespaceIPAddresses(processID int) ([]*net.IPNet, error) {
	// Ignore if this process is running in the root namespace
	netnsRoot, err := netns.GetFromPid(1)
	if err != nil {
		return nil, fmt.Errorf("unable to open root namespace: %s", err)
	}
	defer netnsRoot.Close()
	netnsContainer, err := netns.GetFromPid(processID)
	if err != nil {
		return nil, err
	}
	defer netnsContainer.Close()
	if netnsRoot.Equal(netnsContainer) {
		return nil, nil
	}

	var cidrs []*net.IPNet
	err = withNetNS(netnsContainer, func() error {
		cidrs, err = allNonLocalAddresses()
		return err
	})

	return cidrs, err
}

// return all non-local IP addresses from the current namespace
func allNonLocalAddresses() ([]*net.IPNet, error) {
	var cidrs []*net.IPNet

	addrs, err := netlink.AddrList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		// Exclude link-local ipv6 addresses, localhost, etc. Hope this is the correct test.
		if addr.Scope == unix.RT_SCOPE_UNIVERSE {
			cidrs = append(cidrs, addr.IPNet)
		}
	}
	return cidrs, nil
}

// Run the 'work' function in a different network namespace
func withNetNS(ns netns.NsHandle, work func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	oldNs, err := netns.Get()
	if err == nil {
		defer oldNs.Close()

		err = netns.Set(ns)
		if err == nil {
			defer netns.Set(oldNs)

			err = work()
		}
	}

	return err
}
