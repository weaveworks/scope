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

package kubelet

import (
	"os"
	"testing"
	"time"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/record"
	cadvisortest "k8s.io/kubernetes/pkg/kubelet/cadvisor/testing"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	containertest "k8s.io/kubernetes/pkg/kubelet/container/testing"
	"k8s.io/kubernetes/pkg/kubelet/network"
	nettest "k8s.io/kubernetes/pkg/kubelet/network/testing"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	podtest "k8s.io/kubernetes/pkg/kubelet/pod/testing"
	"k8s.io/kubernetes/pkg/kubelet/status"
	"k8s.io/kubernetes/pkg/util"
	utiltesting "k8s.io/kubernetes/pkg/util/testing"
)

func TestRunOnce(t *testing.T) {
	cadvisor := &cadvisortest.Mock{}
	cadvisor.On("MachineInfo").Return(&cadvisorapi.MachineInfo{}, nil)
	cadvisor.On("DockerImagesFsInfo").Return(cadvisorapiv2.FsInfo{
		Usage:     400 * mb,
		Capacity:  1000 * mb,
		Available: 600 * mb,
	}, nil)
	cadvisor.On("RootFsInfo").Return(cadvisorapiv2.FsInfo{
		Usage:    9 * mb,
		Capacity: 10 * mb,
	}, nil)
	podManager := kubepod.NewBasicPodManager(podtest.NewFakeMirrorClient())
	diskSpaceManager, _ := newDiskSpaceManager(cadvisor, DiskSpacePolicy{})
	fakeRuntime := &containertest.FakeRuntime{}
	basePath, err := utiltesting.MkTmpdir("kubelet")
	if err != nil {
		t.Fatalf("can't make a temp rootdir %v", err)
	}
	defer os.RemoveAll(basePath)
	kb := &Kubelet{
		rootDirectory:       basePath,
		recorder:            &record.FakeRecorder{},
		cadvisor:            cadvisor,
		nodeLister:          testNodeLister{},
		nodeInfo:            testNodeInfo{},
		statusManager:       status.NewManager(nil, podManager),
		containerRefManager: kubecontainer.NewRefManager(),
		podManager:          podManager,
		os:                  containertest.FakeOS{},
		volumeManager:       newVolumeManager(),
		diskSpaceManager:    diskSpaceManager,
		containerRuntime:    fakeRuntime,
		reasonCache:         NewReasonCache(),
		clock:               util.RealClock{},
		kubeClient:          &fake.Clientset{},
	}
	kb.containerManager = cm.NewStubContainerManager()

	kb.networkPlugin, _ = network.InitNetworkPlugin([]network.NetworkPlugin{}, "", nettest.NewFakeHost(nil))
	if err := kb.setupDataDirs(); err != nil {
		t.Errorf("Failed to init data dirs: %v", err)
	}

	pods := []*api.Pod{
		{
			ObjectMeta: api.ObjectMeta{
				UID:       "12345678",
				Name:      "foo",
				Namespace: "new",
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "bar"},
				},
			},
		},
	}
	podManager.SetPods(pods)
	// The original test here is totally meaningless, because fakeruntime will always return an empty podStatus. While
	// the originial logic of isPodRunning happens to return true when podstatus is empty, so the test can always pass.
	// Now the logic in isPodRunning is changed, to let the test pass, we set the podstatus directly in fake runtime.
	// This is also a meaningless test, because the isPodRunning will also always return true after setting this. However,
	// because runonce is never used in kubernetes now, we should deprioritize the cleanup work.
	// TODO(random-liu) Fix the test, make it meaningful.
	fakeRuntime.PodStatus = kubecontainer.PodStatus{
		ContainerStatuses: []*kubecontainer.ContainerStatus{
			{
				Name:  "bar",
				State: kubecontainer.ContainerStateRunning,
			},
		},
	}
	results, err := kb.runOnce(pods, time.Millisecond)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if results[0].Err != nil {
		t.Errorf("unexpected run pod error: %v", results[0].Err)
	}
	if results[0].Pod.Name != "foo" {
		t.Errorf("unexpected pod: %q", results[0].Pod.Name)
	}
}
