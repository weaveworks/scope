package multitenant

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/scope/report"
)

var (
	memcacheRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "memcache_requests_total",
		Help:      "Total count of reports requested from memcache that were not found in our in-memory cache.",
	})

	memcacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "memcache_hits_total",
		Help:      "Total count of reports found in memcache that were not found in our in-memory cache.",
	})

	memcacheRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "memcache_request_duration_seconds",
		Help:      "Total time spent in seconds doing memcache requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
)

func init() {
	prometheus.MustRegister(memcacheRequests)
	prometheus.MustRegister(memcacheHits)
	prometheus.MustRegister(memcacheRequestDuration)
}

// MemcacheClient is a memcache client that gets its server list from SRV
// records, and periodically updates that ServerList.
type MemcacheClient struct {
	client           *memcache.Client
	serverList       *memcache.ServerList
	expiration       int32
	hostname         string
	service          string
	compressionLevel int

	quit chan struct{}
	wait sync.WaitGroup
}

// MemcacheConfig defines how a MemcacheClient should be constructed.
type MemcacheConfig struct {
	Host             string
	Service          string
	Timeout          time.Duration
	UpdateInterval   time.Duration
	Expiration       time.Duration
	CompressionLevel int
}

// NewMemcacheClient creates a new MemcacheClient that gets its server list
// from SRV and updates the server list on a regular basis.
func NewMemcacheClient(config MemcacheConfig) *MemcacheClient {
	var servers memcache.ServerList
	client := memcache.NewFromSelector(&servers)
	client.Timeout = config.Timeout

	newClient := &MemcacheClient{
		client:           client,
		serverList:       &servers,
		expiration:       int32(config.Expiration.Seconds()),
		hostname:         config.Host,
		service:          config.Service,
		compressionLevel: config.CompressionLevel,
		quit:             make(chan struct{}),
	}
	err := newClient.updateMemcacheServers()
	if err != nil {
		log.Errorf("Error setting memcache servers to '%v': %v", config.Host, err)
	}

	newClient.wait.Add(1)
	go newClient.updateLoop(config.UpdateInterval)
	return newClient
}

// Stop the memcache client.
func (c *MemcacheClient) Stop() {
	close(c.quit)
	c.wait.Wait()
}

func (c *MemcacheClient) updateLoop(updateInterval time.Duration) error {
	defer c.wait.Done()
	ticker := time.NewTicker(updateInterval)
	var err error
	for {
		select {
		case <-ticker.C:
			err = c.updateMemcacheServers()
			if err != nil {
				log.Warningf("Error updating memcache servers: %v", err)
			}
		case <-c.quit:
			ticker.Stop()
		}
	}
}

// updateMemcacheServers sets a memcache server list from SRV records. SRV
// priority & weight are ignored.
func (c *MemcacheClient) updateMemcacheServers() error {
	_, addrs, err := net.LookupSRV(c.service, "tcp", c.hostname)
	if err != nil {
		return err
	}
	var servers []string
	for _, srv := range addrs {
		servers = append(servers, fmt.Sprintf("%s:%d", srv.Target, srv.Port))
	}
	// ServerList deterministically maps keys to _index_ of the server list.
	// Since DNS returns records in different order each time, we sort to
	// guarantee best possible match between nodes.
	sort.Strings(servers)
	return c.serverList.SetServers(servers...)
}

func memcacheStatusCode(err error) string {
	// See https://godoc.org/github.com/bradfitz/gomemcache/memcache#pkg-variables
	switch err {
	case nil:
		return "200"
	case memcache.ErrCacheMiss:
		return "404"
	case memcache.ErrMalformedKey:
		return "400"
	default:
		return "500"
	}
}

// FetchReports gets reports from memcache.
func (c *MemcacheClient) FetchReports(ctx context.Context, keys []string) (map[string]report.Report, []string, error) {
	defer memcacheRequests.Add(float64(len(keys)))
	var found map[string]*memcache.Item
	err := instrument.TimeRequestHistogramStatus(ctx, "Memcache.GetMulti", memcacheRequestDuration, memcacheStatusCode, func(_ context.Context) error {
		var err error
		found, err = c.client.GetMulti(keys)
		return err
	})
	if err != nil {
		return nil, keys, err
	}

	// Decode all the reports in parallel.
	type result struct {
		key    string
		report *report.Report
	}
	ch := make(chan result, len(keys))
	var missing []string
	for _, key := range keys {
		item, ok := found[key]
		if !ok {
			missing = append(missing, key)
			continue
		}
		go func(key string) {
			rep, err := report.MakeFromBytes(item.Value)
			if err != nil {
				log.Warningf("Corrupt report in memcache %v: %v", key, err)
				ch <- result{key: key}
				return
			}
			ch <- result{key: key, report: rep}
		}(key)
	}

	reports := map[string]report.Report{}
	lenFound := len(keys) - len(missing)
	for i := 0; i < lenFound; i++ {
		r := <-ch
		if r.report == nil {
			missing = append(missing, r.key)
		} else {
			reports[r.key] = *r.report
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		log.Warningf("Missing %d reports from memcache: %v", len(missing), missing)
	}

	memcacheHits.Add(float64(len(reports)))
	return reports, missing, nil
}

// StoreReportBytes stores a report.
func (c *MemcacheClient) StoreReportBytes(ctx context.Context, key string, rpt []byte) (int, error) {
	err := instrument.TimeRequestHistogramStatus(ctx, "Memcache.Put", memcacheRequestDuration, memcacheStatusCode, func(_ context.Context) error {
		item := memcache.Item{Key: key, Value: rpt, Expiration: c.expiration}
		return c.client.Set(&item)
	})
	return len(rpt), err
}
