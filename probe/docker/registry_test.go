package docker_test

import (
	"reflect"
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

func (c *mockContainer) ID() string {
	return c.c.ID
}

func (c *mockContainer) PID() int {
	return c.c.State.Pid
}

func (c *mockContainer) Image() string {
	return c.c.Image
}

func (c *mockContainer) StartGatheringStats() error {
	return nil
}

func (c *mockContainer) StopGatheringStats() {}

func (c *mockContainer) GetNodeMetadata() report.NodeMetadata {
	return report.NewNodeMetadata(report.Metadata{
		docker.ContainerID:   c.c.ID,
		docker.ContainerName: c.c.Name,
		docker.ImageID:       c.c.Image,
	})
}

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
	return m.containers[id], nil
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

func (m *mockDockerClient) send(event *client.APIEvents) {
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.events {
		c <- event
	}
}

var (
	container1 = &client.Container{
		ID:    "ping",
		Name:  "pong",
		Image: "baz",
		State: client.State{Pid: 1, Running: true},
	}
	container2 = &client.Container{
		ID:    "wiff",
		Name:  "waff",
		Image: "baz",
		State: client.State{Pid: 1, Running: true},
	}
	apiContainer1 = client.APIContainers{ID: "ping"}
	apiImage1     = client.APIImages{ID: "baz", RepoTags: []string{"bang", "not-chosen"}}
	mockClient    = mockDockerClient{
		apiContainers: []client.APIContainers{apiContainer1},
		containers:    map[string]*client.Container{"ping": container1},
		apiImages:     []client.APIImages{apiImage1},
	}
)

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
	mdc := mockClient // take a copy
	setupStubs(&mdc, func() {
		registry, _ := docker.NewRegistry(10 * time.Second)
		defer registry.Stop()
		runtime.Gosched()

		{
			want := []docker.Container{&mockContainer{container1}}
			test.Poll(t, 10*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		{
			have := allImages(registry)
			want := []*client.APIImages{&apiImage1}
			if !reflect.DeepEqual(want, have) {
				t.Errorf("%s", test.Diff(want, have))
			}
		}
	})
}

func TestRegistryEvents(t *testing.T) {
	mdc := mockClient // take a copy
	setupStubs(&mdc, func() {
		registry, _ := docker.NewRegistry(10 * time.Second)
		defer registry.Stop()
		runtime.Gosched()

		check := func(want []docker.Container) {
			test.Poll(t, 10*time.Millisecond, want, func() interface{} {
				return allContainers(registry)
			})
		}

		{
			mdc.Lock()
			mdc.containers["wiff"] = container2
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.StartEvent, ID: "wiff"})
			runtime.Gosched()

			want := []docker.Container{&mockContainer{container1}, &mockContainer{container2}}
			check(want)
		}

		{
			mdc.Lock()
			delete(mdc.containers, "wiff")
			mdc.Unlock()
			mdc.send(&client.APIEvents{Status: docker.DieEvent, ID: "wiff"})
			runtime.Gosched()

			want := []docker.Container{&mockContainer{container1}}
			check(want)
		}

		{
			mdc.Lock()
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
