/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package fake

import (
	core "k8s.io/kubernetes/pkg/client/testing/core"
	v1beta1 "k8s.io/kubernetes/pkg/client/typed/generated/extensions/v1beta1"
)

type FakeExtensions struct {
	*core.Fake
}

func (c *FakeExtensions) DaemonSets(namespace string) v1beta1.DaemonSetInterface {
	return &FakeDaemonSets{c, namespace}
}

func (c *FakeExtensions) Deployments(namespace string) v1beta1.DeploymentInterface {
	return &FakeDeployments{c, namespace}
}

func (c *FakeExtensions) HorizontalPodAutoscalers(namespace string) v1beta1.HorizontalPodAutoscalerInterface {
	return &FakeHorizontalPodAutoscalers{c, namespace}
}

func (c *FakeExtensions) Ingresses(namespace string) v1beta1.IngressInterface {
	return &FakeIngresses{c, namespace}
}

func (c *FakeExtensions) Jobs(namespace string) v1beta1.JobInterface {
	return &FakeJobs{c, namespace}
}

func (c *FakeExtensions) ReplicaSets(namespace string) v1beta1.ReplicaSetInterface {
	return &FakeReplicaSets{c, namespace}
}

func (c *FakeExtensions) Scales(namespace string) v1beta1.ScaleInterface {
	return &FakeScales{c, namespace}
}

func (c *FakeExtensions) ThirdPartyResources(namespace string) v1beta1.ThirdPartyResourceInterface {
	return &FakeThirdPartyResources{c, namespace}
}
