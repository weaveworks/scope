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
	DeploymentID                  = "kubernetes_deployment_id"
	DeploymentName                = "kubernetes_deployment_name"
	DeploymentCreated             = "kubernetes_deployment_created"
	DeploymentObservedGeneration  = "kubernetes_deployment_observed_generation"
	DeploymentDesiredReplicas     = "kubernetes_deployment_desired_replicas"
	DeploymentReplicas            = "kubernetes_deployment_replicas"
	DeploymentUpdatedReplicas     = "kubernetes_deployment_updated_replicas"
	DeploymentAvailableReplicas   = "kubernetes_deployment_available_replicas"
	DeploymentUnavailableReplicas = "kubernetes_deployment_unavailable_replicas"
	DeploymentLabelPrefix         = "kubernetes_deployment_labels_"
	DeploymentStrategy            = "kubernetes_deployment_strategy"
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
		panic(err)
	}
	return selector
}

func (d *deployment) GetNode(probeID string) report.Node {
	n := report.MakeNodeWith(report.MakeDeploymentNodeID(d.UID()), map[string]string{
		DeploymentID:                  d.ID(),
		DeploymentName:                d.Name(),
		Namespace:                     d.Namespace(),
		DeploymentCreated:             d.Created(),
		DeploymentObservedGeneration:  fmt.Sprint(d.Status.ObservedGeneration),
		DeploymentDesiredReplicas:     fmt.Sprint(d.Spec.Replicas),
		DeploymentReplicas:            fmt.Sprint(d.Status.Replicas),
		DeploymentUpdatedReplicas:     fmt.Sprint(d.Status.UpdatedReplicas),
		DeploymentAvailableReplicas:   fmt.Sprint(d.Status.AvailableReplicas),
		DeploymentUnavailableReplicas: fmt.Sprint(d.Status.UnavailableReplicas),
		DeploymentStrategy:            string(d.Spec.Strategy.Type),
		report.ControlProbeID:         probeID,
	})
	n = n.AddTable(DeploymentLabelPrefix, d.ObjectMeta.Labels)
	return n
}
