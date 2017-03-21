package instrument

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
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
// "500".  It will also emit an OpenTracing span if you have a global tracer configured.
//
// If you want more complicated logic for translating errors into statuses,
// use 'TimeRequestHistogramStatus'.
func TimeRequestHistogram(ctx context.Context, method string, metric *prometheus.HistogramVec, f func(context.Context) error) error {
	return TimeRequestHistogramStatus(ctx, method, metric, ErrorCode, f)
}

// TimeRequestHistogramStatus runs 'f' and records how long it took in the given
// Prometheus histogram metric.  It will also emit an OpenTracing span if you have
// a global tracer configured.
//
// toStatusCode is a function that translates errors returned by 'f' into
// HTTP-like status codes.
func TimeRequestHistogramStatus(ctx context.Context, method string, metric *prometheus.HistogramVec, toStatusCode func(error) string, f func(context.Context) error) error {
	if toStatusCode == nil {
		toStatusCode = ErrorCode
	}

	sp, newCtx := opentracing.StartSpanFromContext(ctx, method)
	ext.SpanKindRPCClient.Set(sp)
	startTime := time.Now()

	err := f(newCtx)

	if err != nil {
		ext.Error.Set(sp, true)
	}
	sp.Finish()
	if metric != nil {
		metric.WithLabelValues(method, toStatusCode(err)).Observe(time.Now().Sub(startTime).Seconds())
	}
	return err
}
