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

	"context"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/user"
	"github.com/weaveworks/scope/report"
	"golang.org/x/sync/errgroup"
)

// if StoreInterval is set, reports are merged into here and held until flushed to store
type pendingEntry struct {
	sync.Mutex
	report *report.Report
	older  []*report.Report
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

func (c *awsCollector) isCollector() bool {
	return c.cfg.StoreInterval != 0
}

func (c *awsCollector) hasReportsFromLive(ctx context.Context, userid string) (bool, error) {
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

func (c *awsCollector) reportsFromLive(ctx context.Context, userid string) ([]report.Report, error) {
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
	// TODO: resolve c.collectorAddress periodically instead of every time we make a call
	addrs := resolve(c.cfg.CollectorAddr)
	reports := make([]*report.Report, len(addrs))
	// make a call to each collector and fetch its data for this userid
	g, ctx := errgroup.WithContext(ctx)
	for i, addr := range addrs {
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
	ret := make([]report.Report, 0, len(addrs))
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
