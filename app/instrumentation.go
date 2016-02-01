package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterInstrumentationRoutes exposes the metrics scope has collected, so
// that prometheus can scrape them.
func RegisterInstrumentationRoutes(c Collector, router *mux.Router) {
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
