package main

import (
	"time"

	"github.com/weaveworks/scope/scope/report"
)

// Copy/paste from app/report_lifo.go

// Reporter XXX
type Reporter interface {
	Report() report.Report
}

type timedReport struct {
	report.Report
	Timestamp time.Time
}

// ReportLIFO XXX
type ReportLIFO struct {
	reports  []timedReport
	requests chan chan report.Report
	quit     chan chan struct{}
}

type reporter interface {
	Reports() <-chan report.Report
}

// NewReportLIFO XXX
func NewReportLIFO(r reporter, maxAge time.Duration) *ReportLIFO {
	l := ReportLIFO{
		reports:  []timedReport{},
		requests: make(chan chan report.Report),
		quit:     make(chan chan struct{}),
	}

	go func() {
		for {
			select {
			case report := <-r.Reports():
				report = report.SquashRemote()
				tr := timedReport{
					Timestamp: time.Now(),
					Report:    report,
				}
				l.reports = append(l.reports, tr)
				l.reports = cleanOld(l.reports, time.Now().Add(-maxAge))

			case req := <-l.requests:
				report := report.NewReport()
				for _, r := range l.reports {
					report.Merge(r.Report)
				}
				req <- report

			case q := <-l.quit:
				close(q)
				return
			}
		}
	}()
	return &l
}

// Stop XXX
func (r *ReportLIFO) Stop() {
	q := make(chan struct{})
	r.quit <- q
	<-q
}

// Report XXX
func (r *ReportLIFO) Report() report.Report {
	req := make(chan report.Report)
	r.requests <- req
	return <-req
}

func cleanOld(reports []timedReport, threshold time.Time) []timedReport {
	res := make([]timedReport, 0, len(reports))
	for _, tr := range reports {
		if tr.Timestamp.Before(threshold) {
			continue
		}
		res = append(res, tr)
	}
	return res
}
