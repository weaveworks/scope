package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

const (
	PID    = "pid"
	Name   = "name"
	Domain = "domain"
)

func TestMergeEdgeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.EdgeMetadatas
	}{
		"Empty a": {
			a: report.EdgeMetadatas{},
			b: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				},
			},
		},
		"Empty b": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				},
			},
			b: report.EdgeMetadatas{},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				},
			},
		},
		"Host merge": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				},
			},
			b: report.EdgeMetadatas{
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				},
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				},
			},
		},
		"Edge merge": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(1000),
					MaxConnCountTCP:   newu64(7),
				},
			},
			b: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(9),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(13),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(1002),
					MaxConnCountTCP:   newu64(9),
				},
			},
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func TestFlattenEdgeMetadata(t *testing.T) {
	have := (report.EdgeMetadata{
		EgressPacketCount: newu64(1),
		MaxConnCountTCP:   newu64(2),
	}).Flatten(report.EdgeMetadata{
		EgressPacketCount: newu64(4),
		EgressByteCount:   newu64(8),
		MaxConnCountTCP:   newu64(16),
	})
	want := report.EdgeMetadata{
		EgressPacketCount: newu64(1 + 4),
		EgressByteCount:   newu64(8),
		MaxConnCountTCP:   newu64(2 + 16), // flatten should sum MaxConnCountTCP
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestMergeNodes(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.Nodes
	}{
		"Empty a": {
			a: report.Nodes{},
			b: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Empty b": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Simple merge": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{
				":192.168.1.2:12345": report.MakeNodeWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
				":192.168.1.2:12345": report.MakeNodeWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Merge conflict": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{ // <-- same ID
					PID:    "0",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Counters": {
			a: report.Nodes{
				"1": report.MakeNode().WithCounters(map[string]int{
					"a": 13,
					"b": 57,
					"c": 89,
				}),
			},
			b: report.Nodes{
				"1": report.MakeNode().WithCounters(map[string]int{
					"a": 78,
					"b": 3,
					"d": 47,
				}),
			},
			want: report.Nodes{
				"1": report.MakeNode().WithCounters(map[string]int{
					"a": 91,
					"b": 60,
					"c": 89,
					"d": 47,
				}),
			},
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}

func newu64(value uint64) *uint64 { return &value }
