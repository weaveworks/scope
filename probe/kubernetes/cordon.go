/*
Copied from

	https://github.com/kubernetes/kubectl/blob/master/pkg/drain/cordon.go

at commit f9460c53339c4bb60b20031e3c6125e8bac679e2 to add node cordon feature.	
*/
/*
Copyright 2019 The Kubernetes Authors.

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

package kubernetes

import (
	"encoding/json"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

// CordonHelper wraps functionality to cordon/uncordon nodes
type cordonHelper struct {
	node    *apiv1.Node
	desired bool
}

// NewCordonHelper returns a new CordonHelper
func newCordonHelper(node *apiv1.Node) *cordonHelper {
	return &cordonHelper{
		node: node,
	}
}

// UpdateIfRequired returns true if c.node.Spec.Unschedulable isn't already set,
// or false when no change is needed
func (c *cordonHelper) updateIfRequired(desired bool) bool {
	c.desired = desired

	return c.node.Spec.Unschedulable != c.desired
}

// PatchOrReplace uses given clientset to update the node status, either by patching or
// updating the given node object; it may return error if the object cannot be encoded as
// JSON, or if either patch or update calls fail; it will also return a second error
// whenever creating a patch has failed
func (c *cordonHelper) patchOrReplace(clientset kubernetes.Interface, serverDryRun bool) (error, error) {
	client := clientset.CoreV1().Nodes()

	oldData, err := json.Marshal(c.node)
	if err != nil {
		return err, nil
	}

	c.node.Spec.Unschedulable = c.desired

	newData, err := json.Marshal(c.node)
	if err != nil {
		return err, nil
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, c.node)
	if patchErr == nil {
		_, err = client.Patch(c.node.Name, types.StrategicMergePatchType, patchBytes)
	} else {
		updateOptions := metav1.UpdateOptions{}
		if serverDryRun {
			updateOptions.DryRun = []string{metav1.DryRunAll}
		}
		_, err = client.Update(c.node)
	}
	return err, patchErr
}
