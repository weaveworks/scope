package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	publishTicks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cello",
			Subsystem: "probe",
			Name:      "publish_ticks",
			Help:      "Number of publish ticks observed.",
		},
		[]string{},
	)
	spyDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "cello",
			Subsystem: "probe",
			Name:      "spy_time_nanoseconds",
			Help:      "Total time spent spying on active connections.",
			MaxAge:    10 * time.Second, // like statsd
		},
		[]string{},
	)
)

func makePrometheusHandler() http.Handler {
	prometheus.MustRegister(publishTicks)
	prometheus.MustRegister(spyDuration)
	return prometheus.Handler()
}
