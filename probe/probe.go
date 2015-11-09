package probe

import (
	"log"
	"sync"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

const (
	reportBufferSize = 16
)

// Probe sits there, generating and publishing reports.
type Probe struct {
	spyInterval, publishInterval time.Duration
	publisher                    xfer.Publisher

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
	Tag(r report.Report) (report.Report, error)
}

// Reporter generates Reports.
type Reporter interface {
	Report() (report.Report, error)
}

// Ticker is something which will be invoked every spyDuration.
// It's useful for things that should be updated on that interval.
// For example, cached shared state between Taggers and Reporters.
type Ticker interface {
	Tick() error
}

// New makes a new Probe.
func New(spyInterval, publishInterval time.Duration, publisher xfer.Publisher) *Probe {
	result := &Probe{
		spyInterval:     spyInterval,
		publishInterval: publishInterval,
		publisher:       publisher,
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
	p.shortcutReports <- rpt
}

func (p *Probe) spyLoop() {
	defer p.done.Done()
	spyTick := time.Tick(p.spyInterval)

	for {
		select {
		case <-spyTick:
			start := time.Now()
			for _, ticker := range p.tickers {
				if err := ticker.Tick(); err != nil {
					log.Printf("error doing ticker: %v", err)
				}
			}

			rpt := p.report()
			rpt = p.tag(rpt)
			p.spiedReports <- rpt

			if took := time.Since(start); took > p.spyInterval {
				log.Printf("report generation took too long (%s)", took)
			}

		case <-p.quit:
			return
		}
	}
}

func (p *Probe) report() report.Report {
	reports := make(chan report.Report, len(p.reporters))
	for _, rep := range p.reporters {
		go func(rep Reporter) {
			newReport, err := rep.Report()
			if err != nil {
				log.Printf("error generating report: %v", err)
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
		r, err = tagger.Tag(r)
		if err != nil {
			log.Printf("error applying tagger: %v", err)
		}
	}
	return r
}

func condense(rpt report.Report, rs chan report.Report) report.Report {
	for {
		select {
		case r := <-rs:
			rpt = rpt.Merge(r)
		default:
			return rpt
		}
	}
}

func (p *Probe) publishLoop() {
	defer p.done.Done()
	var (
		pubTick = time.Tick(p.publishInterval)
		rptPub  = xfer.NewReportPublisher(p.publisher)
	)

	for {
		select {
		case <-pubTick:
			rpt := condense(report.MakeReport(), p.spiedReports)
			if err := rptPub.Publish(rpt); err != nil {
				log.Printf("publish: %v", err)
			}

		case rpt := <-p.shortcutReports:
			rpt = condense(rpt, p.shortcutReports)
			if err := rptPub.Publish(rpt); err != nil {
				log.Printf("publish: %v", err)
			}

		case <-p.quit:
			return
		}
	}
}
