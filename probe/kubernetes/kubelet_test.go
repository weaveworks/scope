package kubernetes_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/weaveworks/scope/probe/kubernetes"
)

const kubeletPodsJSONFile = "kubelet_pods.json"

// obtained with jq .items[].metadata.uid < kubelet_pods.json
var expectedPodUIDs = []string{
	"af1b5325-d8cf-11e6-84fa-0800278a0c83",
}

func TestGetLocalPodUIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/pods/" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			b, err := ioutil.ReadFile(kubeletPodsJSONFile)
			if err != nil {
				t.Fatalf("unexpected error reading json file: %v", err)
			}
			w.Write(b)
		},
	))
	defer server.Close()

	uids, err := kubernetes.GetLocalPodUIDs(server.URL.Host)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expectedPodUIDs) != len(uids) {
		t.Errorf("nnexpected length in pod UIDs (%d): expected %d", len(uids), len(expectedPodUIDs))
	}
	for _, expectedUID := range expectedPodUIDs {
		if _, ok := uids[expectedUID]; !ok {
			t.Errorf("uid not found: %s", expectedUID)
		}
	}
}
