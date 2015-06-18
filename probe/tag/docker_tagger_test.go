package tag

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
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
		pid1NodeID          = report.MakeProcessNodeID("somehost.com", "1")
		pid2NodeID          = report.MakeProcessNodeID("somehost.com", "2")
		processNodeMetadata = report.NodeMetadata{
			ContainerID: "foo",
		}
		wantContainerTopology = report.Topology{
			Adjacency:     report.Adjacency{},
			EdgeMetadatas: report.EdgeMetadatas{},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeContainerNodeID("", "foo"): report.NodeMetadata{
					ContainerID:   "foo",
					ContainerName: "bar",
					ImageID:       "baz",
				},
			},
		}
		wantContainerImageTopology = report.Topology{
			Adjacency:     report.Adjacency{},
			EdgeMetadatas: report.EdgeMetadatas{},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeContainerNodeID("", "baz"): report.NodeMetadata{
					ImageID:   "baz",
					ImageName: "bang",
				},
			},
		}
	)

	r := report.MakeReport()
	r.Process.NodeMetadatas[pid1NodeID] = report.NodeMetadata{"pid": "1"}
	r.Process.NodeMetadatas[pid2NodeID] = report.NodeMetadata{"pid": "2"}

	dockerTagger, _ := NewDockerTagger("/irrelevant", 10*time.Second)
	runtime.Gosched()
	for _, nodeID := range []string{pid1NodeID, pid2NodeID} {
		want := processNodeMetadata.Copy()
		have := dockerTagger.Tag(r).Process.NodeMetadatas[nodeID].Copy()
		delete(have, "pid")
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: want %+v, have %+v", nodeID, want, have)
		}
	}

	haveContainerTopology := dockerTagger.ContainerTopology("")
	if !reflect.DeepEqual(wantContainerTopology, haveContainerTopology) {
		t.Errorf("%s", test.Diff(wantContainerTopology, haveContainerTopology))
	}

	haveContainerImageTopology := dockerTagger.ContainerImageTopology("")
	if !reflect.DeepEqual(wantContainerImageTopology, haveContainerImageTopology) {
		t.Errorf("%s", test.Diff(wantContainerImageTopology, haveContainerImageTopology))
	}
}
