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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"
)

var (
	dynamoRequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "dynamo_request_duration_nanoseconds",
		Help:      "Time spent doing DynamoDB requests.",
	}, []string{"method", "status_code"})
)

func init() {
	prometheus.MustRegister(dynamoRequestDuration)
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
}

// NewDynamoDBCollector the reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewDynamoDBCollector(config *aws.Config, userIDer UserIDer, tableName string) DynamoDBCollector {
	return &dynamoDBCollector{
		db:        dynamodb.New(session.New(config)),
		userIDer:  userIDer,
		tableName: tableName,
		merger:    app.NewSmartMerger(),
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

func (c *dynamoDBCollector) getRows(userid string, row int64, start, end time.Time) ([]report.Report, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
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
	})
	duration := time.Now().Sub(startTime)
	if err != nil {
		dynamoRequestDuration.WithLabelValues("Query", "500").Observe(float64(duration.Nanoseconds()))
		return nil, err
	}
	dynamoRequestDuration.WithLabelValues("Query", "200").Observe(float64(duration.Nanoseconds()))
	reports := []report.Report{}
	for _, item := range resp.Items {
		b := item[reportField].B
		if b == nil {
			log.Errorf("Empty row!")
			continue
		}
		buf := bytes.NewBuffer(b)
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
	}
	return reports, nil
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
		reports1, err := c.getRows(userid, rowStart, start, now)
		if err != nil {
			return report.MakeReport(), err
		}

		reports2, err := c.getRows(userid, rowEnd, start, now)
		if err != nil {
			return report.MakeReport(), err
		}

		reports = append(reports1, reports2...)
	} else {
		if reports, err = c.getRows(userid, rowEnd, start, now); err != nil {
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

	now := time.Now()
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(now.UnixNano()/time.Hour.Nanoseconds(), 10))
	startTime := time.Now()
	_, err = c.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			hourField: {
				S: aws.String(rowKey),
			},
			tsField: {
				N: aws.String(strconv.FormatInt(now.UnixNano(), 10)),
			},
			reportField: {
				B: buf.Bytes(),
			},
		},
	})
	duration := time.Now().Sub(startTime)
	if err != nil {
		dynamoRequestDuration.WithLabelValues("PutItem", "500").Observe(float64(duration.Nanoseconds()))
		return err
	}
	dynamoRequestDuration.WithLabelValues("PutItem", "200").Observe(float64(duration.Nanoseconds()))
	return nil
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}
