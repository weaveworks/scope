package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/report"
)

// We merge all reports received within the specified interval, and
// discard the orignals. Higher figures improve the performance of
// Report(), but at the expense of lower time resolution, since time
// is effectively advancing in quantiles.
//
// The current figure is identical to the default
// probe.publishInterval, which results in performance improvements
// as soon as there is more than one probe.
const reportQuantisationInterval = 3 * time.Second

// Reporter is something that can produce reports on demand. It's a convenient
// interface for parts of the app, and several experimental components.
type Reporter interface {
	Report(context.Context, time.Time) (report.Report, error)
	WaitOn(context.Context, chan struct{})
	UnWait(context.Context, chan struct{})
}

// Adder is something that can accept reports. It's a convenient interface for
// parts of the app, and several experimental components.  It takes the following
// arguments:
// - context.Context: the request context
// - report.Report: the deserialised report
// - []byte: the serialised report (as gzip'd msgpack)
type Adder interface {
	Add(context.Context, report.Report, []byte) error
}

// A Collector is a Reporter and an Adder
type Collector interface {
	Reporter
	Adder
}

// Collector receives published reports from multiple producers. It yields a
// single merged report, representing all collected reports.
type collector struct {
	mtx        sync.Mutex
	reports    []report.Report
	timestamps []time.Time
	window     time.Duration
	cached     *report.Report
	merger     Merger
	waitableCondition
}

type waitableCondition struct {
	sync.Mutex
	waiters map[chan struct{}]struct{}
}

func (wc *waitableCondition) WaitOn(_ context.Context, waiter chan struct{}) {
	wc.Lock()
	wc.waiters[waiter] = struct{}{}
	wc.Unlock()
}

func (wc *waitableCondition) UnWait(_ context.Context, waiter chan struct{}) {
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
		merger: NewSmartMerger(),
	}
}

// Add adds a report to the collector's internal state. It implements Adder.
func (c *collector) Add(_ context.Context, rpt report.Report, _ []byte) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reports = append(c.reports, rpt)
	c.timestamps = append(c.timestamps, mtime.Now())

	c.clean()
	c.cached = nil
	if rpt.Shortcut {
		c.Broadcast()
	}
	return nil
}

// Report returns a merged report over all added reports. It implements
// Reporter.
func (c *collector) Report(_ context.Context, timestamp time.Time) (report.Report, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// If the oldest report is still within range,
	// and there is a cached report, return that.
	if c.cached != nil && len(c.reports) > 0 {
		oldest := timestamp.Add(-c.window)
		if c.timestamps[0].After(oldest) {
			return *c.cached, nil
		}
	}

	c.clean()
	c.quantise()

	rpt := c.merger.Merge(c.reports).Upgrade()
	c.cached = &rpt
	return rpt, nil
}

// remove reports older than the app.window
func (c *collector) clean() {
	var (
		cleanedReports    = make([]report.Report, 0, len(c.reports))
		cleanedTimestamps = make([]time.Time, 0, len(c.timestamps))
		oldest            = mtime.Now().Add(-c.window)
	)
	for i, r := range c.reports {
		if c.timestamps[i].After(oldest) {
			cleanedReports = append(cleanedReports, r)
			cleanedTimestamps = append(cleanedTimestamps, c.timestamps[i])
		}
	}
	c.reports = cleanedReports
	c.timestamps = cleanedTimestamps
}

// Merge reports received within the same reportQuantisationInterval.
//
// Quantisation is relative to the time of the first report in a given
// interval, rather than absolute time. So, for example, with a
// reportQuantisationInterval of 3s and reports with timestamps [0, 1,
// 2, 5, 6, 7], the result contains merged reports with
// timestamps/content of [0:{0,1,2}, 5:{5,6,7}].
func (c *collector) quantise() {
	if len(c.reports) == 0 {
		return
	}
	var (
		quantisedReports    = make([]report.Report, 0, len(c.reports))
		quantisedTimestamps = make([]time.Time, 0, len(c.timestamps))
	)
	quantumStartIdx := 0
	quantumStartTimestamp := c.timestamps[0]
	for i, t := range c.timestamps {
		if t.Sub(quantumStartTimestamp) < reportQuantisationInterval {
			continue
		}
		quantisedReports = append(quantisedReports, c.merger.Merge(c.reports[quantumStartIdx:i]))
		quantisedTimestamps = append(quantisedTimestamps, quantumStartTimestamp)
		quantumStartIdx = i
		quantumStartTimestamp = t
	}
	c.reports = append(quantisedReports, c.merger.Merge(c.reports[quantumStartIdx:]))
	c.timestamps = append(quantisedTimestamps, c.timestamps[quantumStartIdx])
}

// StaticCollector always returns the given report.
type StaticCollector report.Report

// Report returns a merged report over all added reports. It implements
// Reporter.
func (c StaticCollector) Report(context.Context, time.Time) (report.Report, error) {
	return report.Report(c), nil
}

// Add adds a report to the collector's internal state. It implements Adder.
func (c StaticCollector) Add(context.Context, report.Report, []byte) error { return nil }

// WaitOn lets other components wait on a new report being received. It
// implements Reporter.
func (c StaticCollector) WaitOn(context.Context, chan struct{}) {}

// UnWait lets other components stop waiting on a new report being received. It
// implements Reporter.
func (c StaticCollector) UnWait(context.Context, chan struct{}) {}

// NewFileCollector reads and parses the files at path (a file or
// directory) as reports.  If there are multiple files, and they all
// have names representing "nanoseconds since epoch" timestamps,
// e.g. "1488557088545489008.msgpack.gz", then the collector will
// return merged reports resulting from replaying the file reports in
// a loop at a sequence and speed determined by the timestamps.
// Otherwise the collector always returns the merger of all reports.
func NewFileCollector(path string, window time.Duration) (Collector, error) {
	var (
		timestamps []time.Time
		reports    []report.Report
	)
	allTimestamped := true
	if err := filepath.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			t, err := timestampFromFilepath(p)
			if err != nil {
				allTimestamped = false
			}
			timestamps = append(timestamps, t)

			rpt, err := report.MakeFromFile(p)
			if err != nil {
				return err
			}
			reports = append(reports, rpt)
			return nil
		}); err != nil {
		return nil, err
	}
	if len(reports) > 1 && allTimestamped {
		collector := NewCollector(window)
		go replay(collector, timestamps, reports)
		return collector, nil
	}
	return StaticCollector(NewSmartMerger().Merge(reports).Upgrade()), nil
}

func timestampFromFilepath(path string) (time.Time, error) {
	name := filepath.Base(path)
	for {
		ext := filepath.Ext(name)
		if ext == "" {
			break
		}
		name = strings.TrimSuffix(name, ext)
	}
	nanosecondsSinceEpoch, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("filename '%s' is not a number (representing nanoseconds since epoch): %v", name, err)
	}
	return time.Unix(0, nanosecondsSinceEpoch), nil
}

func replay(a Adder, timestamps []time.Time, reports []report.Report) {
	// calculate delays between report n and n+1
	l := len(timestamps)
	delays := make([]time.Duration, l, l)
	for i, t := range timestamps[0 : l-1] {
		delays[i] = timestamps[i+1].Sub(t)
		if delays[i] < 0 {
			panic(fmt.Errorf("replay timestamps are not in order! %v", timestamps))
		}
	}
	// We don't know how long to wait before looping round, so make a
	// good guess.
	delays[l-1] = timestamps[l-1].Sub(timestamps[0]) / time.Duration(l)

	due := time.Now()
	for {
		for i, r := range reports {
			a.Add(nil, r, nil)
			due = due.Add(delays[i])
			delay := due.Sub(time.Now())
			if delay > 0 {
				time.Sleep(delay)
			}
		}
	}
}
