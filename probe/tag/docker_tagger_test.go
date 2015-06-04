package tag

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
)

func TestDockerTagger(t *testing.T) {
	oldPIDTree, oldDockerClient := newPIDTree, newDockerClient
	defer func() { newPIDTree, newDockerClient = oldPIDTree, oldDockerClient }()

	newPIDTree = func(procRoot string) (*pidTree, error) {
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
			apiImages: []docker.APIImages{{ID: "baz", RepoTags: []string{"bang", "not-chosen"}}},
		}, nil
	}

	var (
		hostID              = "somehost.com"
		ipAddress           = "192.168.1.1"
		endpoint1NodeID     = report.MakeEndpointNodeID(hostID, ipAddress, "12345")
		endpoint2NodeID     = report.MakeEndpointNodeID(hostID, ipAddress, "67890")
		process1NodeID      = report.MakeProcessNodeID(hostID, "1")
		process2NodeID      = report.MakeProcessNodeID(hostID, "2")
		processNodeMetadata = report.NodeMetadata{
			"docker_container_id":   "foo",
			"docker_container_name": "bar",
			"docker_image_id":       "baz",
			"docker_image_name":     "bang",
		}
	)

	r := report.MakeReport()
	r.Endpoint.NodeMetadatas[endpoint1NodeID] = report.NodeMetadata{"process_node_id": process1NodeID}
	r.Endpoint.NodeMetadatas[endpoint2NodeID] = report.NodeMetadata{"process_node_id": process2NodeID}
	r.Process.NodeMetadatas[process1NodeID] = processNodeMetadata.Copy().Merge(report.NodeMetadata{"pid": "1"})
	r.Process.NodeMetadatas[process2NodeID] = processNodeMetadata.Copy().Merge(report.NodeMetadata{"pid": "2"})

	dockerTagger, err := NewDockerTagger("/irrelevant", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	runtime.Gosched()

	for _, endpointNodeID := range []string{endpoint1NodeID, endpoint2NodeID} {
		want := processNodeMetadata.Copy()
		have := dockerTagger.Tag(r, report.SelectEndpoint, endpointNodeID).Copy()
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: want %+v, have %+v", endpointNodeID, want, have)
		}
	}
}

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

func (m mockDockerClient) AddEventListener(chan<- *docker.APIEvents) error {
	return nil
}

func (m mockDockerClient) RemoveEventListener(chan *docker.APIEvents) error {
	return nil
}
