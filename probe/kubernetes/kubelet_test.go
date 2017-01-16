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
	"afeb8017-d8cf-11e6-84fa-0800278a0c83",
	"b57a1ada-d8cf-11e6-84fa-0800278a0c83",
	"00915ccb-dbd5-11e6-84fa-0800278a0c83",
	"009a3d10-dbd5-11e6-84fa-0800278a0c83",
	"fa8320c0-d8b0-11e6-828c-0800278a0c83",
	"fa95ebae-d8b0-11e6-828c-0800278a0c83",
	"af274a9e-d8cf-11e6-84fa-0800278a0c83",
	"af628fd4-d8cf-11e6-84fa-0800278a0c83",
	"af7f1d33-d8cf-11e6-84fa-0800278a0c83",
	"b0104177-d8cf-11e6-84fa-0800278a0c83",
	"b64d652b-d8cf-11e6-84fa-0800278a0c83",
	"b687c058-d8cf-11e6-84fa-0800278a0c83",
	"b0037e94-d8cf-11e6-84fa-0800278a0c83",
	"b56298ba-d8cf-11e6-84fa-0800278a0c83",
	"b613f0e9-d8cf-11e6-84fa-0800278a0c83",
	"fa2eadd4-d8d1-11e6-84fa-0800278a0c83",
	"af973584-d8cf-11e6-84fa-0800278a0c83",
	"b079da0b-d8cf-11e6-84fa-0800278a0c83",
	"b4e30a10-d8cf-11e6-84fa-0800278a0c83",
	"b5025f49-d8cf-11e6-84fa-0800278a0c83",
	"b52170e8-d8cf-11e6-84fa-0800278a0c83",
	"b53e9d66-d8cf-11e6-84fa-0800278a0c83",
	"b5c4c8c3-d8cf-11e6-84fa-0800278a0c83",
	"014fb8f91f3d52450a942179a984bc15",
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
	var savedKubeletURL string
	savedKubeletURL, kubernetes.KubeletURL = kubernetes.KubeletURL, server.URL
	defer func() { kubernetes.KubeletURL = savedKubeletURL }()

	uids, err := kubernetes.GetLocalPodUIDs()
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
