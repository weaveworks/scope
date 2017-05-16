package app

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/ghost/handlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/hostname"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/report"
)

var (
	// Version - set at buildtime.
	Version = "dev"

	// UniqueID - set at runtime.
	UniqueID = "0"
)

// contextKey is a wrapper type for use in context.WithValue() to satisfy golint
// https://github.com/golang/go/issues/17293
// https://github.com/golang/lint/pull/245
type contextKey string

// RequestCtxKey is key used for request entry in context
const RequestCtxKey contextKey = contextKey("request")

// CtxHandlerFunc is a http.HandlerFunc, with added contexts
type CtxHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func requestContextDecorator(f CtxHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(context.Background(), RequestCtxKey, r)
		f(ctx, w, r)
	}
}

// URLMatcher uses request.RequestURI (the raw, unparsed request) to attempt
// to match pattern.  It does this as go's URL.Parse method is broken, and
// mistakenly unescapes the Path before parsing it.  This breaks %2F (encoded
// forward slashes) in the paths.
func URLMatcher(pattern string) mux.MatcherFunc {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		vars, match := matchURL(r, pattern)
		if match {
			rm.Vars = vars
		}
		return match
	}
}

func matchURL(r *http.Request, pattern string) (map[string]string, bool) {
	matchParts := strings.Split(pattern, "/")
	path := strings.SplitN(r.RequestURI, "?", 2)[0]
	parts := strings.Split(path, "/")
	if len(parts) != len(matchParts) {
		return nil, false
	}

	vars := map[string]string{}
	for i, part := range parts {
		unescaped, err := url.QueryUnescape(part)
		if err != nil {
			return nil, false
		}
		match := matchParts[i]
		if strings.HasPrefix(match, "{") && strings.HasSuffix(match, "}") {
			vars[strings.Trim(match, "{}")] = unescaped
		} else if matchParts[i] != unescaped {
			return nil, false
		}
	}
	return vars, true
}

func gzipHandler(h http.HandlerFunc) http.HandlerFunc {
	return handlers.GZIPHandlerFunc(h, nil)
}

// RegisterTopologyRoutes registers the various topology routes with a http mux.
func RegisterTopologyRoutes(router *mux.Router, r Reporter) {
	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api",
		gzipHandler(requestContextDecorator(apiHandler(r))))
	get.HandleFunc("/api/topology",
		gzipHandler(requestContextDecorator(topologyRegistry.makeTopologyList(r))))
	get.
		HandleFunc("/api/topology/{topology}",
			gzipHandler(requestContextDecorator(topologyRegistry.captureRenderer(r, handleTopology)))).
		Name("api_topology_topology")
	get.
		HandleFunc("/api/topology/{topology}/ws",
			requestContextDecorator(captureReporter(r, handleWebsocket))). // NB not gzip!
		Name("api_topology_topology_ws")
	get.
		MatcherFunc(URLMatcher("/api/topology/{topology}/{id}")).HandlerFunc(
		gzipHandler(requestContextDecorator(topologyRegistry.captureRenderer(r, handleNode)))).
		Name("api_topology_topology_id")
	get.HandleFunc("/api/report",
		gzipHandler(requestContextDecorator(makeRawReportHandler(r))))
	get.HandleFunc("/api/probes",
		gzipHandler(requestContextDecorator(makeProbeHandler(r))))
}

// RegisterReportPostHandler registers the handler for report submission
func RegisterReportPostHandler(a Adder, router *mux.Router) {
	post := router.Methods("POST").Subrouter()
	post.HandleFunc("/api/report", requestContextDecorator(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var (
			rpt    report.Report
			buf    bytes.Buffer
			reader = io.TeeReader(r.Body, &buf)
		)

		gzipped := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
		if !gzipped {
			reader = io.TeeReader(r.Body, gzip.NewWriter(&buf))
		}

		contentType := r.Header.Get("Content-Type")
		isMsgpack := strings.HasPrefix(contentType, "application/msgpack")
		var handle codec.Handle
		switch {
		case strings.HasPrefix(contentType, "application/json"):
			handle = &codec.JsonHandle{}
		case isMsgpack:
			handle = &codec.MsgpackHandle{}
		default:
			respondWith(w, http.StatusBadRequest, fmt.Errorf("Unsupported Content-Type: %v", contentType))
			return
		}

		if err := rpt.ReadBinary(reader, gzipped, handle); err != nil {
			respondWith(w, http.StatusBadRequest, err)
			return
		}

		// a.Add(..., buf) assumes buf is gzip'd msgpack
		if !isMsgpack {
			buf = bytes.Buffer{}
			rpt.WriteBinary(&buf, gzip.DefaultCompression)
		}

		if err := a.Add(ctx, rpt, buf.Bytes()); err != nil {
			log.Errorf("Error Adding report: %v", err)
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

var newVersion = struct {
	sync.Mutex
	*xfer.NewVersionInfo
}{}

// NewVersion is called to expose new version information to /api
func NewVersion(version, downloadURL string) {
	newVersion.Lock()
	defer newVersion.Unlock()
	newVersion.NewVersionInfo = &xfer.NewVersionInfo{
		Version:     version,
		DownloadURL: downloadURL,
	}
}

func apiHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		report, err := rep.Report(ctx, time.Now())
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		newVersion.Lock()
		defer newVersion.Unlock()
		respondWith(w, http.StatusOK, xfer.Details{
			ID:         UniqueID,
			Version:    Version,
			Hostname:   hostname.Get(),
			Plugins:    report.Plugins,
			NewVersion: newVersion.NewVersionInfo,
		})
	}
}
