package multitenant

import (
	"bytes"
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
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bluele/gcache"
	"github.com/nats-io/nats"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField              = "hour"
	tsField                = "ts"
	reportField            = "report"
	reportCacheSize        = (15 / 3) * 10 * 5 // (window size * report rate) * number of hosts per user * number of users
	reportCacheExpiration  = 15 * time.Second
	memcacheExpiration     = 15 // seconds
	memcacheUpdateInterval = 1 * time.Minute
	natsTimeout            = 10 * time.Second
)

var (
	dynamoRequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "dynamo_request_duration_seconds",
		Help:      "Time in seconds spent doing DynamoDB requests.",
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

	reportSize = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "report_size_bytes_total",
		Help:      "Total compressed size of reports received in bytes.",
	})

	s3RequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "s3_request_duration_seconds",
		Help:      "Time in seconds spent doing S3 requests.",
	}, []string{"method", "status_code"})

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
	prometheus.MustRegister(reportSize)
	prometheus.MustRegister(s3RequestDuration)
	prometheus.MustRegister(natsRequests)
}

// DynamoDBCollector is a Collector which can also CreateTables
type DynamoDBCollector interface {
	app.Collector
	CreateTables() error
}

// ReportStore is a thing that we can get reports from.
type ReportStore interface {
	FetchReports([]string) ([]report.Report, []string, error)
}

type dynamoDBCollector struct {
	userIDer   UserIDer
	db         *dynamodb.DynamoDB
	s3         *s3.S3
	tableName  string
	bucketName string
	merger     app.Merger
	inProcess  inProcessStore
	memcache   *MemcacheClient

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

// NewDynamoDBCollector the reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewDynamoDBCollector(
	userIDer UserIDer,
	dynamoDBConfig, s3Config *aws.Config,
	tableName, bucketName, natsHost, memcachedHost string,
	memcachedTimeout time.Duration, memcachedService string,
) (DynamoDBCollector, error) {
	var nc *nats.Conn
	if natsHost != "" {
		var err error
		nc, err = nats.Connect(natsHost)
		if err != nil {
			return nil, err
		}
	}

	var memcacheClient *MemcacheClient
	if memcachedHost != "" {
		var err error
		memcacheClient, err = NewMemcacheClient(memcachedHost, memcachedTimeout, memcachedService, memcacheUpdateInterval, memcacheExpiration)
		if err != nil {
			// TODO(jml): Ideally, we wouldn't abort here, we would instead
			// log errors when we try to use the memcache & fail to do so, as
			// aborting here introduces ordering dependencies into our
			// deployment.
			//
			// Note: this error only happens when either the memcachedHost or
			// any of the SRV records that it points to fail to resolve.
			return nil, err
		}
	}

	return &dynamoDBCollector{
		db:         dynamodb.New(session.New(dynamoDBConfig)),
		s3:         s3.New(session.New(s3Config)),
		userIDer:   userIDer,
		tableName:  tableName,
		bucketName: bucketName,
		merger:     app.NewSmartMerger(),
		inProcess:  newInProcessStore(reportCacheSize, reportCacheExpiration),
		memcache:   memcacheClient,
		nats:       nc,
		waiters:    map[watchKey]*nats.Subscription{},
	}, nil
}

// CreateDynamoDBTables creates the required tables in dynamodb
func (c *dynamoDBCollector) CreateTables() error {
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
func (c *dynamoDBCollector) getReportKeys(rowKey string, start, end time.Time) ([]string, error) {
	var resp *dynamodb.QueryOutput
	err := timeRequest("Query", dynamoRequestDuration, func() error {
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

// Fetch multiple reports in parallel from S3.
func (c *dynamoDBCollector) getNonCached(reportKeys []string) ([]report.Report, error) {
	type result struct {
		key    string
		report *report.Report
		err    error
	}

	ch := make(chan result, len(reportKeys))

	for _, reportKey := range reportKeys {
		go func(reportKey string) {
			r := result{key: reportKey}
			r.report, r.err = c.getNonCachedReport(reportKey)
			ch <- r
		}(reportKey)
	}

	reports := []report.Report{}
	for range reportKeys {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		reports = append(reports, *r.report)
		c.inProcess.StoreReport(r.key, *r.report)
	}
	return reports, nil
}

// Fetch a single report from S3.
func (c *dynamoDBCollector) getNonCachedReport(reportKey string) (*report.Report, error) {
	var resp *s3.GetObjectOutput
	err := timeRequest("Get", s3RequestDuration, func() error {
		var err error
		resp, err = c.s3.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(reportKey),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return report.MakeFromBinary(resp.Body)
}

func (c *dynamoDBCollector) getReports(userid string, row int64, start, end time.Time) ([]report.Report, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	missing, err := c.getReportKeys(rowKey, start, end)
	if err != nil {
		return nil, err
	}

	stores := []ReportStore{c.inProcess}
	if c.memcache != nil {
		stores = append(stores, c.memcache)
	}
	var reports []report.Report
	for _, store := range stores {
		if store == nil {
			continue
		}
		found, missing, err := store.FetchReports(missing)
		if err != nil {
			log.Warningf("Error fetching from cache: %v", err)
		}
		reports = append(reports, found...)
		if len(missing) == 0 {
			return reports, nil
		}
	}

	fetchedReports, err := c.getNonCached(missing)
	if err != nil {
		return nil, err
	}

	return append(reports, fetchedReports...), nil
}

func (c *dynamoDBCollector) Report(ctx context.Context) (report.Report, error) {
	var (
		now              = time.Now()
		start            = now.Add(-15 * time.Second)
		rowStart, rowEnd = start.UnixNano() / time.Hour.Nanoseconds(), now.UnixNano() / time.Hour.Nanoseconds()
		userid, err      = c.userIDer(ctx)
		reports          []report.Report
	)
	if err != nil {
		return report.MakeReport(), err
	}

	// Queries will only every span 2 rows max.
	if rowStart != rowEnd {
		reports1, err := c.getReports(userid, rowStart, start, now)
		if err != nil {
			return report.MakeReport(), err
		}

		reports2, err := c.getReports(userid, rowEnd, start, now)
		if err != nil {
			return report.MakeReport(), err
		}

		reports = append(reports1, reports2...)
	} else {
		if reports, err = c.getReports(userid, rowEnd, start, now); err != nil {
			return report.MakeReport(), err
		}
	}

	return c.merger.Merge(reports), nil
}

func (c *dynamoDBCollector) Add(ctx context.Context, rep report.Report) error {
	userid, err := c.userIDer(ctx)
	if err != nil {
		return err
	}

	// first, encode the report into a buffer and record its size
	var buf bytes.Buffer
	rep.WriteBinary(&buf)
	reportSize.Add(float64(buf.Len()))

	// second, put the report on s3
	now := time.Now()
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(now.UnixNano()/time.Hour.Nanoseconds(), 10))
	colKey := strconv.FormatInt(now.UnixNano(), 10)
	rowKeyHash := md5.New()
	if _, err := io.WriteString(rowKeyHash, rowKey); err != nil {
		return err
	}
	s3Key := fmt.Sprintf("%x/%s", rowKeyHash.Sum(nil), colKey)
	err = timeRequest("Put", s3RequestDuration, func() error {
		var err error
		_, err = c.s3.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(buf.Bytes()),
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(s3Key),
		})
		return err
	})
	if err != nil {
		return err
	}

	// third, put it in memcache
	if c.memcache != nil {
		err = c.memcache.StoreBytes(s3Key, buf.Bytes())
		if err != nil {
			// NOTE: We don't abort here because failing to store in memcache
			// doesn't actually break anything else -- it's just an
			// optimization.
			log.Warningf("Could not store %v in memcache: %v", s3Key, err)
		}
	}

	// fourth, put the key in dynamodb
	dynamoValueSize.WithLabelValues("PutItem").
		Add(float64(len(s3Key)))

	var resp *dynamodb.PutItemOutput
	err = timeRequest("PutItem", dynamoRequestDuration, func() error {
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
					S: aws.String(s3Key),
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
		err := c.nats.Publish(userid, []byte(s3Key))
		natsRequests.WithLabelValues("Publish", errorCode(err)).Add(1)
		if err != nil {
			log.Errorf("Error sending shortcut report: %v", err)
		}
	}

	return nil
}

func (c *dynamoDBCollector) WaitOn(ctx context.Context, waiter chan struct{}) {
	userid, err := c.userIDer(ctx)
	if err != nil {
		log.Errorf("Error getting user id in WaitOn: %v", err)
		return
	}

	if c.nats == nil {
		return
	}

	sub, err := c.nats.SubscribeSync(userid)
	natsRequests.WithLabelValues("SubscribeSync", errorCode(err)).Add(1)
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
			natsRequests.WithLabelValues("NextMsg", errorCode(err)).Add(1)
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

func (c *dynamoDBCollector) UnWait(ctx context.Context, waiter chan struct{}) {
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
	natsRequests.WithLabelValues("Unsubscribe", errorCode(err)).Add(1)
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
func (c inProcessStore) FetchReports(keys []string) ([]report.Report, []string, error) {
	found := []report.Report{}
	missing := []string{}
	for _, key := range keys {
		rpt, err := c.cache.Get(key)
		if err == nil {
			found = append(found, rpt.(report.Report))
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
