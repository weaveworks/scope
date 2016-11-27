package app

import (
	"fmt"

	"github.com/spaolacci/murmur3"

	"github.com/weaveworks/scope/report"
)

// Merger is the type for a thing that can merge reports.
type Merger interface {
	Merge([]report.Report) report.Report
}

type dumbMerger struct{}

// MakeDumbMerger makes a Merger which merges together reports in the simplest possible way.
func MakeDumbMerger() Merger {
	return dumbMerger{}
}

func (dumbMerger) Merge(reports []report.Report) report.Report {
	rpt := report.MakeReport()
	id := murmur3.New64()
	for _, r := range reports {
		rpt = rpt.Merge(r)
		id.Write([]byte(r.ID))
	}
	rpt.ID = fmt.Sprintf("%x", id.Sum64())
	return rpt
}

type smartMerger struct{}

// NewSmartMerger makes a Merger which merges reports in
// parallel. Speed up comes from the fact that a) most merges are
// between small reports, and b) we take advantage of available cores.
func NewSmartMerger() Merger {
	return smartMerger{}
}

func (smartMerger) Merge(reports []report.Report) report.Report {
	l := len(reports)
	switch l {
	case 0:
		return report.MakeReport()
	case 1:
		return reports[0]
	}
	c := make(chan report.Report, l)
	for _, r := range reports {
		c <- r
	}
	for ; l > 1; l-- {
		left, right := <-c, <-c
		go func() {
			c <- left.Merge(right)
		}()
	}
	return <-c
}
