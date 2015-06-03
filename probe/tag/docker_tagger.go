package tag

import (
	"log"
	"strconv"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
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
}

type dockerTagger struct {
	sync.RWMutex
	procRoot   string
	containers map[int]*docker.Container
	images     map[string]*docker.APIImages
	quit       chan struct{}
}

// NewDockerTagger returns a tagger that tags Docker container information to
// nodes with a process_node_id.
func NewDockerTagger(procRoot string, interval time.Duration) Tagger {
	t := dockerTagger{
		procRoot:   procRoot,
		containers: map[int]*docker.Container{},
		images:     map[string]*docker.APIImages{},
		quit:       make(chan struct{}),
	}
	t.update()
	go t.loop(interval)
	return &t
}

func (t *dockerTagger) Tag(r report.Report, ts report.TopologySelector, id string) report.NodeMetadata {
	// Cross-reference the process.
	myNodeMetadata, ok := ts(r).NodeMetadatas[id]
	if !ok {
		//log.Printf("dockerTagger: %q: missing", id)
		return report.NodeMetadata{}
	}
	processNodeID, ok := myNodeMetadata["process_node_id"]
	if !ok {
		//log.Printf("dockerTagger: %q: no process node ID", id)
		return report.NodeMetadata{}
	}
	processNodeMetadata, ok := r.Process.NodeMetadatas[processNodeID]
	if !ok {
		//log.Printf("dockerTagger: %q: process node ID missing", id)
		return report.NodeMetadata{}
	}
	pidStr, ok := processNodeMetadata["pid"]
	if !ok {
		//log.Printf("dockerTagger: %q: process node has no PID", id)
		return report.NodeMetadata{}
	}
	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		//log.Printf("dockerTagger: %q: bad process node PID (%v)", id, err)
		return report.NodeMetadata{}
	}

	t.RLock()
	container, ok := t.containers[int(pid)]
	t.RUnlock()

	if !ok {
		return report.NodeMetadata{}
	}

	md := report.NodeMetadata{
		"docker_container_id":   container.ID,
		"docker_container_name": container.Name,
		"docker_image_id":       container.Image,
	}

	t.RLock()
	image, ok := t.images[container.Image]
	t.RUnlock()

	if ok && len(image.RepoTags) > 0 {
		md["docker_image_name"] = image.RepoTags[0]
	}

	return md
}

func (t *dockerTagger) Stop() {
	close(t.quit)
}

func (t *dockerTagger) loop(d time.Duration) {
	for range time.Tick(d) {
		t.update()
	}
}

func (t *dockerTagger) update() {
	pidTree, err := newPIDTree(t.procRoot)
	if err != nil {
		log.Printf("docker tagger: %s", err)
		return
	}

	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClient(endpoint)
	if err != nil {
		log.Printf("docker tagger: %s", err)
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Printf("docker tagger: %s", err)
		return
	}

	pmap := map[int]*docker.Container{}
	for _, container := range containers {
		info, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Printf("docker tagger: %s", err)
			continue
		}

		if !info.State.Running {
			continue
		}

		pids, err := pidTree.allChildren(info.State.Pid)
		if err != nil {
			log.Printf("docker tagger: %s", err)
			continue
		}
		for _, pid := range pids {
			pmap[pid] = info
		}
	}

	imageList, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		log.Printf("docker tagger: %s", err)
		return
	}

	imageMap := map[string]*docker.APIImages{}
	for i := range imageList {
		image := &imageList[i]
		imageMap[image.ID] = image
	}

	t.Lock()
	t.containers = pmap
	t.images = imageMap
	t.Unlock()
}
