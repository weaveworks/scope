package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Log middleware logs http requests
type Log struct {
	LogSuccess bool // LogSuccess true -> log successful queries; false -> only log failed queries
}

// Wrap implements Middleware
func (l Log) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		uri := r.RequestURI // capture the URI before running next, as it may get rewritten
		i := &interceptor{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(i, r)
		if l.LogSuccess || !(100 <= i.statusCode && i.statusCode < 400) {
			log.Infof("%s %s (%d) %s", r.Method, uri, i.statusCode, time.Since(begin))
		}
	})
}

// Logging middleware logs each HTTP request method, path, response code and
// duration for all HTTP requests.
var Logging = Log{
	LogSuccess: true,
}

// LogFailed middleware logs each HTTP request method, path, response code and
// duration for non-2xx HTTP requests.
var LogFailed = Log{
	LogSuccess: false,
}

// interceptor implements WriteHeader to intercept status codes. WriteHeader
// may not be called on success, so initialize statusCode with the status you
// want to report on success, i.e. http.StatusOK.
//
// interceptor also implements net.Hijacker, to let the downstream Handler
// hijack the connection. This is needed by the app-mapper's proxy.
type interceptor struct {
	http.ResponseWriter
	statusCode int
	recorded   bool
}

func (i *interceptor) WriteHeader(code int) {
	if !i.recorded {
		i.statusCode = code
		i.recorded = true
	}
	i.ResponseWriter.WriteHeader(code)
}

func (i *interceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := i.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("interceptor: can't cast parent ResponseWriter to Hijacker")
	}
	return hj.Hijack()
}
