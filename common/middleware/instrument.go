package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Instrument struct {
	RouteMatcher RouteMatcher
	Duration     *prometheus.SummaryVec
}

// RouteMatcher is implemented by mux.Router.
type RouteMatcher interface {
	Match(*http.Request, *mux.RouteMatch) bool
}

func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		interceptor := &interceptor{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(interceptor, r)
		var (
			route  = i.getRouteName(r)
			status = strconv.Itoa(interceptor.statusCode)
			took   = time.Since(begin)
		)
		i.Duration.WithLabelValues(r.Method, route, status).Observe(float64(took.Nanoseconds()))
	})
}

func (i Instrument) getRouteName(r *http.Request) string {
	var routeMatch mux.RouteMatch
	if !i.RouteMatcher.Match(r, &routeMatch) {
		return "unmatched_path"
	}
	name := routeMatch.Route.GetName()
	if name == "" {
		return "unnamed_path"
	}
	return name
}
