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
	FullyLabeledReplicas = "kubernetes_fully_labeled_replicas"
)

// ReplicaSet represents a Kubernetes replica set
type ReplicaSet interface {
	Meta
	Selector() (labels.Selector, error)
	AddParent(topology, id string)
	GetNode(probeID string) report.Node
}

type replicaSet struct {
	*extensions.ReplicaSet
	Meta
	parents report.Sets
	Node    *api.Node
}

// NewReplicaSet creates a new ReplicaSet
func NewReplicaSet(r *extensions.ReplicaSet) ReplicaSet {
	return &replicaSet{
		ReplicaSet: r,
		Meta:       meta{r.ObjectMeta},
		parents:    report.MakeSets(),
	}
}

func (r *replicaSet) Selector() (labels.Selector, error) {
	selector, err := unversioned.LabelSelectorAsSelector(r.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (r *replicaSet) AddParent(topology, id string) {
	r.parents = r.parents.Add(topology, report.MakeStringSet(id))
}

func (r *replicaSet) GetNode(probeID string) report.Node {
	return r.MetaNode(report.MakeReplicaSetNodeID(r.UID())).WithLatests(map[string]string{
		ObservedGeneration:    fmt.Sprint(r.Status.ObservedGeneration),
		Replicas:              fmt.Sprint(r.Status.Replicas),
		DesiredReplicas:       fmt.Sprint(r.Spec.Replicas),
		FullyLabeledReplicas:  fmt.Sprint(r.Status.FullyLabeledReplicas),
		report.ControlProbeID: probeID,
	}).WithParents(r.parents).WithLatestActiveControls(ScaleUp, ScaleDown)
}
