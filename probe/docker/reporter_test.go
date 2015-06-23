package docker_test

import (
	"reflect"
	"testing"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockRegistry struct {
	containersByPID map[int]docker.Container
	images          map[string]*client.APIImages
}

func (r *mockRegistry) Stop() {}

func (r *mockRegistry) LockedPIDLookup(f func(func(int) docker.Container)) {
	f(func(pid int) docker.Container {
		return r.containersByPID[pid]
	})
}

func (r *mockRegistry) WalkContainers(f func(docker.Container)) {
	for _, c := range r.containersByPID {
		f(c)
	}
}

func (r *mockRegistry) WalkImages(f func(*client.APIImages)) {
	for _, i := range r.images {
		f(i)
	}
}

var (
	mockRegistryInstance = &mockRegistry{
		containersByPID: map[int]docker.Container{
			1: &mockContainer{container1},
		},
		images: map[string]*client.APIImages{
			"baz": &apiImage1,
		},
	}
)

func TestReporter(t *testing.T) {
	want := report.MakeReport()
	want.Container = report.Topology{
		Adjacency:     report.Adjacency{},
		EdgeMetadatas: report.EdgeMetadatas{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeContainerNodeID("", "ping"): report.NodeMetadata{
				docker.ContainerID:   "ping",
				docker.ContainerName: "pong",
				docker.ImageID:       "baz",
			},
		},
	}
	want.ContainerImage = report.Topology{
		Adjacency:     report.Adjacency{},
		EdgeMetadatas: report.EdgeMetadatas{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeContainerNodeID("", "baz"): report.NodeMetadata{
				docker.ImageID:   "baz",
				docker.ImageName: "bang",
			},
		},
	}

	reporter := docker.NewReporter(mockRegistryInstance, "")
	have, _ := reporter.Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
