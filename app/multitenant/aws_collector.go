package multitenant

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bluele/gcache"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"
	natsTimeout = 10 * time.Second

	reportQuantisationInterval = 3 * time.Second
	// Grace period allows for some gap between the timestamp on reports
	// (assigned when they arrive at collector) and them appearing in DynamoDB query
	gracePeriod = 500 * time.Millisecond
)

var (
	dynamoRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "dynamo_request_duration_seconds",
		Help:      "Time in seconds spent doing DynamoDB requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
	dynamoConsumedCapacity = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_consumed_capacity_total",
		Help:      "Total count of capacity units consumed per operation.",
	}, []string{"method"})
	dynamoValueSize = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_value_size_bytes_total",
		Help:      "Total size of data read / written from DynamoDB in bytes.",
	}, []string{"method"})

	inProcessCacheRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "in_process_cache_requests_total",
		Help:      "Total count of reports requested from the in-process cache.",
	})

	inProcessCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "in_process_cache_hits_total",
		Help:      "Total count of reports found in the in-process cache.",
	})

	reportSizeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "report_size_bytes",
		Help:      "Distribution of memcache report sizes",
		Buckets:   prometheus.ExponentialBuckets(4096, 2.0, 10),
	})
	reportsPerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "reports_stored_total",
		Help:      "Total count of stored reports per user.",
	}, []string{"user"})
	reportSizePerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "reports_bytes_total",
		Help:      "Total bytes stored in reports per user.",
	}, []string{"user"})
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

	flushDuration = instrument.NewHistogramCollectorFromOpts(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "flush_duration_seconds",
		Help:      "Time in seconds spent flushing merged reports.",
		Buckets:   prometheus.DefBuckets,
	})
)

func registerAWSCollectorMetrics() {
	prometheus.MustRegister(dynamoRequestDuration)
	prometheus.MustRegister(dynamoConsumedCapacity)
	prometheus.MustRegister(dynamoValueSize)
	prometheus.MustRegister(inProcessCacheRequests)
	prometheus.MustRegister(inProcessCacheHits)
	flushDuration.Register()
}

var registerAWSCollectorMetricsOnce sync.Once

// AWSCollector is a Collector which can also CreateTables
type AWSCollector interface {
	app.Collector
	CreateTables() error
}

// ReportStore is a thing that we can get reports from.
type ReportStore interface {
	FetchReports(context.Context, []string) (map[string]report.Report, []string, error)
}

// AWSCollectorConfig has everything we need to make an AWS collector.
type AWSCollectorConfig struct {
	DynamoDBConfig *aws.Config
	DynamoTable    string
	S3Store        *S3Store
}

type awsCollector struct {
	liveCollector
	awsCfg    AWSCollectorConfig
	db        *dynamodb.DynamoDB
	inProcess inProcessStore
}

// Shortcut reports:
// When the UI connects a WS to the query service, a goroutine periodically
// published rendered reports to that ws.  This process can be interrupted by
// "shortcut" reports, causing the query service to push a render report
// immediately. This whole process is controlled by the aforementioned
// goroutine registering a channel with the collector.  We store these
// registered channels in a map keyed by the userid and the channel itself,
// which in go is hashable.  We then listen on a NATS topic for any shortcut
// reports coming from the collection service.
type watchKey struct {
	userid string
	c      chan struct{}
}

// NewAWSCollector the elastic reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewAWSCollector(liveConfig LiveCollectorConfig, config AWSCollectorConfig) (AWSCollector, error) {
	registerAWSCollectorMetricsOnce.Do(registerAWSCollectorMetrics)

	// (window * report rate) * number of hosts per user * number of users
	reportCacheSize := (int(liveConfig.Window.Seconds()) / 3) * 10 * 5
	c := &awsCollector{
		liveCollector: liveCollector{cfg: liveConfig},
		awsCfg:        config,
		db:            dynamodb.New(session.New(config.DynamoDBConfig)),
		inProcess:     newInProcessStore(reportCacheSize, liveConfig.Window+reportQuantisationInterval),
	}
	err := c.liveCollector.init()
	if err != nil {
		return nil, err
	}
	c.tickCallbacks = append(c.tickCallbacks, c.flushPending)
	return c, nil
}

// Range over all users (instances) that have pending reports and send to store
func (c *awsCollector) flushPending(ctx context.Context) {
	instrument.CollectedRequest(ctx, "FlushPending", flushDuration, nil, func(ctx context.Context) error {
		type queueEntry struct {
			userid string
			buf    []byte
		}
		queue := make(chan queueEntry)
		const numParallel = 10
		var group sync.WaitGroup
		group.Add(numParallel)
		// Run n parallel goroutines fetching reports from the queue and flushing them
		for i := 0; i < numParallel; i++ {
			go func() {
				for entry := range queue {
					rowKey, colKey, reportKey := calculateReportKeys(entry.userid, time.Now())
					err := c.persistReport(ctx, entry.userid, rowKey, colKey, reportKey, entry.buf)
					if err != nil {
						log.Errorf("Could not persist combined report: %v", err)
					}
				}
				group.Done()
			}()
		}
		c.pending.Range(func(key, value interface{}) bool {
			userid := key.(string)
			entry := value.(*pendingEntry)

			entry.Lock()
			rpt := entry.older[0]
			entry.Unlock()

			if rpt != nil {
				// serialise reports on one goroutine to limit CPU usage
				buf, err := rpt.WriteBinary()
				if err != nil {
					log.Errorf("Could not serialise combined report: %v", err)
					return true
				}
				queue <- queueEntry{userid: userid, buf: buf.Bytes()}
			}
			return true
		})
		close(queue)
		group.Wait()
		return nil
	})
}

// Close will flush pending data
func (c *awsCollector) Close() {
	c.liveCollector.Close()
	c.flushPending(context.Background())
}

// CreateTables creates the required tables in dynamodb
func (c *awsCollector) CreateTables() error {
	// see if tableName exists
	resp, err := c.db.ListTables(&dynamodb.ListTablesInput{
		Limit: aws.Int64(10),
	})
	if err != nil {
		return err
	}
	for _, s := range resp.TableNames {
		if *s == c.awsCfg.DynamoTable {
			return nil
		}
	}

	params := &dynamodb.CreateTableInput{
		TableName: aws.String(c.awsCfg.DynamoTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(hourField),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(tsField),
				AttributeType: aws.String("N"),
			},
			// Don't need to specify non-key attributes in schema
			//{
			//	AttributeName: aws.String(reportField),
			//	AttributeType: aws.String("S"),
			//},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(hourField),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(tsField),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(5),
		},
	}
	log.Infof("Creating table %s", c.awsCfg.DynamoTable)
	_, err = c.db.CreateTable(params)
	return err
}

type keyInfo struct {
	key string
	ts  int64
}

// reportKeysInRange returns the s3 keys for reports in the specified range
func (c *awsCollector) reportKeysInRange(ctx context.Context, userid string, row int64, start, end time.Time) ([]keyInfo, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	var resp *dynamodb.QueryOutput
	err := instrument.TimeRequestHistogram(ctx, "DynamoDB.Query", dynamoRequestDuration, func(_ context.Context) error {
		var err error
		resp, err = c.db.Query(&dynamodb.QueryInput{
			TableName: aws.String(c.awsCfg.DynamoTable),
			KeyConditions: map[string]*dynamodb.Condition{
				hourField: {
					AttributeValueList: []*dynamodb.AttributeValue{
						{S: aws.String(rowKey)},
					},
					ComparisonOperator: aws.String("EQ"),
				},
				tsField: {
					AttributeValueList: []*dynamodb.AttributeValue{
						{N: aws.String(strconv.FormatInt(start.UnixNano(), 10))},
						{N: aws.String(strconv.FormatInt(end.UnixNano(), 10))},
					},
					ComparisonOperator: aws.String("BETWEEN"),
				},
			},
			ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		})
		return err
	})
	if resp.ConsumedCapacity != nil {
		dynamoConsumedCapacity.WithLabelValues("Query").
			Add(float64(*resp.ConsumedCapacity.CapacityUnits))
	}
	if err != nil {
		return nil, err
	}

	result := []keyInfo{}
	for _, item := range resp.Items {
		reportKey := item[reportField].S
		tsValue := item[tsField].N
		if reportKey == nil || tsValue == nil {
			log.Errorf("Empty row!")
			continue
		}
		dynamoValueSize.WithLabelValues("BatchGetItem").
			Add(float64(len(*reportKey)))
		ts, _ := strconv.ParseInt(*tsValue, 10, 64)
		result = append(result, keyInfo{key: *reportKey, ts: ts})
	}
	return result, nil
}

// getReportKeys returns the S3 for reports in the interval [start, end].
func (c *awsCollector) getReportKeys(ctx context.Context, userid string, start, end time.Time) ([]keyInfo, error) {
	var (
		rowStart = start.UnixNano() / time.Hour.Nanoseconds()
		rowEnd   = end.UnixNano() / time.Hour.Nanoseconds()
		err      error
	)

	// Queries will only every span 2 rows max.
	var reportKeys []keyInfo
	if rowStart != rowEnd {
		reportKeys1, err := c.reportKeysInRange(ctx, userid, rowStart, start, end)
		if err != nil {
			return nil, err
		}
		reportKeys2, err := c.reportKeysInRange(ctx, userid, rowEnd, start, end)
		if err != nil {
			return nil, err
		}
		reportKeys = append(reportKeys, reportKeys1...)
		reportKeys = append(reportKeys, reportKeys2...)
	} else {
		if reportKeys, err = c.reportKeysInRange(ctx, userid, rowEnd, start, end); err != nil {
			return nil, err
		}
	}

	return reportKeys, nil
}

func (c *awsCollector) getReports(ctx context.Context, userid string, reportKeys []string) ([]report.Report, error) {
	missing := reportKeys

	stores := []ReportStore{c.inProcess}
	if c.cfg.MemcacheClient != nil {
		stores = append(stores, c.cfg.MemcacheClient)
	}
	stores = append(stores, c.awsCfg.S3Store)

	var reports []report.Report
	for _, store := range stores {
		if store == nil {
			continue
		}
		found, newMissing, err := store.FetchReports(ctx, missing)
		missing = newMissing
		if err != nil {
			log.Warningf("Error fetching from cache: %v", err)
		}
		for key, report := range found {
			report = c.massageReport(userid, report)
			c.inProcess.StoreReport(key, report)
			reports = append(reports, report)
		}
		if len(missing) == 0 {
			return reports, nil
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("Error fetching from s3, still have missing reports: %v", missing)
	}
	return reports, nil
}

// If we are running as a Query service, fetch data and merge into a report
// If we are running as a Collector and the request is for live data, merge in-memory data and return
func (c *awsCollector) Report(ctx context.Context, timestamp time.Time) (report.Report, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "awsCollector.Report")
	defer span.Finish()
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return report.MakeReport(), err
	}
	span.SetTag("userid", userid)
	var reports []report.Report
	if time.Since(timestamp) < c.cfg.Window {
		reports, err = c.reportsFromLive(ctx, userid)
	} else {
		reports, err = c.reportsFromStore(ctx, userid, timestamp)
	}
	if err != nil {
		return report.MakeReport(), err
	}
	span.LogFields(otlog.Int("merging", len(reports)))
	return c.merger.Merge(reports), nil
}

/*
Given a timestamp in the past, fetch reports within the window from store or cache

S3 stores original reports from one probe at the timestamp they arrived at collector.
Collector also sends every report to memcached.
The in-memory cache stores:
 - individual reports deserialised, under S3 key for report
 - sets of reports in interval [t,t+3) merged, under key "instance:t"
   - so to check the cache for reports from 14:31:00 to 14:31:15 you would request 5 keys 3 seconds apart
*/
func (c *awsCollector) reportsFromStore(ctx context.Context, userid string, timestamp time.Time) ([]report.Report, error) {
	span := opentracing.SpanFromContext(ctx)
	end := timestamp
	start := end.Add(-c.cfg.Window)
	reportKeys, err := c.getReportKeys(ctx, userid, start, end)
	if err != nil {
		return nil, err
	}
	span.LogFields(otlog.Int("keys", len(reportKeys)), otlog.String("timestamp", timestamp.String()))

	var reports []report.Report
	// Fetch a merged report for each time quantum covering the window
	startTS, endTS := start.UnixNano(), end.UnixNano()
	ts := startTS - (startTS % reportQuantisationInterval.Nanoseconds())
	for ; ts+(reportQuantisationInterval+gracePeriod).Nanoseconds() < endTS; ts += reportQuantisationInterval.Nanoseconds() {
		quantumReport, err := c.reportForQuantum(ctx, userid, reportKeys, ts)
		if err != nil {
			return nil, err
		}
		reports = append(reports, quantumReport)
	}
	// Fetch individual reports for the period after the last quantum
	last, err := c.reportsForKeysInRange(ctx, userid, reportKeys, ts, endTS)
	if err != nil {
		return nil, err
	}
	reports = append(reports, last...)
	return reports, nil
}

// Fetch a merged report either from cache or from store which we put in cache
func (c *awsCollector) reportForQuantum(ctx context.Context, userid string, reportKeys []keyInfo, start int64) (report.Report, error) {
	key := fmt.Sprintf("%s:%d", userid, start)
	cached, _, err := c.inProcess.FetchReports(ctx, []string{key})
	if len(cached) == 1 {
		return cached[key], nil
	}
	reports, err := c.reportsForKeysInRange(ctx, userid, reportKeys, start, start+reportQuantisationInterval.Nanoseconds())
	if err != nil {
		return report.MakeReport(), err
	}
	merged := c.merger.Merge(reports)
	c.inProcess.StoreReport(key, merged)
	return merged, nil
}

// Find the keys relating to this time period then fetch from memcached and/or S3
func (c *awsCollector) reportsForKeysInRange(ctx context.Context, userid string, reportKeys []keyInfo, start, end int64) ([]report.Report, error) {
	var keys []string
	for _, k := range reportKeys {
		if k.ts >= start && k.ts < end {
			keys = append(keys, k.key)
		}
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.LogFields(otlog.Int("fetching", len(keys)), otlog.Int64("start", start), otlog.Int64("end", end))
	}
	log.Debugf("Fetching %d reports from %v to %v", len(keys), start, end)
	return c.getReports(ctx, userid, keys)
}

func (c *awsCollector) HasReports(ctx context.Context, timestamp time.Time) (bool, error) {
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return false, err
	}
	if time.Since(timestamp) < c.cfg.Window {
		has, err := c.hasReportsFromLive(ctx, userid)
		return has, err
	}
	start := timestamp.Add(-c.cfg.Window)
	reportKeys, err := c.getReportKeys(ctx, userid, start, timestamp)
	return len(reportKeys) > 0, err
}

func (c *awsCollector) HasHistoricReports() bool {
	return true
}

// AdminSummary returns a string with some internal information about
// the report, which may be useful to troubleshoot.
func (c *awsCollector) AdminSummary(ctx context.Context, timestamp time.Time) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "awsCollector.AdminSummary")
	defer span.Finish()
	userid, err := c.cfg.UserIDer(ctx)
	if err != nil {
		return "", err
	}
	end := timestamp
	start := end.Add(-c.cfg.Window)
	reportKeys, err := c.getReportKeys(ctx, userid, start, end)
	if err != nil {
		return "", err
	}
	reports, err := c.reportsForKeysInRange(ctx, userid, reportKeys, start.UnixNano(), end.UnixNano())
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i := range reports {
		// TODO: print the key - note reports may be in a different order from reportKeys
		b.WriteString(reports[i].Summary())
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// calculateDynamoKeys generates the row & column keys for Dynamo.
func calculateDynamoKeys(userid string, now time.Time) (string, string) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(now.UnixNano()/time.Hour.Nanoseconds(), 10))
	colKey := strconv.FormatInt(now.UnixNano(), 10)
	return rowKey, colKey
}

// calculateReportKeys returns DynamoDB row & col keys, and S3/memcached key that we will use for a report
func calculateReportKeys(userid string, now time.Time) (string, string, string) {
	rowKey, colKey := calculateDynamoKeys(userid, now)
	rowKeyHash := md5.New()
	_, _ = io.WriteString(rowKeyHash, rowKey) // hash write doesn't error
	return rowKey, colKey, fmt.Sprintf("%x/%s", rowKeyHash.Sum(nil), colKey)
}

func (c *awsCollector) persistReport(ctx context.Context, userid, rowKey, colKey, reportKey string, buf []byte) error {
	// Put in S3 and cache before index, so it is fetchable before it is discoverable
	reportSize, err := c.awsCfg.S3Store.StoreReportBytes(ctx, reportKey, buf)
	if err != nil {
		return err
	}
	if c.cfg.MemcacheClient != nil {
		_, err = c.cfg.MemcacheClient.StoreReportBytes(ctx, reportKey, buf)
		if err != nil {
			// NOTE: We don't abort here because failing to store in memcache
			// doesn't actually break anything else -- it's just an
			// optimization.
			log.Warningf("Could not store %v in memcache: %v", reportKey, err)
		}
	}

	dynamoValueSize.WithLabelValues("PutItem").Add(float64(len(reportKey)))

	err = instrument.TimeRequestHistogram(ctx, "DynamoDB.PutItem", dynamoRequestDuration, func(_ context.Context) error {
		resp, err := c.putItemInDynamo(rowKey, colKey, reportKey)
		if resp.ConsumedCapacity != nil {
			dynamoConsumedCapacity.WithLabelValues("PutItem").
				Add(float64(*resp.ConsumedCapacity.CapacityUnits))
		}
		return err
	})
	if err != nil {
		return err
	}

	reportSizeHistogram.Observe(float64(reportSize))
	reportSizePerUser.WithLabelValues(userid).Add(float64(reportSize))
	reportsPerUser.WithLabelValues(userid).Inc()

	return nil
}

func (c *awsCollector) putItemInDynamo(rowKey, colKey, reportKey string) (*dynamodb.PutItemOutput, error) {
	// Back off on ProvisionedThroughputExceededException
	const (
		maxRetries            = 5
		throuputExceededError = "ProvisionedThroughputExceededException"
	)
	var (
		resp    *dynamodb.PutItemOutput
		err     error
		retries = 0
		backoff = 50 * time.Millisecond
	)
	for {
		resp, err = c.db.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(c.awsCfg.DynamoTable),
			Item: map[string]*dynamodb.AttributeValue{
				hourField: {
					S: aws.String(rowKey),
				},
				tsField: {
					N: aws.String(colKey),
				},
				reportField: {
					S: aws.String(reportKey),
				},
			},
			ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		})
		if err != nil && retries < maxRetries {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == throuputExceededError {
				time.Sleep(backoff)
				retries++
				backoff *= 2
				continue
			}
		}
		break
	}
	return resp, err
}

type inProcessStore struct {
	cache gcache.Cache
}

// newInProcessStore creates an in-process store for reports.
func newInProcessStore(size int, expiration time.Duration) inProcessStore {
	return inProcessStore{gcache.New(size).LRU().Expiration(expiration).Build()}
}

// FetchReports retrieves the given reports from the store.
func (c inProcessStore) FetchReports(_ context.Context, keys []string) (map[string]report.Report, []string, error) {
	found := map[string]report.Report{}
	missing := []string{}
	for _, key := range keys {
		rpt, err := c.cache.Get(key)
		if err == nil {
			found[key] = rpt.(report.Report)
		} else {
			missing = append(missing, key)
		}
	}
	inProcessCacheHits.Add(float64(len(found)))
	inProcessCacheRequests.Add(float64(len(keys)))
	return found, missing, nil
}

// StoreReport stores a report in the store.
func (c inProcessStore) StoreReport(key string, report report.Report) {
	c.cache.Set(key, report)
}
