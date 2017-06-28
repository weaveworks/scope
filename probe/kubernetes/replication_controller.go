package kubernetes

import (
	"fmt"

	"github.com/weaveworks/scope/report"

	"k8s.io/apimachinery/pkg/labels"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// ReplicationController represents a Kubernetes replication controller
type ReplicationController interface {
	Meta
	Selector() (labels.Selector, error)
	AddParent(topology, id string)
	GetNode(probeID string) report.Node
}

// replicationController implements both ReplicationController and ReplicaSet
type replicationController struct {
	*apiv1.ReplicationController
	Meta
	parents report.Sets
	Node    *apiv1.Node
}

// NewReplicationController creates a new ReplicationController
func NewReplicationController(r *apiv1.ReplicationController) ReplicationController {
	return &replicationController{
		ReplicationController: r,
		Meta:    meta{r.ObjectMeta},
		parents: report.MakeSets(),
	}
}

func (r *replicationController) Selector() (labels.Selector, error) {
	if r.Spec.Selector == nil {
		return labels.Nothing(), nil
	}
	return labels.SelectorFromSet(labels.Set(r.Spec.Selector)), nil
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
