package multitenant

// Collect reports from probes per-tenant, and supply them to queriers on demand

import (
	"sync"

	"context"

	"github.com/weaveworks/scope/report"
)

// if StoreInterval is set, reports are merged into here and held until flushed to store
type pendingEntry struct {
	sync.Mutex
	report *report.Report
}

// We are building up a report in memory; merge into that and it will be saved shortly
// NOTE: may retain a reference to rep; must not be used by caller after this.
func (c *awsCollector) addToLive(ctx context.Context, userid string, rep report.Report) {
	entry := &pendingEntry{}
	if e, found := c.pending.LoadOrStore(userid, entry); found {
		entry = e.(*pendingEntry)
	}
	entry.Lock()
	if entry.report == nil {
		entry.report = &rep
	} else {
		entry.report.UnsafeMerge(rep)
	}
	entry.Unlock()
}

