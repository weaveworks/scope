package endpoint

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	ReverseDNSNames = report.ReverseDNSNames
	SnoopedDNSNames = report.SnoopedDNSNames
	CopyOf          = report.CopyOf
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
	ProcessCache *process.CachingWalker
	Scanner      procspy.ConnectionScanner
	DNSSnooper   *DNSSnooper
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

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Endpoint" }
