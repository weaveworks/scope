package kubernetes

import (
	"strconv"
	"strings"

	"github.com/weaveworks/scope/report"

	apiv1 "k8s.io/api/core/v1"
)

// These constants are keys used in node metadata
const (
	State           = report.KubernetesState
	IsInHostNetwork = report.KubernetesIsInHostNetwork
	RestartCount    = report.KubernetesRestartCount
)

// Pod represents a Kubernetes pod
type Pod interface {
	Meta
	AddParent(topology, id string)
	NodeName() string
	GetNode(probeID string) report.Node
	RestartCount() uint
	ContainerNames() []string
	VolumeClaimNames() []string
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
	if p.ObjectMeta.DeletionTimestamp != nil {
		return "Terminating"
	}

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

func (p *pod) VolumeClaimNames() []string {
	var claimNames []string
	for _, volume := range p.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			claimNames = append(claimNames, volume.VolumeSource.PersistentVolumeClaim.ClaimName)
		}
	}
	return claimNames
}

func (p *pod) GetNode(probeID string) report.Node {
	latests := map[string]string{
		State:                 p.State(),
		IP:                    p.Status.PodIP,
		report.ControlProbeID: probeID,
		RestartCount:          strconv.FormatUint(uint64(p.RestartCount()), 10),
	}

	if len(p.VolumeClaimNames()) > 0 {
		// PVC name consist of lower case alphanumeric characters, "-" or "."
		// and must start and end with an alphanumeric character.
		latests[VolumeClaim] = strings.Join(p.VolumeClaimNames(), report.ScopeDelim)
	}

	if p.Pod.Spec.HostNetwork {
		latests[IsInHostNetwork] = "true"
	}

	return p.MetaNode(report.MakePodNodeID(p.UID())).WithLatests(latests).
		WithParents(p.parents).
		WithLatestActiveControls(GetLogs, DeletePod, Describe)
}

func (p *pod) ContainerNames() []string {
	containerNames := make([]string, 0, len(p.Pod.Spec.Containers))
	for _, c := range p.Pod.Spec.Containers {
		containerNames = append(containerNames, c.Name)
	}
	return containerNames
}
