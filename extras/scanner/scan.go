package main

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type bigScanner struct {
	segments  int
	tableName string

	dynamoDB *dynamodb.DynamoDB
}

var (
	pagesPerDot int
)

func bigScan(config *aws.Config, segments, ppd int) {
	var (
		scanner bigScanner
	)

	scanner.segments = segments
	pagesPerDot = ppd // hack!

	session := session.New(config)
	scanner.dynamoDB = dynamodb.New(session)

	var group sync.WaitGroup
	group.Add(scanner.segments)
	totals := newBigSummary()
	var totalsMutex sync.Mutex

	for segment := 0; segment < scanner.segments; segment++ {
		go func(segment int) {
			handler := newHandler()
			err := scanner.segmentScan(segment, handler)
			checkFatal(err)
			totalsMutex.Lock()
			totals.accumulate(handler.summary)
			totalsMutex.Unlock()
			group.Done()
		}(segment)
	}
	group.Wait()
	fmt.Printf("\n")
	totals.print()
}

func (sc bigScanner) segmentScan(segment int, handler handler) error {
	input := &dynamodb.ScanInput{
		TableName:            aws.String(sc.tableName),
		ProjectionExpression: aws.String("#h"),
		// Need to do this because "hour" is a reserved word
		ExpressionAttributeNames: map[string]*string{"#h": aws.String(hourField)},
		Segment:                  aws.Int64(int64(segment)),
		TotalSegments:            aws.Int64(int64(sc.segments)),
		//ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}

	return sc.dynamoDB.ScanPages(input, handler.handlePage)
}

type bigSummary struct {
	counts map[string]int
}

func newBigSummary() bigSummary {
	return bigSummary{
		counts: map[string]int{},
	}
}

func (s *bigSummary) accumulate(b bigSummary) {
	for k, v := range b.counts {
		s.counts[k] += v
	}
}

func (s bigSummary) print() {
	for user, count := range s.counts {
		fmt.Printf("%s, %d\n", user, count)
	}
}

type handler struct {
	pages   int
	summary bigSummary
}

func newHandler() handler {
	return handler{
		summary: newBigSummary(),
	}
}

func (h *handler) reset() {
	h.summary.counts = map[string]int{}
}

func (h *handler) handlePage(page *dynamodb.ScanOutput, lastPage bool) bool {
	h.pages++
	if pagesPerDot > 0 && h.pages%pagesPerDot == 0 {
		fmt.Printf(".")
	}
	for _, m := range page.Items {
		v := m[hourField]
		if v.S != nil {
			key := *v.S
			h.summary.counts[key]++
		}
	}
	return true
}
