package kubernetes

import (
	"fmt"

	"github.com/weaveworks/scope/report"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apiv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// These constants are keys used in node metadata
const (
	UpdatedReplicas     = report.KubernetesUpdatedReplicas
	AvailableReplicas   = report.KubernetesAvailableReplicas
	UnavailableReplicas = report.KubernetesUnavailableReplicas
	Strategy            = report.KubernetesStrategy
)

// Deployment represents a Kubernetes deployment
type Deployment interface {
	Meta
	Selector() (labels.Selector, error)
	GetNode(probeID string) report.Node
}

type deployment struct {
	*apiv1beta1.Deployment
	Meta
	Node *apiv1.Node
}

// NewDeployment creates a new Deployment
func NewDeployment(d *apiv1beta1.Deployment) Deployment {
	return &deployment{Deployment: d, Meta: meta{d.ObjectMeta}}
}

func (d *deployment) Selector() (labels.Selector, error) {
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (d *deployment) GetNode(probeID string) report.Node {
	// Spec.Replicas can be omitted, and the pointer will be nil. It defaults to 1.
	desiredReplicas := 1
	if d.Spec.Replicas != nil {
		desiredReplicas = int(*d.Spec.Replicas)
	}
	return d.MetaNode(report.MakeDeploymentNodeID(d.UID())).WithLatests(map[string]string{
		ObservedGeneration:    fmt.Sprint(d.Status.ObservedGeneration),
		DesiredReplicas:       fmt.Sprint(desiredReplicas),
		Replicas:              fmt.Sprint(d.Status.Replicas),
		UpdatedReplicas:       fmt.Sprint(d.Status.UpdatedReplicas),
		AvailableReplicas:     fmt.Sprint(d.Status.AvailableReplicas),
		UnavailableReplicas:   fmt.Sprint(d.Status.UnavailableReplicas),
		Strategy:              string(d.Spec.Strategy.Type),
		report.ControlProbeID: probeID,
		NodeType:              "Deployment",
	}).WithLatestActiveControls(ScaleUp, ScaleDown)
}
