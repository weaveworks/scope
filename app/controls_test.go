package app_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
)

func TestControl(t *testing.T) {
	router := mux.NewRouter()
	app.RegisterControlRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	ip, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}

	probeConfig := appclient.ProbeConfig{
		ProbeID: "foo",
	}
	controlHandler := xfer.ControlHandlerFunc(func(req xfer.Request) xfer.Response {
		if req.NodeID != "nodeid" {
			t.Fatalf("'%s' != 'nodeid'", req.NodeID)
		}

		if req.Control != "control" {
			t.Fatalf("'%s' != 'control'", req.Control)
		}

		return xfer.Response{
			Value: "foo",
		}
	})
	client, err := appclient.NewAppClient(probeConfig, ip+":"+port, ip+":"+port, controlHandler)
	if err != nil {
		t.Fatal(err)
	}
	client.ControlConnection()
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	httpClient := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := httpClient.Post(server.URL+"/api/control/foo/nodeid/control", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	var response xfer.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Value != "foo" {
		t.Fatalf("'%s' != 'foo'", response.Value)
	}
}
