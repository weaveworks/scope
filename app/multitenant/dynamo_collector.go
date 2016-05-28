package multitenant

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
	dynamoReportSize = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_report_size_bytes",
		Help:      "Size of reports read / written from dynamodb.",
	}, []string{"method"})
)

func init() {
	prometheus.MustRegister(dynamoRequestDuration)
	prometheus.MustRegister(dynamoCacheHits)
	prometheus.MustRegister(dynamoCacheMiss)
	prometheus.MustRegister(dynamoConsumedCapacity)
	prometheus.MustRegister(dynamoReportSize)
}

// DynamoDBCollector is a Collector which can also CreateTables
type DynamoDBCollector interface {
	app.Collector
	CreateTables() error
}

type dynamoDBCollector struct {
	userIDer  UserIDer
	db        *dynamodb.DynamoDB
	tableName string
	merger    app.Merger
	cache     gcache.Cache
}

// NewDynamoDBCollector the reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewDynamoDBCollector(config *aws.Config, userIDer UserIDer, tableName string) DynamoDBCollector {
	return &dynamoDBCollector{
		db:        dynamodb.New(session.New(config)),
		userIDer:  userIDer,
		tableName: tableName,
		merger:    app.NewSmartMerger(),
		cache:     gcache.New(cacheSize).LRU().Expiration(cacheExpiration).Build(),
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
			//	AttributeType: aws.String("B"),
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

// getReportTimestamps gets the column keys for reports in this range
func (c *dynamoDBCollector) getReportTimestamps(rowKey string, start, end time.Time) ([]string, error) {
	startTime := time.Now()
	resp, err := c.db.Query(&dynamodb.QueryInput{
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
		ProjectionExpression:   aws.String(tsField),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	})
	duration := time.Now().Sub(startTime)
	dynamoRequestDuration.WithLabelValues("Query", errorCode(err)).
		Observe(float64(duration.Nanoseconds()))
	if resp.ConsumedCapacity != nil {
		dynamoConsumedCapacity.WithLabelValues("Query").
			Add(float64(*resp.ConsumedCapacity.CapacityUnits))
	}
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, item := range resp.Items {
		ts := item[tsField].N
		if ts == nil {
			log.Errorf("Empty row!")
			continue
		}
		result = append(result, *ts)
	}
	return result, nil
}

func (c *dynamoDBCollector) getCached(rowKey string, columnKeys []string) ([]report.Report, []string) {
	foundReports := []report.Report{}
	missingReports := []string{}
	for _, columnKey := range columnKeys {
		rpt, err := c.cache.Get(struct{ row, column string }{rowKey, columnKey})
		if err == nil {
			foundReports = append(foundReports, rpt.(report.Report))
		} else {
			missingReports = append(missingReports, columnKey)
		}
	}
	return foundReports, missingReports
}

func (c *dynamoDBCollector) getNonCached(rowKey string, columnKeys []string) ([]report.Report, error) {
	keys := []map[string]*dynamodb.AttributeValue{}
	for _, columnKey := range columnKeys {
		keys = append(keys, map[string]*dynamodb.AttributeValue{
			hourField: {S: aws.String(rowKey)},
			tsField:   {N: aws.String(columnKey)},
		})
	}

	startTime := time.Now()
	resp, err := c.db.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			c.tableName: {
				Keys: keys,
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	})
	duration := time.Now().Sub(startTime)
	dynamoRequestDuration.WithLabelValues("BatchGetItem", errorCode(err)).
		Observe(float64(duration.Nanoseconds()))
	for _, capacity := range resp.ConsumedCapacity {
		dynamoConsumedCapacity.WithLabelValues("BatchGetItem").
			Add(float64(*capacity.CapacityUnits))
	}
	if err != nil {
		return nil, err
	}

	reports := []report.Report{}
	for _, table := range resp.Responses {
		for _, entry := range table {
			columnKey, ok := entry[tsField]
			if !ok {
				log.Errorf("Entry doesn't contain columnKey!")
				continue
			}

			b, ok := entry[reportField]
			if !ok {
				log.Errorf("Entry doesn't contain report!")
				continue
			}

			dynamoReportSize.WithLabelValues("BatchGetItem").
				Add(float64(len(b.B)))
			buf := bytes.NewBuffer(b.B)
			reader, err := gzip.NewReader(buf)
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

			c.cache.Set(struct{ row, column string }{rowKey, *columnKey.N}, rep)
		}
	}
	return reports, nil
}

func (c *dynamoDBCollector) getReports(userid string, row int64, start, end time.Time) ([]report.Report, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	cols, err := c.getReportTimestamps(rowKey, start, end)
	if err != nil {
		return nil, err
	}

	cachedReports, missing := c.getCached(rowKey, cols)
	dynamoCacheHits.Add(float64(len(cachedReports)))
	dynamoCacheMiss.Add(float64(len(missing)))
	if len(missing) == 0 {
		return cachedReports, nil
	}

	fetchedReport, err := c.getNonCached(rowKey, missing)
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

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if err := codec.NewEncoder(writer, &codec.MsgpackHandle{}).Encode(&rep); err != nil {
		return err
	}
	writer.Close()
	bytes := buf.Bytes()
	dynamoReportSize.WithLabelValues("PutItem").Add(float64(len(bytes)))

	now := time.Now()
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(now.UnixNano()/time.Hour.Nanoseconds(), 10))
	startTime := time.Now()
	resp, err := c.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			hourField: {
				S: aws.String(rowKey),
			},
			tsField: {
				N: aws.String(strconv.FormatInt(now.UnixNano(), 10)),
			},
			reportField: {
				B: bytes,
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	})
	duration := time.Now().Sub(startTime)
	dynamoRequestDuration.WithLabelValues("PutItem", errorCode(err)).
		Observe(float64(duration.Nanoseconds()))
	if resp.ConsumedCapacity != nil {
		dynamoConsumedCapacity.WithLabelValues("PutItem").
			Add(float64(*resp.ConsumedCapacity.CapacityUnits))
	}
	return err
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}
