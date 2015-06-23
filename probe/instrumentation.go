package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/scope/probe/endpoint"
)

var (
	publishTicks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "scope",
			Subsystem: "probe",
			Name:      "publish_ticks",
			Help:      "Number of publish ticks observed.",
		},
		[]string{},
	)
)

func makePrometheusHandler() http.Handler {
	prometheus.MustRegister(publishTicks)
	prometheus.MustRegister(endpoint.SpyDuration)
	return prometheus.Handler()
}
