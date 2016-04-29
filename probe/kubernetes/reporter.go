package kubernetes

import (
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// Exposed for testing
var (
	PodMetadataTemplates = report.MetadataTemplates{
		PodID:      {ID: PodID, Label: "ID", From: report.FromLatest, Priority: 1},
		PodState:   {ID: PodState, Label: "State", From: report.FromLatest, Priority: 2},
		PodIP:      {ID: PodIP, Label: "IP", From: report.FromLatest, Priority: 3},
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 5},
		PodCreated: {ID: PodCreated, Label: "Created", From: report.FromLatest, Priority: 6},
	}

	ServiceMetadataTemplates = report.MetadataTemplates{
		ServiceID:       {ID: ServiceID, Label: "ID", From: report.FromLatest, Priority: 1},
		Namespace:       {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		ServiceCreated:  {ID: ServiceCreated, Label: "Created", From: report.FromLatest, Priority: 3},
		ServicePublicIP: {ID: ServicePublicIP, Label: "Public IP", From: report.FromLatest, Priority: 4},
		ServiceIP:       {ID: ServiceIP, Label: "Internal IP", From: report.FromLatest, Priority: 5},
	}

	PodTableTemplates = report.TableTemplates{
		PodLabelPrefix: {ID: PodLabelPrefix, Label: "Kubernetes Labels", Prefix: PodLabelPrefix},
	}

	ServiceTableTemplates = report.TableTemplates{
		ServiceLabelPrefix: {ID: ServiceLabelPrefix, Label: "Kubernetes Labels", Prefix: ServiceLabelPrefix},
	}
)

// Reporter generate Reports containing Container and ContainerImage topologies
type Reporter struct {
	client  Client
	pipes   controls.PipeClient
	probeID string
	probe   *probe.Probe
}

// NewReporter makes a new Reporter
func NewReporter(client Client, pipes controls.PipeClient, probeID string, probe *probe.Probe) *Reporter {
	reporter := &Reporter{
		client:  client,
		pipes:   pipes,
		probeID: probeID,
		probe:   probe,
	}
	reporter.registerControls()
	client.WatchPods(reporter.podEvent)
	return reporter
}

// Stop unregisters controls.
func (r *Reporter) Stop() {
	r.deregisterControls()
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "K8s" }

func (r *Reporter) podEvent(e Event, pod Pod) {
	switch e {
	case ADD:
		rpt := report.MakeReport()
		rpt.Shortcut = true
		rpt.Pod.AddNode(pod.GetNode(r.probeID))
		r.probe.Publish(rpt)
	case DELETE:
		rpt := report.MakeReport()
		rpt.Shortcut = true
		rpt.Pod.AddNode(
			report.MakeNodeWith(
				report.MakePodNodeID(pod.UID()),
				map[string]string{PodState: StateDeleted},
			),
		)
		r.probe.Publish(rpt)
	}
}

// Tag adds pod parents to container nodes.
func (r *Reporter) Tag(rpt report.Report) (report.Report, error) {
	for id, n := range rpt.Container.Nodes {
		uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
		if !ok {
			continue
		}
		rpt.Container.Nodes[id] = n.WithParents(report.EmptySets.Add(
			report.Pod,
			report.EmptyStringSet.Add(report.MakePodNodeID(uid)),
		))
	}
	return rpt, nil
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
		result = report.MakeTopology().
			WithMetadataTemplates(ServiceMetadataTemplates).
			WithTableTemplates(ServiceTableTemplates)
		services = []Service{}
	)
	err := r.client.WalkServices(func(s Service) error {
		result = result.AddNode(s.GetNode())
		services = append(services, s)
		return nil
	})
	return result, services, err
}

// GetNodeName return the k8s node name for the current machine.
// It is exported for testing.
var GetNodeName = func(r *Reporter) (string, error) {
	uuidBytes, err := ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
	if os.IsNotExist(err) {
		uuidBytes, err = ioutil.ReadFile("/sys/hypervisor/uuid")
	}
	if err != nil {
		return "", err
	}
	uuid := strings.Trim(string(uuidBytes), "\n")
	nodeName := ""
	err = r.client.WalkNodes(func(node *api.Node) error {
		if node.Status.NodeInfo.SystemUUID == string(uuid) {
			nodeName = node.ObjectMeta.Name
		}
		return nil
	})
	return nodeName, err
}

func (r *Reporter) podTopology(services []Service) (report.Topology, error) {
	var (
		pods = report.MakeTopology().
			WithMetadataTemplates(PodMetadataTemplates).
			WithTableTemplates(PodTableTemplates)
		selectors = map[string]labels.Selector{}
	)
	pods.Controls.AddControl(report.Control{
		ID:    GetLogs,
		Human: "Get logs",
		Icon:  "fa-desktop",
		Rank:  0,
	})
	pods.Controls.AddControl(report.Control{
		ID:    DeletePod,
		Human: "Delete",
		Icon:  "fa-trash-o",
		Rank:  1,
	})
	for _, service := range services {
		selectors[service.ID()] = service.Selector()
	}

	thisNodeName, err := GetNodeName(r)
	if err != nil {
		return pods, err
	}
	err = r.client.WalkPods(func(p Pod) error {
		if p.NodeName() != thisNodeName {
			return nil
		}
		for serviceID, selector := range selectors {
			if selector.Matches(p.Labels()) {
				p.AddServiceID(serviceID)
			}
		}
		pods = pods.AddNode(p.GetNode(r.probeID))
		return nil
	})
	return pods, err
}
