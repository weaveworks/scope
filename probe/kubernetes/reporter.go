package kubernetes

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	IP                 = report.KubernetesIP
	ObservedGeneration = report.KubernetesObservedGeneration
	Replicas           = report.KubernetesReplicas
	DesiredReplicas    = report.KubernetesDesiredReplicas
	NodeType           = report.KubernetesNodeType
	Type               = report.KubernetesType
	Ports              = report.KubernetesPorts
	VolumeClaim        = report.KubernetesVolumeClaim
	StorageClassName   = report.KubernetesStorageClassName
	AccessModes        = report.KubernetesAccessModes
	ReclaimPolicy      = report.KubernetesReclaimPolicy
	Status             = report.KubernetesStatus
	Message            = report.KubernetesMessage
	VolumeName         = report.KubernetesVolumeName
	Provisioner        = report.KubernetesProvisioner
	StorageDriver      = report.KubernetesStorageDriver
	VolumeSnapshotName = report.KubernetesVolumeSnapshotName
	SnapshotData       = report.KubernetesSnapshotData
	VolumeCapacity     = report.KubernetesVolumeCapacity
)

// Exposed for testing
var (
	PodMetadataTemplates = report.MetadataTemplates{
		State:            {ID: State, Label: "State", From: report.FromLatest, Priority: 2},
		IP:               {ID: IP, Label: "IP", From: report.FromLatest, Datatype: report.IP, Priority: 3},
		report.Container: {ID: report.Container, Label: "# Containers", From: report.FromCounters, Datatype: report.Number, Priority: 4},
		Namespace:        {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 5},
		Created:          {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 6},
		RestartCount:     {ID: RestartCount, Label: "Restart #", From: report.FromLatest, Priority: 7},
	}

	PodMetricTemplates = docker.ContainerMetricTemplates

	ServiceMetadataTemplates = report.MetadataTemplates{
		Namespace:  {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:    {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 3},
		PublicIP:   {ID: PublicIP, Label: "Public IP", From: report.FromLatest, Datatype: report.IP, Priority: 4},
		IP:         {ID: IP, Label: "Internal IP", From: report.FromLatest, Datatype: report.IP, Priority: 5},
		report.Pod: {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: report.Number, Priority: 6},
		Type:       {ID: Type, Label: "Type", From: report.FromLatest, Priority: 7},
		Ports:      {ID: Ports, Label: "Ports", From: report.FromLatest, Priority: 8},
	}

	ServiceMetricTemplates = PodMetricTemplates

	DeploymentMetadataTemplates = report.MetadataTemplates{
		NodeType:           {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed gen.", From: report.FromLatest, Datatype: report.Number, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired replicas", From: report.FromLatest, Datatype: report.Number, Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: report.Number, Priority: 6},
		Strategy:           {ID: Strategy, Label: "Strategy", From: report.FromLatest, Priority: 7},
	}

	DeploymentMetricTemplates = PodMetricTemplates

	DaemonSetMetadataTemplates = report.MetadataTemplates{
		NodeType:        {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:       {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:         {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 3},
		DesiredReplicas: {ID: DesiredReplicas, Label: "Desired replicas", From: report.FromLatest, Datatype: report.Number, Priority: 4},
		report.Pod:      {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: report.Number, Priority: 5},
	}

	DaemonSetMetricTemplates = PodMetricTemplates

	StatefulSetMetadataTemplates = report.MetadataTemplates{
		NodeType:           {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:          {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:            {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 3},
		ObservedGeneration: {ID: ObservedGeneration, Label: "Observed gen.", From: report.FromLatest, Datatype: report.Number, Priority: 4},
		DesiredReplicas:    {ID: DesiredReplicas, Label: "Desired replicas", From: report.FromLatest, Datatype: report.Number, Priority: 5},
		report.Pod:         {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: report.Number, Priority: 6},
	}

	StatefulSetMetricTemplates = PodMetricTemplates

	CronJobMetadataTemplates = report.MetadataTemplates{
		NodeType:      {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:     {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Created:       {ID: Created, Label: "Created", From: report.FromLatest, Datatype: report.DateTime, Priority: 3},
		Schedule:      {ID: Schedule, Label: "Schedule", From: report.FromLatest, Priority: 4},
		LastScheduled: {ID: LastScheduled, Label: "Last scheduled", From: report.FromLatest, Datatype: report.DateTime, Priority: 5},
		Suspended:     {ID: Suspended, Label: "Suspended", From: report.FromLatest, Priority: 6},
		ActiveJobs:    {ID: ActiveJobs, Label: "# Jobs", From: report.FromLatest, Datatype: report.Number, Priority: 7},
		report.Pod:    {ID: report.Pod, Label: "# Pods", From: report.FromCounters, Datatype: report.Number, Priority: 8},
	}

	CronJobMetricTemplates = PodMetricTemplates

	PersistentVolumeMetadataTemplates = report.MetadataTemplates{
		NodeType:         {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		VolumeClaim:      {ID: VolumeClaim, Label: "Volume claim", From: report.FromLatest, Priority: 2},
		StorageClassName: {ID: StorageClassName, Label: "Storage class", From: report.FromLatest, Priority: 3},
		AccessModes:      {ID: AccessModes, Label: "Access modes", From: report.FromLatest, Priority: 5},
		Status:           {ID: Status, Label: "Status", From: report.FromLatest, Priority: 6},
		StorageDriver:    {ID: StorageDriver, Label: "Storage driver", From: report.FromLatest, Priority: 7},
	}

	PersistentVolumeClaimMetadataTemplates = report.MetadataTemplates{
		NodeType:         {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:        {ID: Namespace, Label: "Namespace", From: report.FromLatest, Priority: 2},
		Status:           {ID: Status, Label: "Status", From: report.FromLatest, Priority: 3},
		VolumeName:       {ID: VolumeName, Label: "Volume", From: report.FromLatest, Priority: 4},
		StorageClassName: {ID: StorageClassName, Label: "Storage class", From: report.FromLatest, Priority: 5},
		VolumeCapacity:   {ID: VolumeCapacity, Label: "Capacity", From: report.FromLatest, Priority: 6},
	}

	StorageClassMetadataTemplates = report.MetadataTemplates{
		NodeType:    {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Name:        {ID: Name, Label: "Name", From: report.FromLatest, Priority: 2},
		Provisioner: {ID: Provisioner, Label: "Provisioner", From: report.FromLatest, Priority: 3},
	}

	VolumeSnapshotMetadataTemplates = report.MetadataTemplates{
		NodeType:     {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		Namespace:    {ID: Namespace, Label: "Name", From: report.FromLatest, Priority: 2},
		VolumeClaim:  {ID: VolumeClaim, Label: "Persistent volume claim", From: report.FromLatest, Priority: 3},
		SnapshotData: {ID: SnapshotData, Label: "Volume snapshot data", From: report.FromLatest, Priority: 4},
	}

	VolumeSnapshotDataMetadataTemplates = report.MetadataTemplates{
		NodeType:           {ID: NodeType, Label: "Type", From: report.FromLatest, Priority: 1},
		VolumeName:         {ID: VolumeName, Label: "Persistent volume", From: report.FromLatest, Priority: 2},
		VolumeSnapshotName: {ID: VolumeSnapshotName, Label: "Volume snapshot", From: report.FromLatest, Priority: 3},
	}

	TableTemplates = report.TableTemplates{
		LabelPrefix: {
			ID:     LabelPrefix,
			Label:  "Kubernetes labels",
			Type:   report.PropertyListType,
			Prefix: LabelPrefix,
		},
	}

	ScalingControls = []report.Control{
		{
			ID:    ScaleDown,
			Human: "Scale down",
			Icon:  "fa fa-minus",
			Rank:  0,
		},
		{
			ID:    ScaleUp,
			Human: "Scale up",
			Icon:  "fa fa-plus",
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
	hostname        string
	handlerRegistry *controls.HandlerRegistry
	nodeName        string
	kubeletPort     uint
}

// NewReporter makes a new Reporter
func NewReporter(client Client, pipes controls.PipeClient, probeID string, hostname string, probe *probe.Probe, handlerRegistry *controls.HandlerRegistry, nodeName string, kubeletPort uint) *Reporter {
	reporter := &Reporter{
		client:          client,
		pipes:           pipes,
		probeID:         probeID,
		probe:           probe,
		hostname:        hostname,
		handlerRegistry: handlerRegistry,
		nodeName:        nodeName,
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
	// filter out non-local pods, if we have been given a node name to report on
	if r.nodeName != "" && pod.NodeName() != r.nodeName {
		return
	}
	switch e {
	case ADD:
		rpt := report.MakeReport()
		rpt.Shortcut = true
		rpt.Pod.AddNode(pod.GetNode(r.probeID).WithLatests(map[string]string{report.Host: r.hostname}))
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
	return strings.Contains(imageName, "google_containers/pause") ||
		strings.Contains(imageName, "k8s.gcr.io/pause") ||
		strings.Contains(imageName, "eks/pause")
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

// Tagger adds pod parents to container nodes.
type Tagger struct {
}

// Name of this tagger, for metrics gathering
func (Tagger) Name() string { return "K8s" }

// Tag adds pod parents to container nodes.
func (r *Tagger) Tag(rpt report.Report) (report.Report, error) {
	for id, n := range rpt.Container.Nodes {
		uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
		if !ok {
			continue
		}

		// Tag the pause containers with "does-not-make-connections"
		if isPauseContainer(n, rpt) {
			n = n.WithLatest(report.DoesNotMakeConnections, mtime.Now(), "")
		}

		rpt.Container.Nodes[id] = n.WithParent(report.Pod, report.MakePodNodeID(uid))
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
	//daemonSetTopology, daemonSets, err := r.daemonSetTopology()
	//if err != nil {
	//	return result, err
	//}
	//statefulSetTopology, statefulSets, err := r.statefulSetTopology()
	//if err != nil {
	//	return result, err
	//}
	//cronJobTopology, cronJobs, err := r.cronJobTopology()
	//if err != nil {
	//	return result, err
	//}
	//deploymentTopology, deployments, err := r.deploymentTopology()
	//if err != nil {
	//	return result, err
	//}
	podTopology, err := r.podTopology(services)
	if err != nil {
		return result, err
	}

	//namespaceTopology, err := r.namespaceTopology()
	//if err != nil {
	//	return result, err
	//}
	//persistentVolumeTopology, _, err := r.persistentVolumeTopology()
	//if err != nil {
	//	return result, err
	//}
	//persistentVolumeClaimTopology, _, err := r.persistentVolumeClaimTopology()
	//if err != nil {
	//	return result, err
	//}
	//storageClassTopology, _, err := r.storageClassTopology()
	//if err != nil {
	//	return result, err
	//}
	//volumeSnapshotTopology, _, err := r.volumeSnapshotTopology()
	//if err != nil {
	//	return result, err
	//}
	//volumeSnapshotDataTopology, _, err := r.volumeSnapshotDataTopology()
	//if err != nil {
	//	return result, err
	//}
	result.Pod = result.Pod.Merge(podTopology)
	result.Service = result.Service.Merge(serviceTopology)
	//result.DaemonSet = result.DaemonSet.Merge(daemonSetTopology)
	//result.StatefulSet = result.StatefulSet.Merge(statefulSetTopology)
	//result.CronJob = result.CronJob.Merge(cronJobTopology)
	//result.Deployment = result.Deployment.Merge(deploymentTopology)
	//result.Namespace = result.Namespace.Merge(namespaceTopology)
	//result.PersistentVolume = result.PersistentVolume.Merge(persistentVolumeTopology)
	//result.PersistentVolumeClaim = result.PersistentVolumeClaim.Merge(persistentVolumeClaimTopology)
	//result.StorageClass = result.StorageClass.Merge(storageClassTopology)
	//result.VolumeSnapshot = result.VolumeSnapshot.Merge(volumeSnapshotTopology)
	//result.VolumeSnapshotData = result.VolumeSnapshotData.Merge(volumeSnapshotDataTopology)
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
		result.AddNode(s.GetNode(r.probeID))
		services = append(services, s)
		return nil
	})
	return result, services, err
}

func (r *Reporter) deploymentTopology() (report.Topology, []Deployment, error) {
	var (
		result = report.MakeTopology().
			WithMetadataTemplates(DeploymentMetadataTemplates).
			WithMetricTemplates(DeploymentMetricTemplates).
			WithTableTemplates(TableTemplates)
		deployments = []Deployment{}
	)
	result.Controls.AddControls(ScalingControls)

	err := r.client.WalkDeployments(func(d Deployment) error {
		result.AddNode(d.GetNode(r.probeID))
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
		result.AddNode(d.GetNode(r.probeID))
		daemonSets = append(daemonSets, d)
		return nil
	})
	return result, daemonSets, err
}

func (r *Reporter) statefulSetTopology() (report.Topology, []StatefulSet, error) {
	statefulSets := []StatefulSet{}
	result := report.MakeTopology().
		WithMetadataTemplates(StatefulSetMetadataTemplates).
		WithMetricTemplates(StatefulSetMetricTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkStatefulSets(func(s StatefulSet) error {
		result.AddNode(s.GetNode(r.probeID))
		statefulSets = append(statefulSets, s)
		return nil
	})
	return result, statefulSets, err
}

func (r *Reporter) cronJobTopology() (report.Topology, []CronJob, error) {
	cronJobs := []CronJob{}
	result := report.MakeTopology().
		WithMetadataTemplates(CronJobMetadataTemplates).
		WithMetricTemplates(CronJobMetricTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkCronJobs(func(c CronJob) error {
		result.AddNode(c.GetNode(r.probeID))
		cronJobs = append(cronJobs, c)
		return nil
	})
	return result, cronJobs, err
}

func (r *Reporter) persistentVolumeTopology() (report.Topology, []PersistentVolume, error) {
	persistentVolumes := []PersistentVolume{}
	result := report.MakeTopology().
		WithMetadataTemplates(PersistentVolumeMetadataTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkPersistentVolumes(func(p PersistentVolume) error {
		result.AddNode(p.GetNode())
		persistentVolumes = append(persistentVolumes, p)
		return nil
	})
	return result, persistentVolumes, err
}

func (r *Reporter) persistentVolumeClaimTopology() (report.Topology, []PersistentVolumeClaim, error) {
	persistentVolumeClaims := []PersistentVolumeClaim{}
	result := report.MakeTopology().
		WithMetadataTemplates(PersistentVolumeClaimMetadataTemplates).
		WithTableTemplates(TableTemplates)
	result.Controls.AddControl(report.Control{
		ID:    CreateVolumeSnapshot,
		Human: "Create snapshot",
		Icon:  "fa fa-camera",
		Rank:  0,
	})
	err := r.client.WalkPersistentVolumeClaims(func(p PersistentVolumeClaim) error {
		result.AddNode(p.GetNode(r.probeID))
		persistentVolumeClaims = append(persistentVolumeClaims, p)
		return nil
	})
	return result, persistentVolumeClaims, err
}

func (r *Reporter) storageClassTopology() (report.Topology, []StorageClass, error) {
	storageClasses := []StorageClass{}
	result := report.MakeTopology().
		WithMetadataTemplates(StorageClassMetadataTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkStorageClasses(func(p StorageClass) error {
		result.AddNode(p.GetNode())
		storageClasses = append(storageClasses, p)
		return nil
	})
	return result, storageClasses, err
}

func (r *Reporter) volumeSnapshotTopology() (report.Topology, []VolumeSnapshot, error) {
	volumeSnapshots := []VolumeSnapshot{}
	result := report.MakeTopology().
		WithMetadataTemplates(VolumeSnapshotMetadataTemplates).
		WithTableTemplates(TableTemplates)
	result.Controls.AddControl(report.Control{
		ID:    CloneVolumeSnapshot,
		Human: "Clone snapshot",
		Icon:  "far fa-clone",
		Rank:  0,
	})
	result.Controls.AddControl(report.Control{
		ID:    DeleteVolumeSnapshot,
		Human: "Delete",
		Icon:  "far fa-trash-alt",
		Rank:  1,
	})
	err := r.client.WalkVolumeSnapshots(func(p VolumeSnapshot) error {
		result.AddNode(p.GetNode(r.probeID))
		volumeSnapshots = append(volumeSnapshots, p)
		return nil
	})
	return result, volumeSnapshots, err
}

func (r *Reporter) volumeSnapshotDataTopology() (report.Topology, []VolumeSnapshotData, error) {
	volumeSnapshotData := []VolumeSnapshotData{}
	result := report.MakeTopology().
		WithMetadataTemplates(VolumeSnapshotDataMetadataTemplates).
		WithTableTemplates(TableTemplates)
	err := r.client.WalkVolumeSnapshotData(func(p VolumeSnapshotData) error {
		result.AddNode(p.GetNode(r.probeID))
		volumeSnapshotData = append(volumeSnapshotData, p)
		return nil
	})
	return result, volumeSnapshotData, err
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

func (r *Reporter) podTopology(services []Service) (report.Topology, error) {
	var (
		pods = report.MakeTopology().
			WithMetadataTemplates(PodMetadataTemplates).
			WithMetricTemplates(PodMetricTemplates).
			WithTableTemplates(TableTemplates)
		selectors = []func(labelledChild){}
	)
	//pods.Controls.AddControl(report.Control{
	//	ID:    GetLogs,
	//	Human: "Get logs",
	//	Icon:  "fa fa-desktop",
	//	Rank:  0,
	//})
	//pods.Controls.AddControl(report.Control{
	//	ID:    DeletePod,
	//	Human: "Delete",
	//	Icon:  "far fa-trash-alt",
	//	Rank:  1,
	//})
	for _, service := range services {
		selectors = append(selectors, match(
			service.Namespace(),
			service.Selector(),
			report.Service,
			report.MakeServiceNodeID(service.UID()),
		))
	}
	//for _, deployment := range deployments {
	//	selector, err := deployment.Selector()
	//	if err != nil {
	//		return pods, err
	//	}
	//	selectors = append(selectors, match(
	//		deployment.Namespace(),
	//		selector,
	//		report.Deployment,
	//		report.MakeDeploymentNodeID(deployment.UID()),
	//	))
	//}
	//for _, daemonSet := range daemonSets {
	//	selector, err := daemonSet.Selector()
	//	if err != nil {
	//		return pods, err
	//	}
	//	selectors = append(selectors, match(
	//		daemonSet.Namespace(),
	//		selector,
	//		report.DaemonSet,
	//		report.MakeDaemonSetNodeID(daemonSet.UID()),
	//	))
	//}
	//for _, statefulSet := range statefulSets {
	//	selector, err := statefulSet.Selector()
	//	if err != nil {
	//		return pods, err
	//	}
	//	selectors = append(selectors, match(
	//		statefulSet.Namespace(),
	//		selector,
	//		report.StatefulSet,
	//		report.MakeStatefulSetNodeID(statefulSet.UID()),
	//	))
	//}
	//for _, cronJob := range cronJobs {
	//	cronJobSelectors, err := cronJob.Selectors()
	//	if err != nil {
	//		return pods, err
	//	}
	//	for _, selector := range cronJobSelectors {
	//		selectors = append(selectors, match(
	//			cronJob.Namespace(),
	//			selector,
	//			report.CronJob,
	//			report.MakeCronJobNodeID(cronJob.UID()),
	//		))
	//	}
	//}

	var localPodUIDs map[string]struct{}
	if r.nodeName == "" && r.kubeletPort != 0 {
		// We don't know the node name: fall back to obtaining the local pods from kubelet
		var err error
		localPodUIDs, err = GetLocalPodUIDs(fmt.Sprintf("127.0.0.1:%d", r.kubeletPort))
		if err != nil {
			log.Warnf("No node name and cannot obtain local pods, reporting all (which may impact performance): %v", err)
		}
	}
	err := r.client.WalkPods(func(p Pod) error {
		// filter out non-local pods: we only want to report local ones for performance reasons.
		if r.nodeName != "" {
			if p.NodeName() != r.nodeName {
				return nil
			}
		} else if localPodUIDs != nil {
			if _, ok := localPodUIDs[p.UID()]; !ok {
				return nil
			}
		}
		for _, selector := range selectors {
			selector(p)
		}
		pods.AddNode(p.GetNode(r.probeID).WithLatests(map[string]string{report.Host: r.hostname}))
		return nil
	})
	return pods, err
}

func (r *Reporter) namespaceTopology() (report.Topology, error) {
	result := report.MakeTopology()
	err := r.client.WalkNamespaces(func(ns NamespaceResource) error {
		result.AddNode(ns.GetNode())
		return nil
	})
	return result, err
}
