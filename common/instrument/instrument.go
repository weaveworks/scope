package instrument

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ErrorCode converts an error in to an http-style error-code.
func ErrorCode(err error) string {
	if err == nil {
		return "200"
	}
	return "500"
}

// TimeRequestHistogram runs 'f' and records how long it took in the given Prometheus
// histogram metric. If 'f' returns successfully, record a "200". Otherwise, record
// "500".
//
// If you want more complicated logic for translating errors into statuses,
// use 'TimeRequestStatus'.
func TimeRequestHistogram(method string, metric *prometheus.HistogramVec, f func() error) error {
	return TimeRequestHistogramStatus(method, metric, ErrorCode, f)
}

// TimeRequestHistogramStatus runs 'f' and records how long it took in the given
// Prometheus histogram metric.
//
// toStatusCode is a function that translates errors returned by 'f' into
// HTTP-like status codes.
func TimeRequestHistogramStatus(method string, metric *prometheus.HistogramVec, toStatusCode func(error) string, f func() error) error {
	if toStatusCode == nil {
		toStatusCode = ErrorCode
	}
	startTime := time.Now()
	err := f()
	duration := time.Now().Sub(startTime)
	metric.WithLabelValues(method, toStatusCode(err)).Observe(duration.Seconds())
	return err
}
