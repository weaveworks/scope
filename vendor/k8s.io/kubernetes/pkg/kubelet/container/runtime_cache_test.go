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

package container_test

import (
	"reflect"
	"testing"
	"time"

	. "k8s.io/kubernetes/pkg/kubelet/container"
	ctest "k8s.io/kubernetes/pkg/kubelet/container/testing"
)

func TestGetPods(t *testing.T) {
	runtime := &ctest.FakeRuntime{}
	expected := []*Pod{{ID: "1111"}, {ID: "2222"}, {ID: "3333"}}
	runtime.PodList = expected
	cache := NewTestRuntimeCache(runtime)
	actual, err := cache.GetPods()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

func TestForceUpdateIfOlder(t *testing.T) {
	runtime := &ctest.FakeRuntime{}
	cache := NewTestRuntimeCache(runtime)

	// Cache old pods.
	oldpods := []*Pod{{ID: "1111"}}
	runtime.PodList = oldpods
	cache.UpdateCacheWithLock()

	// Update the runtime to new pods.
	newpods := []*Pod{{ID: "1111"}, {ID: "2222"}, {ID: "3333"}}
	runtime.PodList = newpods

	// An older timestamp should not force an update.
	cache.ForceUpdateIfOlder(time.Now().Add(-20 * time.Minute))
	actual := cache.GetCachedPods()
	if !reflect.DeepEqual(oldpods, actual) {
		t.Errorf("expected %#v, got %#v", oldpods, actual)
	}

	// A newer timestamp should force an update.
	cache.ForceUpdateIfOlder(time.Now().Add(20 * time.Second))
	actual = cache.GetCachedPods()
	if !reflect.DeepEqual(newpods, actual) {
		t.Errorf("expected %#v, got %#v", newpods, actual)
	}
}
