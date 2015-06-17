package render_test

import (
	"fmt"
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

	clientIP        = "10.10.10.20"
	serverIP        = "192.168.1.1"
	clientPort54001 = "54001"
	clientPort54002 = "54002"
	serverPort      = "80"

	clientHostName = clientHostID
	serverHostName = serverHostID

	clientPID       = "10001"
	serverPID       = "215"
	nonContainerPID = "1234"

	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	randomHostNodeID = report.MakeHostNodeID(randomHostID)

	client54001NodeID    = report.MakeEndpointNodeID(clientHostID, clientIP, clientPort54001) // curl (1)
	client54002NodeID    = report.MakeEndpointNodeID(clientHostID, clientIP, clientPort54002) // curl (2)
	unknownClient1NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54010")    // we want to ensure two unknown clients, connnected
	unknownClient2NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54020")    // to the same server, are deduped.
	unknownClient3NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.11", "54020")    // Check this one isn't deduped
	server80NodeID       = report.MakeEndpointNodeID(serverHostID, serverIP, serverPort)      // apache

	clientAddressNodeID  = report.MakeAddressNodeID(clientHostID, "10.10.10.20")
	serverAddressNodeID  = report.MakeAddressNodeID(serverHostID, "192.168.1.1")
	randomAddressNodeID  = report.MakeAddressNodeID(randomHostID, "172.16.11.9") // only in Address topology
	unknownAddressNodeID = report.MakeAddressNodeID(unknownHostID, "10.10.10.10")

	clientProcessNodeID       = report.MakeProcessNodeID(clientHostID, clientPID)
	serverProcessNodeID       = report.MakeProcessNodeID(serverHostID, serverPID)
	nonContainerProcessNodeID = report.MakeProcessNodeID(serverHostID, nonContainerPID)
)

var (
	rpt = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(client54001NodeID): report.MakeIDList(server80NodeID),
				report.MakeAdjacencyID(client54002NodeID): report.MakeIDList(server80NodeID),
				report.MakeAdjacencyID(server80NodeID):    report.MakeIDList(client54001NodeID, client54002NodeID, unknownClient1NodeID, unknownClient2NodeID, unknownClient3NodeID),
			},
			NodeMetadatas: report.NodeMetadatas{
				// NodeMetadata is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				client54001NodeID: report.NodeMetadata{
					"addr":            clientIP,
					"port":            clientPort54001,
					"pid":             clientPID,
					report.HostNodeID: clientHostNodeID,
				},
				client54002NodeID: report.NodeMetadata{
					"addr":            clientIP,
					"port":            clientPort54002,
					"pid":             clientPID, // should be same as above!
					report.HostNodeID: clientHostNodeID,
				},
				server80NodeID: report.NodeMetadata{
					"addr":            serverIP,
					"port":            serverPort,
					"pid":             serverPID,
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(client54001NodeID, server80NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 100,
					BytesEgress:  10,
				},
				report.MakeEdgeID(client54002NodeID, server80NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 200,
					BytesEgress:  20,
				},

				report.MakeEdgeID(server80NodeID, client54001NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 10,
					BytesEgress:  100,
				},
				report.MakeEdgeID(server80NodeID, client54002NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 20,
					BytesEgress:  200,
				},
				report.MakeEdgeID(server80NodeID, unknownClient1NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 30,
					BytesEgress:  300,
				},
				report.MakeEdgeID(server80NodeID, unknownClient2NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 40,
					BytesEgress:  400,
				},
				report.MakeEdgeID(server80NodeID, unknownClient3NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 50,
					BytesEgress:  500,
				},
			},
		},
		Process: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				clientProcessNodeID: report.NodeMetadata{
					"pid":                 clientPID,
					"comm":                "curl",
					"docker_container_id": "a1b2c3d4e5",
					report.HostNodeID:     clientHostNodeID,
				},
				serverProcessNodeID: report.NodeMetadata{
					"pid":                 serverPID,
					"comm":                "apache",
					"docker_container_id": "5e4d3c2b1a",
					report.HostNodeID:     serverHostNodeID,
				},
				nonContainerProcessNodeID: report.NodeMetadata{
					"pid":             nonContainerPID,
					"comm":            "bash",
					report.HostNodeID: serverHostNodeID,
				},
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

func init() {
	if err := rpt.Validate(); err != nil {
		panic(err)
	}
}

func trimNodeMetadata(rns render.RenderableNodes) render.RenderableNodes {
	result := render.RenderableNodes{}
	for id, rn := range rns {
		rn.NodeMetadata = nil
		result[id] = rn
	}
	return result
}

func TestProcessRenderer(t *testing.T) {
	var (
		clientProcessID       = fmt.Sprintf("pid:%s:%s", clientHostID, clientPID)
		serverProcessID       = fmt.Sprintf("pid:%s:%s", serverHostID, serverPID)
		nonContainerProcessID = fmt.Sprintf("pid:%s:%s", serverHostID, nonContainerPID)
	)

	want := render.RenderableNodes{
		clientProcessID: {
			ID:         clientProcessID,
			LabelMajor: "curl",
			LabelMinor: fmt.Sprintf("%s (%s)", clientHostID, clientPID),
			Rank:       clientPID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(serverProcessID),
			Origins:    report.MakeIDList(client54001NodeID, client54002NodeID, clientProcessNodeID, clientHostNodeID),
			AggregateMetadata: report.AggregateMetadata{
				report.KeyBytesIngress: 300,
				report.KeyBytesEgress:  30,
			},
		},
		serverProcessID: {
			ID:         serverProcessID,
			LabelMajor: "apache",
			LabelMinor: fmt.Sprintf("%s (%s)", serverHostID, serverPID),
			Rank:       serverPID,
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				clientProcessID,
				"pseudo;10.10.10.10;192.168.1.1;80",
				"pseudo;10.10.10.11;192.168.1.1;80",
			),
			Origins: report.MakeIDList(server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		nonContainerProcessID: {
			ID:                nonContainerProcessID,
			LabelMajor:        "bash",
			LabelMinor:        fmt.Sprintf("%s (%s)", serverHostID, nonContainerPID),
			Rank:              nonContainerPID,
			Pseudo:            false,
			Adjacency:         report.MakeIDList(),
			Origins:           report.MakeIDList(nonContainerProcessNodeID, serverHostNodeID),
			AggregateMetadata: report.AggregateMetadata{},
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
			Origins:    report.MakeIDList(client54001NodeID, client54002NodeID, clientProcessNodeID, clientHostNodeID),
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
				"pseudo;10.10.10.10;192.168.1.1;80",
				"pseudo;10.10.10.11;192.168.1.1;80",
			),
			Origins: report.MakeIDList(server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: report.AggregateMetadata{
				report.KeyBytesIngress: 150,
				report.KeyBytesEgress:  1500,
			},
		},
		"bash": {
			ID:                "bash",
			LabelMajor:        "bash",
			LabelMinor:        "",
			Rank:              "bash",
			Pseudo:            false,
			Origins:           report.MakeIDList(nonContainerProcessNodeID, serverHostNodeID),
			AggregateMetadata: report.AggregateMetadata{},
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
