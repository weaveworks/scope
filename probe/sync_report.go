package probe

import (
	"sync"

	"github.com/weaveworks/scope/report"
)

type syncReport struct {
	mtx sync.RWMutex
	rpt report.Report
}

func (r *syncReport) swap(other report.Report) report.Report {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	old := r.rpt
	r.rpt = other
	return old
}

func (r *syncReport) copy() report.Report {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.rpt.Copy()
}
