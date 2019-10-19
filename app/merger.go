package app

import (
	"fmt"
	"sort"

	"github.com/spaolacci/murmur3"

	"github.com/weaveworks/scope/report"
)

// Merger is the type for a thing that can merge reports.
type Merger interface {
	Merge([]report.Report) report.Report
}

type fastMerger struct{}

// NewFastMerger makes a Merger which merges together reports, mutating the one we are building up
func NewFastMerger() Merger {
	return fastMerger{}
}

type reportsByTS []report.Report

func (r reportsByTS) Len() int           { return len(r) }
func (r reportsByTS) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r reportsByTS) Less(i, j int) bool { return r[i].TS.Before(r[j].TS) }

func (fastMerger) Merge(reports []report.Report) report.Report {
	sort.Sort(reportsByTS(reports))
	rpt := report.MakeReport()
	id := murmur3.New64()
	for _, r := range reports {
		rpt.UnsafeMerge(r)
		id.Write([]byte(r.ID))
	}
	rpt.ID = fmt.Sprintf("%x", id.Sum64())
	return rpt
}
