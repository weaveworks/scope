package render_test

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func init() {
	spew.Config.SortKeys = true // :\
}

var (
	clientHostID  = "client.hostname.com"
	serverHostID  = "server.hostname.com"
	randomHostID  = "random.hostname.com"
	unknownHostID = ""

	clientHostName = clientHostID
	serverHostName = serverHostID

	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	randomHostNodeID = report.MakeHostNodeID(randomHostID)

	client54001NodeID = report.MakeEndpointNodeID(clientHostID, "10.10.10.20", "54001") // curl (1)
	client54002NodeID = report.MakeEndpointNodeID(clientHostID, "10.10.10.20", "54002") // curl (2)
	unknownClient1    = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54010") // we want to ensure two unknown clients, connnected
	unknownClient2    = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54020") // to the same server, are deduped.
	unknownClient3    = report.MakeEndpointNodeID(serverHostID, "10.10.10.11", "54020") // Check this one isn't deduped
	server80          = report.MakeEndpointNodeID(serverHostID, "192.168.1.1", "80")    // apache

	clientAddressNodeID  = report.MakeAddressNodeID(clientHostID, "10.10.10.20")
	serverAddressNodeID  = report.MakeAddressNodeID(serverHostID, "192.168.1.1")
	randomAddressNodeID  = report.MakeAddressNodeID(randomHostID, "172.16.11.9") // only in Address topology
	unknownAddressNodeID = report.MakeAddressNodeID(unknownHostID, "10.10.10.10")
)

var (
	rpt = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(client54001NodeID): report.MakeIDList(server80),
				report.MakeAdjacencyID(client54002NodeID): report.MakeIDList(server80),
				report.MakeAdjacencyID(server80):          report.MakeIDList(client54001NodeID, client54002NodeID, unknownClient1, unknownClient2, unknownClient3),
			},
			NodeMetadatas: report.NodeMetadatas{
				// NodeMetadata is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				client54001NodeID: report.NodeMetadata{
					"name":            "curl",
					"domain":          "client-54001-domain",
					"pid":             "10001",
					report.HostNodeID: clientHostNodeID,
					"host_name":       clientHostName,
				},
				client54002NodeID: report.NodeMetadata{
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
				report.MakeEdgeID(client54001NodeID, server80): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 100,
					BytesEgress:  10,
				},
				report.MakeEdgeID(client54002NodeID, server80): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 200,
					BytesEgress:  20,
				},

				report.MakeEdgeID(server80, client54001NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 10,
					BytesEgress:  100,
				},
				report.MakeEdgeID(server80, client54002NodeID): report.EdgeMetadata{
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
		Process: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeProcessNodeID(clientHostID, "4242"): report.NodeMetadata{
					"host_name":             "client.host.com",
					"pid":                   "4242",
					"comm":                  "curl",
					"docker_container_id":   "a1b2c3d4e5",
					"docker_container_name": "fixture-container",
					"docker_image_id":       "0000000000",
					"docker_image_name":     "fixture/container:latest",
				},
				report.MakeProcessNodeID(serverHostID, "215"): report.NodeMetadata{
					"pid":          "215",
					"process_name": "apache",
				},

				"no-container": report.NodeMetadata{},
			},
			EdgeMetadatas: report.EdgeMetadatas{},
		},
		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(clientAddressNodeID): report.MakeIDList(serverAddressNodeID),
				report.MakeAdjacencyID(randomAddressNodeID): report.MakeIDList(serverAddressNodeID),
				report.MakeAdjacencyID(serverAddressNodeID): report.MakeIDList(clientAddressNodeID, unknownAddressNodeID), // no backlink to random
			},
			NodeMetadatas: report.NodeMetadatas{
				clientAddressNodeID: report.NodeMetadata{
					"name":            "client.hostname.com", // hostname
					"host_name":       "client.hostname.com",
					report.HostNodeID: clientHostNodeID,
				},
				randomAddressNodeID: report.NodeMetadata{
					"name":            "random.hostname.com", // hostname
					report.HostNodeID: randomHostNodeID,
				},
				serverAddressNodeID: report.NodeMetadata{
					"name":            "server.hostname.com", // hostname
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(clientAddressNodeID, serverAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(randomAddressNodeID, serverAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  20, // dangling connections, weird but possible
				},
				report.MakeEdgeID(serverAddressNodeID, clientAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(serverAddressNodeID, unknownAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  7,
				},
			},
		},
		Host: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				serverHostNodeID: report.NodeMetadata{
					"host_name":      serverHostName,
					"local_networks": "10.10.10.0/24",
					"os":             "Linux",
					"load":           "0.01 0.01 0.01",
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{},
		},
	}
)

func trimNodeMetadata(rns render.RenderableNodes) render.RenderableNodes {
	result := render.RenderableNodes{}
	for id, rn := range rns {
		rn.NodeMetadata = nil
		result[id] = rn
	}
	return result
}

func TestProcessRenderer(t *testing.T) {
	want := render.RenderableNodes{
		"pid:client-54001-domain:10001": {
			ID:         "pid:client-54001-domain:10001",
			LabelMajor: "curl",
			LabelMinor: "client-54001-domain (10001)",
			Rank:       "10001",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("pid:server-80-domain:215"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54001")),
			AggregateMetadata: report.AggregateMetadata{
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
			AggregateMetadata: report.AggregateMetadata{
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
			AggregateMetadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		"pseudo;10.10.10.10;192.168.1.1;80": {
			ID:                "pseudo;10.10.10.10;192.168.1.1;80",
			LabelMajor:        "10.10.10.10",
			Pseudo:            true,
			AggregateMetadata: report.AggregateMetadata{},
		},
		"pseudo;10.10.10.11;192.168.1.1;80": {
			ID:                "pseudo;10.10.10.11;192.168.1.1;80",
			LabelMajor:        "10.10.10.11",
			Pseudo:            true,
			AggregateMetadata: report.AggregateMetadata{},
		},
	}
	have := render.ProcessRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	// For grouped, I've somewhat arbitrarily chosen to squash together all
	// processes with the same name by removing the PID and domain (host)
	// dimensions from the ID. That could be changed.
	want := render.RenderableNodes{
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "",
			Rank:       "curl",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("apache"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54001"), report.MakeEndpointNodeID("client.hostname.com", "10.10.10.20", "54002")),
			AggregateMetadata: report.AggregateMetadata{
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
			AggregateMetadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		"pseudo;10.10.10.10;apache": {
			ID:                "pseudo;10.10.10.10;apache",
			LabelMajor:        "10.10.10.10",
			Pseudo:            true,
			AggregateMetadata: report.AggregateMetadata{},
		},
		"pseudo;10.10.10.11;apache": {
			ID:                "pseudo;10.10.10.11;apache",
			LabelMajor:        "10.10.10.11",
			Pseudo:            true,
			AggregateMetadata: report.AggregateMetadata{},
		},
	}
	have := render.ProcessNameRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestRenderByNetworkHostname(t *testing.T) {
	want := render.RenderableNodes{
		"host:client.hostname.com": {
			ID:         "host:client.hostname.com",
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "client",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:server.hostname.com"),
			Origins:    report.MakeIDList(report.MakeHostNodeID("client.hostname.com"), report.MakeAddressNodeID("client.hostname.com", "10.10.10.20")),
			AggregateMetadata: report.AggregateMetadata{
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
			AggregateMetadata: report.AggregateMetadata{
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
			AggregateMetadata: report.AggregateMetadata{
				report.KeyMaxConnCountTCP: 10,
			},
		},
		"pseudo;10.10.10.10;192.168.1.1;": {
			ID:                "pseudo;10.10.10.10;192.168.1.1;",
			LabelMajor:        "10.10.10.10",
			LabelMinor:        "", // after first .
			Rank:              "",
			Pseudo:            true,
			Adjacency:         nil,
			Origins:           nil,
			AggregateMetadata: report.AggregateMetadata{},
		},
	}
	have := render.LeafMap{
		Selector: report.SelectAddress,
		Mapper:   render.NetworkHostname,
		Pseudo:   render.GenericPseudoNode,
	}.Render(rpt)
	have = trimNodeMetadata(have)
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
