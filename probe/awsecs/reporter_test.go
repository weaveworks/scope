package awsecs

import (
	"reflect"
	"testing"
)

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

	rpt.Containers = rpt.Containers.AddNode(
		report.MakeNodeWith(
			report.MakeContainerNodeID("test-container"),
			map[string]string{
				docker.LabelPrefix + "com.amazonaws.ecs.task-arn":
					"arn:aws:ecs:us-east-1:123456789012:task/12345678-9abc-def0-1234-56789abcdef0",
				docker.LabelPrefix + "com.amazonaws.ecs.cluster":
					"test-cluster",
				docker.LabelPrefix + "com.amazonaws.ecs.task-definition-family":
					"test-family",
			}
		)
	)
	labelInfo = r.getLabelInfo(rpt)
	expected = map[string]map[string]*taskLabelInfo{
		"test-cluster": map[string]*taskLabelInfo{
			"arn:aws:ecs:us-east-1:123456789012:task/12345678-9abc-def0-1234-56789abcdef0": &taskLabelInfo{
				containerIDs: []string{"test-container"},
				family: "test-family",
			}
		}
	}
	if !reflect.DeepEqual(labelInfo, expected) {
		t.Error("Did not get expected label info: %v != %v", labelInfo, expected)
	}
}
