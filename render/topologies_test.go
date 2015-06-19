package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

// This is an example Report:
//   2 hosts with probes installed - client & server.

var (
	clientHostID  = "client.hostname.com"
	serverHostID  = "server.hostname.com"
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

	client54001NodeID    = report.MakeEndpointNodeID(clientHostID, clientIP, clientPort54001) // curl (1)
	client54002NodeID    = report.MakeEndpointNodeID(clientHostID, clientIP, clientPort54002) // curl (2)
	server80NodeID       = report.MakeEndpointNodeID(serverHostID, serverIP, serverPort)      // apache
	unknownClient1NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54010")    // we want to ensure two unknown clients, connnected
	unknownClient2NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.10", "54020")    // to the same server, are deduped.
	unknownClient3NodeID = report.MakeEndpointNodeID(serverHostID, "10.10.10.11", "54020")    // Check this one isn't deduped
	randomClientNodeID   = report.MakeEndpointNodeID(serverHostID, "51.52.53.54", "12345")    // this should become an internet node

	clientProcessNodeID       = report.MakeProcessNodeID(clientHostID, clientPID)
	serverProcessNodeID       = report.MakeProcessNodeID(serverHostID, serverPID)
	nonContainerProcessNodeID = report.MakeProcessNodeID(serverHostID, nonContainerPID)

	clientContainerID     = "a1b2c3d4e5"
	serverContainerID     = "5e4d3c2b1a"
	clientContainerNodeID = report.MakeContainerNodeID(clientHostID, clientContainerID)
	serverContainerNodeID = report.MakeContainerNodeID(serverHostID, serverContainerID)

	clientContainerImageID     = "imageid123"
	serverContainerImageID     = "imageid456"
	clientContainerImageNodeID = report.MakeContainerNodeID(clientHostID, clientContainerImageID)
	serverContainerImageNodeID = report.MakeContainerNodeID(serverHostID, serverContainerImageID)

	clientAddressNodeID   = report.MakeAddressNodeID(clientHostID, "10.10.10.20")
	serverAddressNodeID   = report.MakeAddressNodeID(serverHostID, "192.168.1.1")
	unknownAddress1NodeID = report.MakeAddressNodeID(serverHostID, "10.10.10.10")
	unknownAddress2NodeID = report.MakeAddressNodeID(serverHostID, "10.10.10.11")
	randomAddressNodeID   = report.MakeAddressNodeID(serverHostID, "51.52.53.54") // this should become an internet node

	unknownPseudoNode1ID = "pseudo;10.10.10.10;192.168.1.1;80"
	unknownPseudoNode2ID = "pseudo;10.10.10.11;192.168.1.1;80"
	unknownPseudoNode1   = render.RenderableNode{
		ID:                unknownPseudoNode1ID,
		LabelMajor:        "10.10.10.10",
		Pseudo:            true,
		AggregateMetadata: render.AggregateMetadata{},
	}
	unknownPseudoNode2 = render.RenderableNode{
		ID:                unknownPseudoNode2ID,
		LabelMajor:        "10.10.10.11",
		Pseudo:            true,
		AggregateMetadata: render.AggregateMetadata{},
	}
	theInternetNode = render.RenderableNode{
		ID:                render.TheInternetID,
		LabelMajor:        render.TheInternetMajor,
		Pseudo:            true,
		AggregateMetadata: render.AggregateMetadata{},
	}
)

var (
	rpt = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(client54001NodeID): report.MakeIDList(server80NodeID),
				report.MakeAdjacencyID(client54002NodeID): report.MakeIDList(server80NodeID),
				report.MakeAdjacencyID(server80NodeID): report.MakeIDList(
					client54001NodeID, client54002NodeID, unknownClient1NodeID, unknownClient2NodeID,
					unknownClient3NodeID, randomClientNodeID),
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
					"pid":              clientPID,
					"comm":             "curl",
					docker.ContainerID: clientContainerID,
					report.HostNodeID:  clientHostNodeID,
				},
				serverProcessNodeID: report.NodeMetadata{
					"pid":              serverPID,
					"comm":             "apache",
					docker.ContainerID: serverContainerID,
					report.HostNodeID:  serverHostNodeID,
				},
				nonContainerProcessNodeID: report.NodeMetadata{
					"pid":             nonContainerPID,
					"comm":            "bash",
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{},
		},
		Container: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				clientContainerNodeID: report.NodeMetadata{
					docker.ContainerID:   clientContainerID,
					docker.ContainerName: "client",
					docker.ImageID:       clientContainerImageID,
					report.HostNodeID:    clientHostNodeID,
				},
				serverContainerNodeID: report.NodeMetadata{
					docker.ContainerID:   serverContainerID,
					docker.ContainerName: "server",
					docker.ImageID:       serverContainerImageID,
					report.HostNodeID:    serverHostNodeID,
				},
			},
		},
		ContainerImage: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				clientContainerImageNodeID: report.NodeMetadata{
					docker.ImageID:    clientContainerImageID,
					docker.ImageName:  "client_image",
					report.HostNodeID: clientHostNodeID,
				},
				serverContainerImageNodeID: report.NodeMetadata{
					docker.ImageID:    serverContainerImageID,
					docker.ImageName:  "server_image",
					report.HostNodeID: serverHostNodeID,
				},
			},
		},
		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(clientAddressNodeID): report.MakeIDList(serverAddressNodeID),
				report.MakeAdjacencyID(serverAddressNodeID): report.MakeIDList(
					clientAddressNodeID, unknownAddress1NodeID, unknownAddress2NodeID, randomAddressNodeID), // no backlinks to unknown/random
			},
			NodeMetadatas: report.NodeMetadatas{
				clientAddressNodeID: report.NodeMetadata{
					"addr":            clientIP,
					report.HostNodeID: clientHostNodeID,
				},
				serverAddressNodeID: report.NodeMetadata{
					"addr":            serverIP,
					report.HostNodeID: serverHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(clientAddressNodeID, serverAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(serverAddressNodeID, clientAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
			},
		},
		Host: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				clientHostNodeID: report.NodeMetadata{
					"host_name":       clientHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: clientHostNodeID,
				},
				serverHostNodeID: report.NodeMetadata{
					"host_name":       serverHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: serverHostNodeID,
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
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 300,
				render.KeyBytesEgress:  30,
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
				unknownPseudoNode1ID,
				unknownPseudoNode2ID,
				render.TheInternetID,
			),
			Origins: report.MakeIDList(server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 150,
				render.KeyBytesEgress:  1500,
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
			AggregateMetadata: render.AggregateMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1,
		unknownPseudoNode2ID: unknownPseudoNode2,
		render.TheInternetID: theInternetNode,
	}
	have := render.ProcessRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + test.Diff(want, have))
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
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 300,
				render.KeyBytesEgress:  30,
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
				unknownPseudoNode1ID,
				unknownPseudoNode2ID,
				render.TheInternetID,
			),
			Origins: report.MakeIDList(server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 150,
				render.KeyBytesEgress:  1500,
			},
		},
		"bash": {
			ID:                "bash",
			LabelMajor:        "bash",
			LabelMinor:        "",
			Rank:              "bash",
			Pseudo:            false,
			Origins:           report.MakeIDList(nonContainerProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1,
		unknownPseudoNode2ID: unknownPseudoNode2,
		render.TheInternetID: theInternetNode,
	}
	have := render.ProcessNameRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + test.Diff(want, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	want := render.RenderableNodes{
		clientContainerID: {
			ID:         clientContainerID,
			LabelMajor: "client",
			LabelMinor: clientHostName,
			Rank:       clientContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(serverContainerID),
			Origins:    report.MakeIDList(clientContainerNodeID, client54001NodeID, client54002NodeID, clientProcessNodeID, clientHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 300,
				render.KeyBytesEgress:  30,
			},
		},
		serverContainerID: {
			ID:         serverContainerID,
			LabelMajor: "server",
			LabelMinor: serverHostName,
			Rank:       serverContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(clientContainerID, render.UncontainedID, render.TheInternetID),
			Origins:    report.MakeIDList(serverContainerNodeID, server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 150,
				render.KeyBytesEgress:  1500,
			},
		},
		render.UncontainedID: {
			ID:                render.UncontainedID,
			LabelMajor:        render.UncontainedMajor,
			LabelMinor:        "",
			Rank:              "",
			Pseudo:            true,
			Origins:           report.MakeIDList(nonContainerProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}
	have := render.ContainerRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	want := render.RenderableNodes{
		clientContainerImageID: {
			ID:         clientContainerImageID,
			LabelMajor: "client_image",
			LabelMinor: "",
			Rank:       clientContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(serverContainerImageID),
			Origins:    report.MakeIDList(clientContainerImageNodeID, clientContainerNodeID, client54001NodeID, client54002NodeID, clientProcessNodeID, clientHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 300,
				render.KeyBytesEgress:  30,
			},
		},
		serverContainerImageID: {
			ID:         serverContainerImageID,
			LabelMajor: "server_image",
			LabelMinor: "",
			Rank:       serverContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(clientContainerImageID, render.UncontainedID, render.TheInternetID),
			Origins:    report.MakeIDList(serverContainerImageNodeID, serverContainerNodeID, server80NodeID, serverProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyBytesIngress: 150,
				render.KeyBytesEgress:  1500,
			},
		},
		render.UncontainedID: {
			ID:                render.UncontainedID,
			LabelMajor:        render.UncontainedMajor,
			LabelMinor:        "",
			Rank:              "",
			Pseudo:            true,
			Origins:           report.MakeIDList(nonContainerProcessNodeID, serverHostNodeID),
			AggregateMetadata: render.AggregateMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}
	have := render.ContainerImageRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + test.Diff(want, have))
	}
}

func TestHostRenderer(t *testing.T) {
	want := render.RenderableNodes{
		"host:server.hostname.com": {
			ID:         "host:server.hostname.com",
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:client.hostname.com", render.TheInternetID, "pseudo;10.10.10.10;192.168.1.1;", "pseudo;10.10.10.11;192.168.1.1;"),
			Origins:    report.MakeIDList(serverHostNodeID, serverAddressNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyMaxConnCountTCP: 3,
			},
		},
		"host:client.hostname.com": {
			ID:         "host:client.hostname.com",
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("host:server.hostname.com"),
			Origins:    report.MakeIDList(clientHostNodeID, clientAddressNodeID),
			AggregateMetadata: render.AggregateMetadata{
				render.KeyMaxConnCountTCP: 3,
			},
		},
		"pseudo;10.10.10.10;192.168.1.1;": {
			ID:                "pseudo;10.10.10.10;192.168.1.1;",
			LabelMajor:        "10.10.10.10",
			Pseudo:            true,
			AggregateMetadata: render.AggregateMetadata{},
		},
		"pseudo;10.10.10.11;192.168.1.1;": {
			ID:                "pseudo;10.10.10.11;192.168.1.1;",
			LabelMajor:        "10.10.10.11",
			Pseudo:            true,
			AggregateMetadata: render.AggregateMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}
	have := render.HostRenderer.Render(rpt)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + test.Diff(want, have))
	}
}
