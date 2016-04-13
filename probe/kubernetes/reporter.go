package kubernetes

import (
	"k8s.io/kubernetes/pkg/labels"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// Exposed for testing
var (
	PodMetadataTemplates = report.MetadataTemplates{
		PodID:      {ID: PodID, Label: "ID", From: report.FromLatest, Priority: 1},
		PodState:   {ID: PodState, Label: "State", From: report.FromLatest, Priority: 2},
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 3},
		PodCreated: {ID: PodCreated, Label: "Created", From: report.FromLatest, Priority: 4},
	}

	ServiceMetadataTemplates = report.MetadataTemplates{
		ServiceID:      {ID: ServiceID, Label: "ID", From: report.FromLatest, Priority: 1},
		Namespace:      {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		ServiceCreated: {ID: ServiceCreated, Label: "Created", From: report.FromLatest, Priority: 3},
	}
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	client Client
}

// NewReporter makes a new Reporter
func NewReporter(client Client) *Reporter {
	return &Reporter{
		client: client,
	}
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "K8s" }

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	serviceTopology, services, err := r.serviceTopology()
	if err != nil {
		return result, err
	}
	podTopology, containerTopology, err := r.podTopology(services)
	if err != nil {
		return result, err
	}
	result.Service = result.Service.Merge(serviceTopology)
	result.Pod = result.Pod.Merge(podTopology)
	result.Container = result.Container.Merge(containerTopology)
	return result, nil
}

func (r *Reporter) serviceTopology() (report.Topology, []Service, error) {
	var (
		result   = report.MakeTopology().WithMetadataTemplates(ServiceMetadataTemplates)
		services = []Service{}
	)
	err := r.client.WalkServices(func(s Service) error {
		nodeID := report.MakeServiceNodeID(s.Namespace(), s.Name())
		result = result.AddNode(nodeID, s.GetNode())
		services = append(services, s)
		return nil
	})
	return result, services, err
}

func (r *Reporter) podTopology(services []Service) (report.Topology, report.Topology, error) {
	var (
		pods       = report.MakeTopology().WithMetadataTemplates(PodMetadataTemplates)
		containers = report.MakeTopology()
		selectors  = map[string]labels.Selector{}
	)
	for _, service := range services {
		selectors[service.ID()] = service.Selector()
	}
	err := r.client.WalkPods(func(p Pod) error {
		for serviceID, selector := range selectors {
			if selector.Matches(p.Labels()) {
				p.AddServiceID(serviceID)
			}
		}
		nodeID := report.MakePodNodeID(p.Namespace(), p.Name())
		pods = pods.AddNode(nodeID, p.GetNode())

		for _, containerID := range p.ContainerIDs() {
			container := report.MakeNodeWith(map[string]string{
				PodID:              p.ID(),
				Namespace:          p.Namespace(),
				docker.ContainerID: containerID,
			}).WithParents(report.EmptySets.Add(report.Pod, report.MakeStringSet(nodeID)))
			containers.AddNode(report.MakeContainerNodeID(containerID), container)
		}
		return nil
	})
	return pods, containers, err
}
