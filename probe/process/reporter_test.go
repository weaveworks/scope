package process_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockWalker struct {
	processes []*process.Process
}

func (m *mockWalker) Walk(f func(*process.Process)) error {
	for _, p := range m.processes {
		f(p)
	}
	return nil
}

func TestReporter(t *testing.T) {
	walker := &mockWalker{
		processes: []*process.Process{
			{PID: 1, PPID: 0, Comm: "init"},
			{PID: 2, PPID: 1, Comm: "bash"},
			{PID: 3, PPID: 1, Comm: "apache", Threads: 2},
			{PID: 4, PPID: 2, Comm: "ping", Cmdline: "ping foo.bar.local"},
		},
	}

	reporter := process.NewReporter(walker, "")
	want := report.MakeReport()
	want.Process = report.Topology{
		Adjacency:     report.Adjacency{},
		EdgeMetadatas: report.EdgeMetadatas{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeProcessNodeID("", "1"): report.NodeMetadata{
				process.PID:     "1",
				process.Comm:    "init",
				process.Cmdline: "",
				process.Threads: "0",
			},
			report.MakeProcessNodeID("", "2"): report.NodeMetadata{
				process.PID:     "2",
				process.Comm:    "bash",
				process.PPID:    "1",
				process.Cmdline: "",
				process.Threads: "0",
			},
			report.MakeProcessNodeID("", "3"): report.NodeMetadata{
				process.PID:     "3",
				process.Comm:    "apache",
				process.PPID:    "1",
				process.Cmdline: "",
				process.Threads: "2",
			},
			report.MakeProcessNodeID("", "4"): report.NodeMetadata{
				process.PID:     "4",
				process.Comm:    "ping",
				process.PPID:    "2",
				process.Cmdline: "ping foo.bar.local",
				process.Threads: "0",
			},
		},
	}

	have, err := reporter.Report()
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%s (%v)", test.Diff(want, have), err)
	}
}
