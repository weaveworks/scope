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

// If you make changes to this file, you should also make the corresponding change in ReplicaSet.

package replication

import (
	"fmt"
	"math/rand"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/securitycontext"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/sets"
	utiltesting "k8s.io/kubernetes/pkg/util/testing"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

var alwaysReady = func() bool { return true }

func getKey(rc *api.ReplicationController, t *testing.T) string {
	if key, err := controller.KeyFunc(rc); err != nil {
		t.Errorf("Unexpected error getting key for rc %v: %v", rc.Name, err)
		return ""
	} else {
		return key
	}
}

func newReplicationController(replicas int) *api.ReplicationController {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{APIVersion: testapi.Default.GroupVersion().String()},
		ObjectMeta: api.ObjectMeta{
			UID:             util.NewUUID(),
			Name:            "foobar",
			Namespace:       api.NamespaceDefault,
			ResourceVersion: "18",
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: replicas,
			Selector: map[string]string{"foo": "bar"},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{
						"name": "foo",
						"type": "production",
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Image: "foo/bar",
							TerminationMessagePath: api.TerminationMessagePathDefault,
							ImagePullPolicy:        api.PullIfNotPresent,
							SecurityContext:        securitycontext.ValidSecurityContextWithContainerDefaults(),
						},
					},
					RestartPolicy: api.RestartPolicyAlways,
					DNSPolicy:     api.DNSDefault,
					NodeSelector: map[string]string{
						"baz": "blah",
					},
				},
			},
		},
	}
	return rc
}

// create count pods with the given phase for the given rc (same selectors and namespace), and add them to the store.
func newPodList(store cache.Store, count int, status api.PodPhase, rc *api.ReplicationController, name string) *api.PodList {
	pods := []api.Pod{}
	for i := 0; i < count; i++ {
		newPod := api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name:      fmt.Sprintf("%s%d", name, i),
				Labels:    rc.Spec.Selector,
				Namespace: rc.Namespace,
			},
			Status: api.PodStatus{Phase: status},
		}
		if store != nil {
			store.Add(&newPod)
		}
		pods = append(pods, newPod)
	}
	return &api.PodList{
		Items: pods,
	}
}

func validateSyncReplication(t *testing.T, fakePodControl *controller.FakePodControl, expectedCreates, expectedDeletes int) {
	if len(fakePodControl.Templates) != expectedCreates {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", expectedCreates, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != expectedDeletes {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", expectedDeletes, len(fakePodControl.DeletePodName))
	}
}

func replicationControllerResourceName() string {
	return "replicationcontrollers"
}

type serverResponse struct {
	statusCode int
	obj        interface{}
}

func TestSyncReplicationControllerDoesNothing(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	// 2 running pods, a controller with 2 replicas, sync is a no-op
	controllerSpec := newReplicationController(2)
	manager.rcStore.Store.Add(controllerSpec)
	newPodList(manager.podStore.Store, 2, api.PodRunning, controllerSpec, "pod")

	manager.podControl = &fakePodControl
	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 0, 0)
}

func TestSyncReplicationControllerDeletes(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady
	manager.podControl = &fakePodControl

	// 2 running pods and a controller with 1 replica, one pod delete expected
	controllerSpec := newReplicationController(1)
	manager.rcStore.Store.Add(controllerSpec)
	newPodList(manager.podStore.Store, 2, api.PodRunning, controllerSpec, "pod")

	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 0, 1)
}

func TestDeleteFinalStateUnknown(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady
	manager.podControl = &fakePodControl

	received := make(chan string)
	manager.syncHandler = func(key string) error {
		received <- key
		return nil
	}

	// The DeletedFinalStateUnknown object should cause the rc manager to insert
	// the controller matching the selectors of the deleted pod into the work queue.
	controllerSpec := newReplicationController(1)
	manager.rcStore.Store.Add(controllerSpec)
	pods := newPodList(nil, 1, api.PodRunning, controllerSpec, "pod")
	manager.deletePod(cache.DeletedFinalStateUnknown{Key: "foo", Obj: &pods.Items[0]})

	go manager.worker()

	expected := getKey(controllerSpec, t)
	select {
	case key := <-received:
		if key != expected {
			t.Errorf("Unexpected sync all for rc %v, expected %v", key, expected)
		}
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("Processing DeleteFinalStateUnknown took longer than expected")
	}
}

func TestSyncReplicationControllerCreates(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	// A controller with 2 replicas and no pods in the store, 2 creates expected
	rc := newReplicationController(2)
	manager.rcStore.Store.Add(rc)

	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.syncReplicationController(getKey(rc, t))
	validateSyncReplication(t, &fakePodControl, 2, 0)
}

func TestStatusUpdatesWithoutReplicasChange(t *testing.T) {
	// Setup a fake server to listen for requests, and run the rc manager in steady state
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   200,
		ResponseBody: "",
	}
	testServer := httptest.NewServer(&fakeHandler)
	// TODO: Uncomment when fix #19254
	// defer testServer.Close()
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: testServer.URL, ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	// Steady state for the replication controller, no Status.Replicas updates expected
	activePods := 5
	rc := newReplicationController(activePods)
	manager.rcStore.Store.Add(rc)
	rc.Status = api.ReplicationControllerStatus{Replicas: activePods}
	newPodList(manager.podStore.Store, activePods, api.PodRunning, rc, "pod")

	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.syncReplicationController(getKey(rc, t))

	validateSyncReplication(t, &fakePodControl, 0, 0)
	if fakeHandler.RequestReceived != nil {
		t.Errorf("Unexpected update when pods and rcs are in a steady state")
	}

	// This response body is just so we don't err out decoding the http response, all
	// we care about is the request body sent below.
	response := runtime.EncodeOrDie(testapi.Default.Codec(), &api.ReplicationController{})
	fakeHandler.ResponseBody = response

	rc.Generation = rc.Generation + 1
	manager.syncReplicationController(getKey(rc, t))

	rc.Status.ObservedGeneration = rc.Generation
	updatedRc := runtime.EncodeOrDie(testapi.Default.Codec(), rc)
	fakeHandler.ValidateRequest(t, testapi.Default.ResourcePath(replicationControllerResourceName(), rc.Namespace, rc.Name)+"/status", "PUT", &updatedRc)
}

func TestControllerUpdateReplicas(t *testing.T) {
	// This is a happy server just to record the PUT request we expect for status.Replicas
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   200,
		ResponseBody: "",
	}
	testServer := httptest.NewServer(&fakeHandler)
	// TODO: Uncomment when fix #19254
	// defer testServer.Close()
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: testServer.URL, ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	// Insufficient number of pods in the system, and Status.Replicas is wrong;
	// Status.Replica should update to match number of pods in system, 1 new pod should be created.
	rc := newReplicationController(5)
	manager.rcStore.Store.Add(rc)
	rc.Status = api.ReplicationControllerStatus{Replicas: 2, FullyLabeledReplicas: 6, ObservedGeneration: 0}
	rc.Generation = 1
	newPodList(manager.podStore.Store, 2, api.PodRunning, rc, "pod")
	rcCopy := *rc
	extraLabelMap := map[string]string{"foo": "bar", "extraKey": "extraValue"}
	rcCopy.Spec.Selector = extraLabelMap
	newPodList(manager.podStore.Store, 2, api.PodRunning, &rcCopy, "podWithExtraLabel")

	// This response body is just so we don't err out decoding the http response
	response := runtime.EncodeOrDie(testapi.Default.Codec(), &api.ReplicationController{})
	fakeHandler.ResponseBody = response

	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl

	manager.syncReplicationController(getKey(rc, t))

	// 1. Status.Replicas should go up from 2->4 even though we created 5-4=1 pod.
	// 2. Status.FullyLabeledReplicas should equal to the number of pods that
	// has the extra labels, i.e., 2.
	// 3. Every update to the status should include the Generation of the spec.
	rc.Status = api.ReplicationControllerStatus{Replicas: 4, ObservedGeneration: 1}

	decRc := runtime.EncodeOrDie(testapi.Default.Codec(), rc)
	fakeHandler.ValidateRequest(t, testapi.Default.ResourcePath(replicationControllerResourceName(), rc.Namespace, rc.Name)+"/status", "PUT", &decRc)
	validateSyncReplication(t, &fakePodControl, 1, 0)
}

func TestSyncReplicationControllerDormancy(t *testing.T) {
	// Setup a test server so we can lie about the current state of pods
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   200,
		ResponseBody: "{}",
	}
	testServer := httptest.NewServer(&fakeHandler)
	// TODO: Uncomment when fix #19254
	// defer testServer.Close()
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: testServer.URL, ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady
	manager.podControl = &fakePodControl

	controllerSpec := newReplicationController(2)
	manager.rcStore.Store.Add(controllerSpec)
	newPodList(manager.podStore.Store, 1, api.PodRunning, controllerSpec, "pod")

	// Creates a replica and sets expectations
	controllerSpec.Status.Replicas = 1
	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 1, 0)

	// Expectations prevents replicas but not an update on status
	controllerSpec.Status.Replicas = 0
	fakePodControl.Clear()
	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 0, 0)

	// Get the key for the controller
	rcKey, err := controller.KeyFunc(controllerSpec)
	if err != nil {
		t.Errorf("Couldn't get key for object %+v: %v", controllerSpec, err)
	}

	// Lowering expectations should lead to a sync that creates a replica, however the
	// fakePodControl error will prevent this, leaving expectations at 0, 0
	manager.expectations.CreationObserved(rcKey)
	controllerSpec.Status.Replicas = 1
	fakePodControl.Clear()
	fakePodControl.Err = fmt.Errorf("Fake Error")

	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 0, 0)

	// This replica should not need a Lowering of expectations, since the previous create failed
	fakePodControl.Err = nil
	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 1, 0)

	// 1 PUT for the rc status during dormancy window.
	// Note that the pod creates go through pod control so they're not recorded.
	fakeHandler.ValidateRequestCount(t, 1)
}

func TestPodControllerLookup(t *testing.T) {
	manager := NewReplicationManager(clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}}), controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady
	testCases := []struct {
		inRCs     []*api.ReplicationController
		pod       *api.Pod
		outRCName string
	}{
		// pods without labels don't match any rcs
		{
			inRCs: []*api.ReplicationController{
				{ObjectMeta: api.ObjectMeta{Name: "basic"}}},
			pod:       &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo1", Namespace: api.NamespaceAll}},
			outRCName: "",
		},
		// Matching labels, not namespace
		{
			inRCs: []*api.ReplicationController{
				{
					ObjectMeta: api.ObjectMeta{Name: "foo"},
					Spec: api.ReplicationControllerSpec{
						Selector: map[string]string{"foo": "bar"},
					},
				},
			},
			pod: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name: "foo2", Namespace: "ns", Labels: map[string]string{"foo": "bar"}}},
			outRCName: "",
		},
		// Matching ns and labels returns the key to the rc, not the rc name
		{
			inRCs: []*api.ReplicationController{
				{
					ObjectMeta: api.ObjectMeta{Name: "bar", Namespace: "ns"},
					Spec: api.ReplicationControllerSpec{
						Selector: map[string]string{"foo": "bar"},
					},
				},
			},
			pod: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name: "foo3", Namespace: "ns", Labels: map[string]string{"foo": "bar"}}},
			outRCName: "bar",
		},
	}
	for _, c := range testCases {
		for _, r := range c.inRCs {
			manager.rcStore.Add(r)
		}
		if rc := manager.getPodController(c.pod); rc != nil {
			if c.outRCName != rc.Name {
				t.Errorf("Got controller %+v expected %+v", rc.Name, c.outRCName)
			}
		} else if c.outRCName != "" {
			t.Errorf("Expected a controller %v pod %v, found none", c.outRCName, c.pod.Name)
		}
	}
}

func TestWatchControllers(t *testing.T) {
	fakeWatch := watch.NewFake()
	c := &fake.Clientset{}
	c.AddWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	var testControllerSpec api.ReplicationController
	received := make(chan string)

	// The update sent through the fakeWatcher should make its way into the workqueue,
	// and eventually into the syncHandler. The handler validates the received controller
	// and closes the received channel to indicate that the test can finish.
	manager.syncHandler = func(key string) error {

		obj, exists, err := manager.rcStore.Store.GetByKey(key)
		if !exists || err != nil {
			t.Errorf("Expected to find controller under key %v", key)
		}
		controllerSpec := *obj.(*api.ReplicationController)
		if !api.Semantic.DeepDerivative(controllerSpec, testControllerSpec) {
			t.Errorf("Expected %#v, but got %#v", testControllerSpec, controllerSpec)
		}
		close(received)
		return nil
	}
	// Start only the rc watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go manager.rcController.Run(stopCh)
	go wait.Until(manager.worker, 10*time.Millisecond, stopCh)

	testControllerSpec.Name = "foo"
	fakeWatch.Add(&testControllerSpec)

	select {
	case <-received:
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("Expected 1 call but got 0")
	}
}

func TestWatchPods(t *testing.T) {
	fakeWatch := watch.NewFake()
	c := &fake.Clientset{}
	c.AddWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	// Put one rc and one pod into the controller's stores
	testControllerSpec := newReplicationController(1)
	manager.rcStore.Store.Add(testControllerSpec)
	received := make(chan string)
	// The pod update sent through the fakeWatcher should figure out the managing rc and
	// send it into the syncHandler.
	manager.syncHandler = func(key string) error {

		obj, exists, err := manager.rcStore.Store.GetByKey(key)
		if !exists || err != nil {
			t.Errorf("Expected to find controller under key %v", key)
		}
		controllerSpec := obj.(*api.ReplicationController)
		if !api.Semantic.DeepDerivative(controllerSpec, testControllerSpec) {
			t.Errorf("\nExpected %#v,\nbut got %#v", testControllerSpec, controllerSpec)
		}
		close(received)
		return nil
	}
	// Start only the pod watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method for the right rc.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go manager.podController.Run(stopCh)
	go wait.Until(manager.worker, 10*time.Millisecond, stopCh)

	pods := newPodList(nil, 1, api.PodRunning, testControllerSpec, "pod")
	testPod := pods.Items[0]
	testPod.Status.Phase = api.PodFailed
	fakeWatch.Add(&testPod)

	select {
	case <-received:
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("Expected 1 call but got 0")
	}
}

func TestUpdatePods(t *testing.T) {
	manager := NewReplicationManager(fake.NewSimpleClientset(), controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	received := make(chan string)

	manager.syncHandler = func(key string) error {
		obj, exists, err := manager.rcStore.Store.GetByKey(key)
		if !exists || err != nil {
			t.Errorf("Expected to find controller under key %v", key)
		}
		received <- obj.(*api.ReplicationController).Name
		return nil
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	go wait.Until(manager.worker, 10*time.Millisecond, stopCh)

	// Put 2 rcs and one pod into the controller's stores
	testControllerSpec1 := newReplicationController(1)
	manager.rcStore.Store.Add(testControllerSpec1)
	testControllerSpec2 := *testControllerSpec1
	testControllerSpec2.Spec.Selector = map[string]string{"bar": "foo"}
	testControllerSpec2.Name = "barfoo"
	manager.rcStore.Store.Add(&testControllerSpec2)

	// Put one pod in the podStore
	pod1 := newPodList(manager.podStore.Store, 1, api.PodRunning, testControllerSpec1, "pod").Items[0]
	pod2 := pod1
	pod2.Labels = testControllerSpec2.Spec.Selector

	// Send an update of the same pod with modified labels, and confirm we get a sync request for
	// both controllers
	manager.updatePod(&pod1, &pod2)

	expected := sets.NewString(testControllerSpec1.Name, testControllerSpec2.Name)
	for _, name := range expected.List() {
		t.Logf("Expecting update for %+v", name)
		select {
		case got := <-received:
			if !expected.Has(got) {
				t.Errorf("Expected keys %#v got %v", expected, got)
			}
		case <-time.After(wait.ForeverTestTimeout):
			t.Errorf("Expected update notifications for controllers within 100ms each")
		}
	}
}

func TestControllerUpdateRequeue(t *testing.T) {
	// This server should force a requeue of the controller because it fails to update status.Replicas.
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   500,
		ResponseBody: "",
	}
	testServer := httptest.NewServer(&fakeHandler)
	// TODO: Uncomment when fix #19254
	// defer testServer.Close()

	c := clientset.NewForConfigOrDie(&restclient.Config{Host: testServer.URL, ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, BurstReplicas, 0)
	manager.podStoreSynced = alwaysReady

	rc := newReplicationController(1)
	manager.rcStore.Store.Add(rc)
	rc.Status = api.ReplicationControllerStatus{Replicas: 2}
	newPodList(manager.podStore.Store, 1, api.PodRunning, rc, "pod")

	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl

	manager.syncReplicationController(getKey(rc, t))

	ch := make(chan interface{})
	go func() {
		item, _ := manager.queue.Get()
		ch <- item
	}()
	select {
	case key := <-ch:
		expectedKey := getKey(rc, t)
		if key != expectedKey {
			t.Errorf("Expected requeue of controller with key %s got %s", expectedKey, key)
		}
	case <-time.After(wait.ForeverTestTimeout):
		manager.queue.ShutDown()
		t.Errorf("Expected to find an rc in the queue, found none.")
	}
	// 1 Update and 1 GET, both of which fail
	fakeHandler.ValidateRequestCount(t, 2)
}

func TestControllerUpdateStatusWithFailure(t *testing.T) {
	rc := newReplicationController(1)
	c := &fake.Clientset{}
	c.AddReactor("get", "replicationcontrollers", func(action core.Action) (bool, runtime.Object, error) {
		return true, rc, nil
	})
	c.AddReactor("*", "*", func(action core.Action) (bool, runtime.Object, error) {
		return true, &api.ReplicationController{}, fmt.Errorf("Fake error")
	})
	fakeRCClient := c.Core().ReplicationControllers("default")
	numReplicas := 10
	updateReplicaCount(fakeRCClient, *rc, numReplicas, 0)
	updates, gets := 0, 0
	for _, a := range c.Actions() {
		if a.GetResource() != "replicationcontrollers" {
			t.Errorf("Unexpected action %+v", a)
			continue
		}

		switch action := a.(type) {
		case core.GetAction:
			gets++
			// Make sure the get is for the right rc even though the update failed.
			if action.GetName() != rc.Name {
				t.Errorf("Expected get for rc %v, got %+v instead", rc.Name, action.GetName())
			}
		case core.UpdateAction:
			updates++
			// Confirm that the update has the right status.Replicas even though the Get
			// returned an rc with replicas=1.
			if c, ok := action.GetObject().(*api.ReplicationController); !ok {
				t.Errorf("Expected an rc as the argument to update, got %T", c)
			} else if c.Status.Replicas != numReplicas {
				t.Errorf("Expected update for rc to contain replicas %v, got %v instead",
					numReplicas, c.Status.Replicas)
			}
		default:
			t.Errorf("Unexpected action %+v", a)
			break
		}
	}
	if gets != 1 || updates != 2 {
		t.Errorf("Expected 1 get and 2 updates, got %d gets %d updates", gets, updates)
	}
}

// TODO: This test is too hairy for a unittest. It should be moved to an E2E suite.
func doTestControllerBurstReplicas(t *testing.T, burstReplicas, numReplicas int) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, burstReplicas, 0)
	manager.podStoreSynced = alwaysReady
	manager.podControl = &fakePodControl

	controllerSpec := newReplicationController(numReplicas)
	manager.rcStore.Store.Add(controllerSpec)

	expectedPods := 0
	pods := newPodList(nil, numReplicas, api.PodPending, controllerSpec, "pod")

	rcKey, err := controller.KeyFunc(controllerSpec)
	if err != nil {
		t.Errorf("Couldn't get key for object %+v: %v", controllerSpec, err)
	}

	// Size up the controller, then size it down, and confirm the expected create/delete pattern
	for _, replicas := range []int{numReplicas, 0} {

		controllerSpec.Spec.Replicas = replicas
		manager.rcStore.Store.Add(controllerSpec)

		for i := 0; i < numReplicas; i += burstReplicas {
			manager.syncReplicationController(getKey(controllerSpec, t))

			// The store accrues active pods. It's also used by the rc to determine how many
			// replicas to create.
			activePods := len(manager.podStore.Store.List())
			if replicas != 0 {
				// This is the number of pods currently "in flight". They were created by the rc manager above,
				// which then puts the rc to sleep till all of them have been observed.
				expectedPods = replicas - activePods
				if expectedPods > burstReplicas {
					expectedPods = burstReplicas
				}
				// This validates the rc manager sync actually created pods
				validateSyncReplication(t, &fakePodControl, expectedPods, 0)

				// This simulates the watch events for all but 1 of the expected pods.
				// None of these should wake the controller because it has expectations==BurstReplicas.
				for i := 0; i < expectedPods-1; i++ {
					manager.podStore.Store.Add(&pods.Items[i])
					manager.addPod(&pods.Items[i])
				}

				podExp, exists, err := manager.expectations.GetExpectations(rcKey)
				if !exists || err != nil {
					t.Fatalf("Did not find expectations for rc.")
				}
				if add, _ := podExp.GetExpectations(); add != 1 {
					t.Fatalf("Expectations are wrong %v", podExp)
				}
			} else {
				expectedPods = (replicas - activePods) * -1
				if expectedPods > burstReplicas {
					expectedPods = burstReplicas
				}
				validateSyncReplication(t, &fakePodControl, 0, expectedPods)

				// To accurately simulate a watch we must delete the exact pods
				// the rc is waiting for.
				expectedDels := manager.expectations.GetUIDs(getKey(controllerSpec, t))
				podsToDelete := []*api.Pod{}
				for _, key := range expectedDels.List() {
					nsName := strings.Split(key, "/")
					podsToDelete = append(podsToDelete, &api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:      nsName[1],
							Namespace: nsName[0],
							Labels:    controllerSpec.Spec.Selector,
						},
					})
				}
				// Don't delete all pods because we confirm that the last pod
				// has exactly one expectation at the end, to verify that we
				// don't double delete.
				for i := range podsToDelete[1:] {
					manager.podStore.Delete(podsToDelete[i])
					manager.deletePod(podsToDelete[i])
				}
				podExp, exists, err := manager.expectations.GetExpectations(rcKey)
				if !exists || err != nil {
					t.Fatalf("Did not find expectations for rc.")
				}
				if _, del := podExp.GetExpectations(); del != 1 {
					t.Fatalf("Expectations are wrong %v", podExp)
				}
			}

			// Check that the rc didn't take any action for all the above pods
			fakePodControl.Clear()
			manager.syncReplicationController(getKey(controllerSpec, t))
			validateSyncReplication(t, &fakePodControl, 0, 0)

			// Create/Delete the last pod
			// The last add pod will decrease the expectation of the rc to 0,
			// which will cause it to create/delete the remaining replicas up to burstReplicas.
			if replicas != 0 {
				manager.podStore.Store.Add(&pods.Items[expectedPods-1])
				manager.addPod(&pods.Items[expectedPods-1])
			} else {
				expectedDel := manager.expectations.GetUIDs(getKey(controllerSpec, t))
				if expectedDel.Len() != 1 {
					t.Fatalf("Waiting on unexpected number of deletes.")
				}
				nsName := strings.Split(expectedDel.List()[0], "/")
				lastPod := &api.Pod{
					ObjectMeta: api.ObjectMeta{
						Name:      nsName[1],
						Namespace: nsName[0],
						Labels:    controllerSpec.Spec.Selector,
					},
				}
				manager.podStore.Store.Delete(lastPod)
				manager.deletePod(lastPod)
			}
			pods.Items = pods.Items[expectedPods:]
		}

		// Confirm that we've created the right number of replicas
		activePods := len(manager.podStore.Store.List())
		if activePods != controllerSpec.Spec.Replicas {
			t.Fatalf("Unexpected number of active pods, expected %d, got %d", controllerSpec.Spec.Replicas, activePods)
		}
		// Replenish the pod list, since we cut it down sizing up
		pods = newPodList(nil, replicas, api.PodRunning, controllerSpec, "pod")
	}
}

func TestControllerBurstReplicas(t *testing.T) {
	doTestControllerBurstReplicas(t, 5, 30)
	doTestControllerBurstReplicas(t, 5, 12)
	doTestControllerBurstReplicas(t, 3, 2)
}

type FakeRCExpectations struct {
	*controller.ControllerExpectations
	satisfied    bool
	expSatisfied func()
}

func (fe FakeRCExpectations) SatisfiedExpectations(controllerKey string) bool {
	fe.expSatisfied()
	return fe.satisfied
}

// TestRCSyncExpectations tests that a pod cannot sneak in between counting active pods
// and checking expectations.
func TestRCSyncExpectations(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, 2, 0)
	manager.podStoreSynced = alwaysReady
	manager.podControl = &fakePodControl

	controllerSpec := newReplicationController(2)
	manager.rcStore.Store.Add(controllerSpec)
	pods := newPodList(nil, 2, api.PodPending, controllerSpec, "pod")
	manager.podStore.Store.Add(&pods.Items[0])
	postExpectationsPod := pods.Items[1]

	manager.expectations = controller.NewUIDTrackingControllerExpectations(FakeRCExpectations{
		controller.NewControllerExpectations(), true, func() {
			// If we check active pods before checking expectataions, the rc
			// will create a new replica because it doesn't see this pod, but
			// has fulfilled its expectations.
			manager.podStore.Store.Add(&postExpectationsPod)
		},
	})
	manager.syncReplicationController(getKey(controllerSpec, t))
	validateSyncReplication(t, &fakePodControl, 0, 0)
}

func TestDeleteControllerAndExpectations(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, 10, 0)
	manager.podStoreSynced = alwaysReady

	rc := newReplicationController(1)
	manager.rcStore.Store.Add(rc)

	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl

	// This should set expectations for the rc
	manager.syncReplicationController(getKey(rc, t))
	validateSyncReplication(t, &fakePodControl, 1, 0)
	fakePodControl.Clear()

	// Get the RC key
	rcKey, err := controller.KeyFunc(rc)
	if err != nil {
		t.Errorf("Couldn't get key for object %+v: %v", rc, err)
	}

	// This is to simulate a concurrent addPod, that has a handle on the expectations
	// as the controller deletes it.
	podExp, exists, err := manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil {
		t.Errorf("No expectations found for rc")
	}
	manager.rcStore.Delete(rc)
	manager.syncReplicationController(getKey(rc, t))

	if _, exists, err = manager.expectations.GetExpectations(rcKey); exists {
		t.Errorf("Found expectaions, expected none since the rc has been deleted.")
	}

	// This should have no effect, since we've deleted the rc.
	podExp.Add(-1, 0)
	manager.podStore.Store.Replace(make([]interface{}, 0), "0")
	manager.syncReplicationController(getKey(rc, t))
	validateSyncReplication(t, &fakePodControl, 0, 0)
}

func TestRCManagerNotReady(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	fakePodControl := controller.FakePodControl{}
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, 2, 0)
	manager.podControl = &fakePodControl
	manager.podStoreSynced = func() bool { return false }

	// Simulates the rc reflector running before the pod reflector. We don't
	// want to end up creating replicas in this case until the pod reflector
	// has synced, so the rc manager should just requeue the rc.
	controllerSpec := newReplicationController(1)
	manager.rcStore.Store.Add(controllerSpec)

	rcKey := getKey(controllerSpec, t)
	manager.syncReplicationController(rcKey)
	validateSyncReplication(t, &fakePodControl, 0, 0)
	queueRC, _ := manager.queue.Get()
	if queueRC != rcKey {
		t.Fatalf("Expected to find key %v in queue, found %v", rcKey, queueRC)
	}

	manager.podStoreSynced = alwaysReady
	manager.syncReplicationController(rcKey)
	validateSyncReplication(t, &fakePodControl, 1, 0)
}

// shuffle returns a new shuffled list of container controllers.
func shuffle(controllers []*api.ReplicationController) []*api.ReplicationController {
	numControllers := len(controllers)
	randIndexes := rand.Perm(numControllers)
	shuffled := make([]*api.ReplicationController, numControllers)
	for i := 0; i < numControllers; i++ {
		shuffled[i] = controllers[randIndexes[i]]
	}
	return shuffled
}

func TestOverlappingRCs(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})

	for i := 0; i < 5; i++ {
		manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, 10, 0)
		manager.podStoreSynced = alwaysReady

		// Create 10 rcs, shuffled them randomly and insert them into the rc manager's store
		var controllers []*api.ReplicationController
		for j := 1; j < 10; j++ {
			controllerSpec := newReplicationController(1)
			controllerSpec.CreationTimestamp = unversioned.Date(2014, time.December, j, 0, 0, 0, 0, time.Local)
			controllerSpec.Name = string(util.NewUUID())
			controllers = append(controllers, controllerSpec)
		}
		shuffledControllers := shuffle(controllers)
		for j := range shuffledControllers {
			manager.rcStore.Store.Add(shuffledControllers[j])
		}
		// Add a pod and make sure only the oldest rc is synced
		pods := newPodList(nil, 1, api.PodPending, controllers[0], "pod")
		rcKey := getKey(controllers[0], t)

		manager.addPod(&pods.Items[0])
		queueRC, _ := manager.queue.Get()
		if queueRC != rcKey {
			t.Fatalf("Expected to find key %v in queue, found %v", rcKey, queueRC)
		}
	}
}

func TestDeletionTimestamp(t *testing.T) {
	c := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(c, controller.NoResyncPeriodFunc, 10, 0)
	manager.podStoreSynced = alwaysReady

	controllerSpec := newReplicationController(1)
	manager.rcStore.Store.Add(controllerSpec)
	rcKey, err := controller.KeyFunc(controllerSpec)
	if err != nil {
		t.Errorf("Couldn't get key for object %+v: %v", controllerSpec, err)
	}
	pod := newPodList(nil, 1, api.PodPending, controllerSpec, "pod").Items[0]
	pod.DeletionTimestamp = &unversioned.Time{Time: time.Now()}
	manager.expectations.ExpectDeletions(rcKey, []string{controller.PodKey(&pod)})

	// A pod added with a deletion timestamp should decrement deletions, not creations.
	manager.addPod(&pod)

	queueRC, _ := manager.queue.Get()
	if queueRC != rcKey {
		t.Fatalf("Expected to find key %v in queue, found %v", rcKey, queueRC)
	}
	manager.queue.Done(rcKey)

	podExp, exists, err := manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil || !podExp.Fulfilled() {
		t.Fatalf("Wrong expectations %+v", podExp)
	}

	// An update from no deletion timestamp to having one should be treated
	// as a deletion.
	oldPod := newPodList(nil, 1, api.PodPending, controllerSpec, "pod").Items[0]
	manager.expectations.ExpectDeletions(rcKey, []string{controller.PodKey(&pod)})
	manager.updatePod(&oldPod, &pod)

	queueRC, _ = manager.queue.Get()
	if queueRC != rcKey {
		t.Fatalf("Expected to find key %v in queue, found %v", rcKey, queueRC)
	}
	manager.queue.Done(rcKey)

	podExp, exists, err = manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil || !podExp.Fulfilled() {
		t.Fatalf("Wrong expectations %+v", podExp)
	}

	// An update to the pod (including an update to the deletion timestamp)
	// should not be counted as a second delete.
	secondPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      "secondPod",
			Labels:    pod.Labels,
		},
	}
	manager.expectations.ExpectDeletions(rcKey, []string{controller.PodKey(secondPod)})
	oldPod.DeletionTimestamp = &unversioned.Time{Time: time.Now()}
	manager.updatePod(&oldPod, &pod)

	podExp, exists, err = manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil || podExp.Fulfilled() {
		t.Fatalf("Wrong expectations %+v", podExp)
	}

	// A pod with a non-nil deletion timestamp should also be ignored by the
	// delete handler, because it's already been counted in the update.
	manager.deletePod(&pod)
	podExp, exists, err = manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil || podExp.Fulfilled() {
		t.Fatalf("Wrong expectations %+v", podExp)
	}

	// Deleting the second pod should clear expectations.
	manager.deletePod(secondPod)

	queueRC, _ = manager.queue.Get()
	if queueRC != rcKey {
		t.Fatalf("Expected to find key %v in queue, found %v", rcKey, queueRC)
	}
	manager.queue.Done(rcKey)

	podExp, exists, err = manager.expectations.GetExpectations(rcKey)
	if !exists || err != nil || !podExp.Fulfilled() {
		t.Fatalf("Wrong expectations %+v", podExp)
	}
}

func BenchmarkGetPodControllerMultiNS(b *testing.B) {
	client := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(client, controller.NoResyncPeriodFunc, BurstReplicas, 0)

	const nsNum = 1000

	pods := []api.Pod{}
	for i := 0; i < nsNum; i++ {
		ns := fmt.Sprintf("ns-%d", i)
		for j := 0; j < 10; j++ {
			rcName := fmt.Sprintf("rc-%d", j)
			for k := 0; k < 10; k++ {
				podName := fmt.Sprintf("pod-%d-%d", j, k)
				pods = append(pods, api.Pod{
					ObjectMeta: api.ObjectMeta{
						Name:      podName,
						Namespace: ns,
						Labels:    map[string]string{"rcName": rcName},
					},
				})
			}
		}
	}

	for i := 0; i < nsNum; i++ {
		ns := fmt.Sprintf("ns-%d", i)
		for j := 0; j < 10; j++ {
			rcName := fmt.Sprintf("rc-%d", j)
			manager.rcStore.Add(&api.ReplicationController{
				ObjectMeta: api.ObjectMeta{Name: rcName, Namespace: ns},
				Spec: api.ReplicationControllerSpec{
					Selector: map[string]string{"rcName": rcName},
				},
			})
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, pod := range pods {
			manager.getPodController(&pod)
		}
	}
}

func BenchmarkGetPodControllerSingleNS(b *testing.B) {
	client := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()}})
	manager := NewReplicationManager(client, controller.NoResyncPeriodFunc, BurstReplicas, 0)

	const rcNum = 1000
	const replicaNum = 3

	pods := []api.Pod{}
	for i := 0; i < rcNum; i++ {
		rcName := fmt.Sprintf("rc-%d", i)
		for j := 0; j < replicaNum; j++ {
			podName := fmt.Sprintf("pod-%d-%d", i, j)
			pods = append(pods, api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name:      podName,
					Namespace: "foo",
					Labels:    map[string]string{"rcName": rcName},
				},
			})
		}
	}

	for i := 0; i < rcNum; i++ {
		rcName := fmt.Sprintf("rc-%d", i)
		manager.rcStore.Add(&api.ReplicationController{
			ObjectMeta: api.ObjectMeta{Name: rcName, Namespace: "foo"},
			Spec: api.ReplicationControllerSpec{
				Selector: map[string]string{"rcName": rcName},
			},
		})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, pod := range pods {
			manager.getPodController(&pod)
		}
	}
}
