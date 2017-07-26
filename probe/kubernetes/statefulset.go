package kubernetes

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/weaveworks/scope/report"
)

// StatefulSet represents a Kubernetes statefulset
type StatefulSet interface {
	Meta
	Selector() (labels.Selector, error)
	GetNode() report.Node
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

func (s *statefulSet) GetNode() report.Node {
	desiredReplicas := 1
	if s.Spec.Replicas != nil {
		desiredReplicas = int(*s.Spec.Replicas)
	}
	latests := map[string]string{
		NodeType:        "StatefulSet",
		DesiredReplicas: fmt.Sprint(desiredReplicas),
		Replicas:        fmt.Sprint(s.Status.Replicas),
	}
	if s.Status.ObservedGeneration != nil {
		latests[ObservedGeneration] = fmt.Sprint(*s.Status.ObservedGeneration)
	}
	return s.MetaNode(report.MakeStatefulSetNodeID(s.UID())).WithLatests(latests)
}
