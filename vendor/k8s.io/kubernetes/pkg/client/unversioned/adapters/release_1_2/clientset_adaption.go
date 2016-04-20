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

package release_1_2

import (
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_2"
	v1core "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_2/typed/core/v1"
	v1beta1extensions "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_2/typed/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/client/typed/discovery"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

// FromUnversionedClient adapts a unversioned.Client to a release_1_2.Clientset.
// This function is temporary. We will remove it when everyone has moved to using
// Clientset. New code should NOT use this function.
func FromUnversionedClient(c *unversioned.Client) *release_1_2.Clientset {
	var clientset release_1_2.Clientset
	if c != nil {
		clientset.CoreClient = v1core.New(c.RESTClient)
	} else {
		clientset.CoreClient = v1core.New(nil)
	}
	if c != nil && c.ExtensionsClient != nil {
		clientset.ExtensionsClient = v1beta1extensions.New(c.ExtensionsClient.RESTClient)
	} else {
		clientset.ExtensionsClient = v1beta1extensions.New(nil)
	}

	if c != nil && c.DiscoveryClient != nil {
		clientset.DiscoveryClient = discovery.NewDiscoveryClient(c.DiscoveryClient.RESTClient)
	} else {
		clientset.DiscoveryClient = discovery.NewDiscoveryClient(nil)
	}

	return &clientset
}
