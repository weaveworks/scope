package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/user"
)

// Log middleware logs http requests
type Log struct {
	LogRequestHeaders bool // LogRequestHeaders true -> dump http headers at debug log level
}

// logWithRequest information from the request and context as fields.
func logWithRequest(r *http.Request) *log.Entry {
	return log.WithFields(user.LogFields(r.Context()))
}

// Wrap implements Middleware
func (l Log) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		uri := r.RequestURI // capture the URI before running next, as it may get rewritten
		// Log headers before running 'next' in case other interceptors change the data.
		headers, err := httputil.DumpRequest(r, false)
		if err != nil {
			headers = nil
			logWithRequest(r).Errorf("Could not dump request headers: %v", err)
		}
		var buf bytes.Buffer
		wrapped := newBadResponseLoggingWriter(w, &buf)
		next.ServeHTTP(wrapped, r)
		statusCode := wrapped.statusCode
		if 100 <= statusCode && statusCode < 500 {
			logWithRequest(r).Debugf("%s %s (%d) %s", r.Method, uri, statusCode, time.Since(begin))
			if l.LogRequestHeaders && headers != nil {
				logWithRequest(r).Debugf("Is websocket request: %v\n%s", IsWSHandshakeRequest(r), string(headers))
			}
		} else {
			logWithRequest(r).Warnf("%s %s (%d) %s", r.Method, uri, statusCode, time.Since(begin))
			if headers != nil {
				logWithRequest(r).Warnf("Is websocket request: %v\n%s", IsWSHandshakeRequest(r), string(headers))
			}
			logWithRequest(r).Warnf("Response: %s", buf.Bytes())
		}
	})
}

// Logging middleware logs each HTTP request method, path, response code and
// duration for all HTTP requests.
var Logging = Log{}

// interceptor implements WriteHeader to intercept status codes. WriteHeader
// may not be called on success, so initialize statusCode with the status you
// want to report on success, i.e. http.StatusOK.
//
// interceptor also implements net.Hijacker, to let the downstream Handler
// hijack the connection. This is needed, for example, for working with websockets.
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
