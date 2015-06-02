package main

import (
	"runtime"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type mockDockerClient struct {
	apiContainers []docker.APIContainers
	containers    map[string]*docker.Container
	apiImages     []docker.APIImages
}

func (m mockDockerClient) ListContainers(options docker.ListContainersOptions) ([]docker.APIContainers, error) {
	return m.apiContainers, nil
}

func (m mockDockerClient) InspectContainer(id string) (*docker.Container, error) {
	return m.containers[id], nil
}

func (m mockDockerClient) ListImages(options docker.ListImagesOptions) ([]docker.APIImages, error) {
	return m.apiImages, nil
}

func (m mockDockerClient) AddEventListener(events chan<- *docker.APIEvents) error {
	return nil
}

func (m mockDockerClient) RemoveEventListener(events chan *docker.APIEvents) error {
	return nil
}

func TestDockerProcessMapper(t *testing.T) {
	oldPIDTreeStub, oldDockerClientStub := newPIDTreeStub, newDockerClient
	defer func() {
		newPIDTreeStub = oldPIDTreeStub
		newDockerClient = oldDockerClientStub
	}()

	newPIDTreeStub = func(procRoot string) (*pidTree, error) {
		pid1 := &process{pid: 1}
		pid2 := &process{pid: 2, ppid: 1, parent: pid1}
		pid1.children = []*process{pid2}

		return &pidTree{
			processes: map[int]*process{
				1: pid1, 2: pid2,
			},
		}, nil
	}

	newDockerClient = func(endpoint string) (dockerClient, error) {
		return mockDockerClient{
			apiContainers: []docker.APIContainers{{ID: "foo"}},
			containers: map[string]*docker.Container{
				"foo": {
					ID:    "foo",
					Name:  "bar",
					Image: "baz",
					State: docker.State{Pid: 1, Running: true},
				},
			},
			apiImages: []docker.APIImages{{ID: "baz", RepoTags: []string{"tag"}}},
		}, nil
	}

	dockerMapper, _ := newDockerMapper("/proc", 10*time.Second)
	dockerIDMapper := dockerMapper.idMapper()
	dockerNameMapper := dockerMapper.nameMapper()
	dockerImageIDMapper := dockerMapper.imageIDMapper()
	dockerImageNameMapper := dockerMapper.imageNameMapper()

	runtime.Gosched()

	for pid, want := range map[uint]struct{ id, name, imageID, imageName string }{
		1: {"foo", "bar", "baz", "tag"},
		2: {"foo", "bar", "baz", "tag"},
	} {
		haveID, err := dockerIDMapper.Map(pid)
		if err != nil || want.id != haveID {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.id, haveID, err)
		}
		haveName, err := dockerNameMapper.Map(pid)
		if err != nil || want.name != haveName {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.name, haveName, err)
		}
		haveImageID, err := dockerImageIDMapper.Map(pid)
		if err != nil || want.imageID != haveImageID {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.imageID, haveImageID, err)
		}
		haveImageName, err := dockerImageNameMapper.Map(pid)
		if err != nil || want.imageName != haveImageName {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.imageName, haveImageName, err)
		}
	}
}
