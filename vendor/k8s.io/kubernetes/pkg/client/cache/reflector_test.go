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

package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type testLW struct {
	ListFunc  func() (runtime.Object, error)
	WatchFunc func(resourceVersion string) (watch.Interface, error)
}

func (t *testLW) List(options api.ListOptions) (runtime.Object, error) {
	return t.ListFunc()
}
func (t *testLW) Watch(options api.ListOptions) (watch.Interface, error) {
	return t.WatchFunc(options.ResourceVersion)
}

func TestCloseWatchChannelOnError(t *testing.T) {
	r := NewReflector(&testLW{}, &api.Pod{}, NewStore(MetaNamespaceKeyFunc), 0)
	pod := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "bar"}}
	fw := watch.NewFake()
	r.listerWatcher = &testLW{
		WatchFunc: func(rv string) (watch.Interface, error) {
			return fw, nil
		},
		ListFunc: func() (runtime.Object, error) {
			return &api.PodList{ListMeta: unversioned.ListMeta{ResourceVersion: "1"}}, nil
		},
	}
	go r.ListAndWatch(wait.NeverStop)
	fw.Error(pod)
	select {
	case _, ok := <-fw.ResultChan():
		if ok {
			t.Errorf("Watch channel left open after cancellation")
		}
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("the cancellation is at least %s late", wait.ForeverTestTimeout.String())
		break
	}
}

func TestRunUntil(t *testing.T) {
	stopCh := make(chan struct{})
	store := NewStore(MetaNamespaceKeyFunc)
	r := NewReflector(&testLW{}, &api.Pod{}, store, 0)
	fw := watch.NewFake()
	r.listerWatcher = &testLW{
		WatchFunc: func(rv string) (watch.Interface, error) {
			return fw, nil
		},
		ListFunc: func() (runtime.Object, error) {
			return &api.PodList{ListMeta: unversioned.ListMeta{ResourceVersion: "1"}}, nil
		},
	}
	r.RunUntil(stopCh)
	// Synchronously add a dummy pod into the watch channel so we
	// know the RunUntil go routine is in the watch handler.
	fw.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "bar"}})
	stopCh <- struct{}{}
	select {
	case _, ok := <-fw.ResultChan():
		if ok {
			t.Errorf("Watch channel left open after stopping the watch")
		}
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("the cancellation is at least %s late", wait.ForeverTestTimeout.String())
		break
	}
}

func TestReflectorResyncChan(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, time.Millisecond)
	a, _ := g.resyncChan()
	b := time.After(wait.ForeverTestTimeout)
	select {
	case <-a:
		t.Logf("got timeout as expected")
	case <-b:
		t.Errorf("resyncChan() is at least 99 milliseconds late??")
	}
}

func BenchmarkReflectorResyncChanMany(b *testing.B) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, 25*time.Millisecond)
	// The improvement to this (calling the timer's Stop() method) makes
	// this benchmark about 40% faster.
	for i := 0; i < b.N; i++ {
		g.resyncPeriod = time.Duration(rand.Float64() * float64(time.Millisecond) * 25)
		_, stop := g.resyncChan()
		stop()
	}
}

func TestReflectorWatchHandlerError(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, 0)
	fw := watch.NewFake()
	go func() {
		fw.Stop()
	}()
	var resumeRV string
	err := g.watchHandler(fw, &resumeRV, neverExitWatch, wait.NeverStop)
	if err == nil {
		t.Errorf("unexpected non-error")
	}
}

func TestReflectorWatchHandler(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, 0)
	fw := watch.NewFake()
	s.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}})
	s.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "bar"}})
	go func() {
		fw.Add(&api.Service{ObjectMeta: api.ObjectMeta{Name: "rejected"}})
		fw.Delete(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}})
		fw.Modify(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "bar", ResourceVersion: "55"}})
		fw.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "baz", ResourceVersion: "32"}})
		fw.Stop()
	}()
	var resumeRV string
	err := g.watchHandler(fw, &resumeRV, neverExitWatch, wait.NeverStop)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	mkPod := func(id string, rv string) *api.Pod {
		return &api.Pod{ObjectMeta: api.ObjectMeta{Name: id, ResourceVersion: rv}}
	}

	table := []struct {
		Pod    *api.Pod
		exists bool
	}{
		{mkPod("foo", ""), false},
		{mkPod("rejected", ""), false},
		{mkPod("bar", "55"), true},
		{mkPod("baz", "32"), true},
	}
	for _, item := range table {
		obj, exists, _ := s.Get(item.Pod)
		if e, a := item.exists, exists; e != a {
			t.Errorf("%v: expected %v, got %v", item.Pod, e, a)
		}
		if !exists {
			continue
		}
		if e, a := item.Pod.ResourceVersion, obj.(*api.Pod).ResourceVersion; e != a {
			t.Errorf("%v: expected %v, got %v", item.Pod, e, a)
		}
	}

	// RV should send the last version we see.
	if e, a := "32", resumeRV; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	// last sync resource version should be the last version synced with store
	if e, a := "32", g.LastSyncResourceVersion(); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestReflectorWatchHandlerTimeout(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, 0)
	fw := watch.NewFake()
	var resumeRV string
	exit := make(chan time.Time, 1)
	exit <- time.Now()
	err := g.watchHandler(fw, &resumeRV, exit, wait.NeverStop)
	if err != errorResyncRequested {
		t.Errorf("expected timeout error, but got %q", err)
	}
}

func TestReflectorStopWatch(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)
	g := NewReflector(&testLW{}, &api.Pod{}, s, 0)
	fw := watch.NewFake()
	var resumeRV string
	stopWatch := make(chan struct{}, 1)
	stopWatch <- struct{}{}
	err := g.watchHandler(fw, &resumeRV, neverExitWatch, stopWatch)
	if err != errorStopRequested {
		t.Errorf("expected stop error, got %q", err)
	}
}

func TestReflectorListAndWatch(t *testing.T) {
	createdFakes := make(chan *watch.FakeWatcher)

	// The ListFunc says that it's at revision 1. Therefore, we expect our WatchFunc
	// to get called at the beginning of the watch with 1, and again with 3 when we
	// inject an error.
	expectedRVs := []string{"1", "3"}
	lw := &testLW{
		WatchFunc: func(rv string) (watch.Interface, error) {
			fw := watch.NewFake()
			if e, a := expectedRVs[0], rv; e != a {
				t.Errorf("Expected rv %v, but got %v", e, a)
			}
			expectedRVs = expectedRVs[1:]
			// channel is not buffered because the for loop below needs to block. But
			// we don't want to block here, so report the new fake via a go routine.
			go func() { createdFakes <- fw }()
			return fw, nil
		},
		ListFunc: func() (runtime.Object, error) {
			return &api.PodList{ListMeta: unversioned.ListMeta{ResourceVersion: "1"}}, nil
		},
	}
	s := NewFIFO(MetaNamespaceKeyFunc)
	r := NewReflector(lw, &api.Pod{}, s, 0)
	go r.ListAndWatch(wait.NeverStop)

	ids := []string{"foo", "bar", "baz", "qux", "zoo"}
	var fw *watch.FakeWatcher
	for i, id := range ids {
		if fw == nil {
			fw = <-createdFakes
		}
		sendingRV := strconv.FormatUint(uint64(i+2), 10)
		fw.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: id, ResourceVersion: sendingRV}})
		if sendingRV == "3" {
			// Inject a failure.
			fw.Stop()
			fw = nil
		}
	}

	// Verify we received the right ids with the right resource versions.
	for i, id := range ids {
		pod := s.Pop().(*api.Pod)
		if e, a := id, pod.Name; e != a {
			t.Errorf("%v: Expected %v, got %v", i, e, a)
		}
		if e, a := strconv.FormatUint(uint64(i+2), 10), pod.ResourceVersion; e != a {
			t.Errorf("%v: Expected %v, got %v", i, e, a)
		}
	}

	if len(expectedRVs) != 0 {
		t.Error("called watchStarter an unexpected number of times")
	}
}

func TestReflectorListAndWatchWithErrors(t *testing.T) {
	mkPod := func(id string, rv string) *api.Pod {
		return &api.Pod{ObjectMeta: api.ObjectMeta{Name: id, ResourceVersion: rv}}
	}
	mkList := func(rv string, pods ...*api.Pod) *api.PodList {
		list := &api.PodList{ListMeta: unversioned.ListMeta{ResourceVersion: rv}}
		for _, pod := range pods {
			list.Items = append(list.Items, *pod)
		}
		return list
	}
	table := []struct {
		list     *api.PodList
		listErr  error
		events   []watch.Event
		watchErr error
	}{
		{
			list: mkList("1"),
			events: []watch.Event{
				{watch.Added, mkPod("foo", "2")},
				{watch.Added, mkPod("bar", "3")},
			},
		}, {
			list: mkList("3", mkPod("foo", "2"), mkPod("bar", "3")),
			events: []watch.Event{
				{watch.Deleted, mkPod("foo", "4")},
				{watch.Added, mkPod("qux", "5")},
			},
		}, {
			listErr: fmt.Errorf("a list error"),
		}, {
			list:     mkList("5", mkPod("bar", "3"), mkPod("qux", "5")),
			watchErr: fmt.Errorf("a watch error"),
		}, {
			list: mkList("5", mkPod("bar", "3"), mkPod("qux", "5")),
			events: []watch.Event{
				{watch.Added, mkPod("baz", "6")},
			},
		}, {
			list: mkList("6", mkPod("bar", "3"), mkPod("qux", "5"), mkPod("baz", "6")),
		},
	}

	s := NewFIFO(MetaNamespaceKeyFunc)
	for line, item := range table {
		if item.list != nil {
			// Test that the list is what currently exists in the store.
			current := s.List()
			checkMap := map[string]string{}
			for _, item := range current {
				pod := item.(*api.Pod)
				checkMap[pod.Name] = pod.ResourceVersion
			}
			for _, pod := range item.list.Items {
				if e, a := pod.ResourceVersion, checkMap[pod.Name]; e != a {
					t.Errorf("%v: expected %v, got %v for pod %v", line, e, a, pod.Name)
				}
			}
			if e, a := len(item.list.Items), len(checkMap); e != a {
				t.Errorf("%v: expected %v, got %v", line, e, a)
			}
		}
		watchRet, watchErr := item.events, item.watchErr
		lw := &testLW{
			WatchFunc: func(rv string) (watch.Interface, error) {
				if watchErr != nil {
					return nil, watchErr
				}
				watchErr = fmt.Errorf("second watch")
				fw := watch.NewFake()
				go func() {
					for _, e := range watchRet {
						fw.Action(e.Type, e.Object)
					}
					fw.Stop()
				}()
				return fw, nil
			},
			ListFunc: func() (runtime.Object, error) {
				return item.list, item.listErr
			},
		}
		r := NewReflector(lw, &api.Pod{}, s, 0)
		r.ListAndWatch(wait.NeverStop)
	}
}

func TestReflectorResync(t *testing.T) {
	s := NewStore(MetaNamespaceKeyFunc)

	currentTime := time.Time{}
	iteration := 0

	lw := &testLW{
		WatchFunc: func(rv string) (watch.Interface, error) {
			if iteration == 0 {
				// Move time, but do not force resync.
				currentTime = currentTime.Add(30 * time.Second)
			} else if iteration == 1 {
				// Move time to force resync.
				currentTime = currentTime.Add(28 * time.Second)
			} else if iteration >= 2 {
				t.Fatalf("should have forced resync earlier")
			}
			iteration++
			fw := watch.NewFake()
			// Send something to the watcher to avoid "watch too short" errors.
			go func() {
				fw.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: strconv.Itoa(iteration)}})
				fw.Stop()
			}()
			return fw, nil
		},
		ListFunc: func() (runtime.Object, error) {
			return &api.PodList{ListMeta: unversioned.ListMeta{ResourceVersion: "0"}}, nil
		},
	}
	resyncPeriod := time.Minute
	r := NewReflector(lw, &api.Pod{}, s, resyncPeriod)
	r.now = func() time.Time { return currentTime }

	r.ListAndWatch(wait.NeverStop)
	if iteration != 2 {
		t.Errorf("exactly 2 iterations were expected, got: %v", iteration)
	}
}
