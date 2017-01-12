package awsecs

import (
	"reflect"
	"testing"
)

const (
	testCluster = "test-cluster"
	testFamily = "test-family"
	testTaskARN = "arn:aws:ecs:us-east-1:123456789012:task/12345678-9abc-def0-1234-56789abcdef0"
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
		report.MakeContainerNodeID("test-container"),
		testContainerData
	)
}

func TestGetLabelInfo(t *testing.T) {
	r := Make()
	rpt, err := r.Report()
	if err != nil {
		t.Fatal("Error making report", err)
	}
	labelInfo := r.getLabelInfo(rpt)
	expected := map[string]map[string]*taskLabelInfo{}
	if !reflect.DeepEqual(labelInfo, expected) {
		t.Error("Empty report did not produce empty label info: %v != %v", labelInfo, expected)
	}

	rpt.Containers = rpt.Containers.AddNode(getTestContainerNode())
	labelInfo = r.getLabelInfo(rpt)
	expected = map[string]map[string]*taskLabelInfo{
		testCluster: map[string]*taskLabelInfo{
			testTaskARN: &taskLabelInfo{
				containerIDs: []string{testContainer},
				family: testFamily,
			}
		}
	}
	if !reflect.DeepEqual(labelInfo, expected) {
		t.Error("Did not get expected label info: %v != %v", labelInfo, expected)
	}
}

// Implements ecsClient
type mockEcsClient {
	t *testing.T
	expectedARNs []string
	info ecsInfo
}

func newMockEcsClient(t *testing.T, expectedARNs []string, info ecsInfo) *ecsClient {
	return &mockEcsClient{
		t,
		expectedARNs,
		info,
	}
}

func (c mockEcsClient) getInfo(taskARNs []string) ecsInfo {
	if !reflect.DeepEqual(taskARNs, c.expectedARNs) {
		c.t.Fatal("getInfo called with wrong ARNs: %v != %v", taskARNs, c.expectedARNs)
	}
	return c.info
}

func TestTagReport(t *testing.T) {
	r := Make()

	r.clientsByCluster[testCluster] = newMockEcsClient(
		t,
		[]string{},
		ecsInfo{
			// TODO fill in values below
			tasks: map[string]ecsTask{},
			services: map[string]ecsService{},
			taskServiceMap: map[string]string{},
		},
	)

	rpt, err := r.Report()
	if err != nil {
		t.Fatal("Error making report")
	}
	rpt = r.Tag(rpt)
	// TODO check it matches
}
