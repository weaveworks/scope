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
	ReplicaSetID                   = "kubernetes_replica_set_id"
	ReplicaSetName                 = "kubernetes_replica_set_name"
	ReplicaSetCreated              = "kubernetes_replica_set_created"
	ReplicaSetObservedGeneration   = "kubernetes_replica_set_observed_generation"
	ReplicaSetReplicas             = "kubernetes_replica_set_replicas"
	ReplicaSetDesiredReplicas      = "kubernetes_replica_set_desired_replicas"
	ReplicaSetFullyLabeledReplicas = "kubernetes_replica_set_fully_labeled_replicas"
	ReplicaSetLabelPrefix          = "kubernetes_replica_set_labels_"
)

// ReplicaSet represents a Kubernetes replica_set
type ReplicaSet interface {
	Meta
	Selector() labels.Selector
	GetNode(probeID string) report.Node
}

type replicaSet struct {
	*extensions.ReplicaSet
	Meta
	Node *api.Node
}

// NewReplicaSet creates a new ReplicaSet
func NewReplicaSet(d *extensions.ReplicaSet) ReplicaSet {
	return &replicaSet{ReplicaSet: d, Meta: meta{d.ObjectMeta}}
}

func (d *replicaSet) Selector() labels.Selector {
	selector, err := unversioned.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		panic(err)
	}
	return selector
}

func (d *replicaSet) GetNode(probeID string) report.Node {
	n := report.MakeNodeWith(report.MakeReplicaSetNodeID(d.UID()), map[string]string{
		ReplicaSetID:                   d.ID(),
		ReplicaSetName:                 d.Name(),
		Namespace:                      d.Namespace(),
		ReplicaSetCreated:              d.Created(),
		ReplicaSetObservedGeneration:   fmt.Sprint(d.Status.ObservedGeneration),
		ReplicaSetReplicas:             fmt.Sprint(d.Status.Replicas),
		ReplicaSetDesiredReplicas:      fmt.Sprint(d.Spec.Replicas),
		ReplicaSetFullyLabeledReplicas: fmt.Sprint(d.Status.FullyLabeledReplicas),
		report.ControlProbeID:          probeID,
	})
	n = n.AddTable(ReplicaSetLabelPrefix, d.ObjectMeta.Labels)
	return n
}
