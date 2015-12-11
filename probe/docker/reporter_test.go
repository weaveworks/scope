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

func (r *mockRegistry) WatchContainerUpdates(_ docker.ContainerUpdateWatcher) {}

func (r *mockRegistry) GetContainer(_ string) (docker.Container, bool) { return nil, false }

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
	want := report.MakeReport()
	want.Container = report.Topology{
		Nodes: report.Nodes{
			report.MakeContainerNodeID("", "ping"): report.MakeNodeWith(map[string]string{
				docker.ContainerID:   "ping",
				docker.ContainerName: "pong",
				docker.ImageID:       "baz",
			}),
		},
		Controls: report.Controls{
			docker.RestartContainer: report.Control{
				ID:    docker.RestartContainer,
				Human: "Restart",
				Icon:  "fa-repeat",
			},
			docker.StartContainer: report.Control{
				ID:    docker.StartContainer,
				Human: "Start",
				Icon:  "fa-play",
			},
			docker.StopContainer: report.Control{
				ID:    docker.StopContainer,
				Human: "Stop",
				Icon:  "fa-stop",
			},
			docker.PauseContainer: report.Control{
				ID:    docker.PauseContainer,
				Human: "Pause",
				Icon:  "fa-pause",
			},
			docker.UnpauseContainer: report.Control{
				ID:    docker.UnpauseContainer,
				Human: "Unpause",
				Icon:  "fa-play",
			},
			docker.AttachContainer: report.Control{
				ID:    docker.AttachContainer,
				Human: "Attach",
				Icon:  "fa-desktop",
			},
			docker.ExecContainer: report.Control{
				ID:    docker.ExecContainer,
				Human: "Exec /bin/sh",
				Icon:  "fa-terminal",
			},
		},
	}
	want.ContainerImage = report.Topology{
		Nodes: report.Nodes{
			report.MakeContainerNodeID("", "baz"): report.MakeNodeWith(map[string]string{
				docker.ImageID:   "baz",
				docker.ImageName: "bang",
			}),
		},
		Controls: report.Controls{},
	}

	reporter := docker.NewReporter(mockRegistryInstance, "", nil)
	have, _ := reporter.Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
