package process_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

type mockWalker struct {
	processes []process.Process
}

func (m *mockWalker) Walk(f func(process.Process, process.Process)) error {
	for _, p := range m.processes {
		f(p, process.Process{})
	}
	return nil
}

var processes = []process.Process{
	{PID: 1, PPID: 0, Name: "init"},
	{PID: 2, PPID: 1, Name: "bash"},
	{PID: 3, PPID: 1, Name: "apache", Threads: 2},
	{PID: 4, PPID: 2, Name: "ping", Cmdline: "ping foo.bar.local"},
	{PID: 5, PPID: 1, Cmdline: "tail -f /var/log/syslog"},
}

func testReporter(t *testing.T, noCommandLineArguments bool, test func(report.Report)) {
	walker := &mockWalker{processes: processes}
	getDeltaTotalJiffies := func() (uint64, float64, error) { return 0, 0., nil }
	now := time.Now()
	mtime.NowForce(now)
	defer mtime.NowReset()

	rpt, err := process.NewReporter(walker, "", getDeltaTotalJiffies, noCommandLineArguments).Report()
	if err != nil {
		t.Error(err)
	}
	test(rpt)
}

func TestInit(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "1")]
		if !ok {
			t.Errorf("Expected report to include the pid 1 init")
		}
		if name, ok := node.Latest.Lookup(process.Name); !ok || name != processes[0].Name {
			t.Errorf("Expected %q got %q", processes[0].Name, name)
		}
	}
	testReporter(t, false, test)
}

func TestProcesses(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "2")]
		if !ok {
			t.Errorf("Expected report to include the pid 2 bash")
		}
		if name, ok := node.Latest.Lookup(process.Name); !ok || name != processes[1].Name {
			t.Errorf("Expected %q got %q", processes[1].Name, name)
		}
		if ppid, ok := node.Latest.Lookup(process.PPID); !ok || ppid != fmt.Sprint(processes[1].PPID) {
			t.Errorf("Expected %d got %q", processes[1].PPID, ppid)
		}
		if memoryUsage, ok := node.Metrics[process.MemoryUsage]; !ok {
			t.Errorf("Expected memory usage metric, but not found")
		} else if sample, ok := memoryUsage.LastSample(); !ok {
			t.Errorf("Expected memory usage metric to have a sample, but there were none")
		} else if sample.Value != 0. {
			t.Errorf("Expected memory usage metric sample %f, got %f", 0., sample.Value)
		}
	}

	testReporter(t, false, test)
}

func TestThreadCounts(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "3")]
		if !ok {
			t.Errorf("Expected report to include the pid 3 apache")
		}
		if threads, ok := node.Latest.Lookup(process.Threads); !ok || threads != fmt.Sprint(processes[2].Threads) {
			t.Errorf("Expected %d got %q", processes[2].Threads, threads)
		}
	}
	testReporter(t, false, test)
}

func TestCmdline(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "4")]
		if !ok {
			t.Errorf("Expected report to include the pid 4 ping")
		}
		if cmdline, ok := node.Latest.Lookup(process.Cmdline); !ok || cmdline != fmt.Sprint(processes[3].Cmdline) {
			t.Errorf("Expected %q got %q", processes[3].Cmdline, cmdline)
		}
	}
	testReporter(t, false, test)
}

func TestAnonymous(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "5")]
		if !ok {
			t.Errorf("Expected report to include the pid 5 tail")
		}
		if name, ok := node.Latest.Lookup(process.Name); ok {
			t.Errorf("Expected no name, but got %q", name)
		}
		if cmdline, ok := node.Latest.Lookup(process.Cmdline); !ok || cmdline != fmt.Sprint(processes[4].Cmdline) {
			t.Errorf("Expected %q got %q", processes[4].Cmdline, cmdline)
		}
	}
	testReporter(t, false, test)
}

func TestCmdlineRemoval(t *testing.T) {
	test := func(rpt report.Report) {
		node, ok := rpt.Process.Nodes[report.MakeProcessNodeID("", "4")]
		if !ok {
			t.Errorf("Expected report to include the pid 4 ping")
		}
		if cmdline, ok := node.Latest.Lookup(process.Cmdline); !ok || cmdline != fmt.Sprint("ping") {
			t.Errorf("Expected %q got %q", "ping", cmdline)
		}
	}
	testReporter(t, true, test)
}
