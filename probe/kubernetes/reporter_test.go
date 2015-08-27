package kubernetes_test

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

var (
	podTypeMeta = unversioned.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	apiPod1 = api.Pod{
		TypeMeta: podTypeMeta,
		ObjectMeta: api.ObjectMeta{
			Name:              "pong-a",
			Namespace:         "ping",
			CreationTimestamp: unversioned.Now(),
			Labels:            map[string]string{"ponger": "true"},
		},
		Status: api.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []api.ContainerStatus{
				{ContainerID: "container1"},
				{ContainerID: "container2"},
			},
		},
	}
	apiPod2 = api.Pod{
		TypeMeta: podTypeMeta,
		ObjectMeta: api.ObjectMeta{
			Name:              "pong-b",
			Namespace:         "ping",
			CreationTimestamp: unversioned.Now(),
			Labels:            map[string]string{"ponger": "true"},
		},
		Status: api.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []api.ContainerStatus{
				{ContainerID: "container3"},
				{ContainerID: "container4"},
			},
		},
	}
	apiService1 = api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:              "pongservice",
			Namespace:         "ping",
			CreationTimestamp: unversioned.Now(),
		},
		Spec: api.ServiceSpec{
			Type:      api.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.1.1",
			Ports: []api.ServicePort{
				{Protocol: "TCP", Port: 6379},
			},
			Selector: map[string]string{"ponger": "true"},
		},
		Status: api.ServiceStatus{
			LoadBalancer: api.LoadBalancerStatus{
				Ingress: []api.LoadBalancerIngress{
					{IP: "10.0.1.2"},
				},
			},
		},
	}
	pod1               = kubernetes.NewPod(&apiPod1)
	pod2               = kubernetes.NewPod(&apiPod2)
	service1           = kubernetes.NewService(&apiService1)
	mockClientInstance = &mockClient{
		pods:     []kubernetes.Pod{pod1, pod2},
		services: []kubernetes.Service{service1},
	}
)

type mockClient struct {
	pods     []kubernetes.Pod
	services []kubernetes.Service
}

func (c *mockClient) Stop() {}
func (c *mockClient) WalkPods(f func(kubernetes.Pod) error) error {
	for _, pod := range c.pods {
		if err := f(pod); err != nil {
			return err
		}
	}
	return nil
}
func (c *mockClient) WalkServices(f func(kubernetes.Service) error) error {
	for _, service := range c.services {
		if err := f(service); err != nil {
			return err
		}
	}
	return nil
}

func TestReporter(t *testing.T) {
	want := report.MakeReport()
	pod1ID := report.MakePodNodeID("ping", "pong-a")
	pod2ID := report.MakePodNodeID("ping", "pong-b")
	want.Pod = report.Topology{
		Nodes: report.Nodes{
			pod1ID: report.MakeNodeWith(map[string]string{
				kubernetes.PodID:           "ping/pong-a",
				kubernetes.PodName:         "pong-a",
				kubernetes.Namespace:       "ping",
				kubernetes.PodCreated:      pod1.Created(),
				kubernetes.PodContainerIDs: "container1 container2",
				kubernetes.ServiceIDs:      "ping/pongservice",
			}),
			pod2ID: report.MakeNodeWith(map[string]string{
				kubernetes.PodID:           "ping/pong-b",
				kubernetes.PodName:         "pong-b",
				kubernetes.Namespace:       "ping",
				kubernetes.PodCreated:      pod1.Created(),
				kubernetes.PodContainerIDs: "container3 container4",
				kubernetes.ServiceIDs:      "ping/pongservice",
			}),
		},
	}
	want.Service = report.Topology{
		Nodes: report.Nodes{
			report.MakeServiceNodeID("ping", "pongservice"): report.MakeNodeWith(map[string]string{
				kubernetes.ServiceID:      "ping/pongservice",
				kubernetes.ServiceName:    "pongservice",
				kubernetes.Namespace:      "ping",
				kubernetes.ServiceCreated: pod1.Created(),
			}),
		},
	}

	reporter := kubernetes.NewReporter(mockClientInstance)
	have, _ := reporter.Report()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
