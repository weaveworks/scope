package docker

import (
	"log"
	"sync"
	"time"

	docker_client "github.com/fsouza/go-dockerclient"
)

// Consts exported for testing.
const (
	CreateEvent  = "create"
	DestroyEvent = "destroy"
	StartEvent   = "start"
	DieEvent     = "die"
	PauseEvent   = "pause"
	UnpauseEvent = "unpause"
	endpoint     = "unix:///var/run/docker.sock"
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
}

type registry struct {
	sync.RWMutex
	quit     chan chan struct{}
	interval time.Duration
	client   Client

	containers      map[string]Container
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
}

func newDockerClient(endpoint string) (Client, error) {
	return docker_client.NewClient(endpoint)
}

// NewRegistry returns a usable Registry. Don't forget to Stop it.
func NewRegistry(interval time.Duration) (Registry, error) {
	client, err := NewDockerClientStub(endpoint)
	if err != nil {
		return nil, err
	}

	r := &registry{
		containers:      map[string]Container{},
		containersByPID: map[int]Container{},
		images:          map[string]*docker_client.APIImages{},

		client:   client,
		interval: interval,
		quit:     make(chan chan struct{}),
	}

	r.registerControls()
	go r.loop()
	return r, nil
}

// Stop stops the Docker registry's event subscriber.
func (r *registry) Stop() {
	ch := make(chan struct{})
	r.quit <- ch
	<-ch
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
	// This ensure any containers that went away inbetween calls to
	// listenForEvents don't hang around.
	r.reset()

	// Next, start listening for events.  We do this before fetching
	// the list of containers so we don't miss containers created
	// after listing but before listening for events.
	events := make(chan *docker_client.APIEvents)
	if err := r.client.AddEventListener(events); err != nil {
		log.Printf("docker registry: %s", err)
		return true
	}
	defer func() {
		if err := r.client.RemoveEventListener(events); err != nil {
			log.Printf("docker registry: %s", err)
		}
	}()

	if err := r.updateContainers(); err != nil {
		log.Printf("docker registry: %s", err)
		return true
	}

	if err := r.updateImages(); err != nil {
		log.Printf("docker registry: %s", err)
		return true
	}

	otherUpdates := time.Tick(r.interval)
	for {
		select {
		case event := <-events:
			r.handleEvent(event)

		case <-otherUpdates:
			if err := r.updateImages(); err != nil {
				log.Printf("docker registry: %s", err)
				return true
			}

		case ch := <-r.quit:
			r.Lock()
			defer r.Unlock()

			for _, c := range r.containers {
				c.StopGatheringStats()
			}
			close(ch)
			return false
		}
	}
}

func (r *registry) reset() {
	r.Lock()
	defer r.Unlock()

	for _, c := range r.containers {
		c.StopGatheringStats()
	}

	r.containers = map[string]Container{}
	r.containersByPID = map[int]Container{}
	r.images = map[string]*docker_client.APIImages{}
}

func (r *registry) updateContainers() error {
	apiContainers, err := r.client.ListContainers(docker_client.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, apiContainer := range apiContainers {
		r.updateContainerState(apiContainer.ID)
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
		r.images[image.ID] = image
	}

	return nil
}

func (r *registry) handleEvent(event *docker_client.APIEvents) {
	switch event.Status {
	case CreateEvent, StartEvent, DieEvent, DestroyEvent, PauseEvent, UnpauseEvent:
		r.updateContainerState(event.ID)
	}
}

func (r *registry) updateContainerState(containerID string) {
	r.Lock()
	defer r.Unlock()

	dockerContainer, err := r.client.InspectContainer(containerID)
	if err != nil {
		// Don't spam the logs if the container was short lived
		if _, ok := err.(*docker_client.NoSuchContainer); !ok {
			log.Printf("Error processing event for container %s: %v", containerID, err)
			return
		}

		// Container doesn't exist anymore, so lets stop and remove it
		container, ok := r.containers[containerID]
		if !ok {
			return
		}

		delete(r.containers, containerID)
		delete(r.containersByPID, container.PID())
		container.StopGatheringStats()
		return
	}

	// Container exists, ensure we have it
	c, ok := r.containers[containerID]
	if !ok {
		c = NewContainerStub(dockerContainer)
		r.containers[containerID] = c
		r.containersByPID[dockerContainer.State.Pid] = c
	} else {
		c.UpdateState(dockerContainer)
	}

	// And finally, ensure we gather stats for it
	if dockerContainer.State.Running {
		if err := c.StartGatheringStats(); err != nil {
			log.Printf("Error gather stats for container: %s", containerID)
			return
		}
	} else {
		c.StopGatheringStats()
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

	for _, container := range r.containers {
		f(container)
	}
}

func (r *registry) getContainer(id string) (Container, bool) {
	r.RLock()
	defer r.RUnlock()
	c, ok := r.containers[id]
	return c, ok
}

// WalkImages runs f on every image of running containers the registry
// knows of.  f may be run on the same image more than once.
func (r *registry) WalkImages(f func(*docker_client.APIImages)) {
	r.RLock()
	defer r.RUnlock()

	// Loop over containers so we only emit images for running containers.
	for _, container := range r.containers {
		image, ok := r.images[container.Image()]
		if ok {
			f(image)
		}
	}
}
