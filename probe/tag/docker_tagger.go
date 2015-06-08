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
	stop  = "stop"
	start = "start"
)

var (
	newDockerClientStub = newDockerClient
	newPIDTreeStub      = newPIDTree
)

// DockerTagger is a tagger that tags Docker container information to process
// nodes that have a PID.
type DockerTagger struct {
	sync.RWMutex
	quit     chan struct{}
	interval time.Duration

	containers      map[string]*docker.Container
	containersByPID map[int]*docker.Container
	images          map[string]*docker.APIImages

	procRoot string
	pidTree  *pidTree
}

// NewDockerTagger returns a usable DockerTagger. Don't forget to Stop it.
func NewDockerTagger(procRoot string, interval time.Duration) (*DockerTagger, error) {
	pidTree, err := newPIDTreeStub(procRoot)
	if err != nil {
		return nil, err
	}

	t := DockerTagger{
		containers:      map[string]*docker.Container{},
		containersByPID: map[int]*docker.Container{},
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

func (t *DockerTagger) update() bool {
	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClientStub(endpoint)
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

	if err := t.updateContainers(client); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	if err := t.updateImages(client); err != nil {
		log.Printf("docker mapper: %s", err)
		return true
	}

	otherUpdates := time.Tick(t.interval)
	for {
		select {
		case event := <-events:
			t.handleEvent(event, client)

		case <-otherUpdates:
			if err := t.updatePIDTree(); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

			if err := t.updateImages(client); err != nil {
				log.Printf("docker mapper: %s", err)
				continue
			}

		case <-t.quit:
			return false
		}
	}
}

func (t *DockerTagger) updateContainers(client dockerClient) error {
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

	t.Lock()
	for _, container := range containers {
		t.containers[container.ID] = container
		t.containersByPID[container.State.Pid] = container
	}
	t.Unlock()

	return nil
}

func (t *DockerTagger) updateImages(client dockerClient) error {
	images, err := client.ListImages(docker.ListImagesOptions{})
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

func (t *DockerTagger) handleEvent(event *docker.APIEvents, client dockerClient) {
	switch event.Status {
	case stop:
		containerID := event.ID
		t.Lock()
		if container, ok := t.containers[containerID]; ok {
			delete(t.containers, containerID)
			delete(t.containersByPID, container.State.Pid)
		} else {
			log.Printf("docker mapper: container %s not found", containerID)
		}
		t.Unlock()

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

		t.Lock()
		t.containers[containerID] = container
		t.containersByPID[container.State.Pid] = container
		t.Unlock()
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

// Containers returns the Containers the DockerTagger knows about.
func (t *DockerTagger) Containers() []*docker.Container {
	containers := []*docker.Container{}

	t.RLock()
	for _, container := range t.containers {
		containers = append(containers, container)
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
			container *docker.Container
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
			"docker_container_id":   container.ID,
			"docker_container_name": strings.TrimPrefix(container.Name, "/"),
			"docker_image_id":       container.Image,
		}

		t.RLock()
		image, ok := t.images[container.Image]
		t.RUnlock()

		if ok && len(image.RepoTags) > 0 {
			md["docker_image_name"] = image.RepoTags[0]
		}

		r.Process.NodeMetadatas[nodeID].Merge(md)
	}

	return r
}
