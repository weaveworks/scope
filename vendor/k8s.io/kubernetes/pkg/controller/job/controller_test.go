/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package job

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/rand"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

var alwaysReady = func() bool { return true }

func newJob(parallelism, completions int) *extensions.Job {
	j := &extensions.Job{
		ObjectMeta: api.ObjectMeta{
			Name:      "foobar",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.JobSpec{
			Selector: &unversioned.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{Image: "foo/bar"},
					},
				},
			},
		},
	}
	// Special case: -1 for either completions or parallelism means leave nil (negative is not allowed
	// in practice by validation.
	if completions >= 0 {
		j.Spec.Completions = &completions
	} else {
		j.Spec.Completions = nil
	}
	if parallelism >= 0 {
		j.Spec.Parallelism = &parallelism
	} else {
		j.Spec.Parallelism = nil
	}
	return j
}

func getKey(job *extensions.Job, t *testing.T) string {
	if key, err := controller.KeyFunc(job); err != nil {
		t.Errorf("Unexpected error getting key for job %v: %v", job.Name, err)
		return ""
	} else {
		return key
	}
}

// create count pods with the given phase for the given job
func newPodList(count int, status api.PodPhase, job *extensions.Job) []api.Pod {
	pods := []api.Pod{}
	for i := 0; i < count; i++ {
		newPod := api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name:      fmt.Sprintf("pod-%v", rand.String(10)),
				Labels:    job.Spec.Selector.MatchLabels,
				Namespace: job.Namespace,
			},
			Status: api.PodStatus{Phase: status},
		}
		pods = append(pods, newPod)
	}
	return pods
}

func TestControllerSyncJob(t *testing.T) {
	testCases := map[string]struct {
		// job setup
		parallelism int
		completions int

		// pod setup
		podControllerError error
		activePods         int
		succeededPods      int
		failedPods         int

		// expectations
		expectedCreations int
		expectedDeletions int
		expectedActive    int
		expectedSucceeded int
		expectedFailed    int
		expectedComplete  bool
	}{
		"job start": {
			2, 5,
			nil, 0, 0, 0,
			2, 0, 2, 0, 0, false,
		},
		"WQ job start": {
			2, -1,
			nil, 0, 0, 0,
			2, 0, 2, 0, 0, false,
		},
		"correct # of pods": {
			2, 5,
			nil, 2, 0, 0,
			0, 0, 2, 0, 0, false,
		},
		"WQ job: correct # of pods": {
			2, -1,
			nil, 2, 0, 0,
			0, 0, 2, 0, 0, false,
		},
		"too few active pods": {
			2, 5,
			nil, 1, 1, 0,
			1, 0, 2, 1, 0, false,
		},
		"too few active pods with a dynamic job": {
			2, -1,
			nil, 1, 0, 0,
			1, 0, 2, 0, 0, false,
		},
		"too few active pods, with controller error": {
			2, 5,
			fmt.Errorf("Fake error"), 1, 1, 0,
			0, 0, 1, 1, 0, false,
		},
		"too many active pods": {
			2, 5,
			nil, 3, 0, 0,
			0, 1, 2, 0, 0, false,
		},
		"too many active pods, with controller error": {
			2, 5,
			fmt.Errorf("Fake error"), 3, 0, 0,
			0, 0, 3, 0, 0, false,
		},
		"failed pod": {
			2, 5,
			nil, 1, 1, 1,
			1, 0, 2, 1, 1, false,
		},
		"job finish": {
			2, 5,
			nil, 0, 5, 0,
			0, 0, 0, 5, 0, true,
		},
		"WQ job finishing": {
			2, -1,
			nil, 1, 1, 0,
			0, 0, 1, 1, 0, false,
		},
		"WQ job all finished": {
			2, -1,
			nil, 0, 2, 0,
			0, 0, 0, 2, 0, true,
		},
		"WQ job all finished despite one failure": {
			2, -1,
			nil, 0, 1, 1,
			0, 0, 0, 1, 1, true,
		},
		"more active pods than completions": {
			2, 5,
			nil, 10, 0, 0,
			0, 8, 2, 0, 0, false,
		},
		"status change": {
			2, 5,
			nil, 2, 2, 0,
			0, 0, 2, 2, 0, false,
		},
	}

	for name, tc := range testCases {
		// job manager setup
		clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
		manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
		fakePodControl := controller.FakePodControl{Err: tc.podControllerError}
		manager.podControl = &fakePodControl
		manager.podStoreSynced = alwaysReady
		var actual *extensions.Job
		manager.updateHandler = func(job *extensions.Job) error {
			actual = job
			return nil
		}

		// job & pods setup
		job := newJob(tc.parallelism, tc.completions)
		manager.jobStore.Store.Add(job)
		for _, pod := range newPodList(tc.activePods, api.PodRunning, job) {
			manager.podStore.Store.Add(&pod)
		}
		for _, pod := range newPodList(tc.succeededPods, api.PodSucceeded, job) {
			manager.podStore.Store.Add(&pod)
		}
		for _, pod := range newPodList(tc.failedPods, api.PodFailed, job) {
			manager.podStore.Store.Add(&pod)
		}

		// run
		err := manager.syncJob(getKey(job, t))
		if err != nil {
			t.Errorf("%s: unexpected error when syncing jobs %v", name, err)
		}

		// validate created/deleted pods
		if len(fakePodControl.Templates) != tc.expectedCreations {
			t.Errorf("%s: unexpected number of creates.  Expected %d, saw %d\n", name, tc.expectedCreations, len(fakePodControl.Templates))
		}
		if len(fakePodControl.DeletePodName) != tc.expectedDeletions {
			t.Errorf("%s: unexpected number of deletes.  Expected %d, saw %d\n", name, tc.expectedDeletions, len(fakePodControl.DeletePodName))
		}
		// validate status
		if actual.Status.Active != tc.expectedActive {
			t.Errorf("%s: unexpected number of active pods.  Expected %d, saw %d\n", name, tc.expectedActive, actual.Status.Active)
		}
		if actual.Status.Succeeded != tc.expectedSucceeded {
			t.Errorf("%s: unexpected number of succeeded pods.  Expected %d, saw %d\n", name, tc.expectedSucceeded, actual.Status.Succeeded)
		}
		if actual.Status.Failed != tc.expectedFailed {
			t.Errorf("%s: unexpected number of failed pods.  Expected %d, saw %d\n", name, tc.expectedFailed, actual.Status.Failed)
		}
		if actual.Status.StartTime == nil {
			t.Errorf("%s: .status.startTime was not set", name)
		}
		// validate conditions
		if tc.expectedComplete && !getCondition(actual, extensions.JobComplete) {
			t.Errorf("%s: expected completion condition.  Got %#v", name, actual.Status.Conditions)
		}
	}
}

func TestSyncJobPastDeadline(t *testing.T) {
	testCases := map[string]struct {
		// job setup
		parallelism           int
		completions           int
		activeDeadlineSeconds int64
		startTime             int64

		// pod setup
		activePods    int
		succeededPods int
		failedPods    int

		// expectations
		expectedDeletions int
		expectedActive    int
		expectedSucceeded int
		expectedFailed    int
	}{
		"activeDeadlineSeconds less than single pod execution": {
			1, 1, 10, 15,
			1, 0, 0,
			1, 0, 0, 1,
		},
		"activeDeadlineSeconds bigger than single pod execution": {
			1, 2, 10, 15,
			1, 1, 0,
			1, 0, 1, 1,
		},
		"activeDeadlineSeconds times-out before any pod starts": {
			1, 1, 10, 10,
			0, 0, 0,
			0, 0, 0, 0,
		},
	}

	for name, tc := range testCases {
		// job manager setup
		clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
		manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
		fakePodControl := controller.FakePodControl{}
		manager.podControl = &fakePodControl
		manager.podStoreSynced = alwaysReady
		var actual *extensions.Job
		manager.updateHandler = func(job *extensions.Job) error {
			actual = job
			return nil
		}

		// job & pods setup
		job := newJob(tc.parallelism, tc.completions)
		job.Spec.ActiveDeadlineSeconds = &tc.activeDeadlineSeconds
		start := unversioned.Unix(unversioned.Now().Time.Unix()-tc.startTime, 0)
		job.Status.StartTime = &start
		manager.jobStore.Store.Add(job)
		for _, pod := range newPodList(tc.activePods, api.PodRunning, job) {
			manager.podStore.Store.Add(&pod)
		}
		for _, pod := range newPodList(tc.succeededPods, api.PodSucceeded, job) {
			manager.podStore.Store.Add(&pod)
		}
		for _, pod := range newPodList(tc.failedPods, api.PodFailed, job) {
			manager.podStore.Store.Add(&pod)
		}

		// run
		err := manager.syncJob(getKey(job, t))
		if err != nil {
			t.Errorf("%s: unexpected error when syncing jobs %v", name, err)
		}

		// validate created/deleted pods
		if len(fakePodControl.Templates) != 0 {
			t.Errorf("%s: unexpected number of creates.  Expected 0, saw %d\n", name, len(fakePodControl.Templates))
		}
		if len(fakePodControl.DeletePodName) != tc.expectedDeletions {
			t.Errorf("%s: unexpected number of deletes.  Expected %d, saw %d\n", name, tc.expectedDeletions, len(fakePodControl.DeletePodName))
		}
		// validate status
		if actual.Status.Active != tc.expectedActive {
			t.Errorf("%s: unexpected number of active pods.  Expected %d, saw %d\n", name, tc.expectedActive, actual.Status.Active)
		}
		if actual.Status.Succeeded != tc.expectedSucceeded {
			t.Errorf("%s: unexpected number of succeeded pods.  Expected %d, saw %d\n", name, tc.expectedSucceeded, actual.Status.Succeeded)
		}
		if actual.Status.Failed != tc.expectedFailed {
			t.Errorf("%s: unexpected number of failed pods.  Expected %d, saw %d\n", name, tc.expectedFailed, actual.Status.Failed)
		}
		if actual.Status.StartTime == nil {
			t.Errorf("%s: .status.startTime was not set", name)
		}
		// validate conditions
		if !getCondition(actual, extensions.JobFailed) {
			t.Errorf("%s: expected fail condition.  Got %#v", name, actual.Status.Conditions)
		}
	}
}

func getCondition(job *extensions.Job, condition extensions.JobConditionType) bool {
	for _, v := range job.Status.Conditions {
		if v.Type == condition && v.Status == api.ConditionTrue {
			return true
		}
	}
	return false
}

func TestSyncPastDeadlineJobFinished(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.podStoreSynced = alwaysReady
	var actual *extensions.Job
	manager.updateHandler = func(job *extensions.Job) error {
		actual = job
		return nil
	}

	job := newJob(1, 1)
	activeDeadlineSeconds := int64(10)
	job.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds
	start := unversioned.Unix(unversioned.Now().Time.Unix()-15, 0)
	job.Status.StartTime = &start
	job.Status.Conditions = append(job.Status.Conditions, newCondition(extensions.JobFailed, "DeadlineExceeded", "Job was active longer than specified deadline"))
	manager.jobStore.Store.Add(job)
	err := manager.syncJob(getKey(job, t))
	if err != nil {
		t.Errorf("Unexpected error when syncing jobs %v", err)
	}
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
	if actual != nil {
		t.Error("Unexpected job modification")
	}
}

func TestSyncJobComplete(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.podStoreSynced = alwaysReady

	job := newJob(1, 1)
	job.Status.Conditions = append(job.Status.Conditions, newCondition(extensions.JobComplete, "", ""))
	manager.jobStore.Store.Add(job)
	err := manager.syncJob(getKey(job, t))
	if err != nil {
		t.Fatalf("Unexpected error when syncing jobs %v", err)
	}
	uncastJob, _, err := manager.jobStore.Store.Get(job)
	if err != nil {
		t.Fatalf("Unexpected error when trying to get job from the store: %v", err)
	}
	actual := uncastJob.(*extensions.Job)
	// Verify that after syncing a complete job, the conditions are the same.
	if got, expected := len(actual.Status.Conditions), 1; got != expected {
		t.Fatalf("Unexpected job status conditions amount; expected %d, got %d", expected, got)
	}
}

func TestSyncJobDeleted(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.podStoreSynced = alwaysReady
	manager.updateHandler = func(job *extensions.Job) error { return nil }
	job := newJob(2, 2)
	err := manager.syncJob(getKey(job, t))
	if err != nil {
		t.Errorf("Unexpected error when syncing jobs %v", err)
	}
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
}

func TestSyncJobUpdateRequeue(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.podStoreSynced = alwaysReady
	manager.updateHandler = func(job *extensions.Job) error { return fmt.Errorf("Fake error") }
	job := newJob(2, 2)
	manager.jobStore.Store.Add(job)
	err := manager.syncJob(getKey(job, t))
	if err != nil {
		t.Errorf("Unxpected error when syncing jobs, got %v", err)
	}
	t.Log("Waiting for a job in the queue")
	key, _ := manager.queue.Get()
	expectedKey := getKey(job, t)
	if key != expectedKey {
		t.Errorf("Expected requeue of job with key %s got %s", expectedKey, key)
	}
}

func TestJobPodLookup(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	manager.podStoreSynced = alwaysReady
	testCases := []struct {
		job *extensions.Job
		pod *api.Pod

		expectedName string
	}{
		// pods without labels don't match any job
		{
			job: &extensions.Job{
				ObjectMeta: api.ObjectMeta{Name: "basic"},
			},
			pod: &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "foo1", Namespace: api.NamespaceAll},
			},
			expectedName: "",
		},
		// matching labels, different namespace
		{
			job: &extensions.Job{
				ObjectMeta: api.ObjectMeta{Name: "foo"},
				Spec: extensions.JobSpec{
					Selector: &unversioned.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			pod: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name:      "foo2",
					Namespace: "ns",
					Labels:    map[string]string{"foo": "bar"},
				},
			},
			expectedName: "",
		},
		// matching ns and labels returns
		{
			job: &extensions.Job{
				ObjectMeta: api.ObjectMeta{Name: "bar", Namespace: "ns"},
				Spec: extensions.JobSpec{
					Selector: &unversioned.LabelSelector{
						MatchExpressions: []unversioned.LabelSelectorRequirement{
							{
								Key:      "foo",
								Operator: unversioned.LabelSelectorOpIn,
								Values:   []string{"bar"},
							},
						},
					},
				},
			},
			pod: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name:      "foo3",
					Namespace: "ns",
					Labels:    map[string]string{"foo": "bar"},
				},
			},
			expectedName: "bar",
		},
	}
	for _, tc := range testCases {
		manager.jobStore.Add(tc.job)
		if job := manager.getPodJob(tc.pod); job != nil {
			if tc.expectedName != job.Name {
				t.Errorf("Got job %+v expected %+v", job.Name, tc.expectedName)
			}
		} else if tc.expectedName != "" {
			t.Errorf("Expected a job %v pod %v, found none", tc.expectedName, tc.pod.Name)
		}
	}
}

type FakeJobExpectations struct {
	*controller.ControllerExpectations
	satisfied    bool
	expSatisfied func()
}

func (fe FakeJobExpectations) SatisfiedExpectations(controllerKey string) bool {
	fe.expSatisfied()
	return fe.satisfied
}

// TestSyncJobExpectations tests that a pod cannot sneak in between counting active pods
// and checking expectations.
func TestSyncJobExpectations(t *testing.T) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.podStoreSynced = alwaysReady
	manager.updateHandler = func(job *extensions.Job) error { return nil }

	job := newJob(2, 2)
	manager.jobStore.Store.Add(job)
	pods := newPodList(2, api.PodPending, job)
	manager.podStore.Store.Add(&pods[0])

	manager.expectations = FakeJobExpectations{
		controller.NewControllerExpectations(), true, func() {
			// If we check active pods before checking expectataions, the job
			// will create a new replica because it doesn't see this pod, but
			// has fulfilled its expectations.
			manager.podStore.Store.Add(&pods[1])
		},
	}
	manager.syncJob(getKey(job, t))
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
}

type FakeWatcher struct {
	w *watch.FakeWatcher
	*testclient.Fake
}

func TestWatchJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	fakeWatch := watch.NewFake()
	clientset.PrependWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	manager.podStoreSynced = alwaysReady

	var testJob extensions.Job
	received := make(chan struct{})

	// The update sent through the fakeWatcher should make its way into the workqueue,
	// and eventually into the syncHandler.
	manager.syncHandler = func(key string) error {

		obj, exists, err := manager.jobStore.Store.GetByKey(key)
		if !exists || err != nil {
			t.Errorf("Expected to find job under key %v", key)
		}
		job := *obj.(*extensions.Job)
		if !api.Semantic.DeepDerivative(job, testJob) {
			t.Errorf("Expected %#v, but got %#v", testJob, job)
		}
		close(received)
		return nil
	}
	// Start only the job watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go manager.jobController.Run(stopCh)
	go wait.Until(manager.worker, 10*time.Millisecond, stopCh)

	// We're sending new job to see if it reaches syncHandler.
	testJob.Name = "foo"
	fakeWatch.Add(&testJob)
	t.Log("Waiting for job to reach syncHandler")
	<-received
}

func TestIsJobFinished(t *testing.T) {
	job := &extensions.Job{
		Status: extensions.JobStatus{
			Conditions: []extensions.JobCondition{{
				Type:   extensions.JobComplete,
				Status: api.ConditionTrue,
			}},
		},
	}

	if !isJobFinished(job) {
		t.Error("Job was expected to be finished")
	}

	job.Status.Conditions[0].Status = api.ConditionFalse
	if isJobFinished(job) {
		t.Error("Job was not expected to be finished")
	}

	job.Status.Conditions[0].Status = api.ConditionUnknown
	if isJobFinished(job) {
		t.Error("Job was not expected to be finished")
	}
}

func TestWatchPods(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	fakeWatch := watch.NewFake()
	clientset.PrependWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))
	manager := NewJobController(clientset, controller.NoResyncPeriodFunc)
	manager.podStoreSynced = alwaysReady

	// Put one job and one pod into the store
	testJob := newJob(2, 2)
	manager.jobStore.Store.Add(testJob)
	received := make(chan struct{})
	// The pod update sent through the fakeWatcher should figure out the managing job and
	// send it into the syncHandler.
	manager.syncHandler = func(key string) error {

		obj, exists, err := manager.jobStore.Store.GetByKey(key)
		if !exists || err != nil {
			t.Errorf("Expected to find job under key %v", key)
		}
		job := obj.(*extensions.Job)
		if !api.Semantic.DeepDerivative(job, testJob) {
			t.Errorf("\nExpected %#v,\nbut got %#v", testJob, job)
		}
		close(received)
		return nil
	}
	// Start only the pod watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method for the right job.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go manager.podController.Run(stopCh)
	go wait.Until(manager.worker, 10*time.Millisecond, stopCh)

	pods := newPodList(1, api.PodRunning, testJob)
	testPod := pods[0]
	testPod.Status.Phase = api.PodFailed
	fakeWatch.Add(&testPod)

	t.Log("Waiting for pod to reach syncHandler")
	<-received
}
