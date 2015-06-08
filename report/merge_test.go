package report

import (
	"reflect"
	"testing"
	"time"
)

func TestMergeAdjacency(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want Adjacency
	}{
		"Empty b": {
			a: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    MakeIDList(":192.168.1.1:12345"),
			},
			b: Adjacency{},
			want: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    MakeIDList(":192.168.1.1:12345"),
			},
		},
		"Empty a": {
			a: Adjacency{},
			b: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    MakeIDList(":192.168.1.1:12345"),
			},
			want: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    MakeIDList(":192.168.1.1:12345"),
			},
		},
		"Same address": {
			a: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:80"),
			},
			b: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(":192.168.1.2:8080"),
			},
			want: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(
					":192.168.1.2:80", ":192.168.1.2:8080",
				),
			},
		},
		"No duplicates": {
			a: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(
					":192.168.1.2:80",
					":192.168.1.2:8080",
					":192.168.1.2:555",
				),
			},
			b: Adjacency{
				"hostA|:192.168.1.1:12345": MakeIDList(
					":192.168.1.2:8080",
					":192.168.1.2:80",
					":192.168.1.2:444",
				),
			},
			want: Adjacency{
				"hostA|:192.168.1.1:12345": []string{
					":192.168.1.2:444",
					":192.168.1.2:555",
					":192.168.1.2:80",
					":192.168.1.2:8080",
				},
			},
		},
		"Double keys": {
			a: Adjacency{
				"key1": MakeIDList("a", "c", "d", "b"),
				"key2": MakeIDList("c", "a"),
			},
			b: Adjacency{
				"key1": MakeIDList("a", "b", "e"),
				"key3": MakeIDList("e", "a", "a", "a", "e"),
			},
			want: Adjacency{
				"key1": MakeIDList("a", "b", "c", "d", "e"),
				"key2": MakeIDList("a", "c"),
				"key3": MakeIDList("a", "e"),
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v\nhave\n\t%#v", name, c.want, have)
		}
	}
}

func TestMergeEdgeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want EdgeMetadatas
	}{
		"Empty a": {
			a: EdgeMetadatas{},
			b: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  2,
				},
			},
			want: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  2,
				},
			},
		},
		"Empty b": {
			a: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:    true,
					BytesEgress:  12,
					BytesIngress: 0,
				},
			},
			b: EdgeMetadatas{},
			want: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:    true,
					BytesEgress:  12,
					BytesIngress: 0,
				},
			},
		},
		"Host merge": {
			a: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  4,
				},
			},
			b: EdgeMetadatas{
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      1,
					BytesIngress:     2,
					WithConnCountTCP: true,
					MaxConnCountTCP:  6,
				},
			},
			want: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  4,
				},
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      1,
					BytesIngress:     2,
					WithConnCountTCP: true,
					MaxConnCountTCP:  6,
				},
			},
		},
		"Edge merge": {
			a: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  7,
				},
			},
			b: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      1,
					BytesIngress:     2,
					WithConnCountTCP: true,
					MaxConnCountTCP:  9,
				},
			},
			want: EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      13,
					BytesIngress:     2,
					WithConnCountTCP: true,
					MaxConnCountTCP:  9,
				},
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}

func TestMergeHostMetadatas(t *testing.T) {
	now := time.Now()

	for name, c := range map[string]struct {
		a, b, want HostMetadatas
	}{
		"Empty a": {
			a: HostMetadatas{},
			b: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
			want: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
		},
		"Empty b": {
			a: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
			b: HostMetadatas{},
			want: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
		},
		"Host merge": {
			a: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
			b: HostMetadatas{
				"hostB": HostMetadata{
					Timestamp: now,
					Hostname:  "host-b",
					OS:        "freedos",
				},
			},
			want: HostMetadatas{
				"hostB": HostMetadata{
					Timestamp: now,
					Hostname:  "host-b",
					OS:        "freedos",
				},
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux",
				},
			},
		},
		"Host conflict": {
			a: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux1",
				},
			},
			b: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now.Add(-10 * time.Second),
					Hostname:  "host-a",
					OS:        "linux0",
				},
			},
			want: HostMetadatas{
				"hostA": HostMetadata{
					Timestamp: now,
					Hostname:  "host-a",
					OS:        "linux1",
				},
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}

func TestMergeNodeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want NodeMetadatas
	}{
		"Empty a": {
			a: NodeMetadatas{},
			b: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			want: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
		},
		"Empty b": {
			a: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			b: NodeMetadatas{},
			want: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
		},
		"Simple merge": {
			a: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			b: NodeMetadatas{
				":192.168.1.2:12345": NodeMetadata{
					"pid":    "42",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			want: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
				":192.168.1.2:12345": NodeMetadata{
					"pid":    "42",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
		},
		"Merge conflict": {
			a: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			b: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{ // <-- same ID
					"pid":    "0",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
			want: NodeMetadatas{
				":192.168.1.1:12345": NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}
