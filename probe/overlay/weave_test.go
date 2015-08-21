package overlay_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/exec"
)

func TestWeaveTaggerOverlayTopology(t *testing.T) {
	oldExecCmd := exec.Command
	defer func() { exec.Command = oldExecCmd }()
	exec.Command = func(name string, args ...string) exec.Cmd {
		return exec.NewMockCmdString(fmt.Sprintf("%s %s %s/24\n", mockContainerID, mockContainerMAC, mockContainerIP))
	}

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
						docker.ContainerID:       mockContainerID,
						overlay.WeaveDNSHostname: mockHostname,
						overlay.WeaveMACAddress:  mockContainerMAC,
						docker.ContainerIPs:      mockContainerIP,
					}),
				},
			},
		}
		have, err := w.Tag(report.Report{
			Container: report.Topology{
				NodeMetadatas: report.NodeMetadatas{
					nodeID: report.MakeNodeMetadataWith(map[string]string{
						docker.ContainerID: mockContainerID,
					}),
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
	mockContainerID       = "83183a667c01"
	mockContainerMAC      = "d6:f2:5a:12:36:a8"
	mockContainerIP       = "10.0.0.123"
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
