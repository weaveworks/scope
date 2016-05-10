package kubernetes

import (
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	IP                 = "kubernetes_ip"
	ObservedGeneration = "kubernetes_observed_generation"
	Replicas           = "kubernetes_replicas"
	DesiredReplicas    = "kubernetes_desired_replicas"
)

// Exposed for testing
var (
	PodMetadataTemplates = report.MetadataTemplates{
		ID:               {ID: ID, Label: "ID", From: report.FromLatest, Priority: 1},
		State:            {ID: State, Label: "State", From: report.FromLatest, Priority: 2},
		IP:               {ID: IP, Label: "IP", From: report.FromLatest, Priority: 3},
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: "number", Priority: 4},
		Namespace:        {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 5},
		Created:          {ID: Created, Label: "Created", From: report.FromLatest, Priority: 6},
	}

	ServiceMetadataTemplates = report.MetadataTemplates{
		ID:         {ID: ID, Label: "ID", From: report.FromLatest, Priority: 1},
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:    {ID: Created, Label: "Created", From: report.FromLatest, Priority: 3},
		PublicIP:   {ID: PublicIP, Label: "Public IP", From: report.FromLatest, Priority: 4},
		IP:         {ID: IP, Label: "Internal IP", From: report.FromLatest, Priority: 5},
		report.Pod: {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	DeploymentMetadataTemplates = report.MetadataTemplates{
		ID:                 {ID: ID, Label: "ID", From: report.FromLatest, Priority: 1},
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
		Strategy:           {ID: Strategy, Label: "Strategy", From: report.FromLatest, Priority: 7},
	}

	ReplicaSetMetadataTemplates = report.MetadataTemplates{
		ID:                 {ID: ID, Label: "ID", From: report.FromLatest, Priority: 1},
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	TableTemplates = report.TableTemplates{
		LabelPrefix: {ID: LabelPrefix, Label: "Kubernetes Labels", Prefix: LabelPrefix},
	}

	ScalingControls = []report.Control{
		{
			ID:    ScaleDown,
			Human: "Scale Down",
			Icon:  "fa-minus",
			Rank:  0,
		},
		{
			ID:    ScaleUp,
			Human: "Scale Up",
			Icon:  "fa-plus",
			Rank:  1,
		},
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
				map[string]string{State: StateDeleted},
			),
		)
		r.probe.Publish(rpt)
	}
}

func isPauseContainer(n report.Node, rpt report.Report) bool {
	containerImageIDs, ok := n.Parents.Lookup(report.ContainerImage)
	if !ok {
		return false
	}
	for _, imageNodeID := range containerImageIDs {
		imageNode, ok := rpt.ContainerImage.Nodes[imageNodeID]
		if !ok {
			continue
		}
		imageName, ok := imageNode.Latest.Lookup(docker.ImageName)
		if !ok {
			continue
		}
		if docker.ImageNameWithoutVersion(imageName) == "google_containers/pause" {
			return true
		}
	}
	return false
}

// Tag adds pod parents to container nodes.
func (r *Reporter) Tag(rpt report.Report) (report.Report, error) {
	for id, n := range rpt.Container.Nodes {
		uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
		if !ok {
			continue
		}

		// Tag the pause containers with "does-not-make-connections"
		if isPauseContainer(n, rpt) {
			n = n.WithLatest(report.DoesNotMakeConnections, mtime.Now(), "")
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
	deploymentTopology, deployments, err := r.deploymentTopology(r.probeID)
	if err != nil {
		return result, err
	}
	replicaSetTopology, replicaSets, err := r.replicaSetTopology(r.probeID, deployments)
	if err != nil {
		return result, err
	}
	podTopology, err := r.podTopology(services, replicaSets)
	if err != nil {
		return result, err
	}
	result.Pod = result.Pod.Merge(podTopology)
	result.Service = result.Service.Merge(serviceTopology)
	result.Deployment = result.Deployment.Merge(deploymentTopology)
	result.ReplicaSet = result.ReplicaSet.Merge(replicaSetTopology)
	return result, nil
}

func (r *Reporter) serviceTopology() (report.Topology, []Service, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(ServiceMetadataTemplates).
			WithTableTemplates(TableTemplates)
		services = []Service{}
	)
	err := r.client.WalkServices(func(s Service) error {
		result = result.AddNode(s.GetNode())
		services = append(services, s)
		return nil
	})
	return result, services, err
}

func (r *Reporter) deploymentTopology(probeID string) (report.Topology, []Deployment, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(DeploymentMetadataTemplates).
			WithTableTemplates(TableTemplates)
		deployments = []Deployment{}
	)
	result.Controls.AddControls(ScalingControls)

	err := r.client.WalkDeployments(func(d Deployment) error {
		result = result.AddNode(d.GetNode(probeID))
		deployments = append(deployments, d)
		return nil
	})
	return result, deployments, err
}

func (r *Reporter) replicaSetTopology(probeID string, deployments []Deployment) (report.Topology, []ReplicaSet, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(ReplicaSetMetadataTemplates).
			WithTableTemplates(TableTemplates)
		replicaSets = []ReplicaSet{}
		selectors   = []func(labelledChild){}
	)
	result.Controls.AddControls(ScalingControls)

	for _, deployment := range deployments {
		selectors = append(selectors, match(
			deployment.Selector(),
			report.Deployment,
			report.MakeDeploymentNodeID(deployment.UID()),
		))
	}

	err := r.client.WalkReplicaSets(func(r ReplicaSet) error {
		for _, selector := range selectors {
			selector(r)
		}
		result = result.AddNode(r.GetNode(probeID))
		replicaSets = append(replicaSets, r)
		return nil
	})
	if err != nil {
		return result, replicaSets, err
	}

	err = r.client.WalkReplicationControllers(func(r ReplicationController) error {
		for _, selector := range selectors {
			selector(r)
		}
		result = result.AddNode(r.GetNode(probeID))
		replicaSets = append(replicaSets, ReplicaSet(r))
		return nil
	})
	return result, replicaSets, err
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

type labelledChild interface {
	Labels() map[string]string
	AddParent(string, string)
}

// Match parses the selectors and adds the target as a parent if the selector matches.
func match(selector labels.Selector, topology, id string) func(labelledChild) {
	return func(c labelledChild) {
		if selector.Matches(labels.Set(c.Labels())) {
			c.AddParent(topology, id)
		}
	}
}

func (r *Reporter) podTopology(services []Service, replicaSets []ReplicaSet) (report.Topology, error) {
	var (
		pods = report.MakeTopology().
			WithMetadataTemplates(PodMetadataTemplates).
			WithTableTemplates(TableTemplates)
		selectors = []func(labelledChild){}
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
		selectors = append(selectors, match(
			service.Selector(),
			report.Service,
			report.MakeServiceNodeID(service.UID()),
		))
	}
	for _, replicaSet := range replicaSets {
		selectors = append(selectors, match(
			replicaSet.Selector(),
			report.ReplicaSet,
			report.MakeReplicaSetNodeID(replicaSet.UID()),
		))
	}

	thisNodeName, err := GetNodeName(r)
	if err != nil {
		return pods, err
	}
	err = r.client.WalkPods(func(p Pod) error {
		if p.NodeName() != thisNodeName {
			return nil
		}
		for _, selector := range selectors {
			selector(p)
		}
		pods = pods.AddNode(p.GetNode(r.probeID))
		return nil
	})
	return pods, err
}
