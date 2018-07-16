package kubernetes

import (
	"fmt"

	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/weaveworks/scope/report"
)

// StatefulSet represents a Kubernetes statefulset
type StatefulSet interface {
	Meta
	Selector() (labels.Selector, error)
	GetNode(probeID string) report.Node
}

type statefulSet struct {
	*v1beta1.StatefulSet
	Meta
}

// NewStatefulSet creates a new statefulset
func NewStatefulSet(s *v1beta1.StatefulSet) StatefulSet {
	return &statefulSet{
		StatefulSet: s,
		Meta:        meta{s.ObjectMeta},
	}
}

func (s *statefulSet) Selector() (labels.Selector, error) {
	selector, err := metav1.LabelSelectorAsSelector(s.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (s *statefulSet) GetNode(probeID string) report.Node {
	desiredReplicas := 1
	if s.Spec.Replicas != nil {
		desiredReplicas = int(*s.Spec.Replicas)
	}
	latests := []string{
		NodeType, "StatefulSet",
		DesiredReplicas, fmt.Sprint(desiredReplicas),
		Replicas, fmt.Sprint(s.Status.Replicas),
		report.ControlProbeID, probeID,
	}
	if s.Status.ObservedGeneration != nil {
		latests = append(latests, ObservedGeneration, fmt.Sprint(*s.Status.ObservedGeneration))
	}
	return s.MetaNode(report.MakeStatefulSetNodeID(s.UID())).WithLatests(latests...)
}
