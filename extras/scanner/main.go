package main

import (
	"context"
	"flag"
	"fmt"
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
	segments        int
	deleters        int
	deleteBatchSize int
	tableName       string
	bucketName      string
	address         string

	writeLimiter *rate.Limiter
	queryLimiter *rate.Limiter

	dynamoDB *dynamodb.DynamoDB
	s3       *s3.S3

	// Readers send items on this chan to be deleted
	delete chan map[string]*dynamodb.AttributeValue
	retry  chan map[string]*dynamodb.AttributeValue
	// Deleters read batches of items from this chan
	batched chan []*dynamodb.WriteRequest
}

const (
	s3deleteBatchSize = 250
)

var (
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
)

func main() {
	var (
		collectorURL string
		s3URL        string

		queryRateLimit float64
		writeRateLimit float64

		orgsFile  string
		startHour int
		stopHour  int

		scanner  scanner
		loglevel string
	)

	flag.StringVar(&collectorURL, "app.collector", "local", "Collector to use (local, dynamodb, or file/directory)")
	flag.StringVar(&s3URL, "app.collector.s3", "local", "S3 URL to use (when collector is dynamodb)")
	flag.Float64Var(&queryRateLimit, "query-rate-limit", 100, "Max rate to query DynamoDB")
	flag.Float64Var(&writeRateLimit, "write-rate-limit", 100, "Rate-limit on throttling from DynamoDB")
	flag.IntVar(&startHour, "start-hour", 406848, "Hour number to start")
	flag.IntVar(&stopHour, "stop-hour", 0, "Hour number to stop (0 for current hour)")
	flag.IntVar(&scanner.segments, "segments", 1, "Number of segments to read in parallel")
	flag.IntVar(&scanner.deleters, "deleters", 1, "Number of deleters to run in parallel")
	flag.IntVar(&scanner.deleteBatchSize, "delete-batch-size", 25, "Number of delete requests to batch up")
	flag.StringVar(&scanner.address, "address", ":6060", "Address to listen on, for profiling, etc.")
	flag.StringVar(&orgsFile, "delete-orgs-file", "", "File containing IDs of orgs to delete")
	flag.StringVar(&loglevel, "log-level", "info", "Debug level: debug, info, warning, error")

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

	dynamoDBConfig = dynamoDBConfig.WithMaxRetries(0) // We do our own retries, with a rate-limiter
	session := session.New(dynamoDBConfig)
	scanner.dynamoDB = dynamodb.New(session)

	// Unbuffered chan so we can tell when batcher has received all items
	scanner.delete = make(chan map[string]*dynamodb.AttributeValue)
	scanner.retry = make(chan map[string]*dynamodb.AttributeValue, 100)
	scanner.batched = make(chan []*dynamodb.WriteRequest)

	var deleteGroup sync.WaitGroup
	deleteGroup.Add(1 + scanner.deleters)
	var pending sync.WaitGroup
	go func() {
		scanner.batcher(&pending)
		deleteGroup.Done()
	}()
	for i := 0; i < scanner.deleters; i++ {
		go func() {
			scanner.deleteLoop(&pending)
			deleteGroup.Done()
		}()
	}

	totals := newSummary()

	if orgsFile != "" {
		scanner.processOrgsFile(context.Background(), orgsFile, startHour, stopHour)
	}

	// Ensure that batcher has received all items so it won't call Add() any more
	scanner.delete <- nil
	// Wait for pending items to be sent to DynamoDB
	pending.Wait()
	// Close chans to signal deleter(s) and batcher to terminate
	close(scanner.batched)
	close(scanner.retry)
	deleteGroup.Wait()

	fmt.Printf("\n")
	totals.print()
}

func (sc *scanner) processOrgsFile(ctx context.Context, orgsFile string, startHour, stopHour int) {
	if stopHour == 0 {
		stopHour = int(time.Now().Unix() / int64(time.Hour/time.Second))
	}

	content, err := ioutil.ReadFile(orgsFile)
	checkFatal(err)
	orgs := strings.Fields(string(content))

	var orgWait sync.WaitGroup
	orgWait.Add(len(orgs))

	for _, org := range orgs {
		go func(org string) {
			sc.processOrg(ctx, org, startHour, stopHour)
			orgWait.Done()
		}(org)
	}
	orgWait.Wait()
}

func (sc *scanner) processOrg(ctx context.Context, org string, startHour, stopHour int) {
	deleted := 0
	for hour := startHour; hour <= stopHour; hour++ {
		keys := sc.getKeys(ctx, org, hour)
		if len(keys) > 0 {
			log.Debugf("deleting org: %s hour: %d num: %d", org, hour, len(keys))
		}
		sc.parallelDelete(ctx, keys)
		deleted += len(keys)
	}
	log.Infof("done %s: %d", org, deleted)
}

func (sc *scanner) getKeys(ctx context.Context, org string, hour int) []map[string]*dynamodb.AttributeValue {
	for {
		sc.queryLimiter.Wait(ctx)
		keys, err := queryDynamo(ctx, sc.dynamoDB, sc.tableName, org, int64(hour))
		if throttled(err) {
			continue
		}
		checkFatal(err)
		return keys
	}
}

func (sc *scanner) parallelDelete(ctx context.Context, keys []map[string]*dynamodb.AttributeValue) {
	var wait sync.WaitGroup
	for start := 0; start < len(keys); start += s3deleteBatchSize {
		end := start + s3deleteBatchSize
		if end > len(keys) {
			end = len(keys)
		}
		wait.Add(1)
		go func(start, end int) {
			sc.deleteFromS3AndDynamoDB(ctx, keys[start:end])
			wait.Done()
		}(start, end)
	}
	wait.Wait()
	s3ItemsDeleted.Add(float64(len(keys)))
}

func (sc *scanner) deleteFromS3AndDynamoDB(ctx context.Context, keys []map[string]*dynamodb.AttributeValue) {
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
	// Now send to be deleted from DynamoDB
	for _, key := range keys {
		delete(key, reportField) // not part of key in dynamoDB
		sc.delete <- key
	}
}

func queryDynamo(ctx context.Context, db *dynamodb.DynamoDB, tableName, userid string, row int64) ([]map[string]*dynamodb.AttributeValue, error) {
	rowKey := fmt.Sprintf("%s-%s", userid, strconv.FormatInt(row, 10))
	var resp *dynamodb.QueryOutput
	err := instrument.TimeRequestHistogram(ctx, "DynamoDB.Query", dynamoRequestDuration, func(_ context.Context) error {
		var err error
		resp, err = db.Query(&dynamodb.QueryInput{
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
	return resp.Items, nil
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

func (sc *scanner) deleteLoop(pending *sync.WaitGroup) {
	for {
		batch, ok := <-sc.batched
		if !ok {
			return
		}
		log.Debug("about to delete", len(batch))
		var ret *dynamodb.BatchWriteItemOutput
		var err error
		instrument.TimeRequestHistogram(context.Background(), "DynamoDB.Delete", dynamoRequestDuration, func(_ context.Context) error {
			ret, err = sc.dynamoDB.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					sc.tableName: batch,
				},
				ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
			})
			return err
		})
		if ret.ConsumedCapacity != nil {
			for _, cc := range ret.ConsumedCapacity {
				dynamoConsumedCapacity.WithLabelValues("BatchWriteItem").
					Add(float64(*cc.CapacityUnits))
			}
		}
		if err != nil {
			if throttled(err) {
				sc.writeLimiter.WaitN(context.Background(), len(batch))
				// Send the whole request back into the batcher
				for _, item := range batch {
					sc.retry <- item.DeleteRequest.Key
				}
			} else {
				log.Error("msg", "unable to delete", "err", err)
				pending.Add(-len(batch))
			}
			continue
		}
		count := 0
		if len(ret.UnprocessedItems) > 0 {
			sc.writeLimiter.WaitN(context.Background(), len(ret.UnprocessedItems))
		}
		// Send unprocessed items back into the batcher
		for _, items := range ret.UnprocessedItems {
			count += len(items)
			for _, item := range items {
				sc.retry <- item.DeleteRequest.Key
			}
		}
		pending.Add(-(len(batch) - count))
	}
}

// Receive individual requests, and batch them up into groups to send to DynamoDB
func (sc *scanner) batcher(pending *sync.WaitGroup) {
	finished := false
	var requests []*dynamodb.WriteRequest
	for {
		// We will allow in new data if the queue isn't too long
		var in chan map[string]*dynamodb.AttributeValue
		if len(requests) < 1000 {
			in = sc.delete
		}
		// We will send out a batch if the queue is big enough, or if we're finishing
		var out chan []*dynamodb.WriteRequest
		outlen := len(requests)
		if len(requests) >= sc.deleteBatchSize {
			out = sc.batched
			outlen = sc.deleteBatchSize
		} else if finished && len(requests) > 0 {
			out = sc.batched
		}
		var keyMap map[string]*dynamodb.AttributeValue
		var ok bool
		select {
		case keyMap = <-in:
			if keyMap == nil { // Nil used as interlock to know we received all previous values
				finished = true
			} else {
				pending.Add(1)
			}
		case keyMap, ok = <-sc.retry:
			if !ok {
				return
			}
		case out <- requests[:outlen]:
			requests = requests[outlen:]
		}
		if keyMap != nil {
			requests = append(requests, &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: keyMap,
				},
			})
		}
	}
}
