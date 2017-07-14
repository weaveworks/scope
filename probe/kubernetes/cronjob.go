package kubernetes

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	Schedule      = "kubernetes_schedule"
	Suspended     = "kubernetes_suspended"
	LastScheduled = "kubernetes_last_scheduled"
	ActiveJobs    = "kubernetes_active_jobs"
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
func NewCronJob(cj *batchv2alpha1.CronJob, jobs []*batchv1.Job) CronJob {
	myJobs := []*batchv1.Job{}
	for _, j := range jobs {
		for _, o := range cj.Status.Active {
			if j.UID == o.UID {
				myJobs = append(myJobs, j)
				break
			}
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
	return cj.MetaNode(report.MakeCronJobNodeID(cj.UID())).WithLatests(map[string]string{
		NodeType:      "Cron Job",
		Schedule:      cj.Spec.Schedule,
		Suspended:     fmt.Sprint(cj.Spec.Suspend != nil && *cj.Spec.Suspend), // nil -> false
		LastScheduled: cj.Status.LastScheduleTime.Format(time.RFC3339Nano),
		ActiveJobs:    fmt.Sprint(len(cj.jobs)),
	})
}
