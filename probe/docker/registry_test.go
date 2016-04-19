package docker_test

import (
	"fmt"
	"net"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockContainer struct {
	c *client.Container
}

func (c *mockContainer) UpdateState(_ *client.Container) {}

func (c *mockContainer) ID() string {
	return c.c.ID
}

func (c *mockContainer) PID() int {
	return c.c.State.Pid
}

func (c *mockContainer) Image() string {
	return c.c.Image
}

func (c *mockContainer) Hostname() string {
	return ""
}

func (c *mockContainer) State() string {
	return "Up 3 minutes"
}

func (c *mockContainer) StateString() string {
	return docker.StateRunning
}

func (c *mockContainer) StartGatheringStats() error {
	return nil
}

func (c *mockContainer) StopGatheringStats() {}

func (c *mockContainer) GetNode(_ string, _ []net.IP) report.Node {
	return report.MakeNodeWith(report.MakeContainerNodeID(c.c.ID), map[string]string{
		docker.ContainerID:   c.c.ID,
		docker.ContainerName: c.c.Name,
		docker.ImageID:       c.c.Image,
	}).WithParents(report.EmptySets.
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.c.Image))),
	)
}

func (c *mockContainer) Container() *client.Container {
	return c.c
}

func (c *mockContainer) HasTTY() bool { return true }

type mockDockerClient struct {
	sync.RWMutex
	apiContainers []client.APIContainers
	containers    map[string]*client.Container
	apiImages     []client.APIImages
	events        []chan<- *client.APIEvents
}

func (m *mockDockerClient) ListContainers(client.ListContainersOptions) ([]client.APIContainers, error) {
	m.RLock()
	defer m.RUnlock()
	return m.apiContainers, nil
}

func (m *mockDockerClient) InspectContainer(id string) (*client.Container, error) {
	m.RLock()
	defer m.RUnlock()
	c, ok := m.containers[id]
	if !ok {
		return nil, &client.NoSuchContainer{}
	}
	return c, nil
}

func (m *mockDockerClient) ListImages(client.ListImagesOptions) ([]client.APIImages, error) {
	m.RLock()
	defer m.RUnlock()
	return m.apiImages, nil
}

func (m *mockDockerClient) AddEventListener(events chan<- *client.APIEvents) error {
	m.Lock()
	defer m.Unlock()
	m.events = append(m.events, events)
	return nil
}

func (m *mockDockerClient) RemoveEventListener(events chan *client.APIEvents) error {
	m.Lock()
	defer m.Unlock()
	for i, c := range m.events {
		if c == events {
			m.events = append(m.events[:i], m.events[i+1:]...)
		}
	}
	return nil
}

func (m *mockDockerClient) StartContainer(_ string, _ *client.HostConfig) error {
	return fmt.Errorf("started")
}

func (m *mockDockerClient) StopContainer(_ string, _ uint) error {
	return fmt.Errorf("stopped")
}

func (m *mockDockerClient) RestartContainer(_ string, _ uint) error {
	return fmt.Errorf("restarted")
}

func (m *mockDockerClient) PauseContainer(_ string) error {
	return fmt.Errorf("paused")
}

func (m *mockDockerClient) UnpauseContainer(_ string) error {
	return fmt.Errorf("unpaused")
}

type mockCloseWaiter struct{}

func (mockCloseWaiter) Close() error { return nil }
func (mockCloseWaiter) Wait() error  { return nil }

func (m *mockDockerClient) AttachToContainerNonBlocking(_ client.AttachToContainerOptions) (client.CloseWaiter, error) {
	return mockCloseWaiter{}, nil
}

func (m *mockDockerClient) CreateExec(client.CreateExecOptions) (*client.Exec, error) {
	return &client.Exec{ID: "id"}, nil
}

func (m *mockDockerClient) StartExecNonBlocking(string, client.StartExecOptions) (client.CloseWaiter, error) {
	return mockCloseWaiter{}, nil
}

func (m *mockDockerClient) send(event *client.APIEvents) {
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.events {
		c <- event
	}
}

var (
	startTime  = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	container1 = &client.Container{
		ID:    "ping",
		Name:  "pong",
		Image: "baz",
		State: client.State{
			Pid:       2,
			Running:   true,
			StartedAt: startTime,
		},
		NetworkSettings: &client.NetworkSettings{
			IPAddress: "1.2.3.4",
			Ports: map[client.Port][]client.PortBinding{
				client.Port("80/tcp"): {
					{
						HostIP:   "1.2.3.4",
						HostPort: "80",
					},
				},
				client.Port("81/tcp"): {},
			},
		},
		Config: &client.Config{
			Labels: map[string]string{
				"foo1": "bar1",
				"foo2": "bar2",
			},
		},
	}
	container2 = &client.Container{
		ID:    "wiff",
		Name:  "waff",
		Image: "baz",
		State: client.State{Pid: 1, Running: true},
		Config: &client.Config{
			Labels: map[string]string{
				"foo1": "bar1",
				"foo2": "bar2",
			},
		},
	}
	apiContainer1 = client.APIContainers{ID: "ping"}
	apiContainer2 = client.APIContainers{ID: "wiff"}
	apiImage1     = client.APIImages{ID: "baz", RepoTags: []string{"bang", "not-chosen"}}
)

func newMockClient() *mockDockerClient {
	return &mockDockerClient{
		apiContainers: []client.APIContainers{apiContainer1},
		containers:    map[string]*client.Container{"ping": container1},
		apiImages:     []client.APIImages{apiImage1},
	}
}

func setupStubs(mdc *mockDockerClient, f func()) {
	oldDockerClient, oldNewContainer := docker.NewDockerClientStub, docker.NewContainerStub
	defer func() { docker.NewDockerClientStub, docker.NewContainerStub = oldDockerClient, oldNewContainer }()

	docker.NewDockerClientStub = func(endpoint string) (docker.Client, error) {
		return mdc, nil
	}

	docker.NewContainerStub = func(c *client.Container) docker.Container {
		return &mockContainer{c}
	}

	f()
}

type containers []docker.Container

func (c containers) Len() int           { return len(c) }
func (c containers) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c containers) Less(i, j int) bool { return c[i].ID() < c[j].ID() }

func allContainers(r docker.Registry) []docker.Container {
	result := []docker.Container{}
	r.WalkContainers(func(c docker.Container) {
		result = append(result, c)
	})
	sort.Sort(containers(result))
	return result
}

func allImages(r docker.Registry) []*client.APIImages {
	result := []*client.APIImages{}
	r.WalkImages(func(i *client.APIImages) {
		result = append(result, i)
	})
	return result
}

func TestRegistry(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry, _ := docker.NewRegistry(10*time.Second, nil, true)
		defer registry.Stop()
		runtime.Gosched()

		{
			want := []docker.Container{&mockContainer{container1}}
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		{
			want := []*client.APIImages{&apiImage1}
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allImages(registry)
			})
		}
	})
}

func TestLookupByPID(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry, _ := docker.NewRegistry(10*time.Second, nil, true)
		defer registry.Stop()

		want := docker.Container(&mockContainer{container1})
		test.Poll(t, 100*time.Millisecond, want, func() interface{} {
			var have docker.Container
			registry.LockedPIDLookup(func(lookup func(int) docker.Container) {
				have = lookup(2)
			})
			return have
		})
	})
}

func TestRegistryEvents(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry, _ := docker.NewRegistry(10*time.Second, nil, true)
		defer registry.Stop()
		runtime.Gosched()

		check := func(want []docker.Container) {
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		{
			mdc.Lock()
			mdc.apiContainers = []client.APIContainers{apiContainer1, apiContainer2}
			mdc.containers["wiff"] = container2
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.StartEvent, ID: "wiff"})
			runtime.Gosched()

			want := []docker.Container{&mockContainer{container1}, &mockContainer{container2}}
			check(want)
		}

		{
			mdc.Lock()
			mdc.apiContainers = []client.APIContainers{apiContainer1}
			delete(mdc.containers, "wiff")
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.DestroyEvent, ID: "wiff"})
			runtime.Gosched()

			want := []docker.Container{&mockContainer{container1}}
			check(want)
		}

		{
			mdc.Lock()
			mdc.apiContainers = []client.APIContainers{}
			delete(mdc.containers, "ping")
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.DieEvent, ID: "ping"})
			runtime.Gosched()

			want := []docker.Container{}
			check(want)
		}

		{
			mdc.send(&client.APIEvents{Status: docker.DieEvent, ID: "doesntexist"})
			runtime.Gosched()

			want := []docker.Container{}
			check(want)
		}
	})
}
