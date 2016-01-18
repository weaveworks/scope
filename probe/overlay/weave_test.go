package overlay_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/weaveworks/scope/common/exec"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
	testExec "github.com/weaveworks/scope/test/exec"
)

func TestWeaveTaggerOverlayTopology(t *testing.T) {
	wait := make(chan struct{})
	oldExecCmd := exec.Command
	defer func() { exec.Command = oldExecCmd }()
	exec.Command = func(name string, args ...string) exec.Cmd {
		close(wait)
		return testExec.NewMockCmdString(fmt.Sprintf("%s %s %s/24\n", mockContainerID, mockContainerMAC, mockContainerIP))
	}

	s := httptest.NewServer(http.HandlerFunc(mockWeaveRouter))
	defer s.Close()

	w := overlay.NewWeave(mockHostID, s.URL)
	defer w.Stop()
	w.Tick()
	<-wait

	{
		// Overlay node should include peer name and nickname
		have, err := w.Report()
		if err != nil {
			t.Fatal(err)
		}

		nodeID := report.MakeOverlayNodeID(mockWeavePeerName)
		node, ok := have.Overlay.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected overlay node %q, but not found", nodeID)
		}
		if peerName, ok := node.Latest.Lookup(overlay.WeavePeerName); !ok || peerName != mockWeavePeerName {
			t.Errorf("Expected weave peer name %q, got %q", mockWeavePeerName, peerName)
		}
		if peerNick, ok := node.Latest.Lookup(overlay.WeavePeerNickName); !ok || peerNick != mockWeavePeerNickName {
			t.Errorf("Expected weave peer nickname %q, got %q", mockWeavePeerNickName, peerNick)
		}
	}

	{
		// Container nodes should be tagged with their overlay info
		nodeID := report.MakeContainerNodeID(mockContainerID)
		have, err := w.Tag(report.Report{
			Container: report.MakeTopology().AddNode(nodeID, report.MakeNodeWith(map[string]string{
				docker.ContainerID: mockContainerID,
			})),
		})
		if err != nil {
			t.Fatal(err)
		}

		node, ok := have.Container.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected container node %q, but not found", nodeID)
		}

		// Should have Weave DNS Hostname
		if have, ok := node.Latest.Lookup(overlay.WeaveDNSHostname); !ok || have != mockHostname {
			t.Errorf("Expected weave dns hostname %q, got %q", mockHostname, have)
		}
		// Should have Weave MAC Address
		if have, ok := node.Latest.Lookup(overlay.WeaveMACAddress); !ok || have != mockContainerMAC {
			t.Errorf("Expected weave mac address %q, got %q", mockContainerMAC, have)
		}
		// Should have Weave container ip
		if have, ok := node.Sets.Lookup(docker.ContainerIPs); !ok || !have.Contains(mockContainerIP) {
			t.Errorf("Expected container ips to include the weave IP %q, got %q", mockContainerIP, have)
		}
		// Should have Weave container ip (with scope)
		if have, ok := node.Sets.Lookup(docker.ContainerIPsWithScopes); !ok || !have.Contains(mockContainerIPWithScope) {
			t.Errorf("Expected container ips to include the weave IP (with scope) %q, got %q", mockContainerIPWithScope, have)
		}
	}
}

const (
	mockHostID               = "host1"
	mockWeavePeerName        = "winnebago"
	mockWeavePeerNickName    = "winny"
	mockContainerID          = "83183a667c01"
	mockContainerMAC         = "d6:f2:5a:12:36:a8"
	mockContainerIP          = "10.0.0.123"
	mockContainerIPWithScope = ";10.0.0.123"
	mockHostname             = "hostname.weave.local"
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
