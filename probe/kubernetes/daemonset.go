package kubernetes

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	MisscheduledReplicas = "kubernetes_misscheduled_replicas"
)

// DaemonSet represents a Kubernetes daemonset
type DaemonSet interface {
	Meta
	Selector() labels.Selector
	GetNode() report.Node
}

type daemonSet struct {
	*extensions.DaemonSet
	Meta
}

// NewDaemonSet creates a new daemonset
func NewDaemonSet(d *extensions.DaemonSet) DaemonSet {
	return &daemonSet{
		DaemonSet: d,
		Meta:      meta{d.ObjectMeta},
	}
}

func (d *daemonSet) Selector() labels.Selector {
	selector, err := unversioned.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		// TODO(paulbellamy): Remove the panic!
		panic(err)
	}
	return selector
}

func (d *daemonSet) GetNode() report.Node {
	return d.MetaNode(report.MakeDaemonSetNodeID(d.UID())).WithLatests(map[string]string{
		DesiredReplicas:      fmt.Sprint(d.Status.DesiredNumberScheduled),
		Replicas:             fmt.Sprint(d.Status.CurrentNumberScheduled),
		MisscheduledReplicas: fmt.Sprint(d.Status.NumberMisscheduled),
	})
}
