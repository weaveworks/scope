package docker

import (
	"log"
	"net"

	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/report"
)

// Keys for use in Node
const (
	ImageID   = "docker_image_id"
	ImageName = "docker_image_name"
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	registry Registry
	hostID   string
	probe    *probe.Probe
}

// NewReporter makes a new Reporter
func NewReporter(registry Registry, hostID string, probe *probe.Probe) *Reporter {
	reporter := &Reporter{
		registry: registry,
		hostID:   hostID,
		probe:    probe,
	}
	registry.WatchContainerUpdates(reporter.ContainerUpdated)
	return reporter
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Docker" }

// ContainerUpdated should be called whenever a container is updated.
func (r *Reporter) ContainerUpdated(c Container) {
	localAddrs, err := report.LocalAddresses()
	if err != nil {
		log.Printf("Error getting local address: %v", err)
		return
	}

	// Publish a 'short cut' report container just this container
	rpt := report.MakeReport()
	rpt.Shortcut = true
	rpt.Container.AddNode(report.MakeContainerNodeID(c.ID()), c.GetNode(r.hostID, localAddrs))
	r.probe.Publish(rpt)
}

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	localAddrs, err := report.LocalAddresses()
	if err != nil {
		return report.MakeReport(), nil
	}

	result := report.MakeReport()
	result.Container = result.Container.Merge(r.containerTopology(localAddrs))
	result.ContainerImage = result.ContainerImage.Merge(r.containerImageTopology())
	return result, nil
}

func (r *Reporter) containerTopology(localAddrs []net.IP) report.Topology {
	result := report.MakeTopology()
	result.Controls.AddControl(report.Control{
		ID:    StopContainer,
		Human: "Stop",
		Icon:  "fa-stop",
	})
	result.Controls.AddControl(report.Control{
		ID:    StartContainer,
		Human: "Start",
		Icon:  "fa-play",
	})
	result.Controls.AddControl(report.Control{
		ID:    RestartContainer,
		Human: "Restart",
		Icon:  "fa-repeat",
	})
	result.Controls.AddControl(report.Control{
		ID:    PauseContainer,
		Human: "Pause",
		Icon:  "fa-pause",
	})
	result.Controls.AddControl(report.Control{
		ID:    UnpauseContainer,
		Human: "Unpause",
		Icon:  "fa-play",
	})
	result.Controls.AddControl(report.Control{
		ID:    AttachContainer,
		Human: "Attach",
		Icon:  "fa-desktop",
	})
	result.Controls.AddControl(report.Control{
		ID:    ExecContainer,
		Human: "Exec /bin/sh",
		Icon:  "fa-terminal",
	})

	r.registry.WalkContainers(func(c Container) {
		nodeID := report.MakeContainerNodeID(c.ID())
		result.AddNode(nodeID, c.GetNode(r.hostID, localAddrs))
	})

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.MakeTopology()

	r.registry.WalkImages(func(image *docker_client.APIImages) {
		node := report.MakeNodeWith(map[string]string{
			ImageID: image.ID,
		})
		node = AddLabels(node, image.Labels)

		if len(image.RepoTags) > 0 {
			node = node.WithLatests(map[string]string{ImageName: image.RepoTags[0]})
		}

		nodeID := report.MakeContainerImageNodeID(image.ID)
		result.AddNode(nodeID, node)
	})

	return result
}
