package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

// RegisterMetricRoutes registers all routes specifically to do with metrics.
// Exposes the metrics scope has collected, so that prometheus can scrape them
// as well as the query endpoints.
func RegisterMetricRoutes(c Collector, m MetricStorage, router *mux.Router) {
	if m != nil {
		m.RegisterRoutes(c, router)
	}
	router.Methods("GET").Path("/api/metrics").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := []string{}
		rpt := c.Report()
		for _, t := range rpt.Topologies() {
			for id, node := range t.Nodes {
				for key, _ := range node.Metrics {
					metrics = append(metrics, fmt.Sprintf("%s{topology=%q,node=%q}", key, t.Name, id))
				}
			}
		}
		json.NewEncoder(w).Encode(metrics)
	})
}

type MetricStorage interface {
	RegisterRoutes(Collector, *mux.Router)
}

func NewMetricStorage(uri string) (MetricStorage, error) {
	u, err := url.Parse(uri)
	if err == nil {
		err = fmt.Errorf("Unknown uri scheme: %q", u.Scheme)
	}
	return nil, err
}

/*
func (m *prometheusMetricStorage) RegisterRoutes(c Collector, router *mux.Router) {
	router.Methods("GET").Path("/metrics").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rpt := c.Report()
		for _, t := range rpt.Topologies() {
			for id, node := range t.Nodes {
				for key, metric := range node.Metrics {
					if last := metric.LastSample(); last != nil {
						fmt.Fprintf(w, "%s{topology=%q,node=%q} %f\n", key, t.Name, id, last.Value)
					}
				}
			}
		}
	})
}
*/
