package kubernetes

import (
	"io"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/wait"
)

// These constants are keys used in node metadata
const (
	Namespace = "kubernetes_namespace"
)

// Client keeps track of running kubernetes pods and services
type Client interface {
	Stop()
	WalkPods(f func(Pod) error) error
	WalkServices(f func(Service) error) error
	WalkNodes(f func(*api.Node) error) error

	WatchPods(f func(Event, Pod))

	GetLogs(namespaceID, podID string) (io.ReadCloser, error)
	DeletePod(namespaceID, podID string) error
}

type client struct {
	quit             chan struct{}
	client           *unversioned.Client
	podReflector     *cache.Reflector
	serviceReflector *cache.Reflector
	nodeReflector    *cache.Reflector
	podStore         *cache.StoreToPodLister
	serviceStore     *cache.StoreToServiceLister
	nodeStore        *cache.StoreToNodeLister

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

	result := &client{
		quit:   make(chan struct{}),
		client: c,
	}

	podListWatch := cache.NewListWatchFromClient(c, "pods", api.NamespaceAll, fields.Everything())
	podStore := NewEventStore(result.triggerPodWatches, cache.MetaNamespaceKeyFunc)
	result.podStore = &cache.StoreToPodLister{Store: podStore}
	result.podReflector = cache.NewReflector(podListWatch, &api.Pod{}, podStore, resyncPeriod)

	serviceListWatch := cache.NewListWatchFromClient(c, "services", api.NamespaceAll, fields.Everything())
	serviceStore := cache.NewStore(cache.MetaNamespaceKeyFunc)
	result.serviceStore = &cache.StoreToServiceLister{Store: serviceStore}
	result.serviceReflector = cache.NewReflector(serviceListWatch, &api.Service{}, serviceStore, resyncPeriod)

	nodeListWatch := cache.NewListWatchFromClient(c, "nodes", api.NamespaceAll, fields.Everything())
	nodeStore := cache.NewStore(cache.MetaNamespaceKeyFunc)
	result.nodeStore = &cache.StoreToNodeLister{Store: nodeStore}
	result.nodeReflector = cache.NewReflector(nodeListWatch, &api.Node{}, nodeStore, resyncPeriod)

	runReflectorUntil(result.podReflector, resyncPeriod, result.quit)
	runReflectorUntil(result.serviceReflector, resyncPeriod, result.quit)
	runReflectorUntil(result.nodeReflector, resyncPeriod, result.quit)
	return result, nil
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

func (c *client) Stop() {
	close(c.quit)
}
