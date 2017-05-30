package multitenant

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bluele/gcache"
	"github.com/nats-io/nats"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"
	natsTimeout = 10 * time.Second
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

	natsRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "nats_requests_total",
		Help:      "Total count of NATS requests.",
	}, []string{"method", "status_code"})
)

func init() {
	prometheus.MustRegister(dynamoRequestDuration)
	prometheus.MustRegister(dynamoConsumedCapacity)
	prometheus.MustRegister(dynamoValueSize)
	prometheus.MustRegister(inProcessCacheRequests)
	prometheus.MustRegister(inProcessCacheHits)
	prometheus.MustRegister(reportSizeHistogram)
	prometheus.MustRegister(natsRequests)
}

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
	UserIDer       UserIDer
	DynamoDBConfig *aws.Config
	DynamoTable    string
	S3Store        *S3Store
	NatsHost       string
	MemcacheClient *MemcacheClient
	Window         time.Duration
}

type awsCollector struct {
	userIDer  UserIDer
	db        *dynamodb.DynamoDB
	s3        *S3Store
	tableName string
	merger    app.Merger
	inProcess inProcessStore
	memcache  *MemcacheClient
	window    time.Duration

	nats        *nats.Conn
	waitersLock sync.Mutex
	waiters     map[watchKey]*nats.Subscription
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
func NewAWSCollector(config AWSCollectorConfig) (AWSCollector, error) {
	var nc *nats.Conn
	if config.NatsHost != "" {
		var err error
		nc, err = nats.Connect(config.NatsHost)
		if err != nil {
			return nil, err
		}
	}

	// (window * report rate) * number of hosts per user * number of users
	reportCacheSize := (int(config.Window.Seconds()) / 3) * 10 * 5
	return &awsCollector{
		db:        dynamodb.New(session.New(config.DynamoDBConfig)),
		s3:        config.S3Store,
		userIDer:  config.UserIDer,
		tableName: config.DynamoTable,
		merger:    app.NewSmartMerger(),
		inProcess: newInProcessStore(reportCacheSize, config.Window),
		memcache:  config.MemcacheClient,
		window:    config.Window,
		nats:      nc,
		waiters:   map[watchKey]*nats.Subscription{},
	}, nil
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
		if *s == c.tableName {
			return nil
		}
	}

	params := &dynamodb.CreateTableInput{
		TableName: aws.String(c.tableName),
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
	log.Infof("Creating table %s", c.tableName)
	_, err = c.db.CreateTable(params)
	return err
}

// getReportKeys gets the s3 keys for reports in this range
func (c *awsCollector) getReportKeys(ctx context.Context, userid string, row int64, start, end time.Time) ([]string, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	var resp *dynamodb.QueryOutput
	err := instrument.TimeRequestHistogram(ctx, "DynamoDB.Query", dynamoRequestDuration, func(_ context.Context) error {
		var err error
		resp, err = c.db.Query(&dynamodb.QueryInput{
			TableName: aws.String(c.tableName),
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

	result := []string{}
	for _, item := range resp.Items {
		reportKey := item[reportField].S
		if reportKey == nil {
			log.Errorf("Empty row!")
			continue
		}
		dynamoValueSize.WithLabelValues("BatchGetItem").
			Add(float64(len(*reportKey)))
		result = append(result, *reportKey)
	}
	return result, nil
}

func (c *awsCollector) getReports(ctx context.Context, reportKeys []string) ([]report.Report, error) {
	missing := reportKeys

	stores := []ReportStore{c.inProcess}
	if c.memcache != nil {
		stores = append(stores, c.memcache)
	}
	stores = append(stores, c.s3)

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

func (c *awsCollector) Report(ctx context.Context, timestamp time.Time) (report.Report, error) {
	var (
		end         = timestamp
		start       = end.Add(-c.window)
		rowStart    = start.UnixNano() / time.Hour.Nanoseconds()
		rowEnd      = end.UnixNano() / time.Hour.Nanoseconds()
		userid, err = c.userIDer(ctx)
	)
	if err != nil {
		return report.MakeReport(), err
	}

	// Queries will only every span 2 rows max.
	var reportKeys []string
	if rowStart != rowEnd {
		reportKeys1, err := c.getReportKeys(ctx, userid, rowStart, start, end)
		if err != nil {
			return report.MakeReport(), err
		}

		reportKeys2, err := c.getReportKeys(ctx, userid, rowEnd, start, end)
		if err != nil {
			return report.MakeReport(), err
		}

		reportKeys = append(reportKeys, reportKeys1...)
		reportKeys = append(reportKeys, reportKeys2...)
	} else {
		if reportKeys, err = c.getReportKeys(ctx, userid, rowEnd, start, end); err != nil {
			return report.MakeReport(), err
		}
	}

	log.Debugf("Fetching %d reports from %v to %v", len(reportKeys), start, end)
	reports, err := c.getReports(ctx, reportKeys)
	if err != nil {
		return report.MakeReport(), err
	}

	return c.merger.Merge(reports).Upgrade(), nil
}

// calculateDynamoKeys generates the row & column keys for Dynamo.
func calculateDynamoKeys(userid string, now time.Time) (string, string) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(now.UnixNano()/time.Hour.Nanoseconds(), 10))
	colKey := strconv.FormatInt(now.UnixNano(), 10)
	return rowKey, colKey
}

// calculateReportKey determines the key we should use for a report.
func calculateReportKey(rowKey, colKey string) (string, error) {
	rowKeyHash := md5.New()
	if _, err := io.WriteString(rowKeyHash, rowKey); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x/%s", rowKeyHash.Sum(nil), colKey), nil
}

func (c *awsCollector) Add(ctx context.Context, rep report.Report, buf []byte) error {
	userid, err := c.userIDer(ctx)
	if err != nil {
		return err
	}

	// first, put the report on s3
	rowKey, colKey := calculateDynamoKeys(userid, time.Now())
	reportKey, err := calculateReportKey(rowKey, colKey)
	if err != nil {
		return err
	}

	reportSize, err := c.s3.StoreReportBytes(ctx, reportKey, buf)
	if err != nil {
		return err
	}
	reportSizeHistogram.Observe(float64(reportSize))

	// third, put it in memcache
	if c.memcache != nil {
		_, err = c.memcache.StoreReportBytes(ctx, reportKey, buf)
		if err != nil {
			// NOTE: We don't abort here because failing to store in memcache
			// doesn't actually break anything else -- it's just an
			// optimization.
			log.Warningf("Could not store %v in memcache: %v", reportKey, err)
		}
	}

	// fourth, put the key in dynamodb
	dynamoValueSize.WithLabelValues("PutItem").
		Add(float64(len(reportKey)))

	var resp *dynamodb.PutItemOutput
	err = instrument.TimeRequestHistogram(ctx, "DynamoDB.PutItem", dynamoRequestDuration, func(_ context.Context) error {
		var err error
		resp, err = c.db.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(c.tableName),
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
		return err
	})
	if resp.ConsumedCapacity != nil {
		dynamoConsumedCapacity.WithLabelValues("PutItem").
			Add(float64(*resp.ConsumedCapacity.CapacityUnits))
	}
	if err != nil {
		return err
	}

	if rep.Shortcut && c.nats != nil {
		err := c.nats.Publish(userid, []byte(reportKey))
		natsRequests.WithLabelValues("Publish", instrument.ErrorCode(err)).Add(1)
		if err != nil {
			log.Errorf("Error sending shortcut report: %v", err)
		}
	}

	return nil
}

func (c *awsCollector) WaitOn(ctx context.Context, waiter chan struct{}) {
	userid, err := c.userIDer(ctx)
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

func (c *awsCollector) UnWait(ctx context.Context, waiter chan struct{}) {
	userid, err := c.userIDer(ctx)
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
