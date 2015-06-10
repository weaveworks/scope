package tag

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
)

type mockDockerClient struct {
	apiContainers []docker.APIContainers
	containers    map[string]*docker.Container
	apiImages     []docker.APIImages
}

func (m mockDockerClient) ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error) {
	return m.apiContainers, nil
}

func (m mockDockerClient) InspectContainer(id string) (*docker.Container, error) {
	return m.containers[id], nil
}

func (m mockDockerClient) ListImages(docker.ListImagesOptions) ([]docker.APIImages, error) {
	return m.apiImages, nil
}

func (m mockDockerClient) AddEventListener(events chan<- *docker.APIEvents) error {
	return nil
}

func (m mockDockerClient) RemoveEventListener(events chan *docker.APIEvents) error {
	return nil
}

func TestDockerTagger(t *testing.T) {
	oldPIDTree, oldDockerClient := newPIDTreeStub, newDockerClientStub
	defer func() { newPIDTreeStub, newDockerClientStub = oldPIDTree, oldDockerClient }()

	newPIDTreeStub = func(procRoot string) (*PIDTree, error) {
		pid1 := &Process{PID: 1}
		pid2 := &Process{PID: 2, PPID: 1, parent: pid1}
		pid1.children = []*Process{pid2}
		return &PIDTree{
			processes: map[int]*Process{
				1: pid1, 2: pid2,
			},
		}, nil
	}

	newDockerClientStub = func(endpoint string) (dockerClient, error) {
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
			apiImages: []docker.APIImages{{ID: "baz", RepoTags: []string{"bang", "not-chosen"}}},
		}, nil
	}

	var (
		endpoint1NodeID     = "somehost.com;192.168.1.1;12345"
		endpoint2NodeID     = "somehost.com;192.168.1.1;67890"
		processNodeMetadata = report.NodeMetadata{
			"docker_container_id":   "foo",
			"docker_container_name": "bar",
			"docker_image_id":       "baz",
			"docker_image_name":     "bang",
		}
	)

	r := report.MakeReport()
	r.Endpoint.NodeMetadatas[endpoint1NodeID] = report.NodeMetadata{"pid": "1"}
	r.Endpoint.NodeMetadatas[endpoint2NodeID] = report.NodeMetadata{"pid": "2"}

	dockerTagger, _ := NewDockerTagger("/irrelevant", 10*time.Second)
	runtime.Gosched()
	for _, endpointNodeID := range []string{endpoint1NodeID, endpoint2NodeID} {
		want := processNodeMetadata.Copy()
		have := dockerTagger.Tag(r).Endpoint.NodeMetadatas[endpointNodeID].Copy()
		delete(have, "pid")
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: want %+v, have %+v", endpointNodeID, want, have)
		}
	}
}
