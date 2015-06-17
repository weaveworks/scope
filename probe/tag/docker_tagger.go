package tag

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
)

const (
	start    = "start"
	die      = "die"
	endpoint = "unix:///var/run/docker.sock"
)

// These constants are keys used in node metadata
// TODO: use these constants in report/{mapping.go, detailed_node.go} - pending some circular references
const (
	ContainerID   = "docker_container_id"
	ContainerName = "docker_container_name"
	ImageID       = "docker_image_id"
	ImageName     = "docker_image_name"
)

var (
	newDockerClientStub = newDockerClient
	newPIDTreeStub      = NewPIDTree
)

// DockerTagger is a tagger that tags Docker container information to process
// nodes that have a PID.
type DockerTagger struct {
	sync.RWMutex
	quit     chan struct{}
	interval time.Duration
	client   dockerClient

	containers      map[string]*dockerContainer
	containersByPID map[int]*dockerContainer
	images          map[string]*docker.APIImages

	procRoot string
	pidTree  *PIDTree
}

// Sub-interface for mocking.
type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	InspectContainer(string) (*docker.Container, error)
	ListImages(docker.ListImagesOptions) ([]docker.APIImages, error)
	AddEventListener(chan<- *docker.APIEvents) error
	RemoveEventListener(chan *docker.APIEvents) error
}

func newDockerClient(endpoint string) (dockerClient, error) {
	return docker.NewClient(endpoint)
}

// NewDockerTagger returns a usable DockerTagger. Don't forget to Stop it.
func NewDockerTagger(procRoot string, interval time.Duration) (*DockerTagger, error) {
	pidTree, err := newPIDTreeStub(procRoot)
	if err != nil {
		return nil, err
	}

	t := DockerTagger{
		containers:      map[string]*dockerContainer{},
		containersByPID: map[int]*dockerContainer{},
		images:          map[string]*docker.APIImages{},

		procRoot: procRoot,
		pidTree:  pidTree,

		interval: interval,
		quit:     make(chan struct{}),
	}

	go t.loop()
	return &t, nil
}

// Stop stops the Docker tagger's event subscriber.
func (t *DockerTagger) Stop() {
	close(t.quit)
}

func (t *DockerTagger) loop() {
	if !t.update() {
		return
	}

	ticker := time.Tick(t.interval)
	for {
		select {
		case <-ticker:
			if !t.update() {
				return
			}

		case <-t.quit:
			return
		}
	}
}

func (t *DockerTagger) update() bool {
	client, err := newDockerClientStub(endpoint)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}
	t.client = client

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

	if err := t.updateContainers(); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	if err := t.updateImages(); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	otherUpdates := time.Tick(t.interval)
	for {
		select {
		case event := <-events:
			t.handleEvent(event)

		case <-otherUpdates:
			if err := t.updatePIDTree(); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

			if err := t.updateImages(); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

		case <-t.quit:
			return false
		}
	}
}

func (t *DockerTagger) updateContainers() error {
	apiContainers, err := t.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, apiContainer := range apiContainers {
		if err := t.addContainer(apiContainer.ID); err != nil {
			log.Printf("docker mapper: %s", err)
		}
	}

	return nil
}

func (t *DockerTagger) updateImages() error {
	images, err := t.client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return err
	}

	t.Lock()
	for i := range images {
		image := &images[i]
		t.images[image.ID] = image
	}
	t.Unlock()

	return nil
}

func (t *DockerTagger) handleEvent(event *docker.APIEvents) {
	switch event.Status {
	case die:
		containerID := event.ID
		t.removeContainer(containerID)

	case start:
		containerID := event.ID
		if err := t.addContainer(containerID); err != nil {
			log.Printf("docker mapper: %s", err)
		}
	}
}

func (t *DockerTagger) updatePIDTree() error {
	pidTree, err := newPIDTreeStub(t.procRoot)
	if err != nil {
		return err
	}

	t.Lock()
	t.pidTree = pidTree
	t.Unlock()
	return nil
}

func (t *DockerTagger) addContainer(containerID string) error {
	container, err := t.client.InspectContainer(containerID)
	if err != nil {
		// Don't spam the logs if the container was short lived
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			return err
		}
		return nil
	}

	if !container.State.Running {
		return fmt.Errorf("docker mapper: container %s not running", containerID)
	}

	t.Lock()
	defer t.Unlock()

	dockerContainer := &dockerContainer{Container: container}

	t.containers[containerID] = dockerContainer
	t.containersByPID[container.State.Pid] = dockerContainer

	return dockerContainer.startGatheringStats(containerID)
}

func (t *DockerTagger) removeContainer(containerID string) {
	t.Lock()
	defer t.Unlock()

	container, ok := t.containers[containerID]
	if !ok {
		return
	}

	delete(t.containers, containerID)
	delete(t.containersByPID, container.State.Pid)
	container.stopGatheringStats(containerID)
}

// Containers returns the Containers the DockerTagger knows about.
func (t *DockerTagger) Containers() []*docker.Container {
	containers := []*docker.Container{}

	t.RLock()
	for _, container := range t.containers {
		containers = append(containers, container.Container)
	}
	t.RUnlock()

	return containers
}

// Tag implements Tagger.
func (t *DockerTagger) Tag(r report.Report) report.Report {
	for nodeID, nodeMetadata := range r.Process.NodeMetadatas {
		pidStr, ok := nodeMetadata["pid"]
		if !ok {
			//log.Printf("dockerTagger: %q: no process node ID", id)
			continue
		}
		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err != nil {
			//log.Printf("dockerTagger: %q: bad process node PID (%v)", id, err)
			continue
		}

		var (
			container *dockerContainer
			candidate = int(pid)
		)

		t.RLock()
		for {
			container, ok = t.containersByPID[candidate]
			if ok {
				break
			}
			candidate, err = t.pidTree.getParent(candidate)
			if err != nil {
				break
			}
		}
		t.RUnlock()

		if !ok {
			continue
		}

		md := report.NodeMetadata{
			ContainerID: container.ID,
			ImageID:     container.Image,
		}

		t.RLock()
		image, ok := t.images[container.Image]
		t.RUnlock()

		if ok && len(image.RepoTags) > 0 {
			md[ImageName] = image.RepoTags[0]
		}

		r.Process.NodeMetadatas[nodeID].Merge(md)
	}

	return r
}

// ContainerTopology produces a Toplogy of Containers
func (t *DockerTagger) ContainerTopology(scope string) report.Topology {
	t.RLock()
	defer t.RUnlock()

	result := report.NewTopology()
	for _, container := range t.containers {
		nmd := report.NodeMetadata{
			ContainerID:   container.ID,
			ContainerName: strings.TrimPrefix(container.Name, "/"),
			ImageID:       container.Image,
		}

		image, ok := t.images[container.Image]
		if ok && len(image.RepoTags) > 0 {
			nmd[ImageName] = image.RepoTags[0]
		}

		nmd.Merge(container.getStats())

		nodeID := report.MakeContainerNodeID(scope, container.ID)
		result.NodeMetadatas[nodeID] = nmd
	}
	return result
}
