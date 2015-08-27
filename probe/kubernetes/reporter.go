package kubernetes

import (
	"github.com/weaveworks/scope/report"
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

// Report generates a Report containing Container and ContainerImage topologies
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	serviceTopology, services, err := r.serviceTopology()
	if err != nil {
		return result, err
	}
	podTopology, err := r.podTopology(services)
	if err != nil {
		return result, err
	}
	result.Service = result.Service.Merge(serviceTopology)
	result.Pod = result.Pod.Merge(podTopology)
	return result, nil
}

func (r *Reporter) serviceTopology() (report.Topology, []Service, error) {
	var (
		result   = report.MakeTopology()
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

func (r *Reporter) podTopology(services []Service) (report.Topology, error) {
	result := report.MakeTopology()
	err := r.client.WalkPods(func(p Pod) error {
		for _, service := range services {
			if service.Selector().Matches(p.Labels()) {
				p.AddServiceID(service.ID())
			}
		}
		nodeID := report.MakePodNodeID(p.Namespace(), p.Name())
		result = result.AddNode(nodeID, p.GetNode())
		return nil
	})
	return result, err
}
