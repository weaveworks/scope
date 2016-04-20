/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package kubectl

import (
	"errors"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
)

type ErrorReplicationControllers struct {
	testclient.FakeReplicationControllers
	invalid bool
}

func (c *ErrorReplicationControllers) Update(controller *api.ReplicationController) (*api.ReplicationController, error) {
	if c.invalid {
		return nil, kerrors.NewInvalid(api.Kind(controller.Kind), controller.Name, nil)
	}
	return nil, errors.New("Replication controller update failure")
}

type ErrorReplicationControllerClient struct {
	testclient.Fake
	invalid bool
}

func (c *ErrorReplicationControllerClient) ReplicationControllers(namespace string) client.ReplicationControllerInterface {
	return &ErrorReplicationControllers{testclient.FakeReplicationControllers{Fake: &c.Fake, Namespace: namespace}, c.invalid}
}

func TestReplicationControllerScaleRetry(t *testing.T) {
	fake := &ErrorReplicationControllerClient{Fake: testclient.Fake{}, invalid: false}
	scaler := ReplicationControllerScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	if err != nil {
		t.Errorf("Did not expect an error on update failure, got %v", err)
	}
	preconditions = ScalePrecondition{3, ""}
	scaleFunc = ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err = scaleFunc()
	if err == nil {
		t.Errorf("Expected error on precondition failure")
	}
}

func TestReplicationControllerScaleInvalid(t *testing.T) {
	fake := &ErrorReplicationControllerClient{Fake: testclient.Fake{}, invalid: true}
	scaler := ReplicationControllerScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	e, ok := err.(ScaleError)
	if err == nil || !ok || e.FailureType != ScaleUpdateInvalidFailure {
		t.Errorf("Expected error on invalid update failure, got %v", err)
	}
}

func TestReplicationControllerScale(t *testing.T) {
	fake := &testclient.Fake{}
	scaler := ReplicationControllerScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 2 {
		t.Errorf("unexpected actions: %v, expected 2 actions (get, update)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "replicationcontrollers" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-replicationController %s", actions[0], name)
	}
	if action, ok := actions[1].(testclient.UpdateAction); !ok || action.GetResource() != "replicationcontrollers" || action.GetObject().(*api.ReplicationController).Spec.Replicas != int(count) {
		t.Errorf("unexpected action %v, expected update-replicationController with replicas = %d", actions[1], count)
	}
}

func TestReplicationControllerScaleFailsPreconditions(t *testing.T) {
	fake := testclient.NewSimpleFake(&api.ReplicationController{
		Spec: api.ReplicationControllerSpec{
			Replicas: 10,
		},
	})
	scaler := ReplicationControllerScaler{fake}
	preconditions := ScalePrecondition{2, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 1 {
		t.Errorf("unexpected actions: %v, expected 1 action (get)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "replicationcontrollers" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-replicationController %s", actions[0], name)
	}
}

func TestValidateReplicationController(t *testing.T) {
	tests := []struct {
		preconditions ScalePrecondition
		controller    api.ReplicationController
		expectError   bool
		test          string
	}{
		{
			preconditions: ScalePrecondition{-1, ""},
			expectError:   false,
			test:          "defaults",
		},
		{
			preconditions: ScalePrecondition{-1, ""},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 10,
				},
			},
			expectError: false,
			test:        "defaults 2",
		},
		{
			preconditions: ScalePrecondition{0, ""},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 0,
				},
			},
			expectError: false,
			test:        "size matches",
		},
		{
			preconditions: ScalePrecondition{-1, "foo"},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 10,
				},
			},
			expectError: false,
			test:        "resource version matches",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 10,
				},
			},
			expectError: false,
			test:        "both match",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 20,
				},
			},
			expectError: true,
			test:        "size different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 10,
				},
			},
			expectError: true,
			test:        "version different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			controller: api.ReplicationController{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: api.ReplicationControllerSpec{
					Replicas: 20,
				},
			},
			expectError: true,
			test:        "both different",
		},
	}
	for _, test := range tests {
		err := test.preconditions.ValidateReplicationController(&test.controller)
		if err != nil && !test.expectError {
			t.Errorf("unexpected error: %v (%s)", err, test.test)
		}
		if err == nil && test.expectError {
			t.Errorf("unexpected non-error: %v (%s)", err, test.test)
		}
	}
}

type ErrorJobs struct {
	testclient.FakeJobs
	invalid bool
}

func (c *ErrorJobs) Update(job *extensions.Job) (*extensions.Job, error) {
	if c.invalid {
		return nil, kerrors.NewInvalid(extensions.Kind(job.Kind), job.Name, nil)
	}
	return nil, errors.New("Job update failure")
}

func (c *ErrorJobs) Get(name string) (*extensions.Job, error) {
	zero := 0
	return &extensions.Job{
		Spec: extensions.JobSpec{
			Parallelism: &zero,
		},
	}, nil
}

type ErrorJobClient struct {
	testclient.FakeExperimental
	invalid bool
}

func (c *ErrorJobClient) Jobs(namespace string) client.JobInterface {
	return &ErrorJobs{testclient.FakeJobs{Fake: &c.FakeExperimental, Namespace: namespace}, c.invalid}
}

func TestJobScaleRetry(t *testing.T) {
	fake := &ErrorJobClient{FakeExperimental: testclient.FakeExperimental{}, invalid: false}
	scaler := JobScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass != false {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	if err != nil {
		t.Errorf("Did not expect an error on update failure, got %v", err)
	}
	preconditions = ScalePrecondition{3, ""}
	scaleFunc = ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err = scaleFunc()
	if err == nil {
		t.Errorf("Expected error on precondition failure")
	}
}

func TestJobScale(t *testing.T) {
	fake := &testclient.FakeExperimental{Fake: &testclient.Fake{}}
	scaler := JobScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 2 {
		t.Errorf("unexpected actions: %v, expected 2 actions (get, update)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "jobs" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-replicationController %s", actions[0], name)
	}
	if action, ok := actions[1].(testclient.UpdateAction); !ok || action.GetResource() != "jobs" || *action.GetObject().(*extensions.Job).Spec.Parallelism != int(count) {
		t.Errorf("unexpected action %v, expected update-job with parallelism = %d", actions[1], count)
	}
}

func TestJobScaleInvalid(t *testing.T) {
	fake := &ErrorJobClient{FakeExperimental: testclient.FakeExperimental{}, invalid: true}
	scaler := JobScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	e, ok := err.(ScaleError)
	if err == nil || !ok || e.FailureType != ScaleUpdateInvalidFailure {
		t.Errorf("Expected error on invalid update failure, got %v", err)
	}
}

func TestJobScaleFailsPreconditions(t *testing.T) {
	ten := 10
	fake := testclient.NewSimpleFake(&extensions.Job{
		Spec: extensions.JobSpec{
			Parallelism: &ten,
		},
	})
	scaler := JobScaler{&testclient.FakeExperimental{fake}}
	preconditions := ScalePrecondition{2, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 1 {
		t.Errorf("unexpected actions: %v, expected 1 actions (get)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "jobs" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-job %s", actions[0], name)
	}
}

func TestValidateJob(t *testing.T) {
	zero, ten, twenty := 0, 10, 20
	tests := []struct {
		preconditions ScalePrecondition
		job           extensions.Job
		expectError   bool
		test          string
	}{
		{
			preconditions: ScalePrecondition{-1, ""},
			expectError:   false,
			test:          "defaults",
		},
		{
			preconditions: ScalePrecondition{-1, ""},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.JobSpec{
					Parallelism: &ten,
				},
			},
			expectError: false,
			test:        "defaults 2",
		},
		{
			preconditions: ScalePrecondition{0, ""},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.JobSpec{
					Parallelism: &zero,
				},
			},
			expectError: false,
			test:        "size matches",
		},
		{
			preconditions: ScalePrecondition{-1, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.JobSpec{
					Parallelism: &ten,
				},
			},
			expectError: false,
			test:        "resource version matches",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.JobSpec{
					Parallelism: &ten,
				},
			},
			expectError: false,
			test:        "both match",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.JobSpec{
					Parallelism: &twenty,
				},
			},
			expectError: true,
			test:        "size different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
			},
			expectError: true,
			test:        "parallelism nil",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: extensions.JobSpec{
					Parallelism: &ten,
				},
			},
			expectError: true,
			test:        "version different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			job: extensions.Job{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: extensions.JobSpec{
					Parallelism: &twenty,
				},
			},
			expectError: true,
			test:        "both different",
		},
	}
	for _, test := range tests {
		err := test.preconditions.ValidateJob(&test.job)
		if err != nil && !test.expectError {
			t.Errorf("unexpected error: %v (%s)", err, test.test)
		}
		if err == nil && test.expectError {
			t.Errorf("unexpected non-error: %v (%s)", err, test.test)
		}
	}
}

type ErrorDeployments struct {
	testclient.FakeDeployments
	invalid bool
}

func (c *ErrorDeployments) Update(deployment *extensions.Deployment) (*extensions.Deployment, error) {
	if c.invalid {
		return nil, kerrors.NewInvalid(extensions.Kind(deployment.Kind), deployment.Name, nil)
	}
	return nil, errors.New("deployment update failure")
}

func (c *ErrorDeployments) Get(name string) (*extensions.Deployment, error) {
	return &extensions.Deployment{
		Spec: extensions.DeploymentSpec{
			Replicas: 0,
		},
	}, nil
}

type ErrorDeploymentClient struct {
	testclient.FakeExperimental
	invalid bool
}

func (c *ErrorDeploymentClient) Deployments(namespace string) client.DeploymentInterface {
	return &ErrorDeployments{testclient.FakeDeployments{Fake: &c.FakeExperimental, Namespace: namespace}, c.invalid}
}

func TestDeploymentScaleRetry(t *testing.T) {
	fake := &ErrorDeploymentClient{FakeExperimental: testclient.FakeExperimental{Fake: &testclient.Fake{}}, invalid: false}
	scaler := &DeploymentScaler{fake}
	preconditions := &ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(scaler, preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass != false {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	if err != nil {
		t.Errorf("Did not expect an error on update failure, got %v", err)
	}
	preconditions = &ScalePrecondition{3, ""}
	scaleFunc = ScaleCondition(scaler, preconditions, namespace, name, count)
	pass, err = scaleFunc()
	if err == nil {
		t.Errorf("Expected error on precondition failure")
	}
}

func TestDeploymentScale(t *testing.T) {
	fake := &testclient.FakeExperimental{Fake: &testclient.Fake{}}
	scaler := DeploymentScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 2 {
		t.Errorf("unexpected actions: %v, expected 2 actions (get, update)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "deployments" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-replicationController %s", actions[0], name)
	}
	if action, ok := actions[1].(testclient.UpdateAction); !ok || action.GetResource() != "deployments" || action.GetObject().(*extensions.Deployment).Spec.Replicas != int(count) {
		t.Errorf("unexpected action %v, expected update-deployment with replicas = %d", actions[1], count)
	}
}

func TestDeploymentScaleInvalid(t *testing.T) {
	fake := &ErrorDeploymentClient{FakeExperimental: testclient.FakeExperimental{Fake: &testclient.Fake{}}, invalid: true}
	scaler := DeploymentScaler{fake}
	preconditions := ScalePrecondition{-1, ""}
	count := uint(3)
	name := "foo"
	namespace := "default"

	scaleFunc := ScaleCondition(&scaler, &preconditions, namespace, name, count)
	pass, err := scaleFunc()
	if pass {
		t.Errorf("Expected an update failure to return pass = false, got pass = %v", pass)
	}
	e, ok := err.(ScaleError)
	if err == nil || !ok || e.FailureType != ScaleUpdateInvalidFailure {
		t.Errorf("Expected error on invalid update failure, got %v", err)
	}
}

func TestDeploymentScaleFailsPreconditions(t *testing.T) {
	fake := testclient.NewSimpleFake(&extensions.Deployment{
		Spec: extensions.DeploymentSpec{
			Replicas: 10,
		},
	})
	scaler := DeploymentScaler{&testclient.FakeExperimental{fake}}
	preconditions := ScalePrecondition{2, ""}
	count := uint(3)
	name := "foo"
	scaler.Scale("default", name, count, &preconditions, nil, nil)

	actions := fake.Actions()
	if len(actions) != 1 {
		t.Errorf("unexpected actions: %v, expected 1 actions (get)", actions)
	}
	if action, ok := actions[0].(testclient.GetAction); !ok || action.GetResource() != "deployments" || action.GetName() != name {
		t.Errorf("unexpected action: %v, expected get-deployment %s", actions[0], name)
	}
}

func TestValidateDeployment(t *testing.T) {
	zero, ten, twenty := 0, 10, 20
	tests := []struct {
		preconditions ScalePrecondition
		deployment    extensions.Deployment
		expectError   bool
		test          string
	}{
		{
			preconditions: ScalePrecondition{-1, ""},
			expectError:   false,
			test:          "defaults",
		},
		{
			preconditions: ScalePrecondition{-1, ""},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: ten,
				},
			},
			expectError: false,
			test:        "defaults 2",
		},
		{
			preconditions: ScalePrecondition{0, ""},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: zero,
				},
			},
			expectError: false,
			test:        "size matches",
		},
		{
			preconditions: ScalePrecondition{-1, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: ten,
				},
			},
			expectError: false,
			test:        "resource version matches",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: ten,
				},
			},
			expectError: false,
			test:        "both match",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: twenty,
				},
			},
			expectError: true,
			test:        "size different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "foo",
				},
			},
			expectError: true,
			test:        "no replicas",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: ten,
				},
			},
			expectError: true,
			test:        "version different",
		},
		{
			preconditions: ScalePrecondition{10, "foo"},
			deployment: extensions.Deployment{
				ObjectMeta: api.ObjectMeta{
					ResourceVersion: "bar",
				},
				Spec: extensions.DeploymentSpec{
					Replicas: twenty,
				},
			},
			expectError: true,
			test:        "both different",
		},
	}
	for _, test := range tests {
		err := test.preconditions.ValidateDeployment(&test.deployment)
		if err != nil && !test.expectError {
			t.Errorf("unexpected error: %v (%s)", err, test.test)
		}
		if err == nil && test.expectError {
			t.Errorf("unexpected non-error: %v (%s)", err, test.test)
		}
	}
}
