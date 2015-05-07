package report

import (
	"reflect"
	"sort"
	"testing"
)

var report = Report{
	Process: Topology{
		Adjacency: Adjacency{
			"hostA|;192.168.1.1;12345": NewIDList(";192.168.1.2;80"),
			"hostA|;192.168.1.1;12346": NewIDList(";192.168.1.2;80"),
			"hostA|;192.168.1.1;8888":  NewIDList(";1.2.3.4;22"),
			"hostB|;192.168.1.2;80":    NewIDList(";192.168.1.1;12345"),
			"hostB|;192.168.1.2;43201": NewIDList(";1.2.3.5;22"),
		},
		EdgeMetadatas: EdgeMetadatas{
			";192.168.1.1;12345|;192.168.1.2;80": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  12,
				BytesIngress: 0,
			},
			";192.168.1.1;12346|;192.168.1.2;80": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  12,
				BytesIngress: 0,
			},
			";192.168.1.1;8888|;1.2.3.4;22": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  200,
				BytesIngress: 0,
			},
			";192.168.1.2;80|;192.168.1.1;12345": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  0,
				BytesIngress: 12,
			},
			";192.168.1.2;43201|;1.2.3.5;22": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  200,
				BytesIngress: 12,
			},
		},
		NodeMetadatas: NodeMetadatas{
			";192.168.1.1;12345": NodeMetadata{
				"pid":    "23128",
				"name":   "curl",
				"domain": "node-a.local",
			},
			";192.168.1.1;12346": NodeMetadata{ // <-- same as :12345
				"pid":    "23128",
				"name":   "curl",
				"domain": "node-a.local",
			},
			";192.168.1.1;8888": NodeMetadata{
				"pid":    "55100",
				"name":   "ssh",
				"domain": "node-a.local",
			},
			";192.168.1.2;80": NodeMetadata{
				"pid":    "215",
				"name":   "apache",
				"domain": "node-b.local",
			},
			";192.168.1.2;43201": NodeMetadata{
				"pid":    "8765",
				"name":   "ssh",
				"domain": "node-b.local",
			},
		},
	},

	Network: Topology{
		Adjacency: Adjacency{
			"hostA|;192.168.1.1": NewIDList(";192.168.1.2", ";1.2.3.4"),
			"hostB|;192.168.1.2": NewIDList(";192.168.1.1", ";1.2.3.5"),
		},
		EdgeMetadatas: EdgeMetadatas{
			";192.168.1.1|;192.168.1.2": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  12,
				BytesIngress: 0,
			},
			";192.168.1.1|;1.2.3.4": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  200,
				BytesIngress: 0,
			},
			";192.168.1.2|;192.168.1.1": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  0,
				BytesIngress: 12,
			},
			";192.168.1.2|;1.2.3.5": EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  200,
				BytesIngress: 12,
			},
		},
		NodeMetadatas: NodeMetadatas{
			";192.168.1.1": NodeMetadata{
				"name": "host-a",
			},
			";192.168.1.2": NodeMetadata{
				"name": "host-b",
			},
		},
	},

	HostMetadatas: HostMetadatas{
		"hostA": HostMetadata{
			Hostname: "node-a.local",
			OS:       "Linux",
		},
		"hostB": HostMetadata{
			Hostname: "node-b.local",
			OS:       "Linux",
		},
	},
}

func TestTopologyProc(t *testing.T) {
	// Process topology with by-processname mapping
	{

		if want, have := map[string]DetailedRenderableNode{
			"proc:node-b.local:apache": {
				RenderableNode: RenderableNode{
					ID:         "proc:node-b.local:apache",
					LabelMajor: "apache",
					LabelMinor: "node-b.local",
					Rank:       "apache",
					Pseudo:     false,
					Adjacency:  NewIDList("proc:node-a.local:curl"),
					Origin:     NewIDList("hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  0,
					"ingress_bytes": 12,
				},
			},
			"proc:node-a.local:curl": {
				RenderableNode: RenderableNode{
					ID:         "proc:node-a.local:curl",
					LabelMajor: "curl",
					LabelMinor: "node-a.local",
					Rank:       "curl",
					Pseudo:     false,
					Adjacency:  NewIDList("proc:node-b.local:apache"),
					Origin:     NewIDList("hostA"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  24,
					"ingress_bytes": 0,
				},
			},
			"proc:node-a.local:ssh": {
				RenderableNode: RenderableNode{
					ID:         "proc:node-a.local:ssh",
					LabelMajor: "ssh",
					LabelMinor: "node-a.local",
					Rank:       "ssh",
					Pseudo:     false,
					Adjacency:  NewIDList("pseudo:;1.2.3.4;22"),
					Origin:     NewIDList("hostA"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  200,
					"ingress_bytes": 0,
				},
			},
			"proc:node-b.local:ssh": {
				RenderableNode: RenderableNode{
					ID:         "proc:node-b.local:ssh",
					LabelMajor: "ssh",
					LabelMinor: "node-b.local",
					Rank:       "ssh",
					Pseudo:     false,
					Adjacency:  NewIDList("pseudo:;1.2.3.5;22"),
					Origin:     NewIDList("hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  200,
					"ingress_bytes": 12,
				},
			},
			"pseudo:;1.2.3.4;22": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.4;22",
					LabelMajor: "1.2.3.4:22",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
			"pseudo:;1.2.3.5;22": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.5;22",
					LabelMajor: "1.2.3.5:22",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
		}, report.Process.RenderBy(ProcessName, false); !reflect.DeepEqual(want, have) {
			t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
		}
	}

	// check EdgeMetadata
	{
		want := EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  0,
			BytesIngress: 12,
		}
		have := report.Process.EdgeMetadata(
			ProcessName,
			false,
			"proc:node-b.local:apache",
			"proc:node-a.local:curl",
		)
		if want != have {
			t.Errorf("Topology error. Want:\n%#v\nHave:\n%#v\n", want, have)
		}
	}
}

func TestTopologyProcClass(t *testing.T) {
	// Process name classes.
	{
		if want, have := map[string]DetailedRenderableNode{
			"proc::apache": {
				RenderableNode: RenderableNode{
					ID:         "proc::apache",
					LabelMajor: "apache",
					LabelMinor: "",
					Rank:       "apache",
					Pseudo:     false,
					Adjacency:  NewIDList("proc::curl"),
					Origin:     NewIDList("hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  0,
					"ingress_bytes": 12,
				},
			},
			"proc::curl": {
				RenderableNode: RenderableNode{
					ID:         "proc::curl",
					LabelMajor: "curl",
					LabelMinor: "",
					Rank:       "curl",
					Pseudo:     false,
					Adjacency:  NewIDList("proc::apache"),
					Origin:     NewIDList("hostA"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  24,
					"ingress_bytes": 0,
				},
			},
			"proc::ssh": {
				RenderableNode: RenderableNode{
					ID:         "proc::ssh",
					LabelMajor: "ssh",
					LabelMinor: "",
					Rank:       "ssh",
					Pseudo:     false,
					Adjacency:  NewIDList("localunknown"),
					Origin:     NewIDList("hostA", "hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  400,
					"ingress_bytes": 12,
				},
			},
			"localunknown": {
				RenderableNode: RenderableNode{
					ID:         "localunknown",
					LabelMajor: "",
					LabelMinor: "",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
		}, report.Process.RenderBy(ProcessName, true); !reflect.DeepEqual(want, have) {
			t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
		}
	}

	// check EdgeMetadata
	{
		want := EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  0,
			BytesIngress: 12,
		}
		have := report.Process.EdgeMetadata(
			ProcessName,
			true, // class view
			"proc::apache",
			"proc::curl",
		)
		if want != have {
			t.Errorf("Topology error. Want:\n%#v\nHave:\n%#v\n", want, have)
		}
	}
}

func TestTopologyHost(t *testing.T) {
	// Network topology with by-hostname mapping
	{
		want := map[string]DetailedRenderableNode{
			"host:host-a": {
				RenderableNode: RenderableNode{
					ID:         "host:host-a",
					LabelMajor: "host-a",
					Rank:       "host-a",
					Pseudo:     false,
					Adjacency: NewIDList(
						"pseudo:;1.2.3.4",
						"host:host-b",
					),
					Origin: NewIDList("hostA"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  212,
					"ingress_bytes": 0,
				},
			},
			"host:host-b": {
				RenderableNode: RenderableNode{
					ID:         "host:host-b",
					LabelMajor: "host-b",
					Rank:       "host-b",
					Pseudo:     false,
					Adjacency: NewIDList(
						"host:host-a",
						"pseudo:;1.2.3.5",
					),
					Origin: NewIDList("hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  200,
					"ingress_bytes": 24,
				},
			},
			"pseudo:;1.2.3.4": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.4",
					LabelMajor: "1.2.3.4",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
			"pseudo:;1.2.3.5": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.5",
					LabelMajor: "1.2.3.5",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
		}

		have := report.Network.RenderBy(NetworkHostname, false)

		sort.Strings(have["net:host-a"].Adjacency)

		if !reflect.DeepEqual(want, have) {
			t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
		}
	}

	// check EdgeMetadata
	{
		want := EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  0,
			BytesIngress: 12,
		}
		have := report.Network.EdgeMetadata(
			NetworkHostname,
			false,
			"host:host-b",
			"host:host-a",
		)
		if want != have {
			t.Errorf("Topology error. Want:\n%#v\nHave:\n%#v\n", want, have)
		}
	}
}

func TestTopologyIP(t *testing.T) {
	// Network topology with by-IP mapping
	{
		want := map[string]DetailedRenderableNode{
			"addr:;192.168.1.1": {
				RenderableNode: RenderableNode{
					ID:         "addr:;192.168.1.1",
					LabelMajor: "192.168.1.1",
					LabelMinor: "host-a",
					Rank:       "192.168.1.1",
					Pseudo:     false,
					Adjacency: NewIDList(
						"pseudo:;1.2.3.4",
						"addr:;192.168.1.2",
					),
					Origin: NewIDList("hostA"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  212,
					"ingress_bytes": 0,
				},
			},
			"addr:;192.168.1.2": {
				RenderableNode: RenderableNode{
					ID:         "addr:;192.168.1.2",
					LabelMajor: "192.168.1.2",
					LabelMinor: "host-b",
					Rank:       "192.168.1.2",
					Pseudo:     false,
					Adjacency: NewIDList(
						"pseudo:;1.2.3.5",
						"addr:;192.168.1.1",
					),
					Origin: NewIDList("hostB"),
				},
				Aggregate: RenderableMetadata{
					"egress_bytes":  200,
					"ingress_bytes": 24,
				},
			},
			"pseudo:;1.2.3.4": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.4",
					LabelMajor: "1.2.3.4",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
			"pseudo:;1.2.3.5": {
				RenderableNode: RenderableNode{
					ID:         "pseudo:;1.2.3.5",
					LabelMajor: "1.2.3.5",
					Pseudo:     true,
				},
				Aggregate: RenderableMetadata{},
			},
		}
		have := report.Network.RenderBy(NetworkIP, false)
		sort.Strings(have["pseudo:;192.168.1.1"].Adjacency)
		if !reflect.DeepEqual(want, have) {
			t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
		}
	}

	// check EdgeMetadata
	{
		want := EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  12,
			BytesIngress: 0,
		}
		have := report.Network.EdgeMetadata(
			NetworkIP,
			false,
			"addr:;192.168.1.1",
			"addr:;192.168.1.2",
		)
		if want != have {
			t.Errorf("Topology error. Want:\n%#v\nHave:\n%#v\n", want, have)
		}
	}
}

func TestTopologyDiff(t *testing.T) {
	// Diff renderable nodes.
	nodea := DetailedRenderableNode{
		RenderableNode: RenderableNode{
			ID:         "nodea",
			LabelMajor: "Node A",
			LabelMinor: "'ts an a",
			Pseudo:     false,
			Adjacency: []string{
				"nodeb",
			},
		},
	}
	nodeap := nodea
	nodeap.Adjacency = []string{
		"nodeb",
		"nodeq", // not the same anymore.
	}
	nodeb := DetailedRenderableNode{
		RenderableNode: RenderableNode{
			ID:         "nodeb",
			LabelMajor: "Node B",
		},
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...DetailedRenderableNode) map[string]DetailedRenderableNode {
		r := map[string]DetailedRenderableNode{}
		for _, n := range ns {
			r[n.RenderableNode.ID] = n
		}
		return r
	}

	for i, c := range []struct {
		have, want Diff
	}{
		{
			// basecase: empty -> something
			have: TopoDiff(nodes(), nodes(nodea, nodeb)),
			want: Diff{
				Add: []DetailedRenderableNode{nodea, nodeb},
			},
		},
		{
			// basecase: something -> empty
			have: TopoDiff(nodes(nodea, nodeb), nodes()),
			want: Diff{
				Remove: []string{"nodea", "nodeb"},
			},
		},
		{
			// add and remove
			have: TopoDiff(nodes(nodea), nodes(nodeb)),
			want: Diff{
				Add:    []DetailedRenderableNode{nodeb},
				Remove: []string{"nodea"},
			},
		},
		{
			// no change.
			have: TopoDiff(nodes(nodea), nodes(nodea)),
			want: Diff{},
		},
		{
			// change a single node
			have: TopoDiff(nodes(nodea), nodes(nodeap)),
			want: Diff{
				Update: []DetailedRenderableNode{nodeap},
			},
		},
	} {
		sort.Strings(c.have.Remove)
		sort.Sort(ByID(c.have.Add))
		sort.Sort(ByID(c.have.Update))
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("case %d: want\n\t%#v, have\n\t%#v", i, c.want, c.have)
		}
	}
}
