package docker

import (
	docker_client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/report"
)

// Keys for use in NodeMetadata
const (
	ImageID   = "docker_image_id"
	ImageName = "docker_image_name"
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	registry Registry
	scope    string
}

// NewReporter makes a new Reporter
func NewReporter(registry Registry, scope string) *Reporter {
	return &Reporter{
		registry: registry,
		scope:    scope,
	}
}

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	result.Container.Merge(r.containerTopology())
	result.ContainerImage.Merge(r.containerImageTopology())
	return result, nil
}

func (r *Reporter) containerTopology() report.Topology {
	result := report.NewTopology()

	r.registry.WalkContainers(func(c Container) {
		nodeID := report.MakeContainerNodeID(r.scope, c.ID())
		result.NodeMetadatas[nodeID] = c.GetNodeMetadata()
	})

	return result
}

func (r *Reporter) containerImageTopology() report.Topology {
	result := report.NewTopology()

	r.registry.WalkImages(func(image *docker_client.APIImages) {
		nmd := report.NodeMetadata{
			ImageID: image.ID,
		}

		if len(image.RepoTags) > 0 {
			nmd[ImageName] = image.RepoTags[0]
		}

		nodeID := report.MakeContainerNodeID(r.scope, image.ID)
		result.NodeMetadatas[nodeID] = nmd
	})

	return result
}
