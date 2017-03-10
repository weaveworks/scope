package endpoint

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	Addr            = "addr" // typically IPv4
	Port            = "port"
	Conntracked     = "conntracked"
	EBPF            = "eBPF"
	Procspied       = "procspied"
	ReverseDNSNames = "reverse_dns_names"
	SnoopedDNSNames = "snooped_dns_names"
)

// ReporterConfig are the config options for the endpoint reporter.
type ReporterConfig struct {
	HostID       string
	HostName     string
	SpyProcs     bool
	UseConntrack bool
	WalkProc     bool
	UseEbpfConn  bool
	ProcRoot     string
	BufferSize   int
	Scanner      procspy.ConnectionScanner
	DNSSnooper   *DNSSnooper
}

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	conf              ReporterConfig
	connectionTracker connectionTracker
	natMapper         natMapper
}

// SpyDuration is an exported prometheus metric
var SpyDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "scope",
		Subsystem: "probe",
		Name:      "spy_duration_seconds",
		Help:      "Time in seconds spent spying on active connections.",
		MaxAge:    10 * time.Second, // like statsd
	},
	[]string{},
)

// NewReporter creates a new Reporter that invokes procspy.Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(conf ReporterConfig) *Reporter {
	return &Reporter{
		conf: conf,
		connectionTracker: newConnectionTracker(connectionTrackerConfig{
			HostID:       conf.HostID,
			HostName:     conf.HostName,
			SpyProcs:     conf.SpyProcs,
			UseConntrack: conf.UseConntrack,
			WalkProc:     conf.WalkProc,
			UseEbpfConn:  conf.UseEbpfConn,
			ProcRoot:     conf.ProcRoot,
			BufferSize:   conf.BufferSize,
			Scanner:      conf.Scanner,
			DNSSnooper:   conf.DNSSnooper,
		}),
		natMapper: makeNATMapper(newConntrackFlowWalker(conf.UseConntrack, conf.ProcRoot, conf.BufferSize, "--any-nat")),
	}
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Endpoint" }

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
	r.natMapper.applyNAT(rpt, r.conf.HostID)
	return rpt, nil
}
