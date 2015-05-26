package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type dockerMapper struct {
	sync.RWMutex
	containers map[int]*docker.Container
	images     map[string]*docker.APIImages
	procRoot   string
}

func newDockerMapper(procRoot string, interval time.Duration) *dockerMapper {
	m := dockerMapper{
		procRoot:   procRoot,
		containers: map[int]*docker.Container{},
	}
	m.update()
	go m.loop(interval)
	return &m
}

func (m *dockerMapper) loop(d time.Duration) {
	for range time.Tick(d) {
		m.update()
	}
}

// for mocking
type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	InspectContainer(string) (*docker.Container, error)
	ListImages(docker.ListImagesOptions) ([]docker.APIImages, error)
}

func newRealDockerClient(endpoint string) (dockerClient, error) {
	return docker.NewClient(endpoint)
}

var (
	newDockerClient = newRealDockerClient
	newPIDTreeStub  = newPIDTree
)

func (m *dockerMapper) update() {
	pidTree, err := newPIDTreeStub(m.procRoot)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClient(endpoint)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	pmap := map[int]*docker.Container{}
	for _, container := range containers {
		info, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			continue
		}

		if !info.State.Running {
			continue
		}

		pids, err := pidTree.allChildren(info.State.Pid)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			continue
		}
		for _, pid := range pids {
			pmap[pid] = info
		}
	}

	imageList, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	imageMap := map[string]*docker.APIImages{}
	for i := range imageList {
		image := &imageList[i]
		imageMap[image.ID] = image
	}

	m.Lock()
	m.containers = pmap
	m.images = imageMap
	m.Unlock()
}

type dockerProcessMapper struct {
	*dockerMapper
	key string
	f   func(*docker.Container) string
}

func (m *dockerProcessMapper) Key() string { return m.key }
func (m *dockerProcessMapper) Map(pid uint) (string, error) {
	m.RLock()
	container, ok := m.containers[int(pid)]
	m.RUnlock()

	if !ok {
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
		return c.Name
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
