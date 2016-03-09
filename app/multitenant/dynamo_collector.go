package multitenant

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
)

const (
	tableName   = "reports"
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"
)

// DynamoDBCollector is a Collector which can also CreateTables
type DynamoDBCollector interface {
	app.Collector
	CreateTables() error
}

type dynamoDBCollector struct {
	userIDer UserIDer
	db       *dynamodb.DynamoDB
}

// NewDynamoDBCollector the reaper of souls
// https://github.com/aws/aws-sdk-go/wiki/common-examples
func NewDynamoDBCollector(url, region string, creds *credentials.Credentials, userIDer UserIDer) DynamoDBCollector {
	return &dynamoDBCollector{
		db: dynamodb.New(session.New(aws.NewConfig().
			WithEndpoint(url).
			WithRegion(region).
			WithCredentials(creds))),
		userIDer: userIDer,
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
		if *s == tableName {
			return nil
		}
	}

	params := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(hourField),
				AttributeType: aws.String("N"),
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
	log.Infof("Creating table %s", tableName)
	_, err = c.db.CreateTable(params)
	return err
}

func (c *dynamoDBCollector) getRows(userid string, row int64, start, end time.Time, input report.Report) (report.Report, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	resp, err := c.db.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			hourField: {
				AttributeValueList: []*dynamodb.AttributeValue{
					{N: aws.String(rowKey)},
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
	if err != nil {
		return report.MakeReport(), err
	}
	result := input
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
		result = result.Merge(rep)
	}
	return result, nil
}

func (c *dynamoDBCollector) Report(ctx context.Context) (report.Report, error) {
	var (
		now              = time.Now()
		start            = now.Add(-15 * time.Second)
		rowStart, rowEnd = start.UnixNano() / time.Hour.Nanoseconds(), now.UnixNano() / time.Hour.Nanoseconds()
		result           = report.MakeReport()
		userid, err      = c.userIDer(ctx)
	)
	if err != nil {
		return report.MakeReport(), err
	}

	// Queries will only every span 2 rows max.
	if rowStart != rowEnd {
		if result, err = c.getRows(userid, rowStart, start, now, result); err != nil {
			return report.MakeReport(), nil
		}
	}
	if result, err = c.getRows(userid, rowEnd, start, now, result); err != nil {
		return report.MakeReport(), nil
	}
	return result, nil
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
	_, err = c.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			hourField: {
				N: aws.String(rowKey),
			},
			tsField: {
				N: aws.String(strconv.FormatInt(now.UnixNano(), 10)),
			},
			reportField: {
				B: buf.Bytes(),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *dynamoDBCollector) WaitOn(context.Context, chan struct{}) {}

func (c *dynamoDBCollector) UnWait(context.Context, chan struct{}) {}
