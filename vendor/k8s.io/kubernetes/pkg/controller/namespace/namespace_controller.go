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

package namespace

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/typed/dynamic"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/runtime"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/util/workqueue"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/golang/glog"
)

// NamespaceController is responsible for performing actions dependent upon a namespace phase
type NamespaceController struct {
	// client that purges namespace content, must have list/delete privileges on all content
	kubeClient clientset.Interface
	// clientPool manages a pool of dynamic clients
	clientPool dynamic.ClientPool
	// store that holds the namespaces
	store cache.Store
	// controller that observes the namespaces
	controller *framework.Controller
	// namespaces that have been queued up for processing by workers
	queue *workqueue.Type
	// list of preferred group versions and their corresponding resource set for namespace deletion
	groupVersionResources []unversioned.GroupVersionResource
	// opCache is a cache to remember if a particular operation is not supported to aid dynamic client.
	opCache operationNotSupportedCache
	// finalizerToken is the finalizer token managed by this controller
	finalizerToken api.FinalizerName
}

// NewNamespaceController creates a new NamespaceController
func NewNamespaceController(
	kubeClient clientset.Interface,
	clientPool dynamic.ClientPool,
	groupVersionResources []unversioned.GroupVersionResource,
	resyncPeriod time.Duration,
	finalizerToken api.FinalizerName) *NamespaceController {
	// create the controller so we can inject the enqueue function
	namespaceController := &NamespaceController{
		kubeClient: kubeClient,
		clientPool: clientPool,
		queue:      workqueue.New(),
		groupVersionResources: groupVersionResources,
		opCache:               operationNotSupportedCache{},
		finalizerToken:        finalizerToken,
	}

	// configure the backing store/controller
	store, controller := framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return kubeClient.Core().Namespaces().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return kubeClient.Core().Namespaces().Watch(options)
			},
		},
		&api.Namespace{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				namespace := obj.(*api.Namespace)
				namespaceController.enqueueNamespace(namespace)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				namespace := newObj.(*api.Namespace)
				namespaceController.enqueueNamespace(namespace)
			},
		},
	)

	namespaceController.store = store
	namespaceController.controller = controller
	return namespaceController
}

// enqueueNamespace adds an object to the controller work queue
// obj could be an *api.Namespace, or a DeletionFinalStateUnknown item.
func (nm *NamespaceController) enqueueNamespace(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	nm.queue.Add(key)
}

// worker processes the queue of namespace objects.
// Each namespace can be in the queue at most once.
// The system ensures that no two workers can process
// the same namespace at the same time.
func (nm *NamespaceController) worker() {
	for {
		func() {
			key, quit := nm.queue.Get()
			if quit {
				return
			}
			defer nm.queue.Done(key)
			if err := nm.syncNamespaceFromKey(key.(string)); err != nil {
				if estimate, ok := err.(*contentRemainingError); ok {
					go func() {
						defer utilruntime.HandleCrash()
						t := estimate.Estimate/2 + 1
						glog.V(4).Infof("Content remaining in namespace %s, waiting %d seconds", key, t)
						time.Sleep(time.Duration(t) * time.Second)
						nm.queue.Add(key)
					}()
				} else {
					// rather than wait for a full resync, re-add the namespace to the queue to be processed
					nm.queue.Add(key)
					utilruntime.HandleError(err)
				}
			}
		}()
	}
}

// syncNamespaceFromKey looks for a namespace with the specified key in its store and synchronizes it
func (nm *NamespaceController) syncNamespaceFromKey(key string) (err error) {
	startTime := time.Now()
	defer glog.V(4).Infof("Finished syncing namespace %q (%v)", key, time.Now().Sub(startTime))

	obj, exists, err := nm.store.GetByKey(key)
	if !exists {
		glog.Infof("Namespace has been deleted %v", key)
		return nil
	}
	if err != nil {
		glog.Infof("Unable to retrieve namespace %v from store: %v", key, err)
		nm.queue.Add(key)
		return err
	}
	namespace := obj.(*api.Namespace)
	return syncNamespace(nm.kubeClient, nm.clientPool, nm.opCache, nm.groupVersionResources, namespace, nm.finalizerToken)
}

// Run starts observing the system with the specified number of workers.
func (nm *NamespaceController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	go nm.controller.Run(stopCh)
	for i := 0; i < workers; i++ {
		go wait.Until(nm.worker, time.Second, stopCh)
	}
	<-stopCh
	glog.Infof("Shutting down NamespaceController")
	nm.queue.ShutDown()
}
