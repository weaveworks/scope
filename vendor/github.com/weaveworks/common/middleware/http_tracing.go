package middleware

import (
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

// Tracer is a middleware which traces incoming requests.
type Tracer struct{}

// Wrap implements Interface
func (t Tracer) Wrap(next http.Handler) http.Handler {
	traceHandler := nethttp.Middleware(opentracing.GlobalTracer(), next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var maybeTracer http.Handler
		// Don't try and trace websocket requests because nethttp.Middleware
		// doesn't support http.Hijack yet
		if IsWSHandshakeRequest(r) {
			maybeTracer = next
		} else {
			maybeTracer = traceHandler
		}
		maybeTracer.ServeHTTP(w, r)
	})
}
