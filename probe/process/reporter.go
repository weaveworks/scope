package process

import (
	"strconv"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/report"
)

// We use these keys in node metadata
const (
	PID     = "pid"
	Comm    = "comm"
	PPID    = "ppid"
	Cmdline = "cmdline"
	Threads = "threads"
)

// Reporter generates Reports containing the Process topology.
type Reporter struct {
	scope  string
	walker proc.Reader
}

// NewReporter makes a new Reporter.
func NewReporter(walker proc.Reader, scope string) *Reporter {
	return &Reporter{
		scope:  scope,
		walker: walker,
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	processes, err := r.processTopology()
	if err != nil {
		return result, err
	}
	result.Process = result.Process.Merge(processes)
	return result, nil
}

func (r *Reporter) processTopology() (report.Topology, error) {
	t := report.MakeTopology()
	err := r.walker.Processes(func(p proc.Process) {
		pidstr := strconv.Itoa(p.PID)
		nodeID := report.MakeProcessNodeID(r.scope, pidstr)
		t.Nodes[nodeID] = report.MakeNode()
		for _, tuple := range []struct{ key, value string }{
			{PID, pidstr},
			{Comm, p.Comm},
			{Cmdline, p.Cmdline},
			{Threads, strconv.Itoa(p.Threads)},
		} {
			if tuple.value != "" {
				t.Nodes[nodeID].Metadata[tuple.key] = tuple.value
			}
		}
		if p.PPID > 0 {
			t.Nodes[nodeID].Metadata[PPID] = strconv.Itoa(p.PPID)
		}
	})

	return t, err
}
