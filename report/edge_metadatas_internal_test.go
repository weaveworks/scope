package report

import (
	"testing"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestEdgeMetadatasAdd(t *testing.T) {
	want := EmptyEdgeMetadatas.
		Add("foo",
		EdgeMetadata{
			EgressPacketCount: newu64(3),
		})
	have := EmptyEdgeMetadatas.
		Add("foo",
		EdgeMetadata{
			EgressPacketCount: newu64(1),
		}).
		Add("foo",
		EdgeMetadata{
			EgressPacketCount: newu64(2),
		})
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
	if emd, ok := have.Lookup("foo"); !ok || *emd.EgressPacketCount != 3 {
		t.Errorf("foo.EgressPacketCount != 3")
	}
	if emd, ok := have.Lookup("bar"); ok || emd.EgressPacketCount != nil {
		t.Errorf("bar.EgressPacketCount != nil")
	}
	have.ForEach(func(k string, emd EdgeMetadata) {
		if k != "foo" || *emd.EgressPacketCount != 3 {
			t.Errorf("foo.EgressPacketCount != 3")
		}
	})
}

func TestEdgeMetadatasAddNil(t *testing.T) {
	have := EdgeMetadatas{}.Add("foo", EdgeMetadata{EgressPacketCount: newu64(1)})
	if have.Size() != 1 {
		t.Errorf("Adding to a zero-value EdgeMetadatas failed, got: %v", have)
	}
}

func TestEdgeMetadatasMerge(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want EdgeMetadatas
	}{
		"nils": {
			a:    EdgeMetadatas{},
			b:    EdgeMetadatas{},
			want: EdgeMetadatas{},
		},
		"Empty a": {
			a: EmptyEdgeMetadatas,
			b: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				}),
		},
		"Empty b": {
			a: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				}),
			b: EmptyEdgeMetadatas,
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				}),
		},
		"Host merge": {
			a: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				}),
			b: EmptyEdgeMetadatas.
				Add("hostQ|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				}).
				Add("hostQ|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				}),
		},
		"Edge merge": {
			a: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(1000),
					MaxConnCountTCP:   newu64(7),
				}),
			b: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(1),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(9),
				}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
				EdgeMetadata{
					EgressPacketCount: newu64(13),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(1002),
					MaxConnCountTCP:   newu64(9),
				}),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func TestEdgeMetadataFlatten(t *testing.T) {
	{
		have := (EdgeMetadata{
			EgressPacketCount: newu64(1),
			MaxConnCountTCP:   newu64(2),
		}).Flatten(EdgeMetadata{
			EgressPacketCount: newu64(4),
			EgressByteCount:   newu64(8),
			MaxConnCountTCP:   newu64(16),
		})
		want := EdgeMetadata{
			EgressPacketCount: newu64(1 + 4),
			EgressByteCount:   newu64(8),
			MaxConnCountTCP:   newu64(2 + 16), // flatten should sum MaxConnCountTCP
		}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

	{
		have := EmptyEdgeMetadatas.
			Add("foo", EdgeMetadata{
			EgressPacketCount: newu64(1),
			MaxConnCountTCP:   newu64(2),
		}).
			Add("bar", EdgeMetadata{
			EgressPacketCount: newu64(3),
			MaxConnCountTCP:   newu64(5),
		}).Flatten()
		want := EdgeMetadata{
			EgressPacketCount: newu64(1 + 3),
			MaxConnCountTCP:   newu64(2 + 5),
		}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

	{
		// Should not panic on nil
		have := EdgeMetadatas{}.Flatten()
		want := EdgeMetadata{}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestEdgeMetadataReversed(t *testing.T) {
	have := EdgeMetadata{
		EgressPacketCount: newu64(1),
	}.Reversed()
	want := EdgeMetadata{
		IngressPacketCount: newu64(1),
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestEdgeMetadatasEncoding(t *testing.T) {
	want := EmptyEdgeMetadatas.
		Add("foo", EdgeMetadata{
		EgressPacketCount: newu64(1),
		MaxConnCountTCP:   newu64(2),
	}).
		Add("bar", EdgeMetadata{
		EgressPacketCount: newu64(3),
		MaxConnCountTCP:   newu64(5),
	})

	{
		gobs, err := want.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyEdgeMetadatas
		have.GobDecode(gobs)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

	{
		json, err := want.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyEdgeMetadatas
		have.UnmarshalJSON(json)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestEdgeMetadatasEncodingNil(t *testing.T) {
	want := EdgeMetadatas{}

	{
		gobs, err := want.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyEdgeMetadatas
		have.GobDecode(gobs)
		if have.psMap == nil {
			t.Error("needed to get back a non-nil psMap for EdgeMetadata")
		}
	}

	{
		json, err := want.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyEdgeMetadatas
		have.UnmarshalJSON(json)
		if have.psMap == nil {
			t.Error("needed to get back a non-nil psMap for EdgeMetadata")
		}
	}
}

func newu64(value uint64) *uint64 { return &value }
