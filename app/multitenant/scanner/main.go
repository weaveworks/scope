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
)

type scanner struct {
	startHour       int
	stopHour        int
	segments        int
	deleters        int
	deleteBatchSize int
	tableName       string
	address         string

	dynamoDB *dynamodb.DynamoDB
	s3       *s3.S3

	// Readers send items on this chan to be deleted
	delete chan map[string]*dynamodb.AttributeValue
	retry  chan map[string]*dynamodb.AttributeValue
	// Deleters read batches of items from this chan
	batched chan []*dynamodb.WriteRequest
	// Keys to delete in s3
	s3chan chan *s3.DeleteObjectInput
}

var (
	pagesPerDot int
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
	dynamoValueSize = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "scope",
		Name:      "dynamo_value_size_bytes_total",
		Help:      "Total size of data read / written from DynamoDB in bytes.",
	}, []string{"method"})
	s3RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "s3_request_duration_seconds",
		Help:      "Time in seconds spent doing S3 requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
)

func main() {
	var (
		collectorURL string
		s3URL        string

		orgsFile string

		scanner  scanner
		loglevel string
	)

	flag.StringVar(&collectorURL, "app.collector", "local", "Collector to use (local, dynamodb, or file/directory)")
	flag.StringVar(&s3URL, "app.collector.s3", "local", "S3 URL to use (when collector is dynamodb)")
	flag.IntVar(&scanner.startHour, "start-hour", 406848, "Hour number to start")
	flag.IntVar(&scanner.stopHour, "stop-hour", 406848, "Hour number to stop (0 for current hour)")
	flag.IntVar(&scanner.segments, "segments", 1, "Number of segments to read in parallel")
	flag.IntVar(&scanner.deleters, "deleters", 1, "Number of deleters to run in parallel")
	flag.IntVar(&scanner.deleteBatchSize, "delete-batch-size", 25, "Number of delete requests to batch up")
	flag.StringVar(&scanner.address, "address", "localhost:6060", "Address to listen on, for profiling, etc.")
	flag.StringVar(&orgsFile, "delete-orgs-file", "", "File containing IDs of orgs to delete")
	flag.StringVar(&loglevel, "log-level", "info", "Debug level: debug, info, warning, error")
	flag.IntVar(&pagesPerDot, "pages-per-dot", 10, "Print a dot per N pages in DynamoDB (0 to disable)")

	flag.Parse()

	parsed, err := url.Parse(collectorURL)
	checkFatal(err)
	s3Address, err := url.Parse(s3URL)
	checkFatal(err)

	dynamoDBConfig, err := awscommon.ConfigFromURL(parsed)
	checkFatal(err)
	s3Config, err := awscommon.ConfigFromURL(s3Address)
	checkFatal(err)
	bucketName := strings.TrimPrefix(s3Address.Path, "/")
	tableName := strings.TrimPrefix(parsed.Path, "/")
	scanner.s3 = s3.New(session.New(s3Config))

	// HTTP listener for profiling
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		checkFatal(http.ListenAndServe(scanner.address, nil))
	}()

	orgs := []string{}
	if orgsFile != "" {
		content, err := ioutil.ReadFile(orgsFile)
		checkFatal(err)
		orgs = strings.Fields(string(content))
	}

	if scanner.stopHour == 0 {
		scanner.stopHour = int(time.Now().Unix() / int64(time.Hour/time.Second))
	}

	session := session.New(dynamoDBConfig)
	scanner.dynamoDB = dynamodb.New(session)

	// Unbuffered chan so we can tell when batcher has received all items
	scanner.delete = make(chan map[string]*dynamodb.AttributeValue)
	scanner.retry = make(chan map[string]*dynamodb.AttributeValue, 100)
	scanner.batched = make(chan []*dynamodb.WriteRequest)
	scanner.s3chan = make(chan *s3.DeleteObjectInput)

	var deleteGroup sync.WaitGroup
	deleteGroup.Add(1 + scanner.deleters*2)
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
		go func() {
			scanner.deleteLoopS3(&pending)
			deleteGroup.Done()
		}()
	}

	totals := newSummary()
	ctx := context.Background()

	for _, org := range orgs {
		for hour := scanner.startHour; hour <= scanner.stopHour; hour++ {
			keys, err := reportKeysInRow(ctx, scanner.dynamoDB, tableName, org, int64(hour))
			checkFatal(err)
			fmt.Printf("%s: %d Keys: %d\n", org, hour, len(keys))
			for _, key := range keys {
				input := &s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(key),
				}
				scanner.s3chan <- input
			}
		}
	}

	// Ensure that batcher has received all items so it won't call Add() any more
	scanner.delete <- nil
	// Wait for pending items to be sent to DynamoDB
	pending.Wait()
	// Close chans to signal deleter(s) and batcher to terminate
	close(scanner.batched)
	close(scanner.retry)
	close(scanner.s3chan)
	deleteGroup.Wait()

	fmt.Printf("\n")
	totals.print()
}

func reportKeysInRow(ctx context.Context, db *dynamodb.DynamoDB, tableName, userid string, row int64) ([]string, error) {
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
		log.Error("msg", "fatal error", "err", err)
		os.Exit(1)
	}
}

func throttled(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && (awsErr.Code() == dynamodb.ErrCodeProvisionedThroughputExceededException)
}

func (sc *scanner) deleteLoopS3(pending *sync.WaitGroup) {
	ctx := context.Background()
	for {
		item, ok := <-sc.s3chan
		if !ok {
			return
		}
		log.Debug("msg", "S3 delete", "key", aws.StringValue(item.Key))
		err := instrument.TimeRequestHistogram(ctx, "S3.Delete", s3RequestDuration, func(_ context.Context) error {
			_, err := sc.s3.DeleteObjectWithContext(ctx, item)
			return err
		})
		checkFatal(err)
	}
}

func (sc *scanner) deleteLoop(pending *sync.WaitGroup) {
	for {
		batch, ok := <-sc.batched
		if !ok {
			return
		}
		log.Debug("msg", "about to delete", "num_requests", len(batch))
		ret, err := sc.dynamoDB.BatchWriteItem(&dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				sc.tableName: batch,
			},
		})
		if err != nil {
			if throttled(err) {
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
