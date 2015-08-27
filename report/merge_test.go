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

func TestMergeNodeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.NodeMetadatas
	}{
		"Empty a": {
			a: report.NodeMetadatas{},
			b: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Empty b": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Simple merge": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{
				":192.168.1.2:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
				":192.168.1.2:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Merge conflict": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{ // <-- same ID
					PID:    "0",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
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
