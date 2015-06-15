package render

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func init() {
	spew.Config.SortKeys = true // :\
}

type mockRenderer struct {
	report.RenderableNodes
	aggregateMetadata report.AggregateMetadata
}

func (m mockRenderer) Render(rpt report.Report) report.RenderableNodes {
	return m.RenderableNodes
}
func (m mockRenderer) AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return m.aggregateMetadata
}

func TestReduceRender(t *testing.T) {
	renderer := Reduce([]Renderer{
		mockRenderer{RenderableNodes: report.RenderableNodes{"foo": {ID: "foo"}}},
		mockRenderer{RenderableNodes: report.RenderableNodes{"bar": {ID: "bar"}}},
	})

	want := report.RenderableNodes{"foo": {ID: "foo"}, "bar": {ID: "bar"}}
	have := renderer.Render(report.MakeReport())

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestReduceEdge(t *testing.T) {
	renderer := Reduce([]Renderer{
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"foo": 1}},
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"bar": 2}},
	})

	want := report.AggregateMetadata{"foo": 1, "bar": 2}
	have := renderer.AggregateMetadata(report.MakeReport(), "", "")

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

var (
	clientHostID  = "client.hostname.com"
	serverHostID  = "server.hostname.com"
	randomHostID  = "random.hostname.com"
	unknownHostID = ""

	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	randomHostNodeID = report.MakeHostNodeID(randomHostID)

	client54001    = report.MakeEndpointNodeID(clientHostID, "10.10.10.20", "54001") // curl (1)
	client54002    = report.MakeEndpointNodeID(clientHostID, "10.10.10.20", "54002") // curl (2)
	unknownClient1 = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54010") // we want to ensure two unknown clients, connnected
	unknownClient2 = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54020") // to the same server, are deduped.
	unknownClient3 = report.MakeEndpointNodeID(serverHostID, "10.10.10.11", "54020") // Check this one isn't deduped
	server80       = report.MakeEndpointNodeID(serverHostID, "192.168.1.1", "80")    // apache

	clientIP  = report.MakeAddressNodeID(clientHostID, "10.10.10.20")
	serverIP  = report.MakeAddressNodeID(serverHostID, "192.168.1.1")
	randomIP  = report.MakeAddressNodeID(randomHostID, "172.16.11.9") // only in Address topology
	unknownIP = report.MakeAddressNodeID(unknownHostID, "10.10.10.10")
)

var (
	rpt = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(client54001): report.MakeIDList(server80),
				report.MakeAdjacencyID(client54002): report.MakeIDList(server80),
				report.MakeAdjacencyID(server80):    report.MakeIDList(client54001, client54002, unknownClient1, unknownClient2, unknownClient3),
			},
			NodeMetadatas: report.NodeMetadatas{
				// NodeMetadata is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				client54001: report.NodeMetadata{
					"name":            "curl",
					"domain":          "client-54001-domain",
					"pid":             "10001",
					report.HostNodeID: clientHostNodeID,
				},
				client54002: report.NodeMetadata{
					"name":            "curl",                // should be same as above!
					"domain":          "client-54002-domain", // may be different than above
					"pid":             "10001",               // should be same as above!
					report.HostNodeID: clientHostNodeID,
				},
				server80: report.NodeMetadata{
					"name":            "apache",
					"domain":          "server-80-domain",
					"pid":             "215",
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(client54001, server80): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 100,
					BytesEgress:  10,
				},
				report.MakeEdgeID(client54002, server80): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 200,
					BytesEgress:  20,
				},

				report.MakeEdgeID(server80, client54001): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 10,
					BytesEgress:  100,
				},
				report.MakeEdgeID(server80, client54002): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 20,
					BytesEgress:  200,
				},
				report.MakeEdgeID(server80, unknownClient1): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 30,
					BytesEgress:  300,
				},
				report.MakeEdgeID(server80, unknownClient2): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 40,
					BytesEgress:  400,
				},
				report.MakeEdgeID(server80, unknownClient3): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 50,
					BytesEgress:  500,
				},
			},
		},
		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(clientIP): report.MakeIDList(serverIP),
				report.MakeAdjacencyID(randomIP): report.MakeIDList(serverIP),
				report.MakeAdjacencyID(serverIP): report.MakeIDList(clientIP, unknownIP), // no backlink to random
			},
			NodeMetadatas: report.NodeMetadatas{
				clientIP: report.NodeMetadata{
					"name":            "client.hostname.com", // hostname
					report.HostNodeID: clientHostNodeID,
				},
				randomIP: report.NodeMetadata{
					"name":            "random.hostname.com", // hostname
					report.HostNodeID: randomHostNodeID,
				},
				serverIP: report.NodeMetadata{
					"name":            "server.hostname.com", // hostname
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(clientIP, serverIP): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(randomIP, serverIP): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  20, // dangling connections, weird but possible
				},
				report.MakeEdgeID(serverIP, clientIP): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(serverIP, unknownIP): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  7,
				},
			},
		},
	}
)

func TestRenderByEndpointPID(t *testing.T) {
	want := report.RenderableNodes{
		"pid:client-54001-domain:10001": {
			ID:         "pid:client-54001-domain:10001",
			LabelMajor: "curl",
			LabelMinor: "client-54001-domain (10001)",
			Rank:       "10001",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("pid:server-80-domain:215"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54001")),
			Metadata: report.AggregateMetadata{
				report.KeyBytesIngress: 100,
				report.KeyBytesEgress:  10,
			},
		},
		"pid:client-54002-domain:10001": {
			ID:         "pid:client-54002-domain:10001",
			LabelMajor: "curl",
			LabelMinor: "client-54002-domain (10001)",
			Rank:       "10001", // same process
			Pseudo:     false,
			Adjacency:  report.MakeIDList("pid:server-80-domain:215"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54002")),
			Metadata: report.AggregateMetadata{
				report.KeyBytesIngress: 200,
				report.KeyBytesEgress:  20,
			},
		},
		"pid:server-80-domain:215": {
			ID:         "pid:server-80-domain:215",
			LabelMajor: "apache",
			LabelMinor: "server-80-domain (215)",
			Rank:       "215",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"pid:client-54001-domain:10001",
				"pid:client-54002-domain:10001",
				"pseudo;10.10.10.10;192.168.1.1;80",
				"pseudo;10.10.10.11;192.168.1.1;80",
			),
			Origins: report.MakeIDList(report.MakeHostNodeID("server.hostname.com"), report.MakeEndpointNodeID("server.hostname.com", "192.168.1.1", "80")),
			Metadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		"pseudo;10.10.10.10;192.168.1.1;80": {
			ID:         "pseudo;10.10.10.10;192.168.1.1;80",
			LabelMajor: "10.10.10.10",
			Pseudo:     true,
			Metadata:   report.AggregateMetadata{},
		},
		"pseudo;10.10.10.11;192.168.1.1;80": {
			ID:         "pseudo;10.10.10.11;192.168.1.1;80",
			LabelMajor: "10.10.10.11",
			Pseudo:     true,
			Metadata:   report.AggregateMetadata{},
		},
	}
	have := renderTopology(rpt.Endpoint, ProcessPID, GenericPseudoNode)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestRenderByEndpointPIDGrouped(t *testing.T) {
	// For grouped, I've somewhat arbitrarily chosen to squash together all
	// processes with the same name by removing the PID and domain (host)
	// dimensions from the ID. That could be changed.
	want := report.RenderableNodes{
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "",
			Rank:       "curl",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("apache"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54001"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54002")),
			Metadata: report.AggregateMetadata{
				report.KeyBytesIngress: 300,
				report.KeyBytesEgress:  30,
			},
		},
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "",
			Rank:       "apache",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"curl",
				"pseudo;10.10.10.10;apache",
				"pseudo;10.10.10.11;apache",
			),
			Origins: report.MakeIDList(report.MakeHostNodeID("server.hostname.com"), report.MakeEndpointNodeID("server.hostname.com", "192.168.1.1", "80")),
			Metadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		"pseudo;10.10.10.10;apache": {
			ID:         "pseudo;10.10.10.10;apache",
			LabelMajor: "10.10.10.10",
			Pseudo:     true,
			Metadata:   report.AggregateMetadata{},
		},
		"pseudo;10.10.10.11;apache": {
			ID:         "pseudo;10.10.10.11;apache",
			LabelMajor: "10.10.10.11",
			Pseudo:     true,
			Metadata:   report.AggregateMetadata{},
		},
	}
	have := renderTopology(rpt.Endpoint, ProcessName, GenericGroupedPseudoNode)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestRenderByNetworkHostname(t *testing.T) {
	want := report.RenderableNodes{
		"host:client.hostname.com": {
			ID:         "host:client.hostname.com",
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "client",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:server.hostname.com"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeAddressNodeID("client.hostname.com", "10.10.10.20")),
			Metadata: report.AggregateMetadata{
				report.KeyMaxConnCountTCP: 3,
			},
		},
		"host:random.hostname.com": {
			ID:         "host:random.hostname.com",
			LabelMajor: "random",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "random",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:server.hostname.com"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("random.hostname.com"), report.MakeAddressNodeID("random.hostname.com", "172.16.11.9")),
			Metadata: report.AggregateMetadata{
				report.KeyMaxConnCountTCP: 20,
			},
		},
		"host:server.hostname.com": {
			ID:         "host:server.hostname.com",
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "server",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:client.hostname.com", "pseudo;10.10.10.10;192.168.1.1;"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("server.hostname.com"), report.MakeAddressNodeID("server.hostname.com", "192.168.1.1")),
			Metadata: report.AggregateMetadata{
				report.KeyMaxConnCountTCP: 10,
			},
		},
		"pseudo;10.10.10.10;192.168.1.1;": {
			ID:         "pseudo;10.10.10.10;192.168.1.1;",
			LabelMajor: "10.10.10.10",
			LabelMinor: "", // after first .
			Rank:       "",
			Pseudo:     true,
			Adjacency:  nil,
			Origins:    nil,
			Metadata:   report.AggregateMetadata{},
		},
	}
	have := renderTopology(rpt.Address, NetworkHostname, GenericPseudoNode)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func diff(want, have interface{}) string {
	text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(spew.Sdump(want)),
		B:        difflib.SplitLines(spew.Sdump(have)),
		FromFile: "want",
		ToFile:   "have",
		Context:  3,
	})
	return "\n" + text
}
