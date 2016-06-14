package multitenant

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func errorCode(err error) string {
	if err == nil {
		return "200"
	}
	return "500"
}

// timeRequest runs 'f' and records how long it took in the given Prometheus
// metric. If 'f' returns successfully, record a "200". Otherwise, record
// "500".
//
// If you want more complicated logic for translating errors into statuses,
// use 'timeRequestStatus'.
func timeRequest(method string, metric *prometheus.SummaryVec, f func() error) error {
	return timeRequestStatus(method, metric, errorCode, f)
}

// timeRequestStatus runs 'f' and records how long it took in the given
// Prometheus metric.
//
// toStatusCode is a function that translates errors returned by 'f' into
// HTTP-like status codes.
func timeRequestStatus(method string, metric *prometheus.SummaryVec, toStatusCode func(error) string, f func() error) error {
	if toStatusCode == nil {
		toStatusCode = errorCode
	}
	startTime := time.Now()
	err := f()
	duration := time.Now().Sub(startTime)
	metric.WithLabelValues(method, toStatusCode(err)).Observe(float64(duration.Nanoseconds()))
	return err
}
