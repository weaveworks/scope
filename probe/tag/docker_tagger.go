package tag

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
)

const (
	dockerEventStart = "start"
	dockerEventStop  = "stop"
)

var (
	newDockerClient = newRealDockerClient
	newPIDTree      = newRealPIDTree
)

func newRealDockerClient(endpoint string) (dockerClient, error) {
	return docker.NewClient(endpoint)
}

// Sub-interface for mocking.
type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	InspectContainer(string) (*docker.Container, error)
	ListImages(docker.ListImagesOptions) ([]docker.APIImages, error)
	AddEventListener(chan<- *docker.APIEvents) error
	RemoveEventListener(chan *docker.APIEvents) error
}

type dockerTagger struct {
	sync.RWMutex
	containersByID  map[string]*docker.Container
	containersByPID map[int]*docker.Container
	imagesByID      map[string]*docker.APIImages

	procRoot string
	pidTree  *pidTree
	interval time.Duration
	quit     chan struct{}
}

// NewDockerTagger returns a tagger that tags Docker container information to
// nodes with a process_node_id.
func NewDockerTagger(procRoot string, interval time.Duration) (Tagger, error) {
	pidTree, err := newPIDTree(procRoot)
	if err != nil {
		return nil, err
	}

	t := dockerTagger{
		containersByID:  map[string]*docker.Container{},
		containersByPID: map[int]*docker.Container{},
		imagesByID:      map[string]*docker.APIImages{},

		procRoot: procRoot,
		pidTree:  pidTree,
		interval: interval,
		quit:     make(chan struct{}),
	}
	go t.loop()
	return &t, nil
}

func (t *dockerTagger) Tag(r report.Report, ts report.TopologySelector, id string) report.NodeMetadata {
	// Look up the specified node
	myNodeMetadata, ok := ts(r).NodeMetadatas[id]
	if !ok {
		return report.NodeMetadata{}
	}

	// Hopefully it has a process node ID
	processNodeID, ok := myNodeMetadata["process_node_id"]
	if !ok {
		return report.NodeMetadata{}
	}

	// Hopefully that's a valid node in the process topology
	processNodeMetadata, ok := r.Process.NodeMetadatas[processNodeID]
	if !ok {
		return report.NodeMetadata{}
	}

	// Hopefully that process node has a PID
	pidStr, ok := processNodeMetadata["pid"]
	if !ok {
		return report.NodeMetadata{}
	}
	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		return report.NodeMetadata{}
	}

	// Use the PID to look up the container
	var (
		container *docker.Container
		candidate = int(pid)
	)
	t.RLock()
	for {
		container, ok = t.containersByPID[candidate]
		if ok {
			break // found
		}
		candidate, err = t.pidTree.getParent(candidate)
		if err != nil {
			break // oh well
		}
	}
	t.RUnlock()

	if err != nil || container == nil {
		return report.NodeMetadata{} // not found :(
	}

	// Build the metadata
	md := report.NodeMetadata{
		"docker_container_id":   container.ID,
		"docker_container_name": strings.TrimPrefix(container.Name, "/"),
		"docker_image_id":       container.Image,
	}

	t.RLock()
	image, ok := t.imagesByID[container.Image]
	t.RUnlock()

	if ok && len(image.RepoTags) > 0 {
		md["docker_image_name"] = image.RepoTags[0]
	}

	return md
}

func (t *dockerTagger) Stop() {
	close(t.quit)
}

func (t *dockerTagger) loop() {
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

// Return false to stop the owning loop gorouting.
func (t *dockerTagger) update() bool {
	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClient(endpoint)
	if err != nil {
		log.Printf("docker tagger: %v", err)
		return true
	}

	events := make(chan *docker.APIEvents)
	if err := client.AddEventListener(events); err != nil {
		log.Printf("docker tagger: %v", err)
		return true
	}
	defer func() {
		if err := client.RemoveEventListener(events); err != nil {
			log.Printf("docker tagger: %v", err)
		}
	}()

	if err := t.updateContainers(client); err != nil {
		log.Printf("docker tagger: %v", err)
		return true
	}

	if err := t.updateImages(client); err != nil {
		log.Printf("docker tagger: %v", err)
		return true
	}

	otherUpdates := time.Tick(t.interval)
	for {
		select {
		case event := <-events:
			t.handleEvent(event, client)

		case <-otherUpdates:
			if err := t.updatePIDTree(); err != nil {
				log.Printf("docker tagger: %v", err)
				continue
			}

			if err := t.updateImages(client); err != nil {
				log.Printf("docker tagger: %v", err)
				continue
			}

		case <-t.quit:
			return false
		}
	}
}

func (t *dockerTagger) updateContainers(client dockerClient) error {
	apiContainers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	containers := []*docker.Container{}
	for _, apiContainer := range apiContainers {
		container, err := client.InspectContainer(apiContainer.ID)
		if err != nil {
			log.Printf("docker tagger: %v", err)
			continue
		}
		if !container.State.Running {
			continue
		}
		containers = append(containers, container)
	}

	t.Lock()
	for _, container := range containers {
		t.containersByID[container.ID] = container
		t.containersByPID[container.State.Pid] = container
	}
	t.Unlock()

	return nil
}

func (t *dockerTagger) updateImages(client dockerClient) error {
	images, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return err
	}

	t.Lock()
	for i := range images {
		image := &images[i]
		t.imagesByID[image.ID] = image
	}
	t.Unlock()

	return nil
}

func (t *dockerTagger) handleEvent(event *docker.APIEvents, client dockerClient) {
	switch event.Status {
	case dockerEventStop:
		containerID := event.ID

		t.Lock()
		if container, ok := t.containersByID[containerID]; ok {
			delete(t.containersByID, containerID)
			delete(t.containersByPID, container.State.Pid)
		} else {
			log.Printf("docker tagger: container %s not found", containerID)
		}
		t.Unlock()

	case dockerEventStart:
		containerID := event.ID
		container, err := client.InspectContainer(containerID)
		if err != nil {
			log.Printf("docker tagger: %v", err)
			return
		}

		if !container.State.Running {
			log.Printf("docker tagger: container %s not running", containerID)
			return
		}

		t.Lock()
		t.containersByID[containerID] = container
		t.containersByPID[container.State.Pid] = container
		t.Unlock()
	}
}

func (t *dockerTagger) updatePIDTree() error {
	pidTree, err := newPIDTree(t.procRoot)
	if err != nil {
		return err
	}

	t.Lock()
	t.pidTree = pidTree
	t.Unlock()

	return nil
}
