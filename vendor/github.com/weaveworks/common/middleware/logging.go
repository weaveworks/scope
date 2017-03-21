package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Log middleware logs http requests
type Log struct {
	LogRequestHeaders bool // LogRequestHeaders true -> dump http headers at debug log level
}

// Wrap implements Middleware
func (l Log) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		uri := r.RequestURI // capture the URI before running next, as it may get rewritten
		if l.LogRequestHeaders {
			// Log headers before running 'next' in case other interceptors change the data.
			headers, err := httputil.DumpRequest(r, false)
			if err != nil {
				log.Warnf("Could not dump request headers: %v", err)
				return
			}
			log.Debugf("Is websocket request: %v\n%s", IsWSHandshakeRequest(r), string(headers))
		}
		i := &interceptor{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(i, r)
		if 100 <= i.statusCode && i.statusCode < 400 {
			log.Debugf("%s %s (%d) %s", r.Method, uri, i.statusCode, time.Since(begin))
		} else {
			log.Warnf("%s %s (%d) %s", r.Method, uri, i.statusCode, time.Since(begin))
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
