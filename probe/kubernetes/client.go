package kubernetes

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
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
}

type client struct {
	quit             chan struct{}
	client           cache.Getter
	podReflector     *cache.Reflector
	serviceReflector *cache.Reflector
	podStore         *cache.StoreToPodLister
	serviceStore     *cache.StoreToServiceLister
}

// NewClient returns a usable Client. Don't forget to Stop it.
func NewClient(addr string, resyncPeriod time.Duration) (Client, error) {
	c, err := unversioned.New(&unversioned.Config{Host: addr})
	if err != nil {
		return nil, err
	}

	podListWatch := cache.NewListWatchFromClient(c, "pods", api.NamespaceAll, fields.Everything())
	podStore := cache.NewStore(cache.MetaNamespaceKeyFunc)
	podReflector := cache.NewReflector(podListWatch, &api.Pod{}, podStore, resyncPeriod)

	serviceListWatch := cache.NewListWatchFromClient(c, "services", api.NamespaceAll, fields.Everything())
	serviceStore := cache.NewStore(cache.MetaNamespaceKeyFunc)
	serviceReflector := cache.NewReflector(serviceListWatch, &api.Service{}, serviceStore, resyncPeriod)

	quit := make(chan struct{})
	podReflector.RunUntil(quit)
	serviceReflector.RunUntil(quit)

	return &client{
		quit:             quit,
		client:           c,
		podReflector:     podReflector,
		podStore:         &cache.StoreToPodLister{Store: podStore},
		serviceReflector: serviceReflector,
		serviceStore:     &cache.StoreToServiceLister{Store: serviceStore},
	}, nil
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

func (c *client) Stop() {
	close(c.quit)
}
