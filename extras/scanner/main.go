package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	awscommon "github.com/weaveworks/common/aws"
	"github.com/weaveworks/common/instrument"
	"golang.org/x/time/rate"
)

type scanner struct {
	startHour  int
	stopHour   int
	segments   int
	tableName  string
	bucketName string
	address    string

	writeLimiter *rate.Limiter
	queryLimiter *rate.Limiter

	dynamoDB *dynamodb.DynamoDB
	s3       *s3.S3
}

const (
	s3deleteBatchSize = 250

	// See http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Limits.html.
	dynamoDBMaxWriteBatchSize = 25
)

var (
	dynamoFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_failures_total",
		Help:      "The total number of errors from DynamoDB.",
	}, []string{"table", "error", "operation"})
	dynamoRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "dynamo_request_duration_seconds",
		Help:      "Time in seconds spent doing DynamoDB requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
	dynamoConsumedCapacity = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_consumed_capacity_total",
		Help:      "Total count of capacity units consumed per operation.",
	}, []string{"method"})
	s3RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "s3_request_duration_seconds",
		Help:      "Time in seconds spent doing S3 requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
	s3ItemsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "s3_items_deleted",
		Help:      "Total number of items deleted.",
	})
	recordsScanned = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "delete_records_scanned_total",
		Help:      "Total number of items deleted.",
	})
)

func main() {
	var (
		collectorURL string
		s3URL        string

		queryRateLimit float64
		writeRateLimit float64

		deleteOrgsFile string
		keepOrgsStr    string
		recordsFile    string

		scanner  scanner
		loglevel string

		justBigScan bool
		pagesPerDot int
	)

	flag.StringVar(&collectorURL, "app.collector", "local", "Collector to use (local, dynamodb, or file/directory)")
	flag.StringVar(&s3URL, "app.collector.s3", "local", "S3 URL to use (when collector is dynamodb)")
	flag.Float64Var(&queryRateLimit, "query-rate-limit", 10, "Max rate to query DynamoDB")
	flag.Float64Var(&writeRateLimit, "write-rate-limit", 10, "Rate-limit on throttling from DynamoDB")
	flag.IntVar(&scanner.startHour, "start-hour", 406848, "Hour number to start")
	flag.IntVar(&scanner.stopHour, "stop-hour", 0, "Hour number to stop (0 for current hour)")
	flag.IntVar(&scanner.segments, "segments", 1, "Number of segments to read in parallel")
	flag.StringVar(&scanner.address, "address", ":6060", "Address to listen on, for profiling, etc.")
	flag.StringVar(&deleteOrgsFile, "delete-orgs-file", "", "File containing IDs of orgs to delete")
	flag.StringVar(&keepOrgsStr, "keep-orgs", "", "IDs of orgs to keep - DELETE EVERYTHING ELSE (space-separated)")
	flag.StringVar(&recordsFile, "delete-records-file", "", "File containing IDs of orgs to delete")
	flag.StringVar(&loglevel, "log-level", "info", "Debug level: debug, info, warning, error")

	flag.BoolVar(&justBigScan, "big-scan", false, "If true, just scan the whole index and print summaries")
	flag.IntVar(&pagesPerDot, "pages-per-dot", 10, "Print a dot per N pages in DynamoDB (0 to disable)")

	flag.Parse()

	level, err := log.ParseLevel(loglevel)
	checkFatal(err)
	log.SetLevel(level)

	parsed, err := url.Parse(collectorURL)
	checkFatal(err)
	s3Address, err := url.Parse(s3URL)
	checkFatal(err)

	dynamoDBConfig, err := awscommon.ConfigFromURL(parsed)
	checkFatal(err)
	s3Config, err := awscommon.ConfigFromURL(s3Address)
	checkFatal(err)
	scanner.bucketName = strings.TrimPrefix(s3Address.Path, "/")
	scanner.tableName = strings.TrimPrefix(parsed.Path, "/")
	scanner.s3 = s3.New(session.New(s3Config))

	scanner.writeLimiter = rate.NewLimiter(rate.Limit(writeRateLimit), 25) // burst size should be the largest batch
	scanner.queryLimiter = rate.NewLimiter(rate.Limit(queryRateLimit), 1)  // we only do one query at a time

	// HTTP listener for profiling
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		checkFatal(http.ListenAndServe(scanner.address, nil))
	}()

	if justBigScan {
		bigScan(dynamoDBConfig, scanner.segments, pagesPerDot)
		return
	}
	if recordsFile == "" {
		checkFatal(fmt.Errorf("Must set one of -delete-records-file or -big-scan."))
	}

	deleteOrgs, keepOrgs := setupOrgs(deleteOrgsFile, keepOrgsStr)

	if scanner.stopHour == 0 {
		scanner.stopHour = int(time.Now().Unix() / int64(time.Hour/time.Second))
	}

	dynamoDBConfig = dynamoDBConfig.WithMaxRetries(0) // We do our own retries, with a rate-limiter
	session := session.New(dynamoDBConfig)
	scanner.dynamoDB = dynamodb.New(session)

	totals := newSummary()

	var recordsReader io.Reader
	fileReader, err := os.Open(recordsFile)
	checkFatal(err)
	defer fileReader.Close()
	if strings.HasSuffix(recordsFile, ".gz") {
		gzipReader, err := gzip.NewReader(fileReader)
		checkFatal(err)
		defer gzipReader.Close()
		recordsReader = gzipReader
	} else {
		recordsReader = fileReader
	}

	// Create multiple goroutines reading off one queue of records to delete
	queue := make(chan string)
	var wait sync.WaitGroup
	wait.Add(scanner.segments)
	for i := 0; i < scanner.segments; i++ {
		go func() {
			for record := range queue {
				scanner.HandleRecord(context.Background(), keepOrgs, deleteOrgs, record)
				recordsScanned.Inc()
			}
			wait.Done()
		}()
	}

	records := bufio.NewScanner(recordsReader)
	for records.Scan() {
		queue <- records.Text()
	}
	checkFatal(records.Err())
	close(queue)
	wait.Wait()

	fmt.Printf("\n")
	totals.print()
}

func setupOrgs(deleteOrgsFile, keepOrgsStr string) (deleteOrgs, keepOrgs map[int]struct{}) {
	if deleteOrgsFile != "" {
		deleteOrgs = map[int]struct{}{}
		content, err := ioutil.ReadFile(deleteOrgsFile)
		checkFatal(err)
		for _, arg := range strings.Fields(string(content)) {
			org, err := strconv.Atoi(arg)
			checkFatal(err)
			deleteOrgs[org] = struct{}{}
		}
	}

	if keepOrgsStr != "" {
		keepOrgs = map[int]struct{}{}
		for _, arg := range strings.Fields(keepOrgsStr) {
			org, err := strconv.Atoi(arg)
			checkFatal(err)
			keepOrgs[org] = struct{}{}
		}
	}
	return
}

func (sc *scanner) HandleRecord(ctx context.Context, keepOrgs, deleteOrgs map[int]struct{}, record string) {
	// Record is something like "1-406768, 4103", which is org-hour, reports
	fields := strings.Split(record, ", ")
	fields = strings.Split(fields[0], "-")
	org, err := strconv.Atoi(fields[0])
	checkFatal(err)
	if keepOrgs != nil {
		if _, found := keepOrgs[org]; found {
			return
		}
	}
	if deleteOrgs != nil {
		if _, found := deleteOrgs[org]; !found {
			return
		}
	}
	hour, err := strconv.Atoi(fields[1])
	checkFatal(err)
	if hour < sc.startHour || hour > sc.stopHour {
		return
	}
	sc.deleteOneOrgHour(ctx, fields[0], hour)
}

func (sc *scanner) processOrg(ctx context.Context, org string) {
	deleted := 0
	for hour := sc.startHour; hour <= sc.stopHour; hour++ {
		deleted += sc.deleteOneOrgHour(ctx, org, hour)
	}
	log.Infof("done %s: %d", org, deleted)
}

func (sc *scanner) deleteOneOrgHour(ctx context.Context, org string, hour int) int {
	var keys []map[string]*dynamodb.AttributeValue
	for {
		sc.queryLimiter.Wait(ctx)
		var err error
		keys, err = queryDynamo(ctx, sc.dynamoDB, sc.tableName, org, int64(hour))
		if throttled(err) {
			continue
		}
		checkFatal(err)
		break
	}
	var wait sync.WaitGroup
	if len(keys) > 0 {
		log.Debugf("deleting org: %s hour: %d num: %d", org, hour, len(keys))
	}
	for start := 0; start < len(keys); start += s3deleteBatchSize {
		end := start + s3deleteBatchSize
		if end > len(keys) {
			end = len(keys)
		}
		wait.Add(1)
		func(batchKeys []map[string]*dynamodb.AttributeValue) {
			sc.deleteFromS3(ctx, batchKeys)
			for _, key := range batchKeys {
				delete(key, reportField) // not part of key in dynamoDB
			}
			sc.deleteFromDynamoDB(batchKeys)
			s3ItemsDeleted.Add(float64(len(batchKeys)))
			wait.Done()
		}(keys[start:end])
	}
	wait.Wait()
	return len(keys)
}

func (sc *scanner) deleteFromS3(ctx context.Context, keys []map[string]*dynamodb.AttributeValue) {
	// Build multiple-object delete request for S3
	d := &s3.Delete{}
	for _, key := range keys {
		reportKey := key[reportField].S
		d.Objects = append(d.Objects, &s3.ObjectIdentifier{Key: reportKey})
	}
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(sc.bucketName),
		Delete: d,
	}
	// Send batch to S3
	err := instrument.TimeRequestHistogram(ctx, "S3.Delete", s3RequestDuration, func(_ context.Context) error {
		_, err := sc.s3.DeleteObjectsWithContext(ctx, input)
		return err
	})
	if err != nil {
		log.Errorf("S3 delete: err %s", err)
	}
}

func queryDynamo(ctx context.Context, db *dynamodb.DynamoDB, tableName, userid string, row int64) ([]map[string]*dynamodb.AttributeValue, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	var result []map[string]*dynamodb.AttributeValue

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			hourField: {
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(rowKey)},
				},
				ComparisonOperator: aws.String("EQ"),
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
	p := request.Pagination{
		NewRequest: func() (*request.Request, error) {
			req, _ := db.QueryRequest(queryInput)
			req.SetContext(ctx)
			return req, nil
		},
	}

	for {
		var haveData bool
		instrument.TimeRequestHistogram(ctx, "DynamoDB.Query", dynamoRequestDuration, func(_ context.Context) error {
			haveData = p.Next()
			return nil
		})
		if !haveData {
			if p.Err() != nil {
				recordDynamoError(tableName, p.Err(), "DynamoDB.Query")
				return nil, p.Err()
			}
			break
		}
		resp := p.Page().(*dynamodb.QueryOutput)
		if resp.ConsumedCapacity != nil {
			dynamoConsumedCapacity.WithLabelValues("Query").
				Add(float64(*resp.ConsumedCapacity.CapacityUnits))
		}
		result = append(result, resp.Items...)
	}
	return result, nil
}

const (
	hourField   = "hour"
	tsField     = "ts"
	reportField = "report"

	hashKey  = "h"
	rangeKey = "r"
	valueKey = "c"
)

type summary struct {
	counts map[int]int
}

func newSummary() summary {
	return summary{
		counts: map[int]int{},
	}
}

func (s *summary) accumulate(b summary) {
	for k, v := range b.counts {
		s.counts[k] += v
	}
}

func (s summary) print() {
	for user, count := range s.counts {
		fmt.Printf("%d, %d\n", user, count)
	}
}

func checkFatal(err error) {
	if err != nil {
		log.Errorf("fatal error: %s", err)
		os.Exit(1)
	}
}

func throttled(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && (awsErr.Code() == dynamodb.ErrCodeProvisionedThroughputExceededException)
}

// input is map from table to attribute-value
func (sc *scanner) deleteFromDynamoDB(batch []map[string]*dynamodb.AttributeValue) {
	var requests []*dynamodb.WriteRequest

	for _, keyMap := range batch {
		requests = append(requests, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: keyMap,
			},
		})
	}
	log.Debug("about to delete ", len(batch))
	var ret *dynamodb.BatchWriteItemOutput
	var err error
	for len(requests) > 0 {
		numToSend := len(requests)
		if numToSend > dynamoDBMaxWriteBatchSize {
			numToSend = dynamoDBMaxWriteBatchSize
		}
		instrument.TimeRequestHistogram(context.Background(), "DynamoDB.Delete", dynamoRequestDuration, func(_ context.Context) error {
			ret, err = sc.dynamoDB.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					sc.tableName: requests[:numToSend],
				},
				ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
			})
			return err
		})
		for _, cc := range ret.ConsumedCapacity {
			dynamoConsumedCapacity.WithLabelValues("BatchWriteItem").
				Add(float64(*cc.CapacityUnits))
		}
		if err != nil {
			recordDynamoError(sc.tableName, err, "DynamoDB.Delete")
			if throttled(err) {
				sc.writeLimiter.WaitN(context.Background(), numToSend)
				// Back round the loop without taking anything away from the batch
				continue
			} else {
				log.Error("msg", "unable to delete", "err", err)
				// drop this batch
			}
		}
		requests = requests[numToSend:]
		// Add unprocessed items onto the end of requests
		for _, v := range ret.UnprocessedItems {
			sc.writeLimiter.WaitN(context.Background(), len(v))
			requests = append(requests, v...)
		}
	}
}

func recordDynamoError(tableName string, err error, operation string) {
	if awsErr, ok := err.(awserr.Error); ok {
		dynamoFailures.WithLabelValues(tableName, awsErr.Code(), operation).Add(float64(1))
	} else {
		dynamoFailures.WithLabelValues(tableName, "other", operation).Add(float64(1))
	}
}
