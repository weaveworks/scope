package multitenant

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bluele/gcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField       = "hour"
	tsField         = "ts"
	reportField     = "report"
	cacheSize       = (15 / 3) * 10 * 5 // (window size * report rate) * number of hosts per user * number of users
	cacheExpiration = 15 * time.Second
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
)

func init() {
	prometheus.MustRegister(dynamoRequestDuration)
	prometheus.MustRegister(dynamoCacheHits)
	prometheus.MustRegister(dynamoCacheMiss)
	prometheus.MustRegister(dynamoConsumedCapacity)
	prometheus.MustRegister(dynamoValueSize)
	prometheus.MustRegister(reportSize)
	prometheus.MustRegister(s3RequestDuration)
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
}

// NewDynamoDBCollector the reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewDynamoDBCollector(dynamoDBConfig, s3Config *aws.Config, userIDer UserIDer, tableName, bucketName string) DynamoDBCollector {
	return &dynamoDBCollector{
		db:         dynamodb.New(session.New(dynamoDBConfig)),
		s3:         s3.New(session.New(s3Config)),
		userIDer:   userIDer,
		tableName:  tableName,
		bucketName: bucketName,
		merger:     app.NewSmartMerger(),
		cache:      gcache.New(cacheSize).LRU().Expiration(cacheExpiration).Build(),
	}
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

func errorCode(err error) string {
	if err == nil {
		return "200"
	}
	return "500"
}

func timeRequest(method string, metric *prometheus.SummaryVec, f func() error) error {
	startTime := time.Now()
	err := f()
	duration := time.Now().Sub(startTime)
	metric.WithLabelValues(method, errorCode(err)).Observe(float64(duration.Nanoseconds()))
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

func (c *dynamoDBCollector) getNonCached(reportKeys []string) ([]report.Report, error) {
	reports := []report.Report{}
	for _, reportKey := range reportKeys {
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
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Errorf("Error gunzipping report: %v", err)
			continue
		}
		rep := report.MakeReport()
		if err := codec.NewDecoder(reader, &codec.MsgpackHandle{}).Decode(&rep); err != nil {
			log.Errorf("Failed to decode report: %v", err)
			continue
		}
		reports = append(reports, rep)
		c.cache.Set(reportKey, rep)
	}
	return reports, nil
}

func (c *dynamoDBCollector) getReports(userid string, row int64, start, end time.Time) ([]report.Report, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	reportKeys, err := c.getReportKeys(rowKey, start, end)
	if err != nil {
		return nil, err
	}

	cachedReports, missing := c.getCached(reportKeys)
	dynamoCacheHits.Add(float64(len(cachedReports)))
	dynamoCacheMiss.Add(float64(len(missing)))
	if len(missing) == 0 {
		return cachedReports, nil
	}

	fetchedReport, err := c.getNonCached(missing)
	if err != nil {
		return nil, err
	}

	return append(cachedReports, fetchedReport...), nil
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
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return err
	}
	if err := codec.NewEncoder(writer, &codec.MsgpackHandle{}).Encode(&rep); err != nil {
		return err
	}
	writer.Close()
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
	return err
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}
