package overlay_test

import (
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/weave"
)

const (
	mockHostID               = "host1"
	mockContainerIPWithScope = ";" + weave.MockContainerIP
)

func runTest(t *testing.T, f func(*overlay.Weave)) {
	w, err := overlay.NewWeave(mockHostID, weave.MockClient{})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Wait until the reporter reports some nodes
	test.Poll(t, 300*time.Millisecond, 1, func() interface{} {
		have, _ := w.Report()
		return len(have.Overlay.Nodes)
	})

	// Give some time for the Backoff collectors to finish
	time.Sleep(time.Second)

	f(w)
}

func TestContainerTopologyTagging(t *testing.T) {
	test := func(w *overlay.Weave) {
		// Container nodes should be tagged with their overlay info
		nodeID := report.MakeContainerNodeID(weave.MockContainerID)
		topology := report.MakeTopology()
		topology.AddNode(report.MakeNodeWith(nodeID, map[string]string{
			docker.ContainerID: weave.MockContainerID,
		}))
		have, err := w.Tag(report.Report{Container: topology})
		if err != nil {
			t.Fatal(err)
		}

		node, ok := have.Container.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected container node %q, but not found", nodeID)
		}

		// Should have Weave DNS Hostname
		if have, ok := node.Latest.Lookup(overlay.WeaveDNSHostname); !ok || have != weave.MockHostname {
			t.Errorf("Expected weave dns hostname %q, got %q", weave.MockHostname, have)
		}
		// Should have Weave MAC Address
		if have, ok := node.Latest.Lookup(overlay.WeaveMACAddress); !ok || have != weave.MockContainerMAC {
			t.Errorf("Expected weave mac address %q, got %q", weave.MockContainerMAC, have)
		}
		// Should have Weave container ip
		if have, ok := node.Sets.Lookup(docker.ContainerIPs); !ok || !have.Contains(weave.MockContainerIP) {
			t.Errorf("Expected container ips to include the weave IP %q, got %q", weave.MockContainerIP, have)
		}
		// Should have Weave container ip (with scope)
		if have, ok := node.Sets.Lookup(docker.ContainerIPsWithScopes); !ok || !have.Contains(mockContainerIPWithScope) {
			t.Errorf("Expected container ips to include the weave IP (with scope) %q, got %q", mockContainerIPWithScope, have)
		}
	}

	runTest(t, test)
}

func TestOverlayTopology(t *testing.T) {
	test := func(w *overlay.Weave) {
		// Overlay node should include peer name and nickname
		have, err := w.Report()
		if err != nil {
			t.Fatal(err)
		}

		nodeID := report.MakeOverlayNodeID(report.WeaveOverlayPeerPrefix, weave.MockWeavePeerName)
		node, ok := have.Overlay.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected overlay node %q, but not found", nodeID)
		}
		if peerName, ok := node.Latest.Lookup(overlay.WeavePeerName); !ok || peerName != weave.MockWeavePeerName {
			t.Errorf("Expected weave peer name %q, got %q", weave.MockWeavePeerName, peerName)
		}
		if peerNick, ok := node.Latest.Lookup(overlay.WeavePeerNickName); !ok || peerNick != weave.MockWeavePeerNickName {
			t.Errorf("Expected weave peer nickname %q, got %q", weave.MockWeavePeerNickName, peerNick)
		}
		if localNetworks, ok := node.Sets.Lookup(host.LocalNetworks); !ok || !reflect.DeepEqual(localNetworks, report.MakeStringSet(weave.MockWeaveDefaultSubnet)) {
			t.Errorf("Expected weave node local_networks %q, got %q", report.MakeStringSet(weave.MockWeaveDefaultSubnet), localNetworks)
		}
		// The weave proxy container is running
		if have, ok := node.Latest.Lookup(overlay.WeaveProxyStatus); !ok || have != "running" {
			t.Errorf("Expected weave proxy status %q, got %q", "running", have)
		}
		if have, ok := node.Latest.Lookup(overlay.WeaveProxyAddress); !ok || have != weave.MockProxyAddress {
			t.Errorf("Expected weave proxy address %q, got %q", weave.MockProxyAddress, have)
		}
		// The weave plugin container is running
		if have, ok := node.Latest.Lookup(overlay.WeavePluginStatus); !ok || have != "running" {
			t.Errorf("Expected weave plugin status %q, got %q", "running", have)
		}
		if have, ok := node.Latest.Lookup(overlay.WeavePluginDriver); !ok || have != weave.MockDriverName {
			t.Errorf("Expected weave proxy address %q, got %q", weave.MockDriverName, have)
		}
		// The mock data indicates ranges are owned by unreachable peers
		if have, ok := node.Latest.Lookup(overlay.WeaveIPAMStatus); !ok || have != "all ranges owned by unreachable peers" {
			t.Errorf("Expected weave IPAM status %q, got %q", "all ranges owned by unreachable peers", have)
		}
	}

	runTest(t, test)
}
