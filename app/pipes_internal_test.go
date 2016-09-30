package app

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/test"
)

func TestPipeTimeout(t *testing.T) {
	router := mux.NewRouter()
	pr := NewLocalPipeRouter().(*localPipeRouter)
	RegisterPipeRoutes(router, pr)
	pr.Stop() // we don't want the loop running in the background

	mtime.NowForce(time.Now())
	defer mtime.NowReset()

	// create a new pipe.
	id := "foo"
	ctx := context.Background()
	pipe, _, err := pr.Get(ctx, id, UIEnd)
	if err != nil {
		t.Fatalf("not ok: %v", err)
	}

	// move time forward such that the new pipe should timeout
	mtime.NowForce(mtime.Now().Add(pipeTimeout))
	pr.timeout()
	if !pipe.Closed() {
		t.Fatalf("pipe didn't timeout")
	}

	// move time forward such that the pipe should be GCd
	mtime.NowForce(mtime.Now().Add(gcTimeout))
	pr.garbageCollect()
	if _, ok := pr.pipes[id]; ok {
		t.Fatalf("pipe not gc'd")
	}
}

type adapter struct {
	c appclient.AppClient
}

func (a adapter) PipeConnection(_, pipeID string, pipe xfer.Pipe) error {
	a.c.PipeConnection(pipeID, pipe)
	return nil
}

func (a adapter) PipeClose(_, pipeID string) error {
	return a.c.PipeClose(pipeID)
}

func TestPipeClose(t *testing.T) {
	router := mux.NewRouter()
	pr := NewLocalPipeRouter()
	RegisterPipeRoutes(router, pr)
	defer pr.Stop()

	server := httptest.NewServer(router)
	defer server.Close()

	ip, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}

	probeConfig := appclient.ProbeConfig{
		ProbeID: "foo",
	}
	url := url.URL{Scheme: "http", Host: ip + ":" + port}
	client, err := appclient.NewAppClient(probeConfig, ip+":"+port, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	// this is the probe end of the pipe
	pipeID, pipe, err := controls.NewPipe(adapter{client}, "appid")
	if err != nil {
		t.Fatal(err)
	}

	// this is a client to the app
	pipeURL := fmt.Sprintf("ws://%s:%s/api/pipe/%s", ip, port, pipeID)
	conn, _, err := websocket.DefaultDialer.Dial(pipeURL, http.Header{})
	if err != nil {
		t.Fatal(err)
	}

	// Send something from pipe -> app -> conn
	local, _ := pipe.Ends()
	msg := []byte("hello world")
	if _, err := local.Write(msg); err != nil {
		t.Fatal(err)
	}

	if _, buf, err := conn.ReadMessage(); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(buf, msg) {
		t.Fatalf("%v != %v", buf, msg)
	}

	// Send something from conn -> app -> probe
	msg = []byte("goodbye, cruel world")
	if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	if n, err := local.Read(buf); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(msg, buf[:n]) {
		t.Fatalf("%v != %v", buf, msg)
	}

	// Now delete the pipe
	if err := pipe.Close(); err != nil {
		t.Fatal(err)
	}

	// the client backs off for 1 second before trying to reconnect the pipe,
	// so we need to wait for longer.
	test.Poll(t, 2*time.Second, true, func() interface{} {
		return pipe.Closed()
	})
}
