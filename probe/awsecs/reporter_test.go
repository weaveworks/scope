package awsecs

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

var (
	testCluster = "test-cluster"
	testFamily = "test-family"
	testTaskARN = "arn:aws:ecs:us-east-1:123456789012:task/12345678-9abc-def0-1234-56789abcdef0"
	testTaskCreatedAt = time.Unix(1483228800, 0)
	testTaskDefinitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/deadbeef-dead-beef-dead-beefdeadbeef"
	testTaskStartedAt = time.Unix(1483228805, 0)
	testDeploymentID = "ecs-svc/1121123211234321"
	testServiceName = "test-service"
	testServiceCount = 1
	testContainer = "test-container"
	testContainerData = map[string]string{
		docker.LabelPrefix + "com.amazonaws.ecs.task-arn":
			testTaskARN,
		docker.LabelPrefix + "com.amazonaws.ecs.cluster":
			testCluster,
		docker.LabelPrefix + "com.amazonaws.ecs.task-definition-family":
			testFamily,
	}
)

func getTestContainerNode() report.Node {
	return report.MakeNodeWith(
		report.MakeContainerNodeID(testContainer),
		testContainerData,
	)
}

func TestGetLabelInfo(t *testing.T) {
	r := Make(1e6, time.Hour)
	rpt, err := r.Report()
	if err != nil {
		t.Fatalf("Error making report", err)
	}
	labelInfo := getLabelInfo(rpt)
	expected := map[string]map[string]*taskLabelInfo{}
	if !reflect.DeepEqual(labelInfo, expected) {
		t.Errorf("Empty report did not produce empty label info: %v != %v", labelInfo, expected)
	}

	rpt.Container = rpt.Container.AddNode(getTestContainerNode())
	labelInfo = getLabelInfo(rpt)
	expected = map[string]map[string]*taskLabelInfo{
		testCluster: map[string]*taskLabelInfo{
			testTaskARN: &taskLabelInfo{
				containerIDs: []string{report.MakeContainerNodeID(testContainer)},
				family: testFamily,
			},
		},
	}
	if !reflect.DeepEqual(labelInfo, expected) {
		t.Errorf("Did not get expected label info: %v != %v", labelInfo, expected)
	}
}

// Implements ecsClient
type mockEcsClient struct {
	t *testing.T
	expectedARNs []string
	info ecsInfo
}

func newMockEcsClient(t *testing.T, expectedARNs []string, info ecsInfo) ecsClient {
	return &mockEcsClient{
		t,
		expectedARNs,
		info,
	}
}

func (c mockEcsClient) getInfo(taskARNs []string) ecsInfo {
	if !reflect.DeepEqual(taskARNs, c.expectedARNs) {
		c.t.Fatalf("getInfo called with wrong ARNs: %v != %v", taskARNs, c.expectedARNs)
	}
	return c.info
}

func TestTagReport(t *testing.T) {
	r := Make(1e6, time.Hour)

	r.clientsByCluster[testCluster] = newMockEcsClient(
		t,
		[]string{testTaskARN},
		ecsInfo{
			tasks: map[string]ecsTask{
				testTaskARN: ecsTask{
					taskARN: testTaskARN,
					createdAt: testTaskCreatedAt,
					taskDefinitionARN: testTaskDefinitionARN,
					startedAt: testTaskStartedAt,
					startedBy: testDeploymentID,
				},
			},
			services: map[string]ecsService{
				testServiceName: ecsService{
					serviceName: testServiceName,
					deploymentIDs: []string{testDeploymentID},
					desiredCount: 1,
					pendingCount: 0,
					runningCount: 1,
					taskDefinitionARN: testTaskDefinitionARN,
				},
			},
			taskServiceMap: map[string]string{
				testTaskARN: testServiceName,
			},
		},
	)

	rpt, err := r.Report()
	if err != nil {
		t.Fatalf("Error making report")
	}
	rpt.Container = rpt.Container.AddNode(getTestContainerNode())
	rpt, err = r.Tag(rpt)
	if err != nil {
		t.Fatalf("Failed to tag: %v", err)
	}

	// Check task node is present and contains expected values
	task, ok := rpt.ECSTask.Nodes[report.MakeECSTaskNodeID(testTaskARN)]
	if !ok {
		t.Fatalf("Result report did not contain task %v: %v", testTaskARN, rpt.ECSTask.Nodes)
	}
	taskExpected := map[string]string{
		TaskFamily: testFamily,
		Cluster: testCluster,
		CreatedAt: testTaskCreatedAt.Format(time.RFC3339Nano),
	}
	for key, expectedValue := range taskExpected {
		value, ok := task.Latest.Lookup(key)
		if !ok {
			t.Errorf("Result task did not contain expected key %v: %v", key, task.Latest)
			continue
		}
		if value != expectedValue {
			t.Errorf("Result task did not contain expected value for key %v: %v != %v", key, value, expectedValue)
		}
	}

	// Check service node is present and contains expected values
	service, ok := rpt.ECSService.Nodes[report.MakeECSServiceNodeID(testServiceName)]
	if !ok {
		t.Fatalf("Result report did not contain service %v: %v", testServiceName, rpt.ECSService.Nodes)
	}
	serviceExpected := map[string]string{
		Cluster: testCluster,
		ServiceDesiredCount: "1",
		ServiceRunningCount: "1",
	}
	for key, expectedValue := range serviceExpected {
		value, ok := service.Latest.Lookup(key)
		if !ok {
			t.Errorf("Result service did not contain expected key %v: %v", key, service.Latest)
			continue
		}
		if value != expectedValue {
			t.Errorf("Result service did not contain expected value for key %v: %v != %v", key, value, expectedValue)
		}
	}

	// Check container node is present and contains expected parents
	container, ok := rpt.Container.Nodes[report.MakeContainerNodeID(testContainer)]
	if !ok {
		t.Fatalf("Result report did not contain container %v: %v", testContainer, rpt.Container.Nodes)
	}
	containerParentsExpected := map[string]string{
		report.ECSTask: report.MakeECSTaskNodeID(testTaskARN),
		report.ECSService: report.MakeECSServiceNodeID(testServiceName),
	}
	for key, expectedValue := range containerParentsExpected {
		values, ok := container.Parents.Lookup(key)
		if !ok {
			t.Errorf("Result container did not have any parents for key %v: %v", key, container.Parents)
		}
		if !values.Contains(expectedValue) {
			t.Errorf("Result container did not contain expected value %v as a parent for key %v: %v", expectedValue, key, values)
		}
	}
}
