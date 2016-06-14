package docker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-radix"
	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Consts exported for testing.
const (
	CreateEvent            = "create"
	DestroyEvent           = "destroy"
	RenameEvent            = "rename"
	StartEvent             = "start"
	DieEvent               = "die"
	PauseEvent             = "pause"
	UnpauseEvent           = "unpause"
	NetworkConnectEvent    = "network:connect"
	NetworkDisconnectEvent = "network:disconnect"
	endpoint               = "unix:///var/run/docker.sock"
)

// Vars exported for testing.
var (
	NewDockerClientStub = newDockerClient
	NewContainerStub    = NewContainer
)

// Registry keeps track of running docker containers and their images
type Registry interface {
	Stop()
	LockedPIDLookup(f func(func(int) Container))
	WalkContainers(f func(Container))
	WalkImages(f func(*docker_client.APIImages))
	WatchContainerUpdates(ContainerUpdateWatcher)
	GetContainer(string) (Container, bool)
	GetContainerByPrefix(string) (Container, bool)
	GetContainerImage(string) (*docker_client.APIImages, bool)
}

// ContainerUpdateWatcher is the type of functions that get called when containers are updated.
type ContainerUpdateWatcher func(report.Node)

type registry struct {
	sync.RWMutex
	quit         chan chan struct{}
	interval     time.Duration
	collectStats bool
	client       Client
	pipes        controls.PipeClient
	hostID       string

	watchers        []ContainerUpdateWatcher
	containers      *radix.Tree
	containersByPID map[int]Container
	images          map[string]*docker_client.APIImages
}

// Client interface for mocking.
type Client interface {
	ListContainers(docker_client.ListContainersOptions) ([]docker_client.APIContainers, error)
	InspectContainer(string) (*docker_client.Container, error)
	ListImages(docker_client.ListImagesOptions) ([]docker_client.APIImages, error)
	AddEventListener(chan<- *docker_client.APIEvents) error
	RemoveEventListener(chan *docker_client.APIEvents) error

	StopContainer(string, uint) error
	StartContainer(string, *docker_client.HostConfig) error
	RestartContainer(string, uint) error
	PauseContainer(string) error
	UnpauseContainer(string) error
	RemoveContainer(docker_client.RemoveContainerOptions) error
	AttachToContainerNonBlocking(docker_client.AttachToContainerOptions) (docker_client.CloseWaiter, error)
	CreateExec(docker_client.CreateExecOptions) (*docker_client.Exec, error)
	StartExecNonBlocking(string, docker_client.StartExecOptions) (docker_client.CloseWaiter, error)
}

func newDockerClient(endpoint string) (Client, error) {
	return docker_client.NewClient(endpoint)
}

// NewRegistry returns a usable Registry. Don't forget to Stop it.
func NewRegistry(interval time.Duration, pipes controls.PipeClient, collectStats bool, hostID string) (Registry, error) {
	client, err := NewDockerClientStub(endpoint)
	if err != nil {
		return nil, err
	}

	r := &registry{
		containers:      radix.New(),
		containersByPID: map[int]Container{},
		images:          map[string]*docker_client.APIImages{},

		client:       client,
		pipes:        pipes,
		interval:     interval,
		collectStats: collectStats,
		hostID:       hostID,
		quit:         make(chan chan struct{}),
	}

	r.registerControls()
	go r.loop()
	return r, nil
}

// Stop stops the Docker registry's event subscriber.
func (r *registry) Stop() {
	r.deregisterControls()
	ch := make(chan struct{})
	r.quit <- ch
	<-ch
}

// WatchContainerUpdates registers a callback to be called
// whenever a container is updated.
func (r *registry) WatchContainerUpdates(f ContainerUpdateWatcher) {
	r.Lock()
	defer r.Unlock()
	r.watchers = append(r.watchers, f)
}

func (r *registry) loop() {
	for {
		// NB listenForEvents blocks.
		// Returning false means we should exit.
		if !r.listenForEvents() {
			return
		}

		// Sleep here so we don't hammer the
		// logs if docker is down
		time.Sleep(r.interval)
	}
}

func (r *registry) listenForEvents() bool {
	// First we empty the store lists.
	// This ensure any containers that went away in between calls to
	// listenForEvents don't hang around.
	r.reset()

	// Next, start listening for events.  We do this before fetching
	// the list of containers so we don't miss containers created
	// after listing but before listening for events.
	events := make(chan *docker_client.APIEvents)
	if err := r.client.AddEventListener(events); err != nil {
		log.Errorf("docker registry: %s", err)
		return true
	}
	defer func() {
		if err := r.client.RemoveEventListener(events); err != nil {
			log.Errorf("docker registry: %s", err)
		}
	}()

	if err := r.updateContainers(); err != nil {
		log.Errorf("docker registry: %s", err)
		return true
	}

	if err := r.updateImages(); err != nil {
		log.Errorf("docker registry: %s", err)
		return true
	}

	otherUpdates := time.Tick(r.interval)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				log.Errorf("docker registry: event listener unexpectedly disconnected")
				return true
			}
			r.handleEvent(event)

		case <-otherUpdates:
			if err := r.updateImages(); err != nil {
				log.Errorf("docker registry: %s", err)
				return true
			}

		case ch := <-r.quit:
			r.Lock()
			defer r.Unlock()

			if r.collectStats {
				r.containers.Walk(func(_ string, c interface{}) bool {
					c.(Container).StopGatheringStats()
					return false
				})
			}
			close(ch)
			return false
		}
	}
}

func (r *registry) reset() {
	r.Lock()
	defer r.Unlock()

	if r.collectStats {
		r.containers.Walk(func(_ string, c interface{}) bool {
			c.(Container).StopGatheringStats()
			return false
		})
	}

	r.containers = radix.New()
	r.containersByPID = map[int]Container{}
	r.images = map[string]*docker_client.APIImages{}
}

func (r *registry) updateContainers() error {
	apiContainers, err := r.client.ListContainers(docker_client.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, apiContainer := range apiContainers {
		r.updateContainerState(apiContainer.ID, nil)
	}

	return nil
}

func (r *registry) updateImages() error {
	images, err := r.client.ListImages(docker_client.ListImagesOptions{})
	if err != nil {
		return err
	}

	r.Lock()
	defer r.Unlock()

	for i := range images {
		image := &images[i]
		r.images[trimImageID(image.ID)] = image
	}

	return nil
}

func (r *registry) handleEvent(event *docker_client.APIEvents) {
	switch event.Status {
	case CreateEvent, RenameEvent, StartEvent, DieEvent, DestroyEvent, PauseEvent, UnpauseEvent, NetworkConnectEvent, NetworkDisconnectEvent:
		r.updateContainerState(event.ID, stateAfterEvent(event.Status))
	}
}

func stateAfterEvent(event string) *string {
	switch event {
	case DestroyEvent:
		return &StateDeleted
	default:
		return nil
	}
}

func (r *registry) updateContainerState(containerID string, intendedState *string) {
	r.Lock()
	defer r.Unlock()

	dockerContainer, err := r.client.InspectContainer(containerID)
	if err != nil {
		// Don't spam the logs if the container was short lived
		if _, ok := err.(*docker_client.NoSuchContainer); !ok {
			log.Errorf("Error processing event for container %s: %v", containerID, err)
			return
		}

		// Container doesn't exist anymore, so lets stop and remove it
		c, ok := r.containers.Get(containerID)
		if !ok {
			return
		}
		container := c.(Container)

		r.containers.Delete(containerID)
		delete(r.containersByPID, container.PID())
		if r.collectStats {
			container.StopGatheringStats()
		}

		if intendedState != nil {
			node := report.MakeNodeWith(report.MakeContainerNodeID(containerID), map[string]string{
				ContainerID:    containerID,
				ContainerState: *intendedState,
			})
			// Trigger anyone watching for updates
			for _, f := range r.watchers {
				f(node)
			}
		}
		return
	}

	// Container exists, ensure we have it
	o, ok := r.containers.Get(containerID)
	var c Container
	if !ok {
		c = NewContainerStub(dockerContainer, r.hostID)
		r.containers.Insert(containerID, c)
	} else {
		c = o.(Container)
		// potentially remove existing pid mapping.
		delete(r.containersByPID, c.PID())
		c.UpdateState(dockerContainer)
	}

	// Update PID index
	if c.PID() > 1 {
		r.containersByPID[c.PID()] = c
	}

	// Trigger anyone watching for updates
	if err != nil {
		node := c.GetNode()
		for _, f := range r.watchers {
			f(node)
		}
	}

	// And finally, ensure we gather stats for it
	if r.collectStats {
		if dockerContainer.State.Running {
			if err := c.StartGatheringStats(); err != nil {
				log.Errorf("Error gather stats for container: %s", containerID)
				return
			}
		} else {
			c.StopGatheringStats()
		}
	}
}

// LockedPIDLookup runs f under a read lock, and gives f a function for
// use doing pid->container lookups.
func (r *registry) LockedPIDLookup(f func(func(int) Container)) {
	r.RLock()
	defer r.RUnlock()

	lookup := func(pid int) Container {
		return r.containersByPID[pid]
	}

	f(lookup)
}

// WalkContainers runs f on every running containers the registry knows of.
func (r *registry) WalkContainers(f func(Container)) {
	r.RLock()
	defer r.RUnlock()

	r.containers.Walk(func(_ string, c interface{}) bool {
		f(c.(Container))
		return false
	})
}

func (r *registry) GetContainer(id string) (Container, bool) {
	r.RLock()
	defer r.RUnlock()
	c, ok := r.containers.Get(id)
	if ok {
		return c.(Container), true
	}
	return nil, false
}

func (r *registry) GetContainerByPrefix(prefix string) (Container, bool) {
	r.RLock()
	defer r.RUnlock()
	out := []interface{}{}
	r.containers.WalkPrefix(prefix, func(_ string, v interface{}) bool {
		out = append(out, v)
		return false
	})
	if len(out) == 1 {
		return out[0].(Container), true
	}
	return nil, false
}

func (r *registry) GetContainerImage(id string) (*docker_client.APIImages, bool) {
	r.RLock()
	defer r.RUnlock()
	image, ok := r.images[id]
	return image, ok
}

// WalkImages runs f on every image of running containers the registry
// knows of.  f may be run on the same image more than once.
func (r *registry) WalkImages(f func(*docker_client.APIImages)) {
	r.RLock()
	defer r.RUnlock()

	// Loop over containers so we only emit images for running containers.
	r.containers.Walk(func(_ string, c interface{}) bool {
		image, ok := r.images[c.(Container).Image()]
		if ok {
			f(image)
		}
		return false
	})
}

// ImageNameWithoutVersion splits the image name apart, returning the name
// without the version, if possible
func ImageNameWithoutVersion(name string) string {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) == 3 {
		name = fmt.Sprintf("%s/%s", parts[1], parts[2])
	}
	parts = strings.SplitN(name, ":", 2)
	return parts[0]
}
