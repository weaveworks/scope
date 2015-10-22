package app

import (
	"sync"
	"time"

	"github.com/weaveworks/scope/report"
)

// Reporter is something that can produce reports on demand. It's a convenient
// interface for parts of the app, and several experimental components.
type Reporter interface {
	Report() report.Report
	WaitOn(chan struct{})
	UnWait(chan struct{})
}

// Adder is something that can accept reports. It's a convenient interface for
// parts of the app, and several experimental components.
type Adder interface {
	Add(report.Report)
}

// A Collector is a Reporter and an Adder
type Collector interface {
	Reporter
	Adder
}

// Collector receives published reports from multiple producers. It yields a
// single merged report, representing all collected reports.
type collector struct {
	mtx     sync.Mutex
	reports []timestampReport
	window  time.Duration
	waitableCondition
}

type waitableCondition struct {
	sync.Mutex
	waiters map[chan struct{}]struct{}
}

func (wc *waitableCondition) WaitOn(waiter chan struct{}) {
	wc.Lock()
	wc.waiters[waiter] = struct{}{}
	wc.Unlock()
}

func (wc *waitableCondition) UnWait(waiter chan struct{}) {
	wc.Lock()
	delete(wc.waiters, waiter)
	wc.Unlock()
}

func (wc *waitableCondition) Broadcast() {
	wc.Lock()
	for waiter := range wc.waiters {
		// Non-block write to channel
		select {
		case waiter <- struct{}{}:
		default:
		}
	}
	wc.Unlock()
}

// NewCollector returns a collector ready for use.
func NewCollector(window time.Duration) Collector {
	return &collector{
		window: window,
		waitableCondition: waitableCondition{
			waiters: map[chan struct{}]struct{}{},
		},
	}
}

var now = time.Now

// Add adds a report to the collector's internal state. It implements Adder.
func (c *collector) Add(rpt report.Report) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reports = append(c.reports, timestampReport{now(), rpt})
	c.reports = clean(c.reports, c.window)
	if rpt.Shortcut {
		c.Broadcast()
	}
}

// Report returns a merged report over all added reports. It implements
// Reporter.
func (c *collector) Report() report.Report {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.reports = clean(c.reports, c.window)

	rpt := report.MakeReport()
	for _, tr := range c.reports {
		rpt = rpt.Merge(tr.report)
	}
	return rpt
}

type timestampReport struct {
	timestamp time.Time
	report    report.Report
}

func clean(reports []timestampReport, window time.Duration) []timestampReport {
	var (
		cleaned = make([]timestampReport, 0, len(reports))
		oldest  = now().Add(-window)
	)
	for _, tr := range reports {
		if tr.timestamp.Before(oldest) {
			continue
		}
		cleaned = append(cleaned, tr)
	}
	return cleaned
}
