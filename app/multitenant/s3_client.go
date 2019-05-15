package multitenant

import (
	"bytes"

	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/scope/report"
)

var (
	s3RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "s3_request_duration_seconds",
		Help:      "Time in seconds spent doing S3 requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
)

// S3Store is an S3 client that stores and retrieves Reports.
type S3Store struct {
	s3         *s3.S3
	bucketName string
}

func init() {
	prometheus.MustRegister(s3RequestDuration)
}

// NewS3Client creates a new S3 client.
func NewS3Client(config *aws.Config, bucketName string) S3Store {
	return S3Store{
		s3:         s3.New(session.New(config)),
		bucketName: bucketName,
	}
}

// FetchReports fetches multiple reports in parallel from S3.
func (store *S3Store) FetchReports(ctx context.Context, keys []string) (map[string]report.Report, []string, error) {
	type result struct {
		key    string
		report *report.Report
		err    error
	}

	ch := make(chan result, len(keys))

	for _, key := range keys {
		go func(key string) {
			r := result{key: key}
			r.report, r.err = store.fetchReport(ctx, key)
			ch <- r
		}(key)
	}

	reports := map[string]report.Report{}
	for range keys {
		r := <-ch
		if r.err != nil {
			return nil, []string{}, r.err
		}
		reports[r.key] = *r.report
	}
	return reports, []string{}, nil
}

func (store *S3Store) fetchReport(ctx context.Context, key string) (*report.Report, error) {
	var resp *s3.GetObjectOutput
	err := instrument.TimeRequestHistogram(ctx, "S3.Get", s3RequestDuration, func(_ context.Context) error {
		var err error
		resp, err = store.s3.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(store.bucketName),
			Key:    aws.String(key),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return report.MakeFromBinary(ctx, resp.Body)
}

// StoreReportBytes stores a report.
func (store *S3Store) StoreReportBytes(ctx context.Context, key string, buf []byte) (int, error) {
	err := instrument.TimeRequestHistogram(ctx, "S3.Put", s3RequestDuration, func(_ context.Context) error {
		_, err := store.s3.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(buf),
			Bucket: aws.String(store.bucketName),
			Key:    aws.String(key),
		})
		return err
	})
	return len(buf), err
}
