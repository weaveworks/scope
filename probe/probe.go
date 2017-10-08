package probe

import (
	"sync"
	"time"

	"github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/report"
)

const (
	reportBufferSize = 16
)

// Probe sits there, generating and publishing reports.
type Probe struct {
	spyInterval, publishInterval time.Duration
	publisher                    *appclient.ReportPublisher

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
	publisher appclient.Publisher,
	noControls bool,
) *Probe {
	result := &Probe{
		spyInterval:     spyInterval,
		publishInterval: publishInterval,
		publisher:       appclient.NewReportPublisher(publisher, noControls),
		quit:            make(chan struct{}),
		spiedReports:    make(chan report.Report, reportBufferSize),
		shortcutReports: make(chan report.Report, reportBufferSize),
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
func (p *Probe) Stop() {
	close(p.quit)
	p.done.Wait()
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
			t := time.Now()
			p.tick()
			rpt := p.report()
			rpt = p.tag(rpt)
			p.spiedReports <- rpt
			metrics.MeasureSince([]string{"Report Generaton"}, t)
		case <-p.quit:
			return
		}
	}
}

func (p *Probe) tick() {
	for _, ticker := range p.tickers {
		t := time.Now()
		err := ticker.Tick()
		metrics.MeasureSince([]string{ticker.Name(), "ticker"}, t)
		if err != nil {
			log.Errorf("error doing ticker: %v", err)
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
			metrics.MeasureSince([]string{rep.Name(), "reporter"}, t)
			if err != nil {
				log.Errorf("error generating report: %v", err)
				newReport = report.MakeReport() // empty is OK to merge
			}
			reports <- newReport
		}(rep)
	}

	result := report.MakeReport()
	for i := 0; i < cap(reports); i++ {
		result = result.Merge(<-reports)
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
		metrics.MeasureSince([]string{tagger.Name(), "tagger"}, t)
		if err != nil {
			log.Errorf("error applying tagger: %v", err)
		}
	}
	return r
}

func (p *Probe) drainAndPublish(rpt report.Report, rs chan report.Report) {
ForLoop:
	for {
		select {
		case r := <-rs:
			rpt = rpt.Merge(r)
		default:
			break ForLoop
		}
	}

	if err := p.publisher.Publish(rpt.BackwardCompatible()); err != nil {
		log.Infof("publish: %v", err)
	}
}

func (p *Probe) publishLoop() {
	defer p.done.Done()
	pubTick := time.Tick(p.publishInterval)

	for {
		select {
		case <-pubTick:
			p.drainAndPublish(report.MakeReport(), p.spiedReports)

		case rpt := <-p.shortcutReports:
			p.drainAndPublish(rpt, p.shortcutReports)

		case <-p.quit:
			return
		}
	}
}
