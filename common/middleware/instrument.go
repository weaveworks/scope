package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// Instrument is a Middleware which records timings for every HTTP request
type Instrument struct {
	RouteMatcher interface {
		Match(*http.Request, *mux.RouteMatch) bool
	}
	Duration *prometheus.SummaryVec
}

func isWSHandshakeRequest(req *http.Request) bool {
	return strings.ToLower(req.Header.Get("Upgrade")) == "websocket" &&
		strings.ToLower(req.Header.Get("Connection")) == "upgrade"
}

// Wrap implements middleware.Interface
func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		isWS := strconv.FormatBool(isWSHandshakeRequest(r))
		interceptor := &interceptor{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(interceptor, r)
		var (
			route  = i.getRouteName(r)
			status = strconv.Itoa(interceptor.statusCode)
			took   = time.Since(begin)
		)
		i.Duration.WithLabelValues(r.Method, route, status, isWS).Observe(float64(took.Nanoseconds()))
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
