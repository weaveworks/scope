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

type smartMerger struct {
}

// NewSmartMerger makes a Merger which merges together reports as
// a binary tree of reports.  Speed up comes from the fact that
// most merges are between small reports.
func NewSmartMerger() Merger {
	return smartMerger{}
}

// Merge merges the reports as a binary tree. Crucially, it
// effectively merges reports in reverse order. Typically reports are
// ordered oldest-to-youngest, so this strategy merges older reports
// "into" younger reports. This order is more efficient for some Merge
// operations, in particular LatestMap.Merge.
func (s smartMerger) Merge(reports []report.Report) report.Report {
	l := len(reports)
	switch l {
	case 0:
		return report.MakeReport()
	case 1:
		return reports[0]
	case 2:
		return reports[1].Merge(reports[0])
	}
	partition := l / 2
	return s.Merge(reports[partition:]).Merge(s.Merge(reports[:partition]))
}
