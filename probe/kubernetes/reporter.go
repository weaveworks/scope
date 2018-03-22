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
	"github.com/weaveworks/scope/probe/host"
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
		State:            {ID: State, Label: "State", From: report.FromLatest, Priority: 2},
		IP:               {ID: IP, Label: "IP", From: report.FromLatest, Priority: 3},
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: "number", Priority: 4},
		Namespace:        {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 5},
		Created:          {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 6},
	}

	PodMetricTemplates = docker.ContainerMetricTemplates

	ServiceMetadataTemplates = report.MetadataTemplates{
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:    {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		PublicIP:   {ID: PublicIP, Label: "Public IP", From: report.FromLatest, Priority: 4},
		IP:         {ID: IP, Label: "Internal IP", From: report.FromLatest, Priority: 5},
		report.Pod: {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	ServiceMetricTemplates = PodMetricTemplates

	DeploymentMetadataTemplates = report.MetadataTemplates{
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
		Strategy:           {ID: Strategy, Label: "Strategy", From: report.FromLatest, Priority: 7},
	}

	DeploymentMetricTemplates = ReplicaSetMetricTemplates

	ReplicaSetMetadataTemplates = report.MetadataTemplates{
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	ReplicaSetMetricTemplates = PodMetricTemplates

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
	client          Client
	pipes           controls.PipeClient
	probeID         string
	probe           *probe.Probe
	hostID          string
	handlerRegistry *controls.HandlerRegistry
}

// NewReporter makes a new Reporter
func NewReporter(client Client, pipes controls.PipeClient, probeID string, hostID string, probe *probe.Probe, handlerRegistry *controls.HandlerRegistry) *Reporter {
	reporter := &Reporter{
		client:          client,
		pipes:           pipes,
		probeID:         probeID,
		probe:           probe,
		hostID:          hostID,
		handlerRegistry: handlerRegistry,
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

// IsPauseImageName indicates whether an image name corresponds to a
// kubernetes pause container image.
func IsPauseImageName(imageName string) bool {
	return strings.Contains(imageName, "google_containers/pause")

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
		return IsPauseImageName(imageName)
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
	hostTopology := r.hostTopology(services)
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
	result.Host = result.Host.Merge(hostTopology)
	result.Deployment = result.Deployment.Merge(deploymentTopology)
	result.ReplicaSet = result.ReplicaSet.Merge(replicaSetTopology)
	return result, nil
}

func (r *Reporter) serviceTopology() (report.Topology, []Service, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(ServiceMetadataTemplates).
			WithMetricTemplates(ServiceMetricTemplates).
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

// FIXME: Hideous hack to remove persistent-connection edges to virtual service
//        IPs attributed to the internet. We add each service IP as a /32 network
//        (the global service-cluster-ip-range is not exposed by the API
//        server so we treat each IP as a /32 network see
//        https://github.com/kubernetes/kubernetes/issues/25533).
//        The right way of fixing this is performing DNAT mapping on persistent
//        connections for which we don't have a robust solution
//        (see https://github.com/weaveworks/scope/issues/1491)
func (r *Reporter) hostTopology(services []Service) report.Topology {
	localNetworks := report.EmptyStringSet
	for _, service := range services {
		localNetworks = localNetworks.Add(service.ClusterIP() + "/32")
	}
	node := report.MakeNode(report.MakeHostNodeID(r.hostID))
	node = node.WithSets(report.EmptySets.
		Add(host.LocalNetworks, localNetworks))
	return report.MakeTopology().AddNode(node)
}

func (r *Reporter) deploymentTopology(probeID string) (report.Topology, []Deployment, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(DeploymentMetadataTemplates).
			WithMetricTemplates(DeploymentMetricTemplates).
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
			WithMetricTemplates(ReplicaSetMetricTemplates).
			WithTableTemplates(TableTemplates)
		replicaSets = []ReplicaSet{}
		selectors   = []func(labelledChild){}
	)
	result.Controls.AddControls(ScalingControls)

	for _, deployment := range deployments {
		selectors = append(selectors, match(
			deployment.Namespace(),
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
	Namespace() string
}

// Match parses the selectors and adds the target as a parent if the selector matches.
func match(namespace string, selector labels.Selector, topology, id string) func(labelledChild) {
	return func(c labelledChild) {
		if namespace == c.Namespace() && selector.Matches(labels.Set(c.Labels())) {
			c.AddParent(topology, id)
		}
	}
}

func (r *Reporter) podTopology(services []Service, replicaSets []ReplicaSet) (report.Topology, error) {
	var (
		pods = report.MakeTopology().
			WithMetadataTemplates(PodMetadataTemplates).
			WithMetricTemplates(PodMetricTemplates).
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
			service.Namespace(),
			service.Selector(),
			report.Service,
			report.MakeServiceNodeID(service.UID()),
		))
	}
	for _, replicaSet := range replicaSets {
		selectors = append(selectors, match(
			replicaSet.Namespace(),
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
