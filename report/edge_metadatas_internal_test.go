package report

import (
	"bytes"
	"testing"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestEdgeMetadatasAdd(t *testing.T) {
	have := EmptyEdgeMetadatas.
		Add("foo",
			EdgeMetadata{
				EgressPacketCount: newu64(1),
			}).
		Add("foo",
			EdgeMetadata{
				EgressPacketCount: newu64(2),
			})
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

func TestEdgeMetadatasDeepEquals(t *testing.T) {
	want := EmptyEdgeMetadatas.
		Add("foo",
			EdgeMetadata{
				EgressPacketCount: newu64(3),
			})
	have := EmptyEdgeMetadatas.
		Add("foo",
			EdgeMetadata{
				EgressPacketCount: newu64(3),
			})
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
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
					}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(1),
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
		"Disjoint a & b": {
			a: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(12),
						EgressByteCount:   newu64(500),
					}),
			b: EmptyEdgeMetadatas.
				Add("hostQ|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(1),
						EgressByteCount:   newu64(2),
					}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(12),
						EgressByteCount:   newu64(500),
					}).
				Add("hostQ|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(1),
						EgressByteCount:   newu64(2),
					}),
		},
		"Overlapping a & b": {
			a: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(12),
						EgressByteCount:   newu64(1000),
					}),
			b: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(1),
						IngressByteCount:  newu64(123),
						EgressByteCount:   newu64(2),
					}),
			want: EmptyEdgeMetadatas.
				Add("hostA|:192.168.1.1:12345|:192.168.1.2:80",
					EdgeMetadata{
						EgressPacketCount: newu64(13),
						IngressByteCount:  newu64(123),
						EgressByteCount:   newu64(1002),
					}),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func TestEdgeMetadataFlatten(t *testing.T) {
	// Test two EdgeMetadatas flatten to the correct values
	{
		have := (EdgeMetadata{
			EgressPacketCount: newu64(1),
		}).Flatten(EdgeMetadata{
			EgressPacketCount: newu64(4),
			EgressByteCount:   newu64(8),
		})
		want := EdgeMetadata{
			EgressPacketCount: newu64(1 + 4),
			EgressByteCount:   newu64(8),
		}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

	// Test an EdgeMetadatas flatten to the correct value (should
	// just sum)
	{
		have := EmptyEdgeMetadatas.
			Add("foo", EdgeMetadata{
				EgressPacketCount: newu64(1),
			}).
			Add("bar", EdgeMetadata{
				EgressPacketCount: newu64(3),
			}).Flatten()
		want := EdgeMetadata{
			EgressPacketCount: newu64(1 + 3),
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
		}).
		Add("bar", EdgeMetadata{
			EgressPacketCount: newu64(3),
		})

	{
		for _, h := range []codec.Handle{
			codec.Handle(&codec.MsgpackHandle{}),
			codec.Handle(&codec.JsonHandle{}),
		} {
			buf := &bytes.Buffer{}
			encoder := codec.NewEncoder(buf, h)
			want.CodecEncodeSelf(encoder)
			decoder := codec.NewDecoder(buf, h)
			have := EmptyEdgeMetadatas
			have.CodecDecodeSelf(decoder)
			if !reflect.DeepEqual(want, have) {
				t.Error(test.Diff(want, have))
			}
		}
	}
}

func TestEdgeMetadatasEncodingNil(t *testing.T) {
	want := EdgeMetadatas{}

	{

		for _, h := range []codec.Handle{
			codec.Handle(&codec.MsgpackHandle{}),
			codec.Handle(&codec.JsonHandle{}),
		} {
			buf := &bytes.Buffer{}
			encoder := codec.NewEncoder(buf, h)
			want.CodecEncodeSelf(encoder)
			decoder := codec.NewDecoder(buf, h)
			have := EmptyEdgeMetadatas
			have.CodecDecodeSelf(decoder)
			if !reflect.DeepEqual(want, have) {
				t.Error(test.Diff(want, have))
			}
		}
	}
}

func newu64(value uint64) *uint64 { return &value }

func TestEdgeMetadataDeepEquals(t *testing.T) {
	for _, c := range []struct {
		name string
		a, b interface{}
		want bool
	}{
		{
			name: "zero values",
			a:    EdgeMetadata{},
			b:    EdgeMetadata{},
			want: true,
		},
		{
			name: "matching, but different pointers",
			a:    EdgeMetadata{EgressPacketCount: newu64(3)},
			b:    EdgeMetadata{EgressPacketCount: newu64(3)},
			want: true,
		},
		{
			name: "mismatching",
			a:    EdgeMetadata{EgressPacketCount: newu64(3)},
			b:    EdgeMetadata{EgressPacketCount: newu64(4)},
			want: false,
		},
	} {
		if have := reflect.DeepEqual(c.a, c.b); have != c.want {
			t.Errorf("reflect.DeepEqual(%v, %v) != %v", c.a, c.b, c.want)
		}
	}
}
