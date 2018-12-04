package kubernetes

import (
	"fmt"

	"github.com/weaveworks/scope/report"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	Succeeded = report.KubernetesJobsSucceeded
)

//Job interface holds methods related to K8s Job
type Job interface {
	Meta
	GetNode() report.Node
}

type job struct {
	*batchv1.Job
	Meta
}

//NewJob returns fresh instance of job
func NewJob(p *batchv1.Job) Job {
	return &job{Job: p, Meta: meta{p.ObjectMeta}}
}

// GetNode returns job as Scope Node
func (p *job) GetNode() report.Node {
	return p.MetaNode(report.MakeJobNodeID(p.UID())).WithLatests(map[string]string{
		NodeType:  "Job",
		Namespace: p.GetNamespace(),
		Succeeded: fmt.Sprint(p.Status.Succeeded),
		// State: p.Status
	})
}
