package docker

import (
	docker_client "github.com/fsouza/go-dockerclient"

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
}

// NewReporter makes a new Reporter
func NewReporter(registry Registry, hostID string) *Reporter {
	return &Reporter{
		registry: registry,
		hostID:   hostID,
	}
}

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	result.Container = result.Container.Merge(r.containerTopology())
	result.ContainerImage = result.ContainerImage.Merge(r.containerImageTopology())
	return result, nil
}

func (r *Reporter) containerTopology() report.Topology {
	result := report.MakeTopology()

	r.registry.WalkContainers(func(c Container) {
		nodeID := report.MakeContainerNodeID(r.hostID, c.ID())
		result.AddNode(nodeID, c.GetNode())
	})

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.MakeTopology()

	r.registry.WalkImages(func(image *docker_client.APIImages) {
		nmd := report.MakeNodeWith(map[string]string{
			ImageID: image.ID,
		})
		AddLabels(nmd, image.Labels)

		if len(image.RepoTags) > 0 {
			nmd.Metadata[ImageName] = image.RepoTags[0]
		}

		nodeID := report.MakeContainerNodeID(r.hostID, image.ID)
		result.AddNode(nodeID, nmd)
	})

	return result
}
