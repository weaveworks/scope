package kubernetes

import (
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	Schedule      = report.KubernetesSchedule
	Suspended     = report.KubernetesSuspended
	LastScheduled = report.KubernetesLastScheduled
	ActiveJobs    = report.KubernetesActiveJobs
)

// CronJob represents a Kubernetes cron job
type CronJob interface {
	Meta
	Selectors() ([]labels.Selector, error)
	GetNode() report.Node
}

type cronJob struct {
	*batchv2alpha1.CronJob
	Meta
	jobs []*batchv1.Job
}

// NewCronJob creates a new cron job. jobs should be all jobs, which will be filtered
// for those matching this cron job.
func NewCronJob(cj *batchv2alpha1.CronJob, jobs map[types.UID]*batchv1.Job) CronJob {
	myJobs := []*batchv1.Job{}
	for _, o := range cj.Status.Active {
		if j, ok := jobs[o.UID]; ok {
			myJobs = append(myJobs, j)
		}
	}
	return &cronJob{
		CronJob: cj,
		Meta:    meta{cj.ObjectMeta},
		jobs:    myJobs,
	}
}

func (cj *cronJob) Selectors() ([]labels.Selector, error) {
	selectors := []labels.Selector{}
	for _, j := range cj.jobs {
		selector, err := metav1.LabelSelectorAsSelector(j.Spec.Selector)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selector)
	}
	return selectors, nil
}

func (cj *cronJob) GetNode() report.Node {
	latest := map[string]string{
		NodeType:   "CronJob",
		Schedule:   cj.Spec.Schedule,
		Suspended:  fmt.Sprint(cj.Spec.Suspend != nil && *cj.Spec.Suspend), // nil -> false
		ActiveJobs: fmt.Sprint(len(cj.jobs)),
	}
	if cj.Status.LastScheduleTime != nil {
		latest[LastScheduled] = cj.Status.LastScheduleTime.Format(time.RFC3339Nano)
	}
	return cj.MetaNode(report.MakeCronJobNodeID(cj.UID())).WithLatests(latest)
}
