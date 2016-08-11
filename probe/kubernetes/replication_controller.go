package kubernetes

import (
	"fmt"

	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

// ReplicationController represents a Kubernetes replication controller
type ReplicationController interface {
	Meta
	Selector() labels.Selector
	AddParent(topology, id string)
	GetNode(probeID string) report.Node
}

type replicationController struct {
	*api.ReplicationController
	Meta
	parents report.Sets
	Node    *api.Node
}

// NewReplicationController creates a new ReplicationController
func NewReplicationController(r *api.ReplicationController) ReplicationController {
	return &replicationController{
		ReplicationController: r,
		Meta:    meta{r.ObjectMeta},
		parents: report.MakeSets(),
	}
}

func (r *replicationController) Selector() labels.Selector {
	if r.Spec.Selector == nil {
		return labels.Nothing()
	}
	return labels.SelectorFromSet(labels.Set(r.Spec.Selector))
}

func (r *replicationController) AddParent(topology, id string) {
	r.parents = r.parents.Add(topology, report.MakeStringSet(id))
}

func (r *replicationController) GetNode(probeID string) report.Node {
	return r.MetaNode(report.MakeReplicaSetNodeID(r.UID())).WithLatests(map[string]string{
		ObservedGeneration:    fmt.Sprint(r.Status.ObservedGeneration),
		Replicas:              fmt.Sprint(r.Status.Replicas),
		DesiredReplicas:       fmt.Sprint(r.Spec.Replicas),
		FullyLabeledReplicas:  fmt.Sprint(r.Status.FullyLabeledReplicas),
		report.ControlProbeID: probeID,
	}).WithParents(r.parents).WithLatestActiveControls(ScaleUp, ScaleDown)
}
