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
	FullyLabeledReplicas = report.KubernetesFullyLabeledReplicas
)

// ReplicaSet represents a Kubernetes replica set
type ReplicaSet interface {
	Meta
	Selector() (labels.Selector, error)
	AddParent(topology, id string)
	GetNode(probeID string) report.Node
}

type replicaSet struct {
	*apiv1beta1.ReplicaSet
	Meta
	parents report.Sets
	Node    *apiv1.Node
}

// NewReplicaSet creates a new ReplicaSet
func NewReplicaSet(r *apiv1beta1.ReplicaSet) ReplicaSet {
	return &replicaSet{
		ReplicaSet: r,
		Meta:       meta{r.ObjectMeta},
		parents:    report.MakeSets(),
	}
}

func (r *replicaSet) Selector() (labels.Selector, error) {
	selector, err := metav1.LabelSelectorAsSelector(r.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (r *replicaSet) AddParent(topology, id string) {
	r.parents = r.parents.Add(topology, report.MakeStringSet(id))
}

func (r *replicaSet) GetNode(probeID string) report.Node {
	// Spec.Replicas can be omitted, and the pointer will be nil. It defaults to 1.
	desiredReplicas := 1
	if r.Spec.Replicas != nil {
		desiredReplicas = int(*r.Spec.Replicas)
	}
	return r.MetaNode(report.MakeReplicaSetNodeID(r.UID())).WithLatests(map[string]string{
		ObservedGeneration:    fmt.Sprint(r.Status.ObservedGeneration),
		Replicas:              fmt.Sprint(r.Status.Replicas),
		DesiredReplicas:       fmt.Sprint(desiredReplicas),
		FullyLabeledReplicas:  fmt.Sprint(r.Status.FullyLabeledReplicas),
		report.ControlProbeID: probeID,
	}).WithParents(r.parents).WithLatestActiveControls(ScaleUp, ScaleDown)
}
