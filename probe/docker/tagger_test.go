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

	docker.NewProcessTreeStub = func(_ process.Walker) (process.Tree, error) {
		return &mockProcessTree{map[int]int{2: 1}}, nil
	}

	var (
		pid1NodeID       = report.MakeProcessNodeID("somehost.com", "1")
		pid2NodeID       = report.MakeProcessNodeID("somehost.com", "2")
		wantNodeMetadata = report.NodeMetadata{docker.ContainerID: "ping"}
	)

	input := report.MakeReport()
	input.Process.NodeMetadatas[pid1NodeID] = report.NodeMetadata{"pid": "1"}
	input.Process.NodeMetadatas[pid2NodeID] = report.NodeMetadata{"pid": "2"}

	want := report.MakeReport()
	want.Process.NodeMetadatas[pid1NodeID] = report.NodeMetadata{"pid": "1"}.Merge(wantNodeMetadata)
	want.Process.NodeMetadatas[pid2NodeID] = report.NodeMetadata{"pid": "2"}.Merge(wantNodeMetadata)

	tagger := docker.NewTagger(mockRegistryInstance, nil)
	have, err := tagger.Tag(input)
	if err != nil {
		t.Errorf("%v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
