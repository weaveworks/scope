package kubernetes

import (
	"strings"
	"time"

	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

// These constants are keys used in node metadata
const (
	PodID           = "kubernetes_pod_id"
	PodName         = "kubernetes_pod_name"
	PodCreated      = "kubernetes_pod_created"
	PodContainerIDs = "kubernetes_pod_container_ids"
	ServiceIDs      = "kubernetes_service_ids"
)

// Pod represents a Kubernetes pod
type Pod interface {
	ID() string
	Name() string
	Namespace() string
	ContainerIDs() []string
	Created() string
	AddServiceID(id string)
	Labels() labels.Labels
	GetNode() report.Node
}

type pod struct {
	*api.Pod
	serviceIDs []string
	Node       *api.Node
}

// NewPod creates a new Pod
func NewPod(p *api.Pod) Pod {
	return &pod{Pod: p}
}

func (p *pod) ID() string {
	return p.ObjectMeta.Namespace + "/" + p.ObjectMeta.Name
}

func (p *pod) Name() string {
	return p.ObjectMeta.Name
}

func (p *pod) Namespace() string {
	return p.ObjectMeta.Namespace
}

func (p *pod) Created() string {
	return p.ObjectMeta.CreationTimestamp.Format(time.RFC822)
}

func (p *pod) ContainerIDs() []string {
	ids := []string{}
	for _, container := range p.Status.ContainerStatuses {
		ids = append(ids, strings.TrimPrefix(container.ContainerID, "docker://"))
	}
	return ids
}

func (p *pod) Labels() labels.Labels {
	return labels.Set(p.ObjectMeta.Labels)
}

func (p *pod) AddServiceID(id string) {
	p.serviceIDs = append(p.serviceIDs, id)
}

func (p *pod) GetNode() report.Node {
	n := report.MakeNodeWith(map[string]string{
		PodID:           p.ID(),
		PodName:         p.Name(),
		Namespace:       p.Namespace(),
		PodCreated:      p.Created(),
		PodContainerIDs: strings.Join(p.ContainerIDs(), " "),
	})
	if len(p.serviceIDs) > 0 {
		n.Metadata[ServiceIDs] = strings.Join(p.serviceIDs, " ")
	}
	return n
}
