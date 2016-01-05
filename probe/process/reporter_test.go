package process_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
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

func TestReporter(t *testing.T) {
	walker := &mockWalker{
		processes: []process.Process{
			{PID: 1, PPID: 0, Name: "init"},
			{PID: 2, PPID: 1, Name: "bash"},
			{PID: 3, PPID: 1, Name: "apache", Threads: 2},
			{PID: 4, PPID: 2, Name: "ping", Cmdline: "ping foo.bar.local"},
			{PID: 5, PPID: 1, Cmdline: "tail -f /var/log/syslog"},
		},
	}
	getDeltaTotalJiffies := func() (uint64, float64, error) { return 0, 0., nil }
	now := time.Now()
	mtime.NowForce(now)
	defer mtime.NowReset()

	reporter := process.NewReporter(walker, "", getDeltaTotalJiffies)
	want := report.MakeReport()
	want.Process = report.MakeTopology().AddNode(
		report.MakeProcessNodeID("", "1"), report.MakeNodeWith(map[string]string{
			process.PID:     "1",
			process.Name:    "init",
			process.Threads: "0",
		}).WithMetric(process.MemoryUsage, report.MakeMetric().Add(now, 0.)),
	).AddNode(
		report.MakeProcessNodeID("", "2"), report.MakeNodeWith(map[string]string{
			process.PID:     "2",
			process.Name:    "bash",
			process.PPID:    "1",
			process.Threads: "0",
		}).WithMetric(process.MemoryUsage, report.MakeMetric().Add(now, 0.)),
	).AddNode(
		report.MakeProcessNodeID("", "3"), report.MakeNodeWith(map[string]string{
			process.PID:     "3",
			process.Name:    "apache",
			process.PPID:    "1",
			process.Threads: "2",
		}).WithMetric(process.MemoryUsage, report.MakeMetric().Add(now, 0.)),
	).AddNode(
		report.MakeProcessNodeID("", "4"), report.MakeNodeWith(map[string]string{
			process.PID:     "4",
			process.Name:    "ping",
			process.PPID:    "2",
			process.Cmdline: "ping foo.bar.local",
			process.Threads: "0",
		}).WithMetric(process.MemoryUsage, report.MakeMetric().Add(now, 0.)),
	).AddNode(
		report.MakeProcessNodeID("", "5"), report.MakeNodeWith(map[string]string{
			process.PID:     "5",
			process.PPID:    "1",
			process.Cmdline: "tail -f /var/log/syslog",
			process.Threads: "0",
		}).WithMetric(process.MemoryUsage, report.MakeMetric().Add(now, 0.)),
	)

	have, err := reporter.Report()
	if err != nil || !reflect.DeepEqual(want, have) {
		t.Errorf("%s (%v)", test.Diff(want, have), err)
	}
}
