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
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nats-io/nats"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField             = "hour"
	tsField               = "ts"
	reportField           = "report"
	reportCacheSize       = (15 / 3) * 10 * 5 // (window size * report rate) * number of hosts per user * number of users
	reportCacheExpiration = 15 * time.Second
	natsTimeout           = 10 * time.Second
)

var (
	dynamoRequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "dynamo_request_duration_nanoseconds",
		Help:      "Time spent doing DynamoDB requests.",
	}, []string{"method", "status_code"})
	dynamoCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_cache_hits",
		Help:      "Reports fetches that hit local cache.",
	})
	dynamoCacheMiss = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_cache_miss",
		Help:      "Reports fetches that miss local cache.",
	})
	dynamoConsumedCapacity = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_consumed_capacity",
		Help:      "The capacity units consumed by operation.",
	}, []string{"method"})
	dynamoValueSize = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_value_size_bytes",
		Help:      "Size of data read / written from dynamodb.",
	}, []string{"method"})

	reportSize = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "report_size_bytes",
		Help:      "Compressed size of reports received.",
	})

	s3RequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "s3_request_duration_nanoseconds",
		Help:      "Time spent doing S3 requests.",
	}, []string{"method", "status_code"})

	natsRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "nats_requests",
		Help:      "Number of NATS requests.",
	}, []string{"method", "status_code"})

	// XXX: jml thinks that maybe there's a simpler way to do these cache
	// hit/miss metrics but brain is too fuzzy right now.
	memcacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "memcache_hits",
		Help:      "Reports that missed our in-memory cache but went to our memcache",
	})

	memcacheMiss = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "memcache_miss",
		Help:      "Reports that missed both our in-memory cache and our memcache",
	})
)

func init() {
	prometheus.MustRegister(dynamoRequestDuration)
	prometheus.MustRegister(dynamoCacheHits)
	prometheus.MustRegister(dynamoCacheMiss)
	prometheus.MustRegister(dynamoConsumedCapacity)
	prometheus.MustRegister(dynamoValueSize)
	prometheus.MustRegister(reportSize)
	prometheus.MustRegister(s3RequestDuration)
	prometheus.MustRegister(natsRequests)
	prometheus.MustRegister(memcacheHits)
	prometheus.MustRegister(memcacheMiss)
}

// DynamoDBCollector is a Collector which can also CreateTables
type DynamoDBCollector interface {
	app.Collector
	CreateTables() error
}

type dynamoDBCollector struct {
	userIDer   UserIDer
	db         *dynamodb.DynamoDB
	s3         *s3.S3
	tableName  string
	bucketName string
	merger     app.Merger
	cache      gcache.Cache
	memcache   *memcache.Client

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
	memcachedTimeout time.Duration,
) (DynamoDBCollector, error) {
	var nc *nats.Conn
	if natsHost != "" {
		var err error
		nc, err = nats.Connect(natsHost)
		if err != nil {
			return nil, err
		}
	}

	var memcacheClient *memcache.Client
	if memcachedHost != "" {
		memcacheClient := memcache.New(memcachedHost)
		memcacheClient.Timeout = memcachedTimeout
	}

	return &dynamoDBCollector{
		db:         dynamodb.New(session.New(dynamoDBConfig)),
		s3:         s3.New(session.New(s3Config)),
		userIDer:   userIDer,
		tableName:  tableName,
		bucketName: bucketName,
		merger:     app.NewSmartMerger(),
		cache:      gcache.New(reportCacheSize).LRU().Expiration(reportCacheExpiration).Build(),
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

func (c *dynamoDBCollector) getCached(reportKeys []string) ([]report.Report, []string) {
	foundReports := []report.Report{}
	missingReports := []string{}
	for _, reportKey := range reportKeys {
		rpt, err := c.cache.Get(reportKey)
		if err == nil {
			foundReports = append(foundReports, rpt.(report.Report))
		} else {
			missingReports = append(missingReports, reportKey)
		}
	}
	return foundReports, missingReports
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
		c.cache.Set(r.key, *r.report)
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
	reportKeys, err := c.getReportKeys(rowKey, start, end)
	if err != nil {
		return nil, err
	}

	// TODO(jml): Refactor this so that we have a standard interface
	// IReportStore for "fetch :: [ReportKeys] -> ([Report], missing, error)"
	// and a list of providers: in-memory cache, memcache, dynamo. Possibly
	// also an "always everything missing" implementation to simplify logic.
	cachedReports, missing := c.getCached(reportKeys)
	dynamoCacheHits.Add(float64(len(cachedReports)))
	dynamoCacheMiss.Add(float64(len(missing)))
	if len(missing) == 0 {
		return cachedReports, nil
	}

	if c.memcache != nil {
		var memcachedReports []report.Report
		memcachedReports, missing, err = c.fetchFromMemcache(missing)
		memcacheHits.Add(float64(len(memcachedReports)))
		memcacheMiss.Add(float64(len(missing)))
		if err != nil {
			// XXX: jml is unclear whether we should abort in this case or
			// just carry on. Aborting is easier to reason about for us, but
			// suppressing is probably OK since failure just means we fetch
			// from S3.
			return nil, err
		}
		cachedReports = append(cachedReports, memcachedReports...)
	}
	if len(missing) == 0 {
		return cachedReports, nil
	}

	fetchedReport, err := c.getNonCached(missing)
	if err != nil {
		return nil, err
	}

	return append(cachedReports, fetchedReport...), nil
}

func (c *dynamoDBCollector) fetchFromMemcache(reportKeys []string) ([]report.Report, []string, error) {
	var reports []report.Report
	return reports, reportKeys, nil
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

	// third, put the key in dynamodb
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
