package kubernetes

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/weaveworks/common/backoff"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apiappsv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	apibatchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	apibatchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	apiextensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Client keeps track of running kubernetes pods and services
type Client interface {
	Stop()
	WalkPods(f func(Pod) error) error
	WalkServices(f func(Service) error) error
	WalkDeployments(f func(Deployment) error) error
	WalkReplicaSets(f func(ReplicaSet) error) error
	WalkDaemonSets(f func(DaemonSet) error) error
	WalkStatefulSets(f func(StatefulSet) error) error
	WalkCronJobs(f func(CronJob) error) error
	WalkReplicationControllers(f func(ReplicationController) error) error
	WalkNodes(f func(*apiv1.Node) error) error

	WatchPods(f func(Event, Pod))

	GetLogs(namespaceID, podID string) (io.ReadCloser, error)
	DeletePod(namespaceID, podID string) error
	ScaleUp(resource, namespaceID, id string) error
	ScaleDown(resource, namespaceID, id string) error
}

type client struct {
	quit                       chan struct{}
	resyncPeriod               time.Duration
	client                     *kubernetes.Clientset
	podStore                   cache.Store
	serviceStore               cache.Store
	deploymentStore            cache.Store
	replicaSetStore            cache.Store
	daemonSetStore             cache.Store
	statefulSetStore           cache.Store
	jobStore                   cache.Store
	cronJobStore               cache.Store
	replicationControllerStore cache.Store
	nodeStore                  cache.Store

	podWatchesMutex sync.Mutex
	podWatches      []func(Event, Pod)
}

// runReflectorUntil runs cache.Reflector#ListAndWatch in an endless loop.
// Errors are logged and retried with exponential backoff.
func runReflectorUntil(r *cache.Reflector, resyncPeriod time.Duration, stopCh <-chan struct{}, msg string) {
	listAndWatch := func() (bool, error) {
		select {
		case <-stopCh:
			return true, nil
		default:
			err := r.ListAndWatch(stopCh)
			return false, err
		}
	}
	bo := backoff.New(listAndWatch, fmt.Sprintf("Kubernetes reflector (%s)", msg))
	bo.SetInitialBackoff(resyncPeriod)
	bo.SetMaxBackoff(5 * time.Minute)
	go bo.Start()
}

// ClientConfig establishes the configuration for the kubernetes client
type ClientConfig struct {
	Interval             time.Duration
	CertificateAuthority string
	ClientCertificate    string
	ClientKey            string
	Cluster              string
	Context              string
	Insecure             bool
	Kubeconfig           string
	Password             string
	Server               string
	Token                string
	User                 string
	Username             string
}

// NewClient returns a usable Client. Don't forget to Stop it.
func NewClient(config ClientConfig) (Client, error) {
	var restConfig *rest.Config
	if config.Server == "" && config.Kubeconfig == "" {
		// If no API server address or kubeconfig was provided, assume we are running
		// inside a pod. Try to connect to the API server through its
		// Service environment variables, using the default Service
		// Account Token.
		var err error
		if restConfig, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		var err error
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.Kubeconfig},
			&clientcmd.ConfigOverrides{
				AuthInfo: clientcmdapi.AuthInfo{
					ClientCertificate: config.ClientCertificate,
					ClientKey:         config.ClientKey,
					Token:             config.Token,
					Username:          config.Username,
					Password:          config.Password,
				},
				ClusterInfo: clientcmdapi.Cluster{
					Server:                config.Server,
					InsecureSkipTLSVerify: config.Insecure,
					CertificateAuthority:  config.CertificateAuthority,
				},
				Context: clientcmdapi.Context{
					Cluster:  config.Cluster,
					AuthInfo: config.User,
				},
				CurrentContext: config.Context,
			},
		).ClientConfig()
		if err != nil {
			return nil, err
		}

	}
	log.Infof("kubernetes: targeting api server %s", restConfig.Host)

	c, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	result := &client{
		quit:         make(chan struct{}),
		resyncPeriod: config.Interval,
		client:       c,
	}

	podStore := NewEventStore(result.triggerPodWatches, cache.MetaNamespaceKeyFunc)
	result.podStore = result.setupStore(c.CoreV1Client.RESTClient(), "pods", &apiv1.Pod{}, podStore)

	result.serviceStore = result.setupStore(c.CoreV1Client.RESTClient(), "services", &apiv1.Service{}, nil)
	result.replicationControllerStore = result.setupStore(c.CoreV1Client.RESTClient(), "replicationcontrollers", &apiv1.ReplicationController{}, nil)
	result.nodeStore = result.setupStore(c.CoreV1Client.RESTClient(), "nodes", &apiv1.Node{}, nil)

	// We list deployments here to check if this version of kubernetes is >= 1.2.
	// We would use NegotiateVersion, but Kubernetes 1.1 "supports"
	// extensions/v1beta1, but not deployments, replicasets or daemonsets.
	if _, err := c.Extensions().Deployments(metav1.NamespaceAll).List(metav1.ListOptions{}); err != nil {
		log.Infof("Deployments, ReplicaSets and DaemonSets are not supported by this Kubernetes version: %v", err)
	} else {
		result.deploymentStore = result.setupStore(c.ExtensionsV1beta1Client.RESTClient(), "deployments", &apiextensionsv1beta1.Deployment{}, nil)
		result.replicaSetStore = result.setupStore(c.ExtensionsV1beta1Client.RESTClient(), "replicasets", &apiextensionsv1beta1.ReplicaSet{}, nil)
		result.daemonSetStore = result.setupStore(c.ExtensionsV1beta1Client.RESTClient(), "daemonsets", &apiextensionsv1beta1.DaemonSet{}, nil)
	}
	// CronJobs and StatefulSets were introduced later. Easiest to use the same technique.
	if _, err := c.BatchV2alpha1().CronJobs(metav1.NamespaceAll).List(metav1.ListOptions{}); err != nil {
		log.Infof("CronJobs are not supported by this Kubernetes version: %v", err)
	} else {
		result.jobStore = result.setupStore(c.BatchV1Client.RESTClient(), "jobs", &apibatchv1.Job{}, nil)
		result.cronJobStore = result.setupStore(c.BatchV2alpha1Client.RESTClient(), "cronjobs", &apibatchv2alpha1.CronJob{}, nil)
	}
	if _, err := c.Apps().StatefulSets(metav1.NamespaceAll).List(metav1.ListOptions{}); err != nil {
		log.Infof("StatefulSets are not supported by this Kubernetes version: %v", err)
	} else {
		result.statefulSetStore = result.setupStore(c.AppsV1beta1Client.RESTClient(), "statefulsets", &apiappsv1beta1.StatefulSet{}, nil)
	}

	return result, nil
}

func (c *client) setupStore(kclient cache.Getter, resource string, itemType interface{}, nonDefaultStore cache.Store) cache.Store {
	lw := cache.NewListWatchFromClient(kclient, resource, metav1.NamespaceAll, fields.Everything())
	store := nonDefaultStore
	if store == nil {
		store = cache.NewStore(cache.MetaNamespaceKeyFunc)
	}
	runReflectorUntil(cache.NewReflector(lw, itemType, store, c.resyncPeriod), c.resyncPeriod, c.quit, resource)
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
		watch(e, NewPod(pod.(*apiv1.Pod)))
	}
}

func (c *client) WalkPods(f func(Pod) error) error {
	for _, m := range c.podStore.List() {
		pod := m.(*apiv1.Pod)
		if err := f(NewPod(pod)); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkServices(f func(Service) error) error {
	for _, m := range c.serviceStore.List() {
		s := m.(*apiv1.Service)
		if err := f(NewService(s)); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkDeployments(f func(Deployment) error) error {
	if c.deploymentStore == nil {
		return nil
	}
	for _, m := range c.deploymentStore.List() {
		d := m.(*apiextensionsv1beta1.Deployment)
		if err := f(NewDeployment(d)); err != nil {
			return err
		}
	}
	return nil
}

// WalkReplicaSets calls f for each replica set
func (c *client) WalkReplicaSets(f func(ReplicaSet) error) error {
	if c.replicaSetStore == nil {
		return nil
	}
	for _, m := range c.replicaSetStore.List() {
		rs := m.(*apiextensionsv1beta1.ReplicaSet)
		if err := f(NewReplicaSet(rs)); err != nil {
			return err
		}
	}
	return nil

}

// WalkReplicationcontrollers calls f for each replication controller
func (c *client) WalkReplicationControllers(f func(ReplicationController) error) error {
	for _, m := range c.replicationControllerStore.List() {
		rc := m.(*apiv1.ReplicationController)
		if err := f(NewReplicationController(rc)); err != nil {
			return err
		}
	}
	return nil
}

// WalkDaemonSets calls f for each daemonset
func (c *client) WalkDaemonSets(f func(DaemonSet) error) error {
	if c.daemonSetStore == nil {
		return nil
	}
	for _, m := range c.daemonSetStore.List() {
		ds := m.(*apiextensionsv1beta1.DaemonSet)
		if err := f(NewDaemonSet(ds)); err != nil {
			return err
		}
	}
	return nil
}

// WalkStatefulSets calls f for each statefulset
func (c *client) WalkStatefulSets(f func(StatefulSet) error) error {
	if c.statefulSetStore == nil {
		return nil
	}
	for _, m := range c.statefulSetStore.List() {
		s := m.(*apiappsv1beta1.StatefulSet)
		if err := f(NewStatefulSet(s)); err != nil {
			return err
		}
	}
	return nil
}

// WalkCronJobs calls f for each cronjob
func (c *client) WalkCronJobs(f func(CronJob) error) error {
	if c.cronJobStore == nil {
		return nil
	}
	// We index jobs by id to make lookup for each cronjob more efficient
	jobs := map[types.UID]*apibatchv1.Job{}
	for _, m := range c.jobStore.List() {
		j := m.(*apibatchv1.Job)
		jobs[j.UID] = j
	}
	for _, m := range c.cronJobStore.List() {
		cj := m.(*apibatchv2alpha1.CronJob)
		if err := f(NewCronJob(cj, jobs)); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) WalkNodes(f func(*apiv1.Node) error) error {
	for _, m := range c.nodeStore.List() {
		node := m.(*apiv1.Node)
		if err := f(node); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) GetLogs(namespaceID, podID string) (io.ReadCloser, error) {
	req := c.client.CoreV1().Pods(namespaceID).GetLogs(
		podID,
		&apiv1.PodLogOptions{
			Follow:     true,
			Timestamps: true,
		},
	)
	return req.Stream()
}

func (c *client) DeletePod(namespaceID, podID string) error {
	return c.client.CoreV1().Pods(namespaceID).Delete(podID, &metav1.DeleteOptions{})
}

func (c *client) ScaleUp(resource, namespaceID, id string) error {
	return c.modifyScale(resource, namespaceID, id, func(scale *apiextensionsv1beta1.Scale) {
		scale.Spec.Replicas++
	})
}

func (c *client) ScaleDown(resource, namespaceID, id string) error {
	return c.modifyScale(resource, namespaceID, id, func(scale *apiextensionsv1beta1.Scale) {
		scale.Spec.Replicas--
	})
}

func (c *client) modifyScale(resource, namespace, id string, f func(*apiextensionsv1beta1.Scale)) error {
	scaler := c.client.Extensions().Scales(namespace)
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
