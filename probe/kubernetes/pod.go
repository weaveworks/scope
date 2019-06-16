package kubernetes

import (
	"strconv"

	"github.com/weaveworks/scope/report"

	apiv1 "k8s.io/api/core/v1"
)

// These constants are keys used in node metadata
const (
	State           = report.KubernetesState
	IsInHostNetwork = report.KubernetesIsInHostNetwork
	RestartCount    = report.KubernetesRestartCount
	PodContainerIDs = report.KubernetesPodContainerIDs
)

// Pod states we handle specially
const (
	StateDeleted = "deleted"
	StateFailed  = "Failed"
)

// Pod represents a Kubernetes pod
type Pod interface {
	Meta
	AddParent(topology, id string)
	NodeName() string
	GetNode(probeID string) report.Node
	RestartCount() uint
	ContainerNames() []string
	ContainerIDs() []string
}

type pod struct {
	*apiv1.Pod
	Meta
	parents report.Sets
	Node    *apiv1.Node
}

// NewPod creates a new Pod
func NewPod(p *apiv1.Pod) Pod {
	return &pod{
		Pod:     p,
		Meta:    meta{p.ObjectMeta},
		parents: report.MakeSets(),
	}
}

func (p *pod) UID() string {
	// Work around for master pod not reporting the right UID.
	if hash, ok := p.ObjectMeta.Annotations["kubernetes.io/config.hash"]; ok {
		return hash
	}
	return p.Meta.UID()
}

func (p *pod) AddParent(topology, id string) {
	p.parents = p.parents.AddString(topology, id)
}

func (p *pod) State() string {
	return string(p.Status.Phase)
}

func (p *pod) NodeName() string {
	return p.Spec.NodeName
}

func (p *pod) RestartCount() uint {
	count := uint(0)
	for _, cs := range p.Status.ContainerStatuses {
		count += uint(cs.RestartCount)
	}
	return count
}

func (p *pod) VolumeClaimName() string {
	var claimName string
	for _, volume := range p.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			claimName = volume.VolumeSource.PersistentVolumeClaim.ClaimName
			break
		}
	}
	return claimName
}

func (p *pod) GetNode(probeID string) report.Node {
	latests := map[string]string{
		State: p.State(),
		IP:    p.Status.PodIP,
		report.ControlProbeID: probeID,
		RestartCount:          strconv.FormatUint(uint64(p.RestartCount()), 10),
	}

	if p.VolumeClaimName() != "" {
		latests[VolumeClaim] = p.VolumeClaimName()
	}

	if p.Pod.Spec.HostNetwork {
		latests[IsInHostNetwork] = "true"
		latests[PodContainerIDs] = report.MakeChildPIDs(p.ContainerIDs())
	}

	return p.MetaNode(report.MakePodNodeID(p.UID())).WithLatests(latests).
		WithParents(p.parents).
		WithLatestActiveControls(GetLogs, DeletePod)
}

func (p *pod) ContainerNames() []string {
	containerNames := make([]string, 0, len(p.Pod.Spec.Containers))
	for _, c := range p.Pod.Spec.Containers {
		containerNames = append(containerNames, c.Name)
	}
	return containerNames
}

func (p *pod) ContainerIDs() []string {
	containerStatuses := make([]string, 0, len(p.Pod.Status.ContainerStatuses))
	for _, c := range p.Pod.Status.ContainerStatuses {
		containerStatuses = append(containerStatuses, c.ContainerID)
	}
	return containerStatuses
}
