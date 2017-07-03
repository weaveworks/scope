package kubernetes

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/labels"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/common/mtime"
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
	NodeType           = "kubernetes_node_type"
)

// Exposed for testing
var (
	PodMetadataTemplates = report.MetadataTemplates{
		State:            {ID: State, Label: "State", From: report.FromLatest, Priority: 2},
		IP:               {ID: IP, Label: "IP", From: report.FromLatest, Datatype: "ip", Priority: 3},
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: "number", Priority: 4},
		Namespace:        {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 5},
		Created:          {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 6},
	}

	PodMetricTemplates = docker.ContainerMetricTemplates

	ServiceMetadataTemplates = report.MetadataTemplates{
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:    {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		PublicIP:   {ID: PublicIP, Label: "Public IP", From: report.FromLatest, Datatype: "ip", Priority: 4},
		IP:         {ID: IP, Label: "Internal IP", From: report.FromLatest, Datatype: "ip", Priority: 5},
		report.Pod: {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	ServiceMetricTemplates = PodMetricTemplates

	DeploymentMetadataTemplates = report.MetadataTemplates{
		NodeType:           {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Datatype: "number", Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
		Strategy:           {ID: Strategy, Label: "Strategy", From: report.FromLatest, Priority: 7},
	}

	DeploymentMetricTemplates = ReplicaSetMetricTemplates

	ReplicaSetMetadataTemplates = report.MetadataTemplates{
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed Gen.", From: report.FromLatest, Datatype: "number", Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 6},
	}

	ReplicaSetMetricTemplates = PodMetricTemplates

	DaemonSetMetadataTemplates = report.MetadataTemplates{
		NodeType:        {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:       {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:         {ID: Created, Label: "Created", From: report.FromLatest, Datatype: "datetime", Priority: 3},
		DesiredReplicas: {ID: DesiredReplicas, Label: "Desired Replicas", From: report.FromLatest, Datatype: "number", Priority: 4},
		report.Pod:      {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: "number", Priority: 5},
	}

	DaemonSetMetricTemplates = PodMetricTemplates

	TableTemplates = report.TableTemplates{
		LabelPrefix: {
			ID:     LabelPrefix,
			Label:  "Kubernetes Labels",
			Type:   report.PropertyListType,
			Prefix: LabelPrefix,
		},
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
	kubeletPort     uint
}

// NewReporter makes a new Reporter
func NewReporter(client Client, pipes controls.PipeClient, probeID string, hostID string, probe *probe.Probe, handlerRegistry *controls.HandlerRegistry, kubeletPort uint) *Reporter {
	reporter := &Reporter{
		client:          client,
		pipes:           pipes,
		probeID:         probeID,
		probe:           probe,
		hostID:          hostID,
		handlerRegistry: handlerRegistry,
		kubeletPort:     kubeletPort,
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

		rpt.Container.Nodes[id] = n.WithParents(report.MakeSets().Add(
			report.Pod,
			report.MakeStringSet().Add(report.MakePodNodeID(uid)),
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
	daemonSetTopology, daemonSets, err := r.daemonSetTopology()
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
	podTopology, err := r.podTopology(services, replicaSets, daemonSets)
	if err != nil {
		return result, err
	}
	result.Pod = result.Pod.Merge(podTopology)
	result.Service = result.Service.Merge(serviceTopology)
	result.Host = result.Host.Merge(hostTopology)
	result.DaemonSet = result.DaemonSet.Merge(daemonSetTopology)
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
	localNetworks := report.MakeStringSet()
	for _, service := range services {
		localNetworks = localNetworks.Add(service.ClusterIP() + "/32")
	}
	node := report.MakeNode(report.MakeHostNodeID(r.hostID))
	node = node.WithSets(report.MakeSets().
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

func (r *Reporter) daemonSetTopology() (report.Topology, []DaemonSet, error) {
	daemonSets := []DaemonSet{}
	result := report.MakeTopology().
		WithMetadataTemplates(DaemonSetMetadataTemplates).
		WithMetricTemplates(DaemonSetMetricTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkDaemonSets(func(d DaemonSet) error {
		result = result.AddNode(d.GetNode())
		daemonSets = append(daemonSets, d)
		return nil
	})
	return result, daemonSets, err
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
		selector, err := deployment.Selector()
		if err != nil {
			return result, replicaSets, err
		}

		selectors = append(selectors, match(
			deployment.Namespace(),
			selector,
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

func (r *Reporter) podTopology(services []Service, replicaSets []ReplicaSet, daemonSets []DaemonSet) (report.Topology, error) {
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
		selector, err := replicaSet.Selector()
		if err != nil {
			return pods, err
		}
		selectors = append(selectors, match(
			replicaSet.Namespace(),
			selector,
			report.ReplicaSet,
			report.MakeReplicaSetNodeID(replicaSet.UID()),
		))
	}
	for _, daemonSet := range daemonSets {
		selector, err := daemonSet.Selector()
		if err != nil {
			return pods, err
		}
		selectors = append(selectors, match(
			daemonSet.Namespace(),
			selector,
			report.DaemonSet,
			report.MakeDaemonSetNodeID(daemonSet.UID()),
		))
	}

	// Obtain the local pods from kubelet since we only want to report those
	// for performance reasons.
	//
	// In theory a simpler approach would be to obtain the current NodeName
	// and filter local pods based on that. However that's hard since
	// 1. reconstructing the NodeName requires cloud provider credentials
	// 2. inferring the NodeName out of the hostname or system uuid is unreliable
	//    (uuids and hostnames can be duplicated across the cluster).
	localPodUIDs, errUIDs := GetLocalPodUIDs(fmt.Sprintf("127.0.0.1:%d", r.kubeletPort))
	if errUIDs != nil {
		log.Warnf("Cannot obtain local pods, reporting all (which may impact performance): %v", errUIDs)
	}
	err := r.client.WalkPods(func(p Pod) error {
		// filter out non-local pods
		if errUIDs == nil {
			if _, ok := localPodUIDs[p.UID()]; !ok {
				return nil
			}
		}
		for _, selector := range selectors {
			selector(p)
		}
		pods = pods.AddNode(p.GetNode(r.probeID))
		return nil
	})
	return pods, err
}
