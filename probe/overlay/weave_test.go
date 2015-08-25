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

	w, err := overlay.NewWeave(mockHostID, s.URL)
	if err != nil {
		t.Fatal(err)
	}

	{
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

	{
		nodeID := report.MakeContainerNodeID(mockHostID, mockContainerID)
		want := report.Report{
			Container: report.Topology{
				NodeMetadatas: report.NodeMetadatas{
					nodeID: report.MakeNodeMetadataWith(map[string]string{
						overlay.WeaveDNSHostname: mockHostname,
					}),
				},
			},
		}
		have, err := w.Tag(report.Report{
			Container: report.Topology{
				NodeMetadatas: report.NodeMetadatas{
					nodeID: report.MakeNodeMetadata(),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

const (
	mockHostID            = "host1"
	mockWeavePeerName     = "winnebago"
	mockWeavePeerNickName = "winny"
	mockContainerID       = "a1b2c3d4e5"
	mockHostname          = "hostname.weave.local"
)

var (
	mockResponse = fmt.Sprintf(`{
		"Router": {
			"Peers": [{
				"Name": "%s",
				"Nickname": "%s"
			}]
		},
		"DNS": {
			"Entries": [{
				"ContainerID": "%s",
				"Hostname": "%s.",
				"Tombstone": 0
			}]
		}
	}`, mockWeavePeerName, mockWeavePeerNickName, mockContainerID, mockHostname)
)

func mockWeaveRouter(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(mockResponse)); err != nil {
		panic(err)
	}
}
