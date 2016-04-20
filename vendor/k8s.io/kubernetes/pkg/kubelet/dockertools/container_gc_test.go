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

package dockertools

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/api"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/types"
)

func newTestContainerGC(t *testing.T) (*containerGC, *FakeDockerClient) {
	fakeDocker := new(FakeDockerClient)
	fakePodGetter := newFakePodGetter()
	gc := NewContainerGC(fakeDocker, fakePodGetter, "")
	return gc, fakeDocker
}

// Makes a stable time object, lower id is earlier time.
func makeTime(id int) time.Time {
	var zero time.Time
	return zero.Add(time.Duration(id) * time.Second)
}

// Makes a container with the specified properties.
func makeContainer(id, uid, name string, running bool, created time.Time) *docker.Container {
	return &docker.Container{
		Name: fmt.Sprintf("/k8s_%s_bar_new_%s_42", name, uid),
		State: docker.State{
			Running: running,
		},
		ID:      id,
		Created: created,
	}
}

// Makes a container with unidentified name and specified properties.
func makeUndefinedContainer(id string, running bool, created time.Time) *docker.Container {
	return &docker.Container{
		Name: "/k8s_unidentified",
		State: docker.State{
			Running: running,
		},
		ID:      id,
		Created: created,
	}
}

func addPods(podGetter podGetter, podUIDs ...types.UID) {
	fakePodGetter := podGetter.(*fakePodGetter)
	for _, uid := range podUIDs {
		fakePodGetter.pods[uid] = &api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name:      "pod" + string(uid),
				Namespace: "test",
				UID:       uid,
			},
		}
	}
}

func verifyStringArrayEqualsAnyOrder(t *testing.T, actual, expected []string) {
	act := make([]string, len(actual))
	exp := make([]string, len(expected))
	copy(act, actual)
	copy(exp, expected)

	sort.StringSlice(act).Sort()
	sort.StringSlice(exp).Sort()

	if !reflect.DeepEqual(exp, act) {
		t.Errorf("Expected(sorted): %#v, Actual(sorted): %#v", exp, act)
	}
}

func TestGarbageCollectZeroMaxContainers(t *testing.T) {
	gc, fakeDocker := newTestContainerGC(t)
	fakeDocker.SetFakeContainers([]*docker.Container{
		makeContainer("1876", "foo", "POD", false, makeTime(0)),
	})
	addPods(gc.podGetter, "foo")

	assert.Nil(t, gc.GarbageCollect(kubecontainer.ContainerGCPolicy{MinAge: time.Minute, MaxPerPodContainer: 1, MaxContainers: 0}))
	assert.Len(t, fakeDocker.Removed, 1)
}

func TestGarbageCollectNoMaxPerPodContainerLimit(t *testing.T) {
	gc, fakeDocker := newTestContainerGC(t)
	fakeDocker.SetFakeContainers([]*docker.Container{
		makeContainer("1876", "foo", "POD", false, makeTime(0)),
		makeContainer("2876", "foo1", "POD", false, makeTime(1)),
		makeContainer("3876", "foo2", "POD", false, makeTime(2)),
		makeContainer("4876", "foo3", "POD", false, makeTime(3)),
		makeContainer("5876", "foo4", "POD", false, makeTime(4)),
	})
	addPods(gc.podGetter, "foo", "foo1", "foo2", "foo3", "foo4")

	assert.Nil(t, gc.GarbageCollect(kubecontainer.ContainerGCPolicy{MinAge: time.Minute, MaxPerPodContainer: -1, MaxContainers: 4}))
	assert.Len(t, fakeDocker.Removed, 1)
}

func TestGarbageCollectNoMaxLimit(t *testing.T) {
	gc, fakeDocker := newTestContainerGC(t)
	fakeDocker.SetFakeContainers([]*docker.Container{
		makeContainer("1876", "foo", "POD", false, makeTime(0)),
		makeContainer("2876", "foo1", "POD", false, makeTime(0)),
		makeContainer("3876", "foo2", "POD", false, makeTime(0)),
		makeContainer("4876", "foo3", "POD", false, makeTime(0)),
		makeContainer("5876", "foo4", "POD", false, makeTime(0)),
	})
	addPods(gc.podGetter, "foo", "foo1", "foo2", "foo3", "foo4")

	assert.Len(t, fakeDocker.Removed, 0)
}

func TestGarbageCollect(t *testing.T) {
	tests := []struct {
		containers      []*docker.Container
		expectedRemoved []string
	}{
		// Don't remove containers started recently.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, time.Now()),
				makeContainer("2876", "foo", "POD", false, time.Now()),
				makeContainer("3876", "foo", "POD", false, time.Now()),
			},
		},
		// Remove oldest containers.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(0)),
				makeContainer("2876", "foo", "POD", false, makeTime(1)),
				makeContainer("3876", "foo", "POD", false, makeTime(2)),
			},
			expectedRemoved: []string{"1876"},
		},
		// Only remove non-running containers.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", true, makeTime(0)),
				makeContainer("2876", "foo", "POD", false, makeTime(1)),
				makeContainer("3876", "foo", "POD", false, makeTime(2)),
				makeContainer("4876", "foo", "POD", false, makeTime(3)),
			},
			expectedRemoved: []string{"2876"},
		},
		// Less than maxContainerCount doesn't delete any.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(0)),
			},
		},
		// maxContainerCount applies per (UID,container) pair.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(0)),
				makeContainer("2876", "foo", "POD", false, makeTime(1)),
				makeContainer("3876", "foo", "POD", false, makeTime(2)),
				makeContainer("1076", "foo", "bar", false, makeTime(0)),
				makeContainer("2076", "foo", "bar", false, makeTime(1)),
				makeContainer("3076", "foo", "bar", false, makeTime(2)),
				makeContainer("1176", "foo2", "POD", false, makeTime(0)),
				makeContainer("2176", "foo2", "POD", false, makeTime(1)),
				makeContainer("3176", "foo2", "POD", false, makeTime(2)),
			},
			expectedRemoved: []string{"1076", "1176", "1876"},
		},
		// Remove non-running unidentified Kubernetes containers.
		{
			containers: []*docker.Container{
				makeUndefinedContainer("1876", true, makeTime(0)),
				makeUndefinedContainer("2876", false, makeTime(0)),
				makeContainer("3876", "foo", "POD", false, makeTime(0)),
			},
			expectedRemoved: []string{"2876"},
		},
		// Max limit applied and tries to keep from every pod.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(0)),
				makeContainer("2876", "foo", "POD", false, makeTime(1)),
				makeContainer("3876", "foo1", "POD", false, makeTime(0)),
				makeContainer("4876", "foo1", "POD", false, makeTime(1)),
				makeContainer("5876", "foo2", "POD", false, makeTime(0)),
				makeContainer("6876", "foo2", "POD", false, makeTime(1)),
				makeContainer("7876", "foo3", "POD", false, makeTime(0)),
				makeContainer("8876", "foo3", "POD", false, makeTime(1)),
				makeContainer("9876", "foo4", "POD", false, makeTime(0)),
				makeContainer("10876", "foo4", "POD", false, makeTime(1)),
			},
			expectedRemoved: []string{"1876", "3876", "5876", "7876", "9876"},
		},
		// If more pods than limit allows, evicts oldest pod.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(1)),
				makeContainer("2876", "foo", "POD", false, makeTime(2)),
				makeContainer("3876", "foo1", "POD", false, makeTime(1)),
				makeContainer("4876", "foo1", "POD", false, makeTime(2)),
				makeContainer("5876", "foo2", "POD", false, makeTime(0)),
				makeContainer("6876", "foo3", "POD", false, makeTime(1)),
				makeContainer("7876", "foo4", "POD", false, makeTime(0)),
				makeContainer("8876", "foo5", "POD", false, makeTime(1)),
				makeContainer("9876", "foo6", "POD", false, makeTime(2)),
				makeContainer("10876", "foo7", "POD", false, makeTime(1)),
			},
			expectedRemoved: []string{"1876", "3876", "5876", "7876"},
		},
		// Containers for deleted pods should be GC'd.
		{
			containers: []*docker.Container{
				makeContainer("1876", "foo", "POD", false, makeTime(1)),
				makeContainer("2876", "foo", "POD", false, makeTime(2)),
				makeContainer("3876", "deleted", "POD", false, makeTime(1)),
				makeContainer("4876", "deleted", "POD", false, makeTime(2)),
				makeContainer("5876", "deleted", "POD", false, time.Now()), // Deleted pods still respect MinAge.
			},
			expectedRemoved: []string{"3876", "4876"},
		},
	}
	for i, test := range tests {
		t.Logf("Running test case with index %d", i)
		gc, fakeDocker := newTestContainerGC(t)
		fakeDocker.SetFakeContainers(test.containers)
		addPods(gc.podGetter, "foo", "foo1", "foo2", "foo3", "foo4", "foo5", "foo6", "foo7")
		assert.Nil(t, gc.GarbageCollect(kubecontainer.ContainerGCPolicy{MinAge: time.Hour, MaxPerPodContainer: 2, MaxContainers: 6}))
		verifyStringArrayEqualsAnyOrder(t, fakeDocker.Removed, test.expectedRemoved)
	}
}
