package kubernetes_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/types"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

var (
	nodeName    = "nodename"
	pod1UID     = "a1b2c3d4e5"
	pod2UID     = "f6g7h8i9j0"
	serviceUID  = "service1234"
	podTypeMeta = unversioned.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	apiPod1 = api.Pod{
		TypeMeta: podTypeMeta,
		ObjectMeta: api.ObjectMeta{
			Name:              "pong-a",
			UID:               types.UID(pod1UID),
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
			UID:               types.UID(pod2UID),
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
			UID:               types.UID(serviceUID),
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
func (c *mockClient) WalkDeployments(f func(kubernetes.Deployment) error) error {
	return nil
}
func (c *mockClient) WalkReplicaSets(f func(kubernetes.ReplicaSet) error) error {
	return nil
}
func (c *mockClient) WalkReplicationControllers(f func(kubernetes.ReplicationController) error) error {
	return nil
}
func (*mockClient) WalkNodes(f func(*api.Node) error) error {
	return nil
}
func (*mockClient) WatchPods(func(kubernetes.Event, kubernetes.Pod)) {}
func (c *mockClient) GetLogs(namespaceID, podName string) (io.ReadCloser, error) {
	r, ok := c.logs[namespaceID+";"+podName]
	if !ok {
		return nil, fmt.Errorf("Not found")
	}
	return r, nil
}
func (c *mockClient) DeletePod(namespaceID, podID string) error {
	return nil
}
func (c *mockClient) ScaleUp(resource, namespaceID, id string) error {
	return nil
}
func (c *mockClient) ScaleDown(resource, namespaceID, id string) error {
	return nil
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

	pod1ID := report.MakePodNodeID(pod1UID)
	pod2ID := report.MakePodNodeID(pod2UID)
	serviceID := report.MakeServiceNodeID(serviceUID)
	rpt, _ := kubernetes.NewReporter(newMockClient(), nil, "", "foo", nil).Report()

	// Reporter should have added the following pods
	for _, pod := range []struct {
		id            string
		parentService string
		latest        map[string]string
	}{
		{pod1ID, serviceID, map[string]string{
			kubernetes.Name:      "pong-a",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   pod1.Created(),
		}},
		{pod2ID, serviceID, map[string]string{
			kubernetes.Name:      "pong-b",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   pod1.Created(),
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
			kubernetes.Name:      "pongservice",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   pod1.Created(),
		} {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected service %s latest %q: %q, got %q", serviceID, k, want, have)
			}
		}
	}
}

func TestTagger(t *testing.T) {
	rpt := report.MakeReport()
	rpt.Container.AddNode(report.MakeNodeWith("container1", map[string]string{
		docker.LabelPrefix + "io.kubernetes.pod.uid": "123456",
	}))

	rpt, err := kubernetes.NewReporter(newMockClient(), nil, "", "", nil).Tag(rpt)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	have, ok := rpt.Container.Nodes["container1"].Parents.Lookup(report.Pod)
	want := report.EmptyStringSet.Add(report.MakePodNodeID("123456"))
	if !ok || !reflect.DeepEqual(have, want) {
		t.Errorf("Expected container to have pod parent %v %v", have, want)
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
	reporter := kubernetes.NewReporter(client, pipes, "", "", nil)

	// Should error on invalid IDs
	{
		resp := reporter.CapturePod(reporter.GetLogs)(xfer.Request{
			NodeID:  "invalidID",
			Control: kubernetes.GetLogs,
		})
		if want := "Invalid ID: invalidID"; resp.Error != want {
			t.Errorf("Expected error on invalid ID: %q, got %q", want, resp.Error)
		}
	}

	// Should pass through errors from k8s (e.g if pod does not exist)
	{
		resp := reporter.CapturePod(reporter.GetLogs)(xfer.Request{
			AppID:   "appID",
			NodeID:  report.MakePodNodeID("notfound"),
			Control: kubernetes.GetLogs,
		})
		if want := "Pod not found: notfound"; resp.Error != want {
			t.Errorf("Expected error on invalid ID: %q, got %q", want, resp.Error)
		}
	}

	podNamespaceAndID := "ping;pong-a"
	pod1Request := xfer.Request{
		AppID:   "appID",
		NodeID:  report.MakePodNodeID(pod1UID),
		Control: kubernetes.GetLogs,
	}

	// Inject our logs content, and watch for it to be closed
	closed := false
	wantContents := "logs: ping/pong-a"
	client.logs[podNamespaceAndID] = &callbackReadCloser{Reader: strings.NewReader(wantContents), close: func() error {
		closed = true
		return nil
	}}

	// Should create a new pipe for the stream
	resp := reporter.CapturePod(reporter.GetLogs)(pod1Request)
	if resp.Pipe == "" {
		t.Errorf("Expected pipe id to be returned, but got %#v", resp)
	}
	pipe, ok := pipes[resp.Pipe]
	if !ok {
		t.Fatalf("Expected pipe %q to have been created, but wasn't", resp.Pipe)
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
