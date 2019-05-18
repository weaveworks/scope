package kubernetes

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/weaveworks/scope/report"
)

//Job represents a Kubernetes job
type Job interface {
	Meta
	Selector() (labels.Selector, error)
	GetNode(probeID string) report.Node
}

type job struct {
	*batchv1.Job
	Meta
}

// NewJob creates a new job.
func NewJob(j *batchv1.Job) Job {
	return &job{
		Job:  j,
		Meta: meta{j.ObjectMeta},
	}
}

func (j *job) Selector() (labels.Selector, error) {
	selector, err := metav1.LabelSelectorAsSelector(j.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func (j *job) GetNode(probeID string) report.Node {
	latests := map[string]string{
		NodeType:              "Job",
		report.ControlProbeID: probeID,
	}
	return j.MetaNode(report.MakeJobNodeID(j.UID())).
		WithLatests(latests).
		WithLatestActiveControls(Describe)
}
