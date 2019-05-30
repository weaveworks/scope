// +build linux

package endpoint

import (
	"time"

	"github.com/weaveworks/scope/report"
)

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	conf              ReporterConfig
	connectionTracker connectionTracker
	natMapper         natMapper
}

// NewReporter creates a new Reporter that invokes procspy.Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(conf ReporterConfig) *Reporter {
	return &Reporter{
		conf:              conf,
		connectionTracker: newConnectionTracker(conf),
		//natMapper:         makeNATMapper(newConntrackFlowWalker(conf.UseConntrack, conf.ProcRoot, conf.BufferSize, true /* natOnly */)),
	}
}

// Stop stop stop
func (r *Reporter) Stop() {
	r.connectionTracker.Stop()
	r.natMapper.stop()
	if r.conf.Scanner != nil {
		r.conf.Scanner.Stop()
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(time.Since(begin).Seconds())
	}(time.Now())

	rpt := report.MakeReport()

	r.connectionTracker.ReportConnections(&rpt)
	//r.natMapper.applyNAT(rpt, r.conf.HostID)
	return rpt, nil
}
