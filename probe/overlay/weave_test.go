package overlay_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestWeaveTaggerOverlayTopology(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(mockWeaveRouter))
	defer s.Close()

	w, err := overlay.NewWeave(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	have, err := w.Report()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := (report.Topology{
		Adjacency:     report.Adjacency{},
		EdgeMetadatas: report.EdgeMetadatas{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeOverlayNodeID(mockWeavePeerName): report.MakeNodeMetadataWith(map[string]string{
				overlay.WeavePeerName:     mockWeavePeerName,
				overlay.WeavePeerNickName: mockWeavePeerNickName,
			}),
		},
	}), have.Overlay; !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

const (
	mockWeavePeerName     = "winnebago"
	mockWeavePeerNickName = "winny"
)

var (
	mockResponse = fmt.Sprintf(`{
		"router": {
			"peers": [{
				"name": "%s",
				"nickname": "%s"
			}]
		}
	}`, mockWeavePeerName, mockWeavePeerNickName)
)

func mockWeaveRouter(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(mockResponse)); err != nil {
		panic(err)
	}
}
