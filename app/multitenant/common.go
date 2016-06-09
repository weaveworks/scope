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

func timeRequest(method string, metric *prometheus.SummaryVec, f func() error) error {
	startTime := time.Now()
	err := f()
	duration := time.Now().Sub(startTime)
	metric.WithLabelValues(method, errorCode(err)).Observe(float64(duration.Nanoseconds()))
	return err
}
