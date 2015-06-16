package tag_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

func TestWeaveTaggerOverlayTopology(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(mockWeaveRouter))
	defer s.Close()

	w, err := tag.NewWeaveTagger(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	if want, have := (report.Topology{
		Adjacency:     report.Adjacency{},
		EdgeMetadatas: report.EdgeMetadatas{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeOverlayNodeID(mockWeavePeerName): {
				tag.WeavePeerName:     mockWeavePeerName,
				tag.WeavePeerNickName: mockWeavePeerNickName,
			},
		},
	}), w.OverlayTopology(); !reflect.DeepEqual(want, have) {
		t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
	}
}

const (
	mockWeavePeerName     = "winnebago"
	mockWeavePeerNickName = "winny"
)

func mockWeaveRouter(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"Peers": []map[string]interface{}{{
			"Name":     mockWeavePeerName,
			"NickName": mockWeavePeerNickName,
		}},
	}); err != nil {
		println(err.Error())
	}
}
