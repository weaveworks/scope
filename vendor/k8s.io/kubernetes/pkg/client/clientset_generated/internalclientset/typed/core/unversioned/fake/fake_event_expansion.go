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

package fake

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/runtime"
)

func (c *FakeEvents) CreateWithEventNamespace(event *api.Event) (*api.Event, error) {
	action := core.NewRootCreateAction("events", event)
	if c.ns != "" {
		action = core.NewCreateAction("events", c.ns, event)
	}
	obj, err := c.Fake.Invokes(action, event)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Event), err
}

// Update replaces an existing event. Returns the copy of the event the server returns, or an error.
func (c *FakeEvents) UpdateWithEventNamespace(event *api.Event) (*api.Event, error) {
	action := core.NewRootUpdateAction("events", event)
	if c.ns != "" {
		action = core.NewUpdateAction("events", c.ns, event)
	}
	obj, err := c.Fake.Invokes(action, event)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Event), err
}

// Patch patches an existing event. Returns the copy of the event the server returns, or an error.
func (c *FakeEvents) Patch(event *api.Event, data []byte) (*api.Event, error) {
	action := core.NewRootPatchAction("events", event)
	if c.ns != "" {
		action = core.NewPatchAction("events", c.ns, event)
	}
	obj, err := c.Fake.Invokes(action, event)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Event), err
}

// Search returns a list of events matching the specified object.
func (c *FakeEvents) Search(objOrRef runtime.Object) (*api.EventList, error) {
	action := core.NewRootListAction("events", api.ListOptions{})
	if c.ns != "" {
		action = core.NewListAction("events", c.ns, api.ListOptions{})
	}
	obj, err := c.Fake.Invokes(action, &api.EventList{})
	if obj == nil {
		return nil, err
	}

	return obj.(*api.EventList), err
}

func (c *FakeEvents) GetFieldSelector(involvedObjectName, involvedObjectNamespace, involvedObjectKind, involvedObjectUID *string) fields.Selector {
	action := core.GenericActionImpl{}
	action.Verb = "get-field-selector"
	action.Resource = "events"

	c.Fake.Invokes(action, nil)
	return fields.Everything()
}
