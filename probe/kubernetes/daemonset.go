package kubernetes

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apiv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	MisscheduledReplicas = "kubernetes_misscheduled_replicas"
)

// DaemonSet represents a Kubernetes daemonset
type DaemonSet interface {
	Meta
	Selector() (labels.Selector, error)
	GetNode() report.Node
}

type daemonSet struct {
	*apiv1beta1.DaemonSet
	Meta
}

// NewDaemonSet creates a new daemonset
func NewDaemonSet(d *apiv1beta1.DaemonSet) DaemonSet {
	return &daemonSet{
		DaemonSet: d,
		Meta:      meta{d.ObjectMeta},
	}
}

func (d *daemonSet) Selector() (labels.Selector, error) {
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (d *daemonSet) GetNode() report.Node {
	return d.MetaNode(report.MakeDaemonSetNodeID(d.UID())).WithLatests(map[string]string{
		DesiredReplicas:      fmt.Sprint(d.Status.DesiredNumberScheduled),
		Replicas:             fmt.Sprint(d.Status.CurrentNumberScheduled),
		MisscheduledReplicas: fmt.Sprint(d.Status.NumberMisscheduled),
		NodeType:             "Daemon Set",
	})
}
