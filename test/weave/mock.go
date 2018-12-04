package weave

import (
	"net"

	"github.com/weaveworks/scope/common/weave"
)

// Constants used for testing
const (
	MockWeavePeerName      = "winnebago"
	MockWeavePeerNickName  = "winny"
	MockWeaveDefaultSubnet = "10.32.0.1/12"
	MockContainerID        = "83183a667c01"
	MockHostname           = "hostname.weave.local"
	MockProxyAddress       = "unix:///foo/bar/weave.sock"
	MockDriverName         = "weave_mock"
)

// MockClient is a mock version of weave.Client
type MockClient struct{}

// Status implements weave.Client
func (MockClient) Status() (weave.Status, error) {
	return weave.Status{
		Router: weave.Router{
			Name: MockWeavePeerName,
			Peers: []weave.Peer{
				{
					Name:     MockWeavePeerName,
					NickName: MockWeavePeerNickName,
				},
			},
		},
		DNS: &weave.DNS{
			Entries: []struct {
				Hostname    string
				ContainerID string
				Tombstone   int64
			}{
				{
					Hostname:    MockHostname + ".",
					ContainerID: MockContainerID,
					Tombstone:   0,
				},
			},
		},
		IPAM: &weave.IPAM{
			DefaultSubnet: MockWeaveDefaultSubnet,
			Entries: []struct {
				Size        uint32
				IsKnownPeer bool
			}{
				{Size: 0, IsKnownPeer: false},
			},
		},
		Proxy: &weave.Proxy{
			Addresses: []string{MockProxyAddress},
		},
		Plugin: &weave.Plugin{
			DriverName: MockDriverName,
		},
	}, nil
}

// AddDNSEntry implements weave.Client
func (MockClient) AddDNSEntry(fqdn, containerid string, ip net.IP) error {
	return nil
}

// Expose implements weave.Client
func (MockClient) Expose() error {
	return nil
}
