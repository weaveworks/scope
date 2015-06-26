package process

import (
	"strconv"

	"github.com/weaveworks/scope/probe/tag"
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

// Reporter generate Reports containing the Process topology
type reporter struct {
	scope  string
	walker Walker
}

// NewReporter makes a new Reporter
func NewReporter(walker Walker, scope string) tag.Reporter {
	return &reporter{
		scope:  scope,
		walker: walker,
	}
}

// Report generates a Report containing the Process topology
func (r *reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	processes, err := r.processTopology()
	if err != nil {
		return result, err
	}
	result.Process.Merge(processes)
	return result, nil
}

func (r *reporter) processTopology() (report.Topology, error) {
	t := report.NewTopology()
	err := r.walker.Walk(func(p *Process) {
		pidstr := strconv.Itoa(p.PID)
		nodeID := report.MakeProcessNodeID(r.scope, pidstr)
		t.NodeMetadatas[nodeID] = report.NodeMetadata{
			PID:     pidstr,
			Comm:    p.Comm,
			Cmdline: p.Cmdline,
			Threads: strconv.Itoa(p.Threads),
		}
		if p.PPID > 0 {
			t.NodeMetadatas[nodeID][PPID] = strconv.Itoa(p.PPID)
		}
	})

	return t, err
}
