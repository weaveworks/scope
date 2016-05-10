package kubernetes

import (
	"io"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/wait"
)

// Client keeps track of running kubernetes pods and services
type Client interface {
	Stop()
	WalkPods(f func(Pod) error) error
	WalkServices(f func(Service) error) error
	WalkDeployments(f func(Deployment) error) error
	WalkReplicaSets(f func(ReplicaSet) error) error
	WalkReplicationControllers(f func(ReplicationController) error) error
	WalkNodes(f func(*api.Node) error) error

	WatchPods(f func(Event, Pod))

	GetLogs(namespaceID, podID string) (io.ReadCloser, error)
	DeletePod(namespaceID, podID string) error
	ScaleUp(resource, namespaceID, id string) error
	ScaleDown(resource, namespaceID, id string) error
}

type client struct {
	quit                       chan struct{}
	resyncPeriod               time.Duration
	client                     *unversioned.Client
	extensionsClient           *unversioned.ExtensionsClient
	podStore                   *cache.StoreToPodLister
	serviceStore               *cache.StoreToServiceLister
	deploymentStore            *cache.StoreToDeploymentLister
	replicaSetStore            *cache.StoreToReplicaSetLister
	replicationControllerStore *cache.StoreToReplicationControllerLister
	nodeStore                  *cache.StoreToNodeLister

	podWatchesMutex sync.Mutex
	podWatches      []func(Event, Pod)
}

// runReflectorUntil is equivalent to cache.Reflector.RunUntil, but it also logs
// errors, which cache.Reflector.RunUntil simply ignores
func runReflectorUntil(r *cache.Reflector, resyncPeriod time.Duration, stopCh <-chan struct{}) {
	loggingListAndWatch := func() {
		if err := r.ListAndWatch(stopCh); err != nil {
			log.Errorf("Kubernetes reflector: %v", err)
		}
	}
	go wait.Until(loggingListAndWatch, resyncPeriod, stopCh)
}

// NewClient returns a usable Client. Don't forget to Stop it.
func NewClient(addr string, resyncPeriod time.Duration) (Client, error) {
	var config *restclient.Config
	if addr != "" {
		config = &restclient.Config{Host: addr}
	} else {
		// If no API server address was provided, assume we are running
		// inside a pod. Try to connect to the API server through its
		// Service environment variables, using the default Service
		// Account Token.
		var err error
		if config, err = restclient.InClusterConfig(); err != nil {
			return nil, err
		}
	}

	c, err := unversioned.New(config)
	if err != nil {
		return nil, err
	}

	ec, err := unversioned.NewExtensions(config)
	if err != nil {
		return nil, err
	}

	result := &client{
		quit:             make(chan struct{}),
		resyncPeriod:     resyncPeriod,
		client:           c,
		extensionsClient: ec,
	}

	result.podStore = &cache.StoreToPodLister{Store: result.setupStore(c, "pods", &api.Pod{})}
	result.serviceStore = &cache.StoreToServiceLister{Store: result.setupStore(c, "services", &api.Service{})}
	result.deploymentStore = &cache.StoreToDeploymentLister{Store: result.setupStore(ec, "deployments", &extensions.Deployment{})}
	result.replicaSetStore = &cache.StoreToReplicaSetLister{Store: result.setupStore(ec, "replicasets", &extensions.ReplicaSet{})}
	result.replicationControllerStore = &cache.StoreToReplicationControllerLister{Store: result.setupStore(c, "replicationcontrollers", &api.ReplicationController{})}
	result.nodeStore = &cache.StoreToNodeLister{Store: result.setupStore(c, "nodes", &api.Node{})}
	return result, nil
}

func (c *client) setupStore(kclient cache.Getter, resource string, itemType interface{}) cache.Store {
	lw := cache.NewListWatchFromClient(kclient, resource, api.NamespaceAll, fields.Everything())
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	runReflectorUntil(cache.NewReflector(lw, itemType, store, c.resyncPeriod), c.resyncPeriod, c.quit)
	return store
}

func (c *client) WatchPods(f func(Event, Pod)) {
	c.podWatchesMutex.Lock()
	defer c.podWatchesMutex.Unlock()
	c.podWatches = append(c.podWatches, f)
}

func (c *client) triggerPodWatches(e Event, pod interface{}) {
	c.podWatchesMutex.Lock()
	defer c.podWatchesMutex.Unlock()
	for _, watch := range c.podWatches {
		watch(e, NewPod(pod.(*api.Pod)))
	}
}

func (c *client) WalkPods(f func(Pod) error) error {
	pods, err := c.podStore.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, pod := range pods {
		if err := f(NewPod(pod)); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkServices(f func(Service) error) error {
	list, err := c.serviceStore.List()
	if err != nil {
		return err
	}
	for i := range list.Items {
		if err := f(NewService(&(list.Items[i]))); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkDeployments(f func(Deployment) error) error {
	list, err := c.deploymentStore.List()
	if err != nil {
		return err
	}
	for i := range list {
		if err := f(NewDeployment(&(list[i]))); err != nil {
			return err
		}
	}
	return nil
}

// WalkReplicaSets calls f for each replica set
func (c *client) WalkReplicaSets(f func(ReplicaSet) error) error {
	list, err := c.replicaSetStore.List()
	if err != nil {
		return err
	}
	for i := range list {
		if err := f(NewReplicaSet(&(list[i]))); err != nil {
			return err
		}
	}
	return nil

}

// WalkReplicationcontrollers calls f for each replication controller
func (c *client) WalkReplicationControllers(f func(ReplicationController) error) error {
	list, err := c.replicationControllerStore.List()
	if err != nil {
		return err
	}
	for i := range list {
		if err := f(NewReplicationController(&(list[i]))); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkNodes(f func(*api.Node) error) error {
	list, err := c.nodeStore.List()
	if err != nil {
		return err
	}
	for i := range list.Items {
		if err := f(&(list.Items[i])); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) GetLogs(namespaceID, podID string) (io.ReadCloser, error) {
	return c.client.RESTClient.Get().
		Namespace(namespaceID).
		Name(podID).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true)).
		Param("previous", strconv.FormatBool(false)).
		Param("timestamps", strconv.FormatBool(true)).
		Stream()
}

func (c *client) DeletePod(namespaceID, podID string) error {
	return c.client.RESTClient.Delete().
		Namespace(namespaceID).
		Name(podID).
		Resource("pods").Do().Error()
}

func (c *client) ScaleUp(resource, namespaceID, id string) error {
	return c.modifyScale(resource, namespaceID, id, func(scale *extensions.Scale) {
		scale.Spec.Replicas++
	})
}

func (c *client) ScaleDown(resource, namespaceID, id string) error {
	return c.modifyScale(resource, namespaceID, id, func(scale *extensions.Scale) {
		scale.Spec.Replicas--
	})
}

func (c *client) modifyScale(resource, namespace, id string, f func(*extensions.Scale)) error {
	scaler := c.extensionsClient.Scales(namespace)
	scale, err := scaler.Get(resource, id)
	if err != nil {
		return err
	}
	f(scale)
	_, err = scaler.Update(resource, scale)
	return err
}

func (c *client) Stop() {
	close(c.quit)
}
