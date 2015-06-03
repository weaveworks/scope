package main

import (
	"time"

	"github.com/weaveworks/scope/report"
)

// Reporter represents something that can yield reports on-demand.
type Reporter interface {
	Report() report.Report
	Stop()
}

var (
	now = time.Now
)

type lifoReporter struct {
	requests chan chan report.Report
	quit     chan struct{}
}

func newLIFOReporter(incoming <-chan report.Report, maxAge time.Duration) Reporter {
	r := &lifoReporter{
		requests: make(chan chan report.Report),
		quit:     make(chan struct{}),
	}
	go r.loop(incoming, maxAge)
	return r
}

func (r *lifoReporter) Report() report.Report {
	c := make(chan report.Report)
	r.requests <- c
	return <-c
}

func (r *lifoReporter) Stop() {
	close(r.quit)
}

func (r *lifoReporter) loop(incoming <-chan report.Report, maxAge time.Duration) {
	reports := []timedReport{}
	for {
		select {
		case report := <-incoming:
			reports = append(reports, timedReport{now(), report.Squash()})
			reports = trim(reports, now().Add(-maxAge))

		case c := <-r.requests:
			r := report.MakeReport()
			for _, tr := range reports {
				r = r.Merge(tr.Report)
			}
			c <- r

		case <-r.quit:
			return
		}
	}
}

type timedReport struct {
	time.Time
	report.Report
}

func trim(in []timedReport, oldest time.Time) []timedReport {
	out := make([]timedReport, 0, len(in))
	for _, tr := range in {
		if tr.Time.Before(oldest) {
			continue
		}
		out = append(out, tr)
	}
	return out
}
