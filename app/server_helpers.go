package app

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/ugorji/go/codec"
	"github.com/weaveworks/scope/report"

	log "github.com/sirupsen/logrus"
)

func respondWith(ctx context.Context, w http.ResponseWriter, code int, response interface{}) {
	if err, ok := response.(error); ok {
		log.Errorf("Error %d: %v", code, err)
		response = err.Error()
	} else if 500 <= code && code < 600 {
		log.Errorf("Non-error %d: %v", code, response)
	} else if ctx.Err() != nil {
		log.Debugf("Context error %v", ctx.Err())
		code = 499
		response = nil
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.LogKV("response-code", code)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(code)
	encoder := codec.NewEncoder(w, &codec.JsonHandle{})
	if err := encoder.Encode(response); err != nil {
		log.Errorf("Error encoding response: %v", err)
	}
}

// Similar to the above function, but respect the request's Accept header.
// Possibly we should do a complete parse of Accept, but for now just rudimentary check
func respondWithReport(ctx context.Context, w http.ResponseWriter, req *http.Request, response report.Report) {
	accept := req.Header.Get("Accept")
	if strings.HasPrefix(accept, "application/msgpack") {
		buf := bytes.Buffer{}
		encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
		if err := encoder.Encode(response); err != nil {
			log.Errorf("Error encoding response: %v", err)
		}
		if span := opentracing.SpanFromContext(ctx); span != nil {
			span.LogKV("encoded-size", len(buf.Bytes()))
		}
		w.Header().Set("Content-Type", "application/msgpack")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	encoder := codec.NewEncoder(w, &codec.JsonHandle{})
	if err := encoder.Encode(response); err != nil {
		log.Errorf("Error encoding response: %v", err)
	}
}
