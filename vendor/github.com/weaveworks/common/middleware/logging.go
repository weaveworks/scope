package middleware

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/user"
)

// Log middleware logs http requests
type Log struct {
	Log               logging.Interface
	LogRequestHeaders bool // LogRequestHeaders true -> dump http headers at debug log level
}

// logWithRequest information from the request and context as fields.
func (l Log) logWithRequest(r *http.Request) logging.Interface {
	return user.LogWith(r.Context(), l.Log)
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
			l.logWithRequest(r).Errorf("Could not dump request headers: %v", err)
		}
		var buf bytes.Buffer
		wrapped := newBadResponseLoggingWriter(w, &buf)
		next.ServeHTTP(wrapped, r)
		statusCode := wrapped.statusCode
		if 100 <= statusCode && statusCode < 500 || statusCode == http.StatusBadGateway || statusCode == http.StatusServiceUnavailable {
			l.logWithRequest(r).Debugf("%s %s (%d) %s", r.Method, uri, statusCode, time.Since(begin))
			if l.LogRequestHeaders && headers != nil {
				l.logWithRequest(r).Debugf("Is websocket request: %v\n%s", IsWSHandshakeRequest(r), string(headers))
			}
		} else {
			l.logWithRequest(r).Warnf("%s %s (%d) %s", r.Method, uri, statusCode, time.Since(begin))
			if headers != nil {
				l.logWithRequest(r).Warnf("Is websocket request: %v\n%s", IsWSHandshakeRequest(r), string(headers))
			}
			l.logWithRequest(r).Warnf("Response: %s", buf.Bytes())
		}
	})
}

// Logging middleware logs each HTTP request method, path, response code and
// duration for all HTTP requests.
var Logging = Log{
	Log: logging.Global(),
}
