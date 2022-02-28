package cri

import (
	"context"
	"fmt"

	client "github.com/weaveworks/scope/cri/runtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	cri client.RuntimeServiceClient
}

// NewReporter makes a new Reporter
func NewReporter(cri client.RuntimeServiceClient) *Reporter {
	reporter := &Reporter{
		cri: cri,
	}

	return reporter
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "CRI" }

// Report generates a Report containing Container topologies
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	containerTopol, err := r.containerTopology()
	if err != nil {
		return report.MakeReport(), err
	}

	result.Container = result.Container.Merge(containerTopol)
	return result, nil
}

func (r *Reporter) containerTopology() (report.Topology, error) {
	result := report.MakeTopology().
		WithMetadataTemplates(docker.ContainerImageMetadataTemplates).
		WithTableTemplates(docker.ContainerImageTableTemplates)

	ctx := context.Background()
	resp, err := r.cri.ListContainers(ctx, &client.ListContainersRequest{})
	if err != nil {
		return result, err
	}

	for _, c := range resp.Containers {
		result.AddNode(getNode(c))
	}

	return result, nil
}

func containerStateString(s client.ContainerState) string {
	switch s {
	case client.ContainerState_CONTAINER_CREATED:
		return report.StateCreated
	case client.ContainerState_CONTAINER_RUNNING:
		return report.StateRunning
	case client.ContainerState_CONTAINER_EXITED:
		return report.StateExited
	}
	return "unknown"
}

func getNode(c *client.Container) report.Node {
	result := report.MakeNodeWith(report.MakeContainerNodeID(c.Id), map[string]string{
		docker.ContainerName:         c.Metadata.Name,
		docker.ContainerID:           c.Id,
		docker.ContainerState:        containerStateString(c.State),
		docker.ContainerRestartCount: fmt.Sprintf("%v", c.Metadata.Attempt),
		docker.ImageID:               c.ImageRef,
		docker.ImageName:             c.Image.Image,
	}).WithParents(report.MakeSets().
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.ImageRef))),
	)
	result = result.AddPrefixPropertyList(docker.LabelPrefix, c.Labels)

	return result
}
