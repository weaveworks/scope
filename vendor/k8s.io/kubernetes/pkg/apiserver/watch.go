/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/httplog"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/streaming"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/util/wsstream"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/kubernetes/pkg/watch/versioned"

	"github.com/emicklei/go-restful"
	"golang.org/x/net/websocket"
)

// nothing will ever be sent down this channel
var neverExitWatch <-chan time.Time = make(chan time.Time)

// timeoutFactory abstracts watch timeout logic for testing
type timeoutFactory interface {
	TimeoutCh() (<-chan time.Time, func() bool)
}

// realTimeoutFactory implements timeoutFactory
type realTimeoutFactory struct {
	timeout time.Duration
}

// TimeoutChan returns a channel which will receive something when the watch times out,
// and a cleanup function to call when this happens.
func (w *realTimeoutFactory) TimeoutCh() (<-chan time.Time, func() bool) {
	if w.timeout == 0 {
		return neverExitWatch, func() bool { return false }
	}
	t := time.NewTimer(w.timeout)
	return t.C, t.Stop
}

type textEncodable interface {
	// EncodesAsText should return true if objects should be transmitted as a WebSocket Text
	// frame (otherwise, they will be sent as a Binary frame).
	EncodesAsText() bool
}

// serveWatch handles serving requests to the server
// TODO: the functionality in this method and in WatchServer.Serve is not cleanly decoupled.
func serveWatch(watcher watch.Interface, scope RequestScope, req *restful.Request, res *restful.Response, timeout time.Duration) {
	// negotiate for the stream serializer
	serializer, mediaType, err := negotiateOutputSerializer(req.Request, scope.StreamSerializer)
	if err != nil {
		scope.err(err, res.ResponseWriter, req.Request)
		return
	}
	encoder := scope.StreamSerializer.EncoderForVersion(serializer, scope.Kind.GroupVersion())

	useTextFraming := false
	if encodable, ok := encoder.(textEncodable); ok && encodable.EncodesAsText() {
		useTextFraming = true
	}

	// find the embedded serializer matching the media type
	embeddedSerializer, ok := scope.Serializer.SerializerForMediaType(mediaType, nil)
	if !ok {
		scope.err(fmt.Errorf("no serializer defined for %q available for embedded encoding", mediaType), res.ResponseWriter, req.Request)
		return
	}
	embeddedEncoder := scope.Serializer.EncoderForVersion(embeddedSerializer, scope.Kind.GroupVersion())

	server := &WatchServer{
		watching: watcher,
		scope:    scope,

		useTextFraming:  useTextFraming,
		mediaType:       mediaType,
		encoder:         encoder,
		embeddedEncoder: embeddedEncoder,
		fixup: func(obj runtime.Object) {
			if err := setSelfLink(obj, req, scope.Namer); err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to set link for object %v: %v", reflect.TypeOf(obj), err))
			}
		},

		t: &realTimeoutFactory{timeout},
	}

	server.ServeHTTP(res.ResponseWriter, req.Request)
}

// WatchServer serves a watch.Interface over a websocket or vanilla HTTP.
type WatchServer struct {
	watching watch.Interface
	scope    RequestScope

	// true if websocket messages should use text framing (as opposed to binary framing)
	useTextFraming bool
	// the media type this watch is being served with
	mediaType string
	// used to encode the watch stream event itself
	encoder runtime.Encoder
	// used to encode the nested object in the watch stream
	embeddedEncoder runtime.Encoder
	fixup           func(runtime.Object)

	t timeoutFactory
}

// Serve serves a series of encoded events via HTTP with Transfer-Encoding: chunked
// or over a websocket connection.
func (s *WatchServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w = httplog.Unlogged(w)

	if wsstream.IsWebSocketRequest(req) {
		w.Header().Set("Content-Type", s.mediaType)
		websocket.Handler(s.HandleWS).ServeHTTP(w, req)
		return
	}

	cn, ok := w.(http.CloseNotifier)
	if !ok {
		err := fmt.Errorf("unable to start watch - can't get http.CloseNotifier: %#v", w)
		utilruntime.HandleError(err)
		s.scope.err(errors.NewInternalError(err), w, req)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		err := fmt.Errorf("unable to start watch - can't get http.Flusher: %#v", w)
		utilruntime.HandleError(err)
		s.scope.err(errors.NewInternalError(err), w, req)
		return
	}

	// get a framed encoder
	f, ok := s.encoder.(streaming.Framer)
	if !ok {
		// programmer error
		err := fmt.Errorf("no streaming support is available for media type %q", s.mediaType)
		utilruntime.HandleError(err)
		s.scope.err(errors.NewBadRequest(err.Error()), w, req)
		return
	}
	framer := f.NewFrameWriter(w)
	if framer == nil {
		// programmer error
		err := fmt.Errorf("no stream framing support is available for media type %q", s.mediaType)
		utilruntime.HandleError(err)
		s.scope.err(errors.NewBadRequest(err.Error()), w, req)
		return
	}
	e := streaming.NewEncoder(framer, s.encoder)

	// ensure the connection times out
	timeoutCh, cleanup := s.t.TimeoutCh()
	defer cleanup()
	defer s.watching.Stop()

	// begin the stream
	w.Header().Set("Content-Type", s.mediaType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	buf := &bytes.Buffer{}
	for {
		select {
		case <-cn.CloseNotify():
			return
		case <-timeoutCh:
			return
		case event, ok := <-s.watching.ResultChan():
			if !ok {
				// End of results.
				return
			}
			obj := event.Object
			s.fixup(obj)
			if err := s.embeddedEncoder.EncodeToStream(obj, buf); err != nil {
				// unexpected error
				utilruntime.HandleError(fmt.Errorf("unable to encode watch object: %v", err))
				return
			}
			event.Object = &runtime.Unknown{
				Raw: buf.Bytes(),
				// ContentType is not required here because we are defaulting to the serializer
				// type
			}
			if err := e.Encode((*versioned.InternalEvent)(&event)); err != nil {
				utilruntime.HandleError(fmt.Errorf("unable to encode watch object: %v", err))
				// client disconnect.
				return
			}
			flusher.Flush()

			buf.Reset()
		}
	}
}

// HandleWS implements a websocket handler.
func (s *WatchServer) HandleWS(ws *websocket.Conn) {
	defer ws.Close()
	done := make(chan struct{})
	go wsstream.IgnoreReceives(ws, 0)
	buf := &bytes.Buffer{}
	streamBuf := &bytes.Buffer{}
	for {
		select {
		case <-done:
			s.watching.Stop()
			return
		case event, ok := <-s.watching.ResultChan():
			if !ok {
				// End of results.
				return
			}
			obj := event.Object
			s.fixup(obj)
			if err := s.embeddedEncoder.EncodeToStream(obj, buf); err != nil {
				// unexpected error
				utilruntime.HandleError(fmt.Errorf("unable to encode watch object: %v", err))
				return
			}
			event.Object = &runtime.Unknown{
				Raw: buf.Bytes(),
				// ContentType is not required here because we are defaulting to the serializer
				// type
			}
			// the internal event will be versioned by the encoder
			internalEvent := versioned.InternalEvent(event)
			if err := s.encoder.EncodeToStream(&internalEvent, streamBuf); err != nil {
				// encoding error
				utilruntime.HandleError(fmt.Errorf("unable to encode event: %v", err))
				s.watching.Stop()
				return
			}
			if s.useTextFraming {
				if err := websocket.Message.Send(ws, streamBuf.String()); err != nil {
					// Client disconnect.
					s.watching.Stop()
					return
				}
			} else {
				if err := websocket.Message.Send(ws, streamBuf.Bytes()); err != nil {
					// Client disconnect.
					s.watching.Stop()
					return
				}
			}
			buf.Reset()
			streamBuf.Reset()
		}
	}
}
