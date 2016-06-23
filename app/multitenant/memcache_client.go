package multitenant

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/scope/report"
)

var (
	memcacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "memcache",
		Help:      "Reports that missed our in-memory cache but went to our memcache",
	}, []string{"result"})

	memcacheRequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "memcache_request_duration_nanoseconds",
		Help:      "Time spent doing memcache requests.",
	}, []string{"method", "status_code"})
)

func init() {
	prometheus.MustRegister(memcacheCounter)
	prometheus.MustRegister(memcacheRequestDuration)
}

// MemcacheClient is a memcache client that gets its server list from SRV
// records, and periodically updates that ServerList.
type MemcacheClient struct {
	client     *memcache.Client
	serverList *memcache.ServerList
	expiration int32
	hostname   string
	service    string

	quit chan struct{}
	wait sync.WaitGroup
}

// NewMemcacheClient creates a new MemcacheClient that gets its server list
// from SRV and updates the server list on a regular basis.
func NewMemcacheClient(host string, timeout time.Duration, service string, updateInterval time.Duration, expiration int32) (*MemcacheClient, error) {
	var servers memcache.ServerList
	client := memcache.NewFromSelector(&servers)
	client.Timeout = timeout

	newClient := &MemcacheClient{
		client:     client,
		serverList: &servers,
		expiration: expiration,
		hostname:   host,
		service:    service,
		quit:       make(chan struct{}),
	}
	err := newClient.updateMemcacheServers()
	if err != nil {
		return nil, err
	}

	newClient.wait.Add(1)
	go newClient.updateLoop(updateInterval)
	return newClient, nil
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

// FetchReports gets reports from memcache.
func (c *MemcacheClient) FetchReports(keys []string) ([]report.Report, []string, error) {
	var found map[string]*memcache.Item
	err := timeRequestStatus("Get", memcacheRequestDuration, memcacheStatusCode, func() error {
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
			rep, err := report.MakeFromBinary(bytes.NewReader(item.Value))
			if err != nil {
				log.Warningf("Corrupt report in memcache %v: %v", key, err)
				ch <- result{key: key}
				return
			}
			ch <- result{report: rep}
		}(key)
	}

	var reports []report.Report
	for i := 0; i < len(keys)-len(missing); i++ {
		r := <-ch
		if r.report == nil {
			missing = append(missing, r.key)
		} else {
			reports = append(reports, *r.report)
		}
	}

	memcacheCounter.WithLabelValues(hitLabel).Add(float64(len(reports)))
	memcacheCounter.WithLabelValues(missLabel).Add(float64(len(missing)))
	return reports, missing, nil
}

// StoreBytes stores a report, expecting the report to be serialized already.
func (c *MemcacheClient) StoreBytes(key string, content []byte) error {
	item := memcache.Item{Key: key, Value: content, Expiration: c.expiration}
	return c.client.Set(&item)
}
