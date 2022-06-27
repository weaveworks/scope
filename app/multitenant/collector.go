package multitenant

// Collect reports from probes per-tenant, and supply them to queriers on demand

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"context"

	"github.com/nats-io/nats"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/common/user"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
	"golang.org/x/sync/errgroup"
)

var (
	topologiesDropped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "topologies_dropped_total",
		Help:      "Total count of topologies dropped for being over limit.",
	}, []string{"user", "topology"})

	natsRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "nats_requests_total",
		Help:      "Total count of NATS requests.",
	}, []string{"method", "status_code"})

	reportReceivedSizeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "report_received_size_bytes",
		Help:      "Distribution of received report sizes",
		Buckets:   prometheus.ExponentialBuckets(4096, 2.0, 10),
	})
	reportsReceivedPerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "reports_received_total",
		Help:      "Total count of received reports per user.",
	}, []string{"user"})
	shortcutsReceivedPerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "shortcut_reports_received_total",
		Help:      "Total count of received shortcut reports per user.",
	}, []string{"user"})
	reportReceivedSizePerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "reports_received_bytes_total",
		Help:      "Total bytes received in reports per user.",
	}, []string{"user"})
)

func registerLiveCollectorMetrics() {
	prometheus.MustRegister(topologiesDropped)
	prometheus.MustRegister(natsRequests)
	prometheus.MustRegister(reportReceivedSizeHistogram)
	prometheus.MustRegister(reportsReceivedPerUser)
	prometheus.MustRegister(shortcutsReceivedPerUser)
	prometheus.MustRegister(reportReceivedSizePerUser)
}

var registerLiveCollectorMetricsOnce sync.Once

// LiveCollectorConfig has everything we need to make a collector for live multitenant data.
type LiveCollectorConfig struct {
	UserIDer       UserIDer
	NatsHost       string
	MemcacheClient *MemcacheClient
	Window         time.Duration
	TickInterval   time.Duration
	MaxTopNodes    int
	CollectorAddr  string
}

type liveCollector struct {
	cfg           LiveCollectorConfig
	merger        app.Merger
	pending       sync.Map
	ticker        *time.Ticker
	tickCallbacks []func(context.Context)

	nats        *nats.Conn
	waitersLock sync.Mutex
	waiters     map[watchKey]*nats.Subscription

	collectors   []string
	lastResolved time.Time
}

// if StoreInterval is set, reports are merged into here and held until flushed to store
type pendingEntry struct {
	sync.Mutex
	report *report.Report
	older  []*report.Report
}

// NewLiveCollector makes a new LiveCollector from the supplied config.
func NewLiveCollector(config LiveCollectorConfig) (app.Collector, error) {
	c := &liveCollector{
		cfg: config,
	}
	return c, c.init()
}

func (c *liveCollector) init() error {
	registerLiveCollectorMetricsOnce.Do(registerLiveCollectorMetrics)
	var nc *nats.Conn
	if c.cfg.NatsHost != "" {
		if c.cfg.MemcacheClient == nil {
			return fmt.Errorf("Must supply memcache client when using nats")
		}
		var err error
		nc, err = nats.Connect(c.cfg.NatsHost)
		if err != nil {
			return err
		}
	}
	c.nats = nc
	c.merger = app.NewFastMerger()
	c.waiters = make(map[watchKey]*nats.Subscription)
	if c.isCollector() {
		if c.cfg.TickInterval == 0 {
			return fmt.Errorf("--app.collector.tick-interval or --app.collector.store-interval must be non-zero for a collector")
		}
		c.ticker = time.NewTicker(c.cfg.TickInterval)
		go c.tickLoop()
	}
	c.tickCallbacks = append(c.tickCallbacks, c.bumpPending)
	return nil
}

func (c *liveCollector) tickLoop() {
	for range c.ticker.C {
		for _, f := range c.tickCallbacks {
			f(context.Background())
		}
	}
}

// Close will close things down
func (c *liveCollector) Close() {
	c.ticker.Stop() // note this doesn't close the chan; goroutine keeps running
}

// Range over all users (instances) that have pending reports and shift the data back in the array
func (c *liveCollector) bumpPending(ctx context.Context) {
	c.pending.Range(func(key, value interface{}) bool {
		entry := value.(*pendingEntry)

		entry.Lock()
		rpt := entry.report
		entry.report = nil
		if entry.older == nil {
			entry.older = make([]*report.Report, c.cfg.Window/c.cfg.TickInterval)
		} else {
			copy(entry.older[1:], entry.older) // move everything down one
		}
		entry.older[0] = rpt
		entry.Unlock()
		return true
	})
}

func (c *liveCollector) HasHistoricReports() bool {
	return false
}

func (c *liveCollector) HasReports(ctx context.Context, timestamp time.Time) (bool, error) {
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return false, err
	}
	if time.Since(timestamp) < c.cfg.Window {
		has, err := c.hasReportsFromLive(ctx, userid)
		return has, err
	}
	return false, nil
}

func (c *liveCollector) Add(ctx context.Context, rep report.Report, buf []byte) error {
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return err
	}

	reportReceivedSizeHistogram.Observe(float64(len(buf)))
	reportReceivedSizePerUser.WithLabelValues(userid).Add(float64(len(buf)))
	reportsReceivedPerUser.WithLabelValues(userid).Inc()

	// Shortcut reports are published to nats but not persisted -
	// we'll get a full report from the same probe in a few seconds
	if rep.Shortcut {
		shortcutsReceivedPerUser.WithLabelValues(userid).Inc()
		if c.nats != nil {
			_, _, reportKey := calculateReportKeys(userid, time.Now())
			_, err = c.cfg.MemcacheClient.StoreReportBytes(ctx, reportKey, buf)
			if err != nil {
				log.Warningf("Could not store shortcut %v in memcache: %v", reportKey, err)
				// No point publishing on nats if cache store failed
				return nil
			}
			err := c.nats.Publish(userid, []byte(reportKey))
			natsRequests.WithLabelValues("Publish", instrument.ErrorCode(err)).Add(1)
			if err != nil {
				log.Errorf("Error sending shortcut report: %v", err)
			}
		}
		return nil
	}

	rep = c.massageReport(userid, rep)
	c.addToLive(ctx, userid, rep)

	return nil
}

// process a report from a probe which may be at an older version or overloaded
func (c *liveCollector) massageReport(userid string, report report.Report) report.Report {
	if c.cfg.MaxTopNodes > 0 {
		max := c.cfg.MaxTopNodes
		if len(report.Host.Nodes) > 1 {
			max = max * len(report.Host.Nodes) // higher limit for merged reports
		}
		var dropped []string
		report, dropped = report.DropTopologiesOver(max)
		for _, name := range dropped {
			topologiesDropped.WithLabelValues(userid, name).Inc()
		}
	}
	report = report.Upgrade()
	return report
}

// We are building up a report in memory; merge into that (for awsCollector it will be saved shortly)
// NOTE: may retain a reference to rep; must not be used by caller after this.
func (c *liveCollector) addToLive(ctx context.Context, userid string, rep report.Report) {
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

func (c *liveCollector) isCollector() bool {
	return c.cfg.CollectorAddr == ""
}

func (c *liveCollector) hasReportsFromLive(ctx context.Context, userid string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "hasReportsFromLive")
	defer span.Finish()
	if c.isCollector() {
		e, found := c.pending.Load(userid)
		if !found {
			return false, nil
		}
		entry := e.(*pendingEntry)
		entry.Lock()
		defer entry.Unlock()
		if entry.report != nil {
			return true, nil
		}
		for _, v := range entry.older {
			if v != nil {
				return true, nil
			}
		}
		return false, nil
	}
	// We are a querier: ask each collector if it has any
	// (serially, since we will bail out on the first one that has reports)
	addrs := resolve(c.cfg.CollectorAddr)
	for _, addr := range addrs {
		body, err := oneCall(ctx, addr, "/api/probes?sparse=true", userid)
		if err != nil {
			return false, err
		}
		var hasReports bool
		decoder := json.NewDecoder(body)
		if err := decoder.Decode(&hasReports); err != nil {
			log.Errorf("Error encoding response: %v", err)
		}
		body.Close()
		if hasReports {
			return true, nil
		}
	}
	return false, nil
}

func (c *liveCollector) Report(ctx context.Context, timestamp time.Time) (report.Report, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "liveCollector.Report")
	defer span.Finish()
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return report.MakeReport(), err
	}
	span.SetTag("userid", userid)
	var reports []report.Report
	if time.Since(timestamp) < c.cfg.Window {
		reports, err = c.reportsFromLive(ctx, userid)
	}
	if err != nil {
		return report.MakeReport(), err
	}
	span.LogFields(otlog.Int("merging", len(reports)))
	return c.merger.Merge(reports), nil
}

func (c *liveCollector) reportsFromLive(ctx context.Context, userid string) ([]report.Report, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportsFromLive")
	defer span.Finish()
	if c.isCollector() {
		e, found := c.pending.Load(userid)
		if !found {
			return nil, nil
		}
		entry := e.(*pendingEntry)
		entry.Lock()
		ret := make([]report.Report, 0, len(entry.older)+1)
		if entry.report != nil {
			ret = append(ret, entry.report.Copy()) // Copy contents because this report is being unsafe-merged to
		}
		for _, v := range entry.older {
			if v != nil {
				ret = append(ret, *v) // no copy because older reports are immutable
			}
		}
		entry.Unlock()
		return ret, nil
	}

	// We are a querier: fetch the most up-to-date reports from collectors
	if time.Since(c.lastResolved) > time.Second*5 {
		c.collectors = resolve(c.cfg.CollectorAddr)
		c.lastResolved = time.Now()
	}
	reports := make([]*report.Report, len(c.collectors))
	// make a call to each collector and fetch its data for this userid
	g, ctx := errgroup.WithContext(ctx)
	for i, addr := range c.collectors {
		i, addr := i, addr // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			body, err := oneCall(ctx, addr, "/api/report", userid)
			if err != nil {
				log.Warnf("error calling '%s': %v", addr, err)
				return nil
			}
			reports[i], err = report.MakeFromBinary(ctx, body, false, true)
			body.Close()
			if err != nil {
				log.Warnf("error decoding: %v", err)
				return nil
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// dereference pointers into the expected return format
	ret := make([]report.Report, 0, len(reports))
	for _, rpt := range reports {
		if rpt != nil {
			ret = append(ret, *rpt)
		}
	}

	return ret, nil
}

func resolve(name string) []string {
	_, addrs, err := net.LookupSRV("", "", name)
	if err != nil {
		log.Warnf("Cannot resolve '%s': %v", name, err)
		return []string{}
	}
	endpoints := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		port := strconv.Itoa(int(addr.Port))
		endpoints = append(endpoints, net.JoinHostPort(addr.Target, port))
	}
	return endpoints
}

func oneCall(ctx context.Context, endpoint, path, userid string) (io.ReadCloser, error) {
	fullPath := "http://" + endpoint + path
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request %s: %w", fullPath, err)
	}
	req = req.WithContext(ctx)
	req.Header.Set(user.OrgIDHeaderName, userid)
	req.Header.Set("Accept", "application/msgpack")
	req.Header.Set("Accept-Encoding", "identity") // disable compression
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		var ht *nethttp.Tracer
		req, ht = nethttp.TraceRequest(parentSpan.Tracer(), req, nethttp.OperationName("Collector Fetch"))
		defer ht.Finish()
	}
	client := &http.Client{Transport: &nethttp.Transport{}}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", fullPath, err)
	}
	if res.StatusCode != http.StatusOK {
		content, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf("error from collector: %s (%s)", res.Status, string(content))
	}

	return res.Body, nil
}

func (c *liveCollector) WaitOn(ctx context.Context, waiter chan struct{}) {
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		log.Errorf("Error getting user id in WaitOn: %v", err)
		return
	}

	if c.nats == nil {
		return
	}

	sub, err := c.nats.SubscribeSync(userid)
	natsRequests.WithLabelValues("SubscribeSync", instrument.ErrorCode(err)).Add(1)
	if err != nil {
		log.Errorf("Error subscribing for shortcuts: %v", err)
		return
	}

	c.waitersLock.Lock()
	c.waiters[watchKey{userid, waiter}] = sub
	c.waitersLock.Unlock()

	go func() {
		for {
			_, err := sub.NextMsg(natsTimeout)
			if err == nats.ErrTimeout {
				continue
			}
			natsRequests.WithLabelValues("NextMsg", instrument.ErrorCode(err)).Add(1)
			if err != nil {
				log.Debugf("NextMsg error: %v", err)
				return
			}
			select {
			case waiter <- struct{}{}:
			default:
			}
		}
	}()
}

func (c *liveCollector) UnWait(ctx context.Context, waiter chan struct{}) {
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		log.Errorf("Error getting user id in WaitOn: %v", err)
		return
	}

	if c.nats == nil {
		return
	}

	c.waitersLock.Lock()
	key := watchKey{userid, waiter}
	sub := c.waiters[key]
	delete(c.waiters, key)
	c.waitersLock.Unlock()

	err = sub.Unsubscribe()
	natsRequests.WithLabelValues("Unsubscribe", instrument.ErrorCode(err)).Add(1)
	if err != nil {
		log.Errorf("Error on unsubscribe: %v", err)
	}
}

// AdminSummary returns a string with some internal information about
// the report, which may be useful to troubleshoot.
func (c *liveCollector) AdminSummary(ctx context.Context, timestamp time.Time) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "liveCollector.AdminSummary")
	defer span.Finish()
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return "", err
	}
	_ = userid
	// TODO: finish implementation
	return "TODO", nil
}
