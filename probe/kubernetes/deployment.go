package kubernetes

import (
	"fmt"

	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/labels"
)

// These constants are keys used in node metadata
const (
	UpdatedReplicas     = "kubernetes_updated_replicas"
	AvailableReplicas   = "kubernetes_available_replicas"
	UnavailableReplicas = "kubernetes_unavailable_replicas"
	Strategy            = "kubernetes_strategy"
)

// Deployment represents a Kubernetes deployment
type Deployment interface {
	Meta
	Selector() labels.Selector
	GetNode(probeID string) report.Node
}

type deployment struct {
	*extensions.Deployment
	Meta
	Node *api.Node
}

// NewDeployment creates a new Deployment
func NewDeployment(d *extensions.Deployment) Deployment {
	return &deployment{Deployment: d, Meta: meta{d.ObjectMeta}}
}

func (d *deployment) Selector() labels.Selector {
	selector, err := unversioned.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		// TODO(paulbellamy): Remove the panic!
		panic(err)
	}
	return selector
}

func (d *deployment) GetNode(probeID string) report.Node {
	return d.MetaNode(report.MakeDeploymentNodeID(d.UID())).WithLatests(map[string]string{
		ObservedGeneration:    fmt.Sprint(d.Status.ObservedGeneration),
		DesiredReplicas:       fmt.Sprint(d.Spec.Replicas),
		Replicas:              fmt.Sprint(d.Status.Replicas),
		UpdatedReplicas:       fmt.Sprint(d.Status.UpdatedReplicas),
		AvailableReplicas:     fmt.Sprint(d.Status.AvailableReplicas),
		UnavailableReplicas:   fmt.Sprint(d.Status.UnavailableReplicas),
		Strategy:              string(d.Spec.Strategy.Type),
		report.ControlProbeID: probeID,
	}).WithLatestActiveControls(ScaleUp, ScaleDown)
}
