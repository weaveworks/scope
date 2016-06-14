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
// metric.
//
// toStatusCode is a function that translates errors returned by 'f' into
// HTTP-like status codes. If 'nil', uses a default function that returns
// "500" for any error and "200" otherwise.
func timeRequest(method string, metric *prometheus.SummaryVec, toStatusCode func(error) string, f func() error) error {
	if toStatusCode == nil {
		toStatusCode = errorCode
	}
	startTime := time.Now()
	err := f()
	duration := time.Now().Sub(startTime)
	metric.WithLabelValues(method, toStatusCode(err)).Observe(float64(duration.Nanoseconds()))
	return err
}
