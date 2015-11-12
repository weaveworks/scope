package process

import (
	"strconv"

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
	walker Walker
}

// NewReporter makes a new Reporter.
func NewReporter(walker Walker, scope string) *Reporter {
	return &Reporter{
		scope:  scope,
		walker: walker,
	}
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Process" }

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
	err := r.walker.Walk(func(p Process) {
		pidstr := strconv.Itoa(p.PID)
		nodeID := report.MakeProcessNodeID(r.scope, pidstr)
		node := report.MakeNode()
		for _, tuple := range []struct{ key, value string }{
			{PID, pidstr},
			{Comm, p.Comm},
			{Cmdline, p.Cmdline},
			{Threads, strconv.Itoa(p.Threads)},
		} {
			if tuple.value != "" {
				node.Metadata[tuple.key] = tuple.value
			}
		}
		if p.PPID > 0 {
			node.Metadata[PPID] = strconv.Itoa(p.PPID)
		}
		t.AddNode(nodeID, node)
	})

	return t, err
}
