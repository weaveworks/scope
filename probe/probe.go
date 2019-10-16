package probe

import (
	"context"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/mtime"
	"golang.org/x/time/rate"

	"github.com/weaveworks/scope/report"
)

const (
	spiedReportBufferSize    = 16
	shortcutReportBufferSize = 1024
)

// ReportPublisher publishes reports, probably to a remote collector.
type ReportPublisher interface {
	Publish(r report.Report) error
}

// Probe sits there, generating and publishing reports.
type Probe struct {
	spyInterval, publishInterval time.Duration
	publisher                    ReportPublisher
	rateLimiter                  *rate.Limiter
	ticksPerFullReport           int
	noControls                   bool

	tickers   []Ticker
	reporters []Reporter
	taggers   []Tagger

	quit chan struct{}
	done sync.WaitGroup

	spiedReports    chan report.Report
	shortcutReports chan report.Report
}

// Tagger tags nodes with value-add node metadata.
type Tagger interface {
	Name() string
	Tag(r report.Report) (report.Report, error)
}

// Reporter generates Reports.
type Reporter interface {
	Name() string
	Report() (report.Report, error)
}

// ReporterFunc uses a function to implement a Reporter
func ReporterFunc(name string, f func() (report.Report, error)) Reporter {
	return reporterFunc{name, f}
}

type reporterFunc struct {
	name string
	f    func() (report.Report, error)
}

func (r reporterFunc) Name() string                   { return r.name }
func (r reporterFunc) Report() (report.Report, error) { return r.f() }

// Ticker is something which will be invoked every spyDuration.
// It's useful for things that should be updated on that interval.
// For example, cached shared state between Taggers and Reporters.
type Ticker interface {
	Name() string
	Tick() error
}

// New makes a new Probe.
func New(
	spyInterval, publishInterval time.Duration,
	publisher ReportPublisher,
	ticksPerFullReport int,
	noControls bool,
) *Probe {
	result := &Probe{
		spyInterval:        spyInterval,
		publishInterval:    publishInterval,
		publisher:          publisher,
		rateLimiter:        rate.NewLimiter(rate.Every(publishInterval/100), 1),
		ticksPerFullReport: ticksPerFullReport,
		noControls:         noControls,
		quit:               make(chan struct{}),
		spiedReports:       make(chan report.Report, spiedReportBufferSize),
		shortcutReports:    make(chan report.Report, shortcutReportBufferSize),
	}
	return result
}

// AddTagger adds a new Tagger to the Probe
func (p *Probe) AddTagger(ts ...Tagger) {
	p.taggers = append(p.taggers, ts...)
}

// AddReporter adds a new Reported to the Probe
func (p *Probe) AddReporter(rs ...Reporter) {
	p.reporters = append(p.reporters, rs...)
}

// AddTicker adds a new Ticker to the Probe
func (p *Probe) AddTicker(ts ...Ticker) {
	p.tickers = append(p.tickers, ts...)
}

// Start starts the probe
func (p *Probe) Start() {
	p.done.Add(2)
	go p.spyLoop()
	go p.publishLoop()
}

// Stop stops the probe
func (p *Probe) Stop() error {
	close(p.quit)
	p.done.Wait()
	return nil
}

// Publish will queue a report for immediate publication,
// bypassing the spy tick
func (p *Probe) Publish(rpt report.Report) {
	rpt = p.tag(rpt)
	p.shortcutReports <- rpt
}

func (p *Probe) spyLoop() {
	defer p.done.Done()
	spyTick := time.Tick(p.spyInterval)

	for {
		select {
		case <-spyTick:
			p.tick()
			rpt := p.report()
			rpt = p.tag(rpt)
			p.spiedReports <- rpt
		case <-p.quit:
			return
		}
	}
}

func (p *Probe) tick() {
	for _, ticker := range p.tickers {
		t := time.Now()
		err := ticker.Tick()
		metrics.MeasureSinceWithLabels([]string{"duration", "seconds"}, t, []metrics.Label{
			{Name: "operation", Value: "ticker"},
			{Name: "module", Value: ticker.Name()},
		})
		if err != nil {
			log.Errorf("Error doing ticker: %v", err)
		}
	}
}

func (p *Probe) report() report.Report {
	reports := make(chan report.Report, len(p.reporters))
	for _, rep := range p.reporters {
		go func(rep Reporter) {
			t := time.Now()
			timer := time.AfterFunc(p.spyInterval, func() { log.Warningf("%v reporter took longer than %v", rep.Name(), p.spyInterval) })
			newReport, err := rep.Report()
			if !timer.Stop() {
				log.Warningf("%v reporter took %v (longer than %v)", rep.Name(), time.Now().Sub(t), p.spyInterval)
			}
			metrics.MeasureSinceWithLabels([]string{"duration", "seconds"}, t, []metrics.Label{
				{Name: "operation", Value: "reporter"},
				{Name: "module", Value: rep.Name()},
			})
			if err != nil {
				log.Errorf("Error generating %s report: %v", rep.Name(), err)
				newReport = report.MakeReport() // empty is OK to merge
			}
			reports <- newReport
		}(rep)
	}

	result := report.MakeReport()
	result.TS = mtime.Now()
	for i := 0; i < cap(reports); i++ {
		result.UnsafeMerge(<-reports)
	}
	return result
}

func (p *Probe) tag(r report.Report) report.Report {
	var err error
	for _, tagger := range p.taggers {
		t := time.Now()
		timer := time.AfterFunc(p.spyInterval, func() { log.Warningf("%v tagger took longer than %v", tagger.Name(), p.spyInterval) })
		r, err = tagger.Tag(r)
		if !timer.Stop() {
			log.Warningf("%v tagger took %v (longer than %v)", tagger.Name(), time.Now().Sub(t), p.spyInterval)
		}
		metrics.MeasureSinceWithLabels([]string{"duration", "seconds"}, t, []metrics.Label{
			{Name: "operation", Value: "tagger"},
			{Name: "module", Value: tagger.Name()},
		})
		if err != nil {
			log.Errorf("Error applying tagger: %v", err)
		}
	}
	return r
}

func (p *Probe) drainAndSanitise(rpt report.Report, rs chan report.Report) report.Report {
	p.rateLimiter.Wait(context.Background())
	rpt = rpt.Copy()
ForLoop:
	for {
		select {
		case r := <-rs:
			rpt.UnsafeMerge(r)
		default:
			break ForLoop
		}
	}

	if p.noControls {
		rpt.WalkTopologies(func(t *report.Topology) {
			t.Controls = report.Controls{}
		})
	}
	return rpt
}

func (p *Probe) publishLoop() {
	defer p.done.Done()
	pubTick := time.Tick(p.publishInterval)
	publishCount := 0
	var lastFullReport report.Report

	for {
		var err error
		select {
		case <-pubTick:
			rpt := p.drainAndSanitise(report.MakeReport(), p.spiedReports)
			fullReport := (publishCount % p.ticksPerFullReport) == 0
			if !fullReport {
				rpt.UnsafeUnMerge(lastFullReport)
			}
			err = p.publisher.Publish(rpt)
			if err == nil {
				if fullReport {
					lastFullReport = rpt
				}
				publishCount++
			} else {
				// If we failed to send then drop back to full report next time
				publishCount = 0
			}

		case rpt := <-p.shortcutReports:
			rpt = p.drainAndSanitise(rpt, p.shortcutReports)
			err = p.publisher.Publish(rpt)

		case <-p.quit:
			return
		}
		if err != nil {
			log.Infof("Publish: %v", err)
		}
	}
}
