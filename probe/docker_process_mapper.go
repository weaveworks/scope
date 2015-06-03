package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	stop  = "stop"
	start = "start"
)

type dockerMapper struct {
	sync.RWMutex
	quit     chan struct{}
	interval time.Duration

	containers      map[string]*docker.Container
	containersByPID map[int]*docker.Container
	images          map[string]*docker.APIImages

	procRoot string
	pidTree  *pidTree
}

func newDockerMapper(procRoot string, interval time.Duration) (*dockerMapper, error) {
	pidTree, err := newPIDTreeStub(procRoot)
	if err != nil {
		return nil, err
	}

	m := dockerMapper{
		containers:      map[string]*docker.Container{},
		containersByPID: map[int]*docker.Container{},
		images:          map[string]*docker.APIImages{},

		procRoot: procRoot,
		pidTree:  pidTree,

		interval: interval,
		quit:     make(chan struct{}),
	}

	go m.loop()
	return &m, nil
}

func (m *dockerMapper) Stop() {
	close(m.quit)
}

func (m *dockerMapper) loop() {
	if !m.update() {
		return
	}

	ticker := time.Tick(m.interval)
	for {
		select {
		case <-ticker:
			if !m.update() {
				return
			}

		case <-m.quit:
			return
		}
	}
}

// for mocking
type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	InspectContainer(string) (*docker.Container, error)
	ListImages(docker.ListImagesOptions) ([]docker.APIImages, error)
	AddEventListener(chan<- *docker.APIEvents) error
	RemoveEventListener(chan *docker.APIEvents) error
}

func newRealDockerClient(endpoint string) (dockerClient, error) {
	return docker.NewClient(endpoint)
}

var (
	newDockerClient = newRealDockerClient
	newPIDTreeStub  = newPIDTree
)

// returns false when stopping.
func (m *dockerMapper) update() bool {
	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClient(endpoint)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	events := make(chan *docker.APIEvents)
	if err := client.AddEventListener(events); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}
	defer func() {
		if err := client.RemoveEventListener(events); err != nil {
			log.Printf("docker mapper: %s", err)
		}
	}()

	if err := m.updateContainers(client); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	if err := m.updateImages(client); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	otherUpdates := time.Tick(m.interval)
	for {
		select {
		case event := <-events:
			m.handleEvent(event, client)

		case <-otherUpdates:
			if err := m.updatePIDTree(); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

			if err := m.updateImages(client); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

		case <-m.quit:
			return false
		}
	}
}

func (m *dockerMapper) updateContainers(client dockerClient) error {
	apiContainers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	containers := []*docker.Container{}
	for _, apiContainer := range apiContainers {
		container, err := client.InspectContainer(apiContainer.ID)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			continue
		}

		if !container.State.Running {
			continue
		}

		containers = append(containers, container)
	}

	m.Lock()
	for _, container := range containers {
		m.containers[container.ID] = container
		m.containersByPID[container.State.Pid] = container
	}
	m.Unlock()

	return nil
}

func (m *dockerMapper) updateImages(client dockerClient) error {
	images, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return err
	}

	m.Lock()
	for i := range images {
		image := &images[i]
		m.images[image.ID] = image
	}
	m.Unlock()

	return nil
}

func (m *dockerMapper) handleEvent(event *docker.APIEvents, client dockerClient) {
	switch event.Status {
	case stop:
		containerID := event.ID
		m.Lock()
		if container, ok := m.containers[containerID]; ok {
			delete(m.containers, containerID)
			delete(m.containersByPID, container.State.Pid)
		} else {
			log.Printf("docker mapper: container %s not found", containerID)
		}
		m.Unlock()

	case start:
		containerID := event.ID
		container, err := client.InspectContainer(containerID)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			return
		}

		if !container.State.Running {
			log.Printf("docker mapper: container %s not running", containerID)
			return
		}

		m.Lock()
		m.containers[containerID] = container
		m.containersByPID[container.State.Pid] = container
		m.Unlock()
	}
}

func (m *dockerMapper) updatePIDTree() error {
	pidTree, err := newPIDTreeStub(m.procRoot)
	if err != nil {
		return err
	}

	m.Lock()
	m.pidTree = pidTree
	m.Unlock()
	return nil
}

type dockerProcessMapper struct {
	*dockerMapper
	key string
	f   func(*docker.Container) string
}

func (m *dockerProcessMapper) Key() string { return m.key }
func (m *dockerProcessMapper) Map(pid uint) (string, error) {
	var (
		container *docker.Container
		ok        bool
		err       error
		candidate = int(pid)
	)

	m.RLock()
	for {
		container, ok = m.containersByPID[candidate]
		if ok {
			break
		}
		candidate, err = m.pidTree.getParent(candidate)
		if err != nil {
			break
		}
	}
	m.RUnlock()

	if err != nil {
		return "", fmt.Errorf("no container found for PID %d", pid)
	}

	return m.f(container), nil
}

func (m *dockerMapper) idMapper() processMapper {
	return &dockerProcessMapper{m, "docker_id", func(c *docker.Container) string {
		return c.ID
	}}
}

func (m *dockerMapper) nameMapper() processMapper {
	return &dockerProcessMapper{m, "docker_name", func(c *docker.Container) string {
		return strings.TrimPrefix(c.Name, "/")
	}}
}

func (m *dockerMapper) imageIDMapper() processMapper {
	return &dockerProcessMapper{m, "docker_image_id", func(c *docker.Container) string {
		return c.Image
	}}
}

func (m *dockerMapper) imageNameMapper() processMapper {
	return &dockerProcessMapper{m, "docker_image_name", func(c *docker.Container) string {
		m.RLock()
		image, ok := m.images[c.Image]
		m.RUnlock()

		if !ok || len(image.RepoTags) == 0 {
			return ""
		}

		return image.RepoTags[0]
	}}
}
