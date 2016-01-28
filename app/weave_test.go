package app_test

import (
	"net"
	"sync"
	"testing"
	"time"

	fsouza "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/test"
)

type mockDockerClient struct{}

func (mockDockerClient) ListContainers(fsouza.ListContainersOptions) ([]fsouza.APIContainers, error) {
	return []fsouza.APIContainers{
		{
			Names: []string{"/" + containerName},
			ID:    containerID,
		},
		{
			Names: []string{"/notme"},
			ID:    "1234abcd",
		},
	}, nil
}

type entry struct {
	containerid string
	ip          net.IP
}

type mockWeaveClient struct {
	sync.Mutex
	published map[string]entry
}

func (m *mockWeaveClient) AddDNSEntry(hostname, containerid string, ip net.IP) error {
	m.Lock()
	defer m.Unlock()
	m.published[hostname] = entry{containerid, ip}
	return nil
}

func (m *mockWeaveClient) Expose() error {
	return nil
}

const (
	hostname      = "foo.weave"
	containerName = "bar"
	containerID   = "a1b2c3d4"
)

var (
	ip = net.ParseIP("1.2.3.4")
)

func TestWeave(t *testing.T) {
	weaveClient := &mockWeaveClient{
		published: map[string]entry{},
	}
	dockerClient := mockDockerClient{}
	interfaces := func() ([]app.Interface, error) {
		return []app.Interface{
			{
				Name: "eth0",
				Addrs: []net.Addr{
					&net.IPAddr{
						IP: ip,
					},
				},
			},
			{
				Name: "docker0",
				Addrs: []net.Addr{
					&net.IPAddr{
						IP: net.ParseIP("4.3.2.1"),
					},
				},
			},
		}, nil
	}
	publisher := app.NewWeavePublisher(
		weaveClient, dockerClient, interfaces,
		hostname, containerName)
	defer publisher.Stop()

	want := map[string]entry{
		hostname: {containerID, ip},
	}
	test.Poll(t, 100*time.Millisecond, want, func() interface{} {
		weaveClient.Lock()
		defer weaveClient.Unlock()
		result := map[string]entry{}
		for k, v := range weaveClient.published {
			result[k] = v
		}
		return result
	})
}
