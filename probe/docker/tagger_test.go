package docker_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockProcessTree struct {
	parents map[int]int
}

func (m *mockProcessTree) GetParent(pid int) (int, error) {
	parent, ok := m.parents[pid]
	if !ok {
		return -1, fmt.Errorf("Not found %d", pid)
	}
	return parent, nil
}

func TestTagger(t *testing.T) {
	oldProcessTree := docker.NewProcessTreeStub
	defer func() { docker.NewProcessTreeStub = oldProcessTree }()

	docker.NewProcessTreeStub = func(_ process.Reader) (process.Tree, error) {
		return &mockProcessTree{map[int]int{2: 1}}, nil
	}

	var (
		pid1NodeID = report.MakeProcessNodeID("somehost.com", "1")
		pid2NodeID = report.MakeProcessNodeID("somehost.com", "2")
		wantNode   = report.MakeNodeWith(map[string]string{docker.ContainerID: "ping"})
	)

	input := report.MakeReport()
	input.Process.AddNode(pid1NodeID, report.MakeNodeWith(map[string]string{"pid": "1"}))
	input.Process.AddNode(pid2NodeID, report.MakeNodeWith(map[string]string{"pid": "2"}))

	want := report.MakeReport()
	want.Process.AddNode(pid1NodeID, report.MakeNodeWith(map[string]string{"pid": "1"}).Merge(wantNode))
	want.Process.AddNode(pid2NodeID, report.MakeNodeWith(map[string]string{"pid": "2"}).Merge(wantNode))

	tagger := docker.NewTagger(mockRegistryInstance, nil)
	have, err := tagger.Tag(input)
	if err != nil {
		t.Errorf("%v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
