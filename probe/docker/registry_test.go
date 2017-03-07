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

	"github.com/weaveworks/common/mtime"
	commonTest "github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func testRegistry() docker.Registry {
	hr := controls.NewDefaultHandlerRegistry()
	registry, _ := docker.NewRegistry(docker.RegistryOptions{
		Interval:        10 * time.Second,
		CollectStats:    true,
		HandlerRegistry: hr,
	})
	return registry
}

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

func (c *mockContainer) StartGatheringStats(docker.StatsGatherer) error {
	return nil
}

func (c *mockContainer) StopGatheringStats() {}

func (c *mockContainer) GetNode() report.Node {
	return report.MakeNodeWith(report.MakeContainerNodeID(c.c.ID), map[string]string{
		docker.ContainerID:   c.c.ID,
		docker.ContainerName: c.c.Name,
		docker.ImageID:       c.c.Image,
	}).WithParents(report.EmptySets.
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.c.Image))),
	)
}

func (c *mockContainer) NetworkMode() (string, bool) {
	return "", false
}
func (c *mockContainer) NetworkInfo([]net.IP) report.Sets {
	return report.EmptySets
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
	networks      []client.Network
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

func (m *mockDockerClient) ListNetworks() ([]client.Network, error) {
	m.RLock()
	defer m.RUnlock()
	return m.networks, nil
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

func (m *mockDockerClient) RemoveContainer(_ client.RemoveContainerOptions) error {
	return fmt.Errorf("remove")
}

func (m *mockDockerClient) Stats(_ client.StatsOptions) error {
	return fmt.Errorf("stats")
}

func (m *mockDockerClient) ResizeExecTTY(id string, height, width int) error {
	return fmt.Errorf("resizeExecTTY")
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
		Path:  "ping",
		Args: []string{
			"foo.bar.local",
		},
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
			Networks: map[string]client.ContainerNetwork{
				"network1": {
					IPAddress: "5.6.7.8",
				},
			},
		},
		Config: &client.Config{
			Env: []string{
				"FOO=secret-bar",
			},
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
	renamedContainer = &client.Container{
		ID:    "renamed",
		Name:  "renamed",
		Image: "baz",
		State: client.State{Pid: 1, Running: true},
		Config: &client.Config{
			Labels: map[string]string{
				"foo1": "bar1",
				"foo2": "bar2",
			},
		},
	}
	apiContainer1       = client.APIContainers{ID: "ping"}
	apiContainer2       = client.APIContainers{ID: "wiff"}
	renamedAPIContainer = client.APIContainers{ID: "renamed"}
	apiImage1           = client.APIImages{
		ID:       "baz",
		RepoTags: []string{"bang", "not-chosen"},
		Labels: map[string]string{
			"imgfoo1": "bar1",
			"imgfoo2": "bar2",
		},
	}
	network1 = client.Network{
		ID:    "deadbeef",
		Name:  "network1",
		Scope: "local",
		IPAM: client.IPAMOptions{
			Config: []client.IPAMConfig{{Subnet: "5.6.7.8/24"}},
		},
	}
)

func newMockClient() *mockDockerClient {
	return &mockDockerClient{
		apiContainers: []client.APIContainers{apiContainer1},
		containers:    map[string]*client.Container{"ping": container1},
		apiImages:     []client.APIImages{apiImage1},
		networks:      []client.Network{network1},
	}
}

func setupStubs(mdc *mockDockerClient, f func()) {
	oldDockerClient, oldNewContainer := docker.NewDockerClientStub, docker.NewContainerStub
	defer func() { docker.NewDockerClientStub, docker.NewContainerStub = oldDockerClient, oldNewContainer }()

	docker.NewDockerClientStub = func(endpoint string) (docker.Client, error) {
		return mdc, nil
	}

	docker.NewContainerStub = func(c *client.Container, _ string, _ bool, _ bool) docker.Container {
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

func allImages(r docker.Registry) []client.APIImages {
	result := []client.APIImages{}
	r.WalkImages(func(i client.APIImages) {
		result = append(result, i)
	})
	return result
}

func allNetworks(r docker.Registry) []client.Network {
	result := []client.Network{}
	r.WalkNetworks(func(i client.Network) {
		result = append(result, i)
	})
	return result
}

func TestRegistry(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry := testRegistry()
		defer registry.Stop()
		runtime.Gosched()

		{
			want := []docker.Container{&mockContainer{container1}}
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		{
			want := []client.APIImages{apiImage1}
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allImages(registry)
			})
		}

		{
			want := []client.Network{network1}
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allNetworks(registry)
			})
		}

	})
}

func TestLookupByPID(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry := testRegistry()
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
		registry := testRegistry()
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

		{
			mdc.Lock()
			mdc.apiContainers = []client.APIContainers{renamedAPIContainer}
			mdc.containers[renamedContainer.ID] = renamedContainer
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.RenameEvent, ID: renamedContainer.ID})
			runtime.Gosched()

			want := []docker.Container{&mockContainer{renamedContainer}}
			check(want)
		}
	})
}

func TestRegistryDelete(t *testing.T) {
	mtime.NowForce(mtime.Now())
	defer mtime.NowReset()

	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry := testRegistry()
		defer registry.Stop()
		runtime.Gosched()

		// Collect all the events.
		mtx := sync.Mutex{}
		nodes := []report.Node{}
		registry.WatchContainerUpdates(func(n report.Node) {
			mtx.Lock()
			defer mtx.Unlock()
			nodes = append(nodes, n)
		})

		check := func(want []docker.Container) {
			test.Poll(t, 100*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		want := []docker.Container{&mockContainer{container1}}
		check(want)

		{
			mdc.Lock()
			mdc.apiContainers = []client.APIContainers{}
			delete(mdc.containers, "ping")
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.DestroyEvent, ID: "ping"})
			runtime.Gosched()

			check([]docker.Container{})

			mtx.Lock()
			want := []report.Node{
				report.MakeNodeWith(report.MakeContainerNodeID("ping"), map[string]string{
					docker.ContainerID:    "ping",
					docker.ContainerState: "deleted",
				}),
			}
			if !reflect.DeepEqual(want, nodes) {
				t.Errorf("Didn't get right container updates: %v", commonTest.Diff(want, nodes))
			}
			nodes = []report.Node{}
			mtx.Unlock()
		}
	})
}

func TestDockerImageName(t *testing.T) {
	for _, input := range []struct{ in, name string }{
		{"foo/bar", "foo/bar"},
		{"foo/bar:baz", "foo/bar"},
		{"reg:123/foo/bar:baz", "foo/bar"},
		{"docker-registry.domain.name:5000/repo/image1:ver", "repo/image1"},
		{"foo", "foo"},
	} {
		name := docker.ImageNameWithoutVersion(input.in)
		if name != input.name {
			t.Fatalf("%s: %s != %s", input.in, name, input.name)
		}
	}
}
