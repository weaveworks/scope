package kubernetes_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/bmizerany/assert"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

var (
	nodeName    = "nodename"
	pod1UID     = "a1b2c3d4e5"
	pod2UID     = "f6g7h8i9j0"
	job1UID     = "abcdef1234"
	serviceUID  = "service1234"
	podTypeMeta = metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	apiPod1 = apiv1.Pod{
		TypeMeta: podTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pong-a",
			UID:               types.UID(pod1UID),
			Namespace:         "ping",
			CreationTimestamp: metav1.Now(),
			Labels:            map[string]string{"ponger": "true"},
		},
		Status: apiv1.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []apiv1.ContainerStatus{
				{ContainerID: "container1"},
				{ContainerID: "container2"},
			},
		},
		Spec: apiv1.PodSpec{
			NodeName:    nodeName,
			HostNetwork: true,
		},
	}
	apiPod2 = apiv1.Pod{
		TypeMeta: podTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pong-b",
			UID:               types.UID(pod2UID),
			Namespace:         "ping",
			CreationTimestamp: metav1.Now(),
			Labels:            map[string]string{"ponger": "true"},
		},
		Status: apiv1.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []apiv1.ContainerStatus{
				{ContainerID: "container3"},
				{ContainerID: "container4"},
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}
	apiJob1 = batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pong",
			UID:               types.UID(job1UID),
			Namespace:         "ping",
			CreationTimestamp: metav1.Now(),
			Labels:            map[string]string{"ponger": "true"},
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"ponger": "true"},
			},
		},
		Status: batchv1.JobStatus{},
	}
	apiService1 = apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pongservice",
			UID:               types.UID(serviceUID),
			Namespace:         "ping",
			CreationTimestamp: metav1.Now(),
		},
		Spec: apiv1.ServiceSpec{
			Type:           apiv1.ServiceTypeLoadBalancer,
			ClusterIP:      "10.0.1.1",
			LoadBalancerIP: "10.0.1.2",
			Ports: []apiv1.ServicePort{
				{Protocol: "TCP", Port: 6379},
			},
			Selector: map[string]string{"ponger": "true"},
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{IP: "10.0.1.2"},
				},
			},
		},
	}
	pod1     = kubernetes.NewPod(&apiPod1)
	pod2     = kubernetes.NewPod(&apiPod2)
	job1     = kubernetes.NewJob(&apiJob1)
	service1 = kubernetes.NewService(&apiService1)
)

func newMockClient() *mockClient {
	return &mockClient{
		pods:     []kubernetes.Pod{pod1, pod2},
		services: []kubernetes.Service{service1},
		jobs:     []kubernetes.Job{job1},
		logs:     map[string]io.ReadCloser{},
	}
}

type mockClient struct {
	pods        []kubernetes.Pod
	services    []kubernetes.Service
	jobs        []kubernetes.Job
	deployments []kubernetes.Deployment
	logs        map[string]io.ReadCloser
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
func (c *mockClient) WalkDaemonSets(f func(kubernetes.DaemonSet) error) error {
	return nil
}
func (c *mockClient) WalkStatefulSets(f func(kubernetes.StatefulSet) error) error {
	return nil
}
func (c *mockClient) WalkCronJobs(f func(kubernetes.CronJob) error) error {
	return nil
}
func (c *mockClient) WalkDeployments(f func(kubernetes.Deployment) error) error {
	for _, deployment := range c.deployments {
		if err := f(deployment); err != nil {
			return err
		}
	}
	return nil
}
func (c *mockClient) WalkNamespaces(f func(kubernetes.NamespaceResource) error) error {
	return nil
}
func (c *mockClient) WalkPersistentVolumes(f func(kubernetes.PersistentVolume) error) error {
	return nil
}
func (c *mockClient) WalkPersistentVolumeClaims(f func(kubernetes.PersistentVolumeClaim) error) error {
	return nil
}
func (c *mockClient) WalkStorageClasses(f func(kubernetes.StorageClass) error) error {
	return nil
}
func (c *mockClient) WalkVolumeSnapshots(f func(kubernetes.VolumeSnapshot) error) error {
	return nil
}
func (c *mockClient) WalkVolumeSnapshotData(f func(kubernetes.VolumeSnapshotData) error) error {
	return nil
}
func (c *mockClient) WalkJobs(f func(kubernetes.Job) error) error {
	for _, job := range c.jobs {
		if err := f(job); err != nil {
			return err
		}
	}
	return nil
}
func (*mockClient) WatchPods(func(kubernetes.Event, kubernetes.Pod)) {}
func (c *mockClient) GetLogs(namespaceID, podName string, _ []string) (io.ReadCloser, error) {
	r, ok := c.logs[namespaceID+";"+podName]
	if !ok {
		return nil, fmt.Errorf("Not found")
	}
	return r, nil
}
func (c *mockClient) DeletePod(namespaceID, podID string) error {
	return nil
}
func (c *mockClient) ScaleUp(namespaceID, id string) error {
	return nil
}
func (c *mockClient) ScaleDown(namespaceID, id string) error {
	return nil
}
func (c *mockClient) CloneVolumeSnapshot(namespaceID, VolumeSnapshotID, persistentVolumeClaimID, capacity string) error {
	return nil
}
func (c *mockClient) CreateVolumeSnapshot(namespaceID, persistentVolumeClaimID, capacity string) error {
	return nil
}
func (c *mockClient) DeleteVolumeSnapshot(namespaceID, VolumeSnapshotID string) error {
	return nil
}
func (c *mockClient) Describe(namespaceID, resourceID string, groupKind schema.GroupKind, restMapping k8smeta.RESTMapping) (io.ReadCloser, error) {
	return nil, nil
}

func (c *mockClient) CordonNode(name string, desired bool) error {
	return nil
}

func (c *mockClient) GetNodes() ([]apiv1.Node, error) {
	return nil, nil
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
	pod1ID := report.MakePodNodeID(pod1UID)
	pod2ID := report.MakePodNodeID(pod2UID)
	job1ID := report.MakeJobNodeID(job1UID)
	serviceID := report.MakeServiceNodeID(serviceUID)
	hr := controls.NewDefaultHandlerRegistry()
	rpt, _ := kubernetes.NewReporter(newMockClient(), nil, "probe-id", "foo", nil, hr, nodeName).Report()

	// Reporter should have added the following pods
	for _, pod := range []struct {
		id            string
		parentService string
		parentJob     string
		latest        map[string]string
	}{
		{pod1ID, serviceID, job1ID, map[string]string{
			kubernetes.Name:      "pong-a",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   pod1.Created(),
		}},
		{pod2ID, serviceID, job1ID, map[string]string{
			kubernetes.Name:      "pong-b",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   pod2.Created(),
		}},
	} {
		node, ok := rpt.Pod.Nodes[pod.id]
		if !ok {
			t.Errorf("Expected report to have pod %q, but not found", pod.id)
		}

		if parents, ok := node.Parents.Lookup(report.Service); !ok || !parents.Contains(pod.parentService) {
			t.Errorf("Expected pod %s to have parent service %q, got %q", pod.id, pod.parentService, parents)
		}

		if parents, ok := node.Parents.Lookup(report.Job); !ok || !parents.Contains(pod.parentJob) {
			t.Errorf("Expected pod %s to have parent service %q, got %q", pod.id, pod.parentJob, parents)
		}

		for k, want := range pod.latest {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected pod %s latest %q: %q, got %q", pod.id, k, want, have)
			}
		}
	}
	// Reporter should have added a job
	{
		node, ok := rpt.Job.Nodes[job1ID]
		assert.Equal(t, ok, true)
		jobLatest := map[string]string{
			kubernetes.Name:      "pong",
			kubernetes.Namespace: "ping",
			kubernetes.Created:   job1.Created(),
		}
		for k, want := range jobLatest {
			hava, ok := node.Latest.Lookup(k)
			assert.Equal(t, ok, true)
			assert.Equal(t, want, hava)
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
			kubernetes.Created:   service1.Created(),
		} {
			if have, ok := node.Latest.Lookup(k); !ok || have != want {
				t.Errorf("Expected service %s latest %q: %q, got %q", serviceID, k, want, have)
			}
		}
	}

	// Reporter should allow controls for k8s topologies by providing a probe ID
	{
		for _, topologyName := range []string{
			report.Container,
			report.CronJob,
			report.DaemonSet,
			report.Deployment,
			report.Pod,
			report.Service,
			report.StatefulSet,
		} {
			topology, ok := rpt.Topology(topologyName)
			if !ok {
				// TODO: this mock report doesn't have nodes for all the topologies yet, so don't fail for now.
				// t.Errorf("Expected report to have nodes in topology %q, but none found", topology)
			}
			for _, n := range topology.Nodes {
				if probeID, ok := n.Latest.Lookup(report.ControlProbeID); !ok || probeID != "probe-id" {
					t.Errorf("Expected node %q to have probeID, but not found", n.ID)
				}
			}
		}
	}

}

func BenchmarkReporter(b *testing.B) {
	hr := controls.NewDefaultHandlerRegistry()
	mockK8s := newMockClient()
	// Add more dummy data
	for i := 0; i < 50; i++ {
		service := apiService1
		service.ObjectMeta.UID = types.UID(fmt.Sprintf("service%d", i))
		mockK8s.services = append(mockK8s.services, kubernetes.NewService(&service))
		pod := apiPod1
		pod.ObjectMeta.UID = types.UID(fmt.Sprintf("pod%d", i))
		mockK8s.pods = append(mockK8s.pods, kubernetes.NewPod(&pod))
		deployment := appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              fmt.Sprintf("deployment%d", i),
				UID:               types.UID(fmt.Sprintf("deployment%d", i)),
				Namespace:         "ping",
				CreationTimestamp: metav1.Now(),
			},
		}
		mockK8s.deployments = append(mockK8s.deployments, kubernetes.NewDeployment(&deployment))
	}
	reporter := kubernetes.NewReporter(mockK8s, nil, "probe-id", "foo", nil, hr, nodeName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reporter.Report()
	}
}

func TestTagger(t *testing.T) {
	rpt := report.MakeReport()
	rpt.ContainerImage.AddNode(report.MakeNodeWith("image1", map[string]string{
		docker.ImageName: "weaveworks/some_interesting_image",
	}))
	rpt.ContainerImage.AddNode(report.MakeNodeWith("pause_image", map[string]string{
		docker.ImageName: "google_containers/pause",
	}))
	rpt.Container.AddNode(report.MakeNodeWith("container1", map[string]string{
		docker.LabelPrefix + "io.kubernetes.pod.uid": "123456",
	}).WithParent(report.ContainerImage, "image1"))
	// This is the first way that Scope identified a pause container - via image name
	rpt.Container.AddNode(report.MakeNodeWith("container2", map[string]string{
		docker.LabelPrefix + "io.kubernetes.pod.uid": "123456",
	}).WithParent(report.ContainerImage, "pause_image"))
	// Second way that Scope identifies a pause container - via docker.type label
	rpt.Container.AddNode(report.MakeNodeWith("container3", map[string]string{
		docker.LabelPrefix + "io.kubernetes.pod.uid":     "123456",
		docker.LabelPrefix + "io.kubernetes.docker.type": "podsandbox",
	}))

	rpt, err := (&kubernetes.Tagger{}).Tag(rpt)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	podParents, ok := rpt.Container.Nodes["container1"].Parents.Lookup(report.Pod)
	want := report.MakeStringSet(report.MakePodNodeID("123456"))
	if !ok || !reflect.DeepEqual(podParents, want) {
		t.Errorf("Expected container1 to have pod parent %v %v", podParents, want)
	}
	_, ok = rpt.Container.Nodes["container1"].Latest.Lookup(report.DoesNotMakeConnections)
	if ok {
		t.Errorf("Expected container1 not to have DoesNotMakeConnections flag")
	}

	_, ok = rpt.Container.Nodes["container2"].Latest.Lookup(report.DoesNotMakeConnections)
	if !ok {
		t.Errorf("Expected pause container to have DoesNotMakeConnections flag")
	}

	_, ok = rpt.Container.Nodes["container3"].Latest.Lookup(report.DoesNotMakeConnections)
	if !ok {
		t.Errorf("Expected pause container to have DoesNotMakeConnections flag")
	}
}

type callbackReadCloser struct {
	io.Reader
	close func() error
}

func (c *callbackReadCloser) Close() error { return c.close() }

func TestReporterGetLogs(t *testing.T) {
	client := newMockClient()
	pipes := mockPipeClient{}
	hr := controls.NewDefaultHandlerRegistry()
	reporter := kubernetes.NewReporter(client, pipes, "", "", nil, hr, nodeName)

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
