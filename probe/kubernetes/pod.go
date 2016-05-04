package kubernetes

import (
	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
)

// These constants are keys used in node metadata
const (
	PodID          = "kubernetes_pod_id"
	PodName        = "kubernetes_pod_name"
	PodCreated     = "kubernetes_pod_created"
	PodState       = "kubernetes_pod_state"
	PodLabelPrefix = "kubernetes_pod_labels_"
	PodIP          = "kubernetes_pod_ip"

	StateDeleted = "deleted"
)

// Pod represents a Kubernetes pod
type Pod interface {
	Meta
	AddParent(topology, id string)
	NodeName() string
	GetNode(probeID string) report.Node
}

type pod struct {
	*api.Pod
	Meta
	parents report.Sets
	Node    *api.Node
}

// NewPod creates a new Pod
func NewPod(p *api.Pod) Pod {
	return &pod{Pod: p, Meta: meta{p.ObjectMeta}, parents: report.MakeSets()}
}

func (p *pod) UID() string {
	// Work around for master pod not reporting the right UID.
	if hash, ok := p.ObjectMeta.Annotations["kubernetes.io/config.hash"]; ok {
		return hash
	}
	return p.Meta.UID()
}

func (p *pod) AddParent(topology, id string) {
	p.parents = p.parents.Add(topology, report.MakeStringSet(id))
}

func (p *pod) State() string {
	return string(p.Status.Phase)
}

func (p *pod) NodeName() string {
	return p.Spec.NodeName
}

func (p *pod) GetNode(probeID string) report.Node {
	n := report.MakeNodeWith(report.MakePodNodeID(p.UID()), map[string]string{
		PodID:      p.ID(),
		PodName:    p.Name(),
		Namespace:  p.Namespace(),
		PodCreated: p.Created(),
		PodState:   p.State(),
		PodIP:      p.Status.PodIP,
		report.ControlProbeID: probeID,
	}).WithParents(p.parents)
	n = n.AddTable(PodLabelPrefix, p.ObjectMeta.Labels)
	n = n.WithControls(GetLogs, DeletePod)
	return n
}
