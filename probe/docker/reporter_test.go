package docker_test

import (
	"testing"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
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

func (r *mockRegistry) WatchContainerUpdates(_ docker.ContainerUpdateWatcher) {}

func (r *mockRegistry) GetContainer(_ string) (docker.Container, bool) { return nil, false }

func (r *mockRegistry) GetContainerByPrefix(_ string) (docker.Container, bool) { return nil, false }

var (
	mockRegistryInstance = &mockRegistry{
		containersByPID: map[int]docker.Container{
			2: &mockContainer{container1},
		},
		images: map[string]*client.APIImages{
			"baz": &apiImage1,
		},
	}
)

func TestReporter(t *testing.T) {
	var controlProbeID = "a1b2c3d4"

	containerImageNodeID := report.MakeContainerImageNodeID("baz")
	rpt, err := docker.NewReporter(mockRegistryInstance, "host1", controlProbeID, nil).Report()
	if err != nil {
		t.Fatal(err)
	}

	// Reporter should add a container
	{
		containerNodeID := report.MakeContainerNodeID("ping")
		node, ok := rpt.Container.Nodes[containerNodeID]
		if !ok {
			t.Fatalf("Expected report to have container image %q, but not found", containerNodeID)
		}

		for k, want := range map[string]string{
			docker.ContainerID:    "ping",
			docker.ContainerName:  "pong",
			docker.ImageID:        "baz",
			report.ControlProbeID: controlProbeID,
		} {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected container %s latest %q: %q, got %q", containerNodeID, k, want, have)
			}
		}

		// container should have controls
		if len(rpt.Container.Controls) == 0 {
			t.Errorf("Container should have some controls")
		}

		// container should have the image as a parent
		if parents, ok := node.Parents.Lookup(report.ContainerImage); !ok || !parents.Contains(containerImageNodeID) {
			t.Errorf("Expected container %s to have parent container image %q, got %q", containerNodeID, containerImageNodeID, parents)
		}
	}

	// Reporter should add a container image
	{
		node, ok := rpt.ContainerImage.Nodes[containerImageNodeID]
		if !ok {
			t.Fatalf("Expected report to have container image %q, but not found", containerImageNodeID)
		}

		for k, want := range map[string]string{
			docker.ImageID:                      "baz",
			docker.ImageName:                    "bang",
			docker.ImageLabelPrefix + "imgfoo1": "bar1",
			docker.ImageLabelPrefix + "imgfoo2": "bar2",
		} {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected container image %s latest %q: %q, got %q", containerImageNodeID, k, want, have)
			}
		}

		// container image should have no controls
		if len(rpt.ContainerImage.Controls) != 0 {
			t.Errorf("Container images should not have any controls")
		}
	}
}
