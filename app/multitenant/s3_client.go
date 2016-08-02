package multitenant

import (
	"bytes"
	"compress/gzip"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/scope/common/instrument"
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
func (store *S3Store) FetchReports(keys []string) (map[string]ReportWithStats, []string, error) {
	type result struct {
		key    string
		report ReportWithStats
		err    error
	}

	ch := make(chan result, len(keys))

	for _, key := range keys {
		go func(key string) {
			r := result{key: key}
			report, readStats, err := store.fetchReport(key)
			r.err = err
			stats := ReportStats{
				ReadStats: readStats,
				Origin:    "s3",
			}
			r.report = ReportWithStats{
				Report:      *report,
				ReportStats: stats,
			}
			ch <- r
		}(key)
	}

	reports := map[string]ReportWithStats{}
	for range keys {
		r := <-ch
		if r.err != nil {
			return nil, []string{}, r.err
		}
		reports[r.key] = r.report
	}
	return reports, []string{}, nil
}

func (store *S3Store) fetchReport(key string) (*report.Report, report.ReadStats, error) {
	var resp *s3.GetObjectOutput
	err := instrument.TimeRequestHistogram("Get", s3RequestDuration, func() error {
		var err error
		resp, err = store.s3.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(store.bucketName),
			Key:    aws.String(key),
		})
		return err
	})
	if err != nil {
		return nil, report.ReadStats{}, err
	}
	return report.MakeFromBinaryWithStats(resp.Body)
}

// StoreReport serializes and stores a report.
//
// Returns the size of the report. This only equals bytes written if err is nil.
func (store *S3Store) StoreReport(key string, report *report.Report) (int, error) {
	var buf bytes.Buffer
	report.WriteBinary(&buf, gzip.BestCompression)
	err := instrument.TimeRequestHistogram("Put", s3RequestDuration, func() error {
		_, err := store.s3.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(buf.Bytes()),
			Bucket: aws.String(store.bucketName),
			Key:    aws.String(key),
		})
		return err
	})
	return buf.Len(), err
}
