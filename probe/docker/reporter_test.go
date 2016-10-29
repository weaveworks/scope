package docker_test

import (
	"testing"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

type mockRegistry struct {
	containersByPID map[int]docker.Container
	images          map[string]client.APIImages
	networks        []client.Network
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

func (r *mockRegistry) WalkImages(f func(client.APIImages)) {
	for _, i := range r.images {
		f(i)
	}
}

func (r *mockRegistry) WalkNetworks(f func(client.Network)) {
	for _, i := range r.networks {
		f(i)
	}
}

func (r *mockRegistry) WatchContainerUpdates(_ docker.ContainerUpdateWatcher) {}

func (r *mockRegistry) GetContainer(_ string) (docker.Container, bool) { return nil, false }

func (r *mockRegistry) GetContainerByPrefix(_ string) (docker.Container, bool) { return nil, false }

func (r *mockRegistry) GetContainerImage(id string) (client.APIImages, bool) {
	image, ok := r.images[id]
	return image, ok
}

func (r *mockRegistry) StartImage(_, _, _ string, _ xfer.Request) xfer.Response {
	return xfer.Response{}
}

var (
	imageID              = "baz"
	mockRegistryInstance = &mockRegistry{
		containersByPID: map[int]docker.Container{
			2: &mockContainer{container1},
		},
		images: map[string]client.APIImages{
			imageID: apiImage1,
		},
		networks: []client.Network{network1},
	}
)

func TestReporter(t *testing.T) {
	var (
		controlProbeID = "a1b2c3d4"
		hostID         = "host1"
	)

	containerImageNodeID := report.MakeContainerImageNodeID(imageID)
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
			docker.ImageID:        imageID,
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
			docker.ImageID:                      imageID,
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

	// Reporter should add a container network
	{
		overlayNodeID := report.MakeOverlayNodeID(report.DockerOverlayPeerPrefix, hostID)
		node, ok := rpt.Overlay.Nodes[overlayNodeID]
		if !ok {
			t.Fatalf("Expected report to have overlay node  %q, but not found", overlayNodeID)
		}

		want := "5.6.7.8/24"
		if have, ok := node.Sets.Lookup(host.LocalNetworks); !ok || len(have) != 1 || have[0] != want {
			t.Fatalf("Expected node to have exactly local network %v but found %v", want, have)
		}

	}
}
