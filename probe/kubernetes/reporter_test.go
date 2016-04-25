package kubernetes_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

var (
	nodeName    = "nodename"
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
		Spec: api.PodSpec{
			NodeName: nodeName,
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
		Spec: api.PodSpec{
			NodeName: nodeName,
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
			Type:           api.ServiceTypeLoadBalancer,
			ClusterIP:      "10.0.1.1",
			LoadBalancerIP: "10.0.1.2",
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
	pod1     = kubernetes.NewPod(&apiPod1)
	pod2     = kubernetes.NewPod(&apiPod2)
	service1 = kubernetes.NewService(&apiService1)
)

func newMockClient() *mockClient {
	return &mockClient{
		pods:     []kubernetes.Pod{pod1, pod2},
		services: []kubernetes.Service{service1},
		logs:     map[string]io.ReadCloser{},
	}
}

type mockClient struct {
	pods     []kubernetes.Pod
	services []kubernetes.Service
	logs     map[string]io.ReadCloser
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
func (*mockClient) WalkNodes(f func(*api.Node) error) error {
	return nil
}
func (c *mockClient) GetLogs(namespaceID, podName string) (io.ReadCloser, error) {
	r, ok := c.logs[report.MakePodNodeID(namespaceID, podName)]
	if !ok {
		return nil, fmt.Errorf("Not found")
	}
	return r, nil
}

type mockPipeClient map[string]xfer.Pipe

func (c mockPipeClient) PipeConnection(appID, id string, pipe xfer.Pipe) error {
	c[id] = pipe
	return nil
}

func (c mockPipeClient) PipeClose(appID, id string) error {
	err := c[id].Close()
	delete(c, id)
	return err
}

func TestReporter(t *testing.T) {
	oldGetNodeName := kubernetes.GetNodeName
	defer func() { kubernetes.GetNodeName = oldGetNodeName }()
	kubernetes.GetNodeName = func(*kubernetes.Reporter) (string, error) {
		return nodeName, nil
	}

	pod1ID := report.MakePodNodeID("ping", "pong-a")
	pod2ID := report.MakePodNodeID("ping", "pong-b")
	serviceID := report.MakeServiceNodeID("ping", "pongservice")
	rpt, _ := kubernetes.NewReporter(newMockClient(), nil, "").Report()

	// Reporter should have added the following pods
	for _, pod := range []struct {
		id            string
		parentService string
		latest        map[string]string
	}{
		{pod1ID, serviceID, map[string]string{
			kubernetes.PodID:           "ping/pong-a",
			kubernetes.PodName:         "pong-a",
			kubernetes.Namespace:       "ping",
			kubernetes.PodCreated:      pod1.Created(),
			kubernetes.PodContainerIDs: "container1 container2",
			kubernetes.ServiceIDs:      "ping/pongservice",
		}},
		{pod2ID, serviceID, map[string]string{
			kubernetes.PodID:           "ping/pong-b",
			kubernetes.PodName:         "pong-b",
			kubernetes.Namespace:       "ping",
			kubernetes.PodCreated:      pod1.Created(),
			kubernetes.PodContainerIDs: "container3 container4",
			kubernetes.ServiceIDs:      "ping/pongservice",
		}},
	} {
		node, ok := rpt.Pod.Nodes[pod.id]
		if !ok {
			t.Errorf("Expected report to have pod %q, but not found", pod.id)
		}

		if parents, ok := node.Parents.Lookup(report.Service); !ok || !parents.Contains(pod.parentService) {
			t.Errorf("Expected pod %s to have parent service %q, got %q", pod.id, pod.parentService, parents)
		}

		for k, want := range pod.latest {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected pod %s latest %q: %q, got %q", pod.id, k, want, have)
			}
		}
	}

	// Reporter should have added a service
	{
		node, ok := rpt.Service.Nodes[serviceID]
		if !ok {
			t.Errorf("Expected report to have service %q, but not found", serviceID)
		}

		for k, want := range map[string]string{
			kubernetes.ServiceID:      "ping/pongservice",
			kubernetes.ServiceName:    "pongservice",
			kubernetes.Namespace:      "ping",
			kubernetes.ServiceCreated: pod1.Created(),
		} {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected service %s latest %q: %q, got %q", serviceID, k, want, have)
			}
		}
	}

	// Reporter should have tagged the containers
	for _, pod := range []struct {
		id, nodeID string
		containers []string
	}{
		{"ping/pong-a", pod1ID, []string{"container1", "container2"}},
		{"ping/pong-b", pod2ID, []string{"container3", "container4"}},
	} {
		for _, containerID := range pod.containers {
			node, ok := rpt.Container.Nodes[report.MakeContainerNodeID(containerID)]
			if !ok {
				t.Errorf("Expected report to have container %q, but not found", containerID)
			}
			// container should have pod id
			if have, ok := node.Latest.Lookup(kubernetes.PodID); !ok || have != pod.id {
				t.Errorf("Expected container %s latest %q: %q, got %q", containerID, kubernetes.PodID, pod.id, have)
			}
			// container should have namespace
			if have, ok := node.Latest.Lookup(kubernetes.Namespace); !ok || have != "ping" {
				t.Errorf("Expected container %s latest %q: %q, got %q", containerID, kubernetes.Namespace, "ping", have)
			}
			// container should have pod parent
			if parents, ok := node.Parents.Lookup(report.Pod); !ok || !parents.Contains(pod.nodeID) {
				t.Errorf("Expected container %s to have parent service %q, got %q", containerID, pod.nodeID, parents)
			}
		}
	}
}

type callbackReadCloser struct {
	io.Reader
	close func() error
}

func (c *callbackReadCloser) Close() error { return c.close() }

func TestReporterGetLogs(t *testing.T) {
	oldGetNodeName := kubernetes.GetNodeName
	defer func() { kubernetes.GetNodeName = oldGetNodeName }()
	kubernetes.GetNodeName = func(*kubernetes.Reporter) (string, error) {
		return nodeName, nil
	}

	client := newMockClient()
	pipes := mockPipeClient{}
	reporter := kubernetes.NewReporter(client, pipes, "")

	// Should error on invalid IDs
	{
		resp := reporter.GetLogs(xfer.Request{
			NodeID:  "invalidID",
			Control: kubernetes.GetLogs,
		})
		if want := "Invalid ID: invalidID"; resp.Error != want {
			t.Errorf("Expected error on invalid ID: %q, got %q", want, resp.Error)
		}
	}

	// Should pass through errors from k8s (e.g if pod does not exist)
	{
		resp := reporter.GetLogs(xfer.Request{
			AppID:   "appID",
			NodeID:  report.MakePodNodeID("not", "found"),
			Control: kubernetes.GetLogs,
		})
		if want := "Not found"; resp.Error != want {
			t.Errorf("Expected error on invalid ID: %q, got %q", want, resp.Error)
		}
	}

	pod1ID := report.MakePodNodeID("ping", "pong-a")
	pod1Request := xfer.Request{
		AppID:   "appID",
		NodeID:  pod1ID,
		Control: kubernetes.GetLogs,
	}

	// Inject our logs content, and watch for it to be closed
	closed := false
	wantContents := "logs: ping/pong-a"
	client.logs[pod1ID] = &callbackReadCloser{Reader: strings.NewReader(wantContents), close: func() error {
		closed = true
		return nil
	}}

	// Should create a new pipe for the stream
	resp := reporter.GetLogs(pod1Request)
	if resp.Pipe == "" {
		t.Errorf("Expected pipe id to be returned, but got %#v", resp)
	}
	pipe, ok := pipes[resp.Pipe]
	if !ok {
		t.Errorf("Expected pipe %q to have been created, but wasn't", resp.Pipe)
	}

	// Should push logs from k8s client into the pipe
	_, readWriter := pipe.Ends()
	contents, err := ioutil.ReadAll(readWriter)
	if err != nil {
		t.Error(err)
	}
	if string(contents) != wantContents {
		t.Errorf("Expected pipe to contain %q, but got %q", wantContents, string(contents))
	}

	// Should close the stream when the pipe closes
	if err := pipe.Close(); err != nil {
		t.Error(err)
	}
	if !closed {
		t.Errorf("Expected pipe to close the underlying log stream")
	}
}
