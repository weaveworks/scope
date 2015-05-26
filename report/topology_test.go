package report

import (
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func init() {
	spew.Config.SortKeys = true // :\
}

const (
	client54001    = ScopeDelim + "10.10.10.20" + ScopeDelim + "54001" // curl (1)
	client54002    = ScopeDelim + "10.10.10.20" + ScopeDelim + "54002" // curl (2)
	unknownClient1 = ScopeDelim + "10.10.10.10" + ScopeDelim + "54010" // we want to ensure two unknown clients, connnected
	unknownClient2 = ScopeDelim + "10.10.10.10" + ScopeDelim + "54020" // to the same server, are deduped.
	unknownClient3 = ScopeDelim + "10.10.10.11" + ScopeDelim + "54020" // Check this one isn't deduped
	server80       = ScopeDelim + "192.168.1.1" + ScopeDelim + "80"    // apache

	clientIP  = ScopeDelim + "10.10.10.20"
	serverIP  = ScopeDelim + "192.168.1.1"
	randomIP  = ScopeDelim + "172.16.11.9" // only in Network topology
	unknownIP = ScopeDelim + "10.10.10.10"
)

var (
	report = Report{
		Process: Topology{
			Adjacency: Adjacency{
				"client.hostname.com" + IDDelim + client54001: NewIDList(server80),
				"client.hostname.com" + IDDelim + client54002: NewIDList(server80),
				"server.hostname.com" + IDDelim + server80:    NewIDList(client54001, client54002, unknownClient1, unknownClient2, unknownClient3),
			},
			NodeMetadatas: NodeMetadatas{
				// NodeMetadata is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				client54001: NodeMetadata{
					"name":   "curl",
					"domain": "client-54001-domain",
					"pid":    "10001",
				},
				client54002: NodeMetadata{
					"name":   "curl",                // should be same as above!
					"domain": "client-54002-domain", // may be different than above
					"pid":    "10001",               // should be same as above!
				},
				server80: NodeMetadata{
					"name":   "apache",
					"domain": "server-80-domain",
					"pid":    "215",
				},
			},
			EdgeMetadatas: EdgeMetadatas{
				client54001 + IDDelim + server80: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 100,
					BytesEgress:  10,
				},
				client54002 + IDDelim + server80: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 200,
					BytesEgress:  20,
				},

				server80 + IDDelim + client54001: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 10,
					BytesEgress:  100,
				},
				server80 + IDDelim + client54002: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 20,
					BytesEgress:  200,
				},
				server80 + IDDelim + unknownClient1: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 30,
					BytesEgress:  300,
				},
				server80 + IDDelim + unknownClient2: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 40,
					BytesEgress:  400,
				},
				server80 + IDDelim + unknownClient3: EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 50,
					BytesEgress:  500,
				},
			},
		},
		Network: Topology{
			Adjacency: Adjacency{
				"client.hostname.com" + IDDelim + clientIP: NewIDList(serverIP),
				"random.hostname.com" + IDDelim + randomIP: NewIDList(serverIP),
				"server.hostname.com" + IDDelim + serverIP: NewIDList(clientIP, unknownIP), // no backlink to random
			},
			NodeMetadatas: NodeMetadatas{
				clientIP: NodeMetadata{
					"name": "client.hostname.com", // hostname
				},
				randomIP: NodeMetadata{
					"name": "random.hostname.com", // hostname
				},
				serverIP: NodeMetadata{
					"name": "server.hostname.com", // hostname
				},
			},
			EdgeMetadatas: EdgeMetadatas{
				clientIP + IDDelim + serverIP: EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				randomIP + IDDelim + serverIP: EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  20, // dangling connections, weird but possible
				},
				serverIP + IDDelim + clientIP: EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				serverIP + IDDelim + unknownIP: EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  7,
				},
			},
		},
	}
)

func TestRenderByProcessPID(t *testing.T) {
	want := map[string]RenderableNode{
		"pid:client-54001-domain:10001": {
			ID:          "pid:client-54001-domain:10001",
			LabelMajor:  "curl",
			LabelMinor:  "client-54001-domain (10001)",
			Rank:        "10001",
			Pseudo:      false,
			Adjacency:   NewIDList("pid:server-80-domain:215"),
			OriginHosts: NewIDList("client.hostname.com"),
			OriginNodes: NewIDList(";10.10.10.20;54001"),
			Metadata: AggregateMetadata{
				KeyBytesIngress: 100,
				KeyBytesEgress:  10,
			},
		},
		"pid:client-54002-domain:10001": {
			ID:          "pid:client-54002-domain:10001",
			LabelMajor:  "curl",
			LabelMinor:  "client-54002-domain (10001)",
			Rank:        "10001", // same process
			Pseudo:      false,
			Adjacency:   NewIDList("pid:server-80-domain:215"),
			OriginHosts: NewIDList("client.hostname.com"),
			OriginNodes: NewIDList(";10.10.10.20;54002"),
			Metadata: AggregateMetadata{
				KeyBytesIngress: 200,
				KeyBytesEgress:  20,
			},
		},
		"pid:server-80-domain:215": {
			ID:         "pid:server-80-domain:215",
			LabelMajor: "apache",
			LabelMinor: "server-80-domain (215)",
			Rank:       "215",
			Pseudo:     false,
			Adjacency: NewIDList(
				"pid:client-54001-domain:10001",
				"pid:client-54002-domain:10001",
				"pseudo:;10.10.10.10;192.168.1.1;80",
				"pseudo:;10.10.10.11;192.168.1.1;80",
			),
			OriginHosts: NewIDList("server.hostname.com"),
			OriginNodes: NewIDList(";192.168.1.1;80"),
			Metadata: AggregateMetadata{
				KeyBytesIngress: 150,
				KeyBytesEgress:  1500,
			},
		},
		"pseudo:;10.10.10.10;192.168.1.1;80": {
			ID:          "pseudo:;10.10.10.10;192.168.1.1;80",
			LabelMajor:  "10.10.10.10",
			LabelMinor:  "",
			Rank:        "",
			Pseudo:      true,
			Adjacency:   nil,
			OriginHosts: nil,
			OriginNodes: nil,
			Metadata:    AggregateMetadata{},
		},
		"pseudo:;10.10.10.11;192.168.1.1;80": {
			ID:          "pseudo:;10.10.10.11;192.168.1.1;80",
			LabelMajor:  "10.10.10.11",
			LabelMinor:  "",
			Rank:        "",
			Pseudo:      true,
			Adjacency:   nil,
			OriginHosts: nil,
			OriginNodes: nil,
			Metadata:    AggregateMetadata{},
		},
	}
	have := report.Process.RenderBy(ProcessPID, false)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestRenderByProcessPIDGrouped(t *testing.T) {
	// For grouped, I've somewhat arbitrarily chosen to squash together all
	// processes with the same name by removing the PID and domain (host)
	// dimensions from the ID. That could be changed.
	want := map[string]RenderableNode{
		"curl": {
			ID:          "curl",
			LabelMajor:  "curl",
			LabelMinor:  "",
			Rank:        "10001",
			Pseudo:      false,
			Adjacency:   NewIDList("apache"),
			OriginHosts: NewIDList("client.hostname.com"),
			OriginNodes: NewIDList(";10.10.10.20;54001", ";10.10.10.20;54002"),
			Metadata: AggregateMetadata{
				KeyBytesIngress: 300,
				KeyBytesEgress:  30,
			},
		},
		"apache": {
			ID:          "apache",
			LabelMajor:  "apache",
			LabelMinor:  "",
			Rank:        "215",
			Pseudo:      false,
			Adjacency:   NewIDList("curl", "localUnknown"),
			OriginHosts: NewIDList("server.hostname.com"),
			OriginNodes: NewIDList(";192.168.1.1;80"),
			Metadata: AggregateMetadata{
				KeyBytesIngress: 150,
				KeyBytesEgress:  1500,
			},
		},
		"localUnknown": {
			ID:          "localUnknown",
			LabelMajor:  "",
			LabelMinor:  "",
			Rank:        "",
			Pseudo:      true,
			Adjacency:   nil,
			OriginHosts: nil,
			OriginNodes: nil,
			Metadata:    AggregateMetadata{},
		},
	}
	have := report.Process.RenderBy(ProcessPID, true)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestRenderByNetworkHostname(t *testing.T) {
	want := map[string]RenderableNode{
		"host:client.hostname.com": {
			ID:          "host:client.hostname.com",
			LabelMajor:  "client",       // before first .
			LabelMinor:  "hostname.com", // after first .
			Rank:        "client",
			Pseudo:      false,
			Adjacency:   NewIDList("host:server.hostname.com"),
			OriginHosts: NewIDList("client.hostname.com"),
			OriginNodes: NewIDList(";10.10.10.20"),
			Metadata: AggregateMetadata{
				KeyMaxConnCountTCP: 3,
			},
		},
		"host:random.hostname.com": {
			ID:          "host:random.hostname.com",
			LabelMajor:  "random",       // before first .
			LabelMinor:  "hostname.com", // after first .
			Rank:        "random",
			Pseudo:      false,
			Adjacency:   NewIDList("host:server.hostname.com"),
			OriginHosts: NewIDList("random.hostname.com"),
			OriginNodes: NewIDList(";172.16.11.9"),
			Metadata: AggregateMetadata{
				KeyMaxConnCountTCP: 20,
			},
		},
		"host:server.hostname.com": {
			ID:          "host:server.hostname.com",
			LabelMajor:  "server",       // before first .
			LabelMinor:  "hostname.com", // after first .
			Rank:        "server",
			Pseudo:      false,
			Adjacency:   NewIDList("host:client.hostname.com", "pseudo:;10.10.10.10;192.168.1.1;"),
			OriginHosts: NewIDList("server.hostname.com"),
			OriginNodes: NewIDList(";192.168.1.1"),
			Metadata: AggregateMetadata{
				KeyMaxConnCountTCP: 10,
			},
		},
		"pseudo:;10.10.10.10;192.168.1.1;": {
			ID:          "pseudo:;10.10.10.10;192.168.1.1;",
			LabelMajor:  "10.10.10.10",
			LabelMinor:  "", // after first .
			Rank:        "",
			Pseudo:      true,
			Adjacency:   nil,
			OriginHosts: nil,
			OriginNodes: nil,
			Metadata:    AggregateMetadata{},
		},
	}
	have := report.Network.RenderBy(NetworkHostname, false)
	if !reflect.DeepEqual(want, have) {
		t.Error("\n" + diff(want, have))
	}
}

func TestTopoDiff(t *testing.T) {
	nodea := RenderableNode{
		ID:         "nodea",
		LabelMajor: "Node A",
		LabelMinor: "'ts an a",
		Pseudo:     false,
		Adjacency: []string{
			"nodeb",
		},
	}
	nodeap := nodea
	nodeap.Adjacency = []string{
		"nodeb",
		"nodeq", // not the same anymore
	}
	nodeb := RenderableNode{
		ID:         "nodeb",
		LabelMajor: "Node B",
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...RenderableNode) map[string]RenderableNode {
		r := map[string]RenderableNode{}
		for _, n := range ns {
			r[n.ID] = n
		}
		return r
	}

	for _, c := range []struct {
		label      string
		have, want Diff
	}{
		{
			label: "basecase: empty -> something",
			have:  TopoDiff(nodes(), nodes(nodea, nodeb)),
			want: Diff{
				Add: []RenderableNode{nodea, nodeb},
			},
		},
		{
			label: "basecase: something -> empty",
			have:  TopoDiff(nodes(nodea, nodeb), nodes()),
			want: Diff{
				Remove: []string{"nodea", "nodeb"},
			},
		},
		{
			label: "add and remove",
			have:  TopoDiff(nodes(nodea), nodes(nodeb)),
			want: Diff{
				Add:    []RenderableNode{nodeb},
				Remove: []string{"nodea"},
			},
		},
		{
			label: "no change",
			have:  TopoDiff(nodes(nodea), nodes(nodea)),
			want:  Diff{},
		},
		{
			label: "change a single node",
			have:  TopoDiff(nodes(nodea), nodes(nodeap)),
			want: Diff{
				Update: []RenderableNode{nodeap},
			},
		},
	} {
		sort.Strings(c.have.Remove)
		sort.Sort(ByID(c.have.Add))
		sort.Sort(ByID(c.have.Update))
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s\n%s", c.label, diff(c.want, c.have))
		}
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
	return text
}
