package report

import (
	"reflect"
	"testing"
)

func TestAggregateMetadata(t *testing.T) {
	for from, want := range map[EdgeMetadata]AggregateMetadata{

		// Simple connection count
		EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  400,
		}: {
			keyMaxConnCountTCP: 400,
		},

		// Connection count rounding
		EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  4,
		}: {
			keyMaxConnCountTCP: 4,
		},

		// 0 connections.
		EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  0,
		}: {
			keyMaxConnCountTCP: 0,
		},

		// Egress
		EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  24,
			BytesIngress: 0,
		}: {
			keyBytesEgress:  24,
			keyBytesIngress: 0,
		},

		// Ingress
		EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  0,
			BytesIngress: 1200,
		}: {
			keyBytesEgress:  0,
			keyBytesIngress: 1200,
		},

		// Nothing there.
		EdgeMetadata{}: {},
	} {
		if have := from.Render(); !reflect.DeepEqual(have, want) {
			t.Errorf("have: %#v, want %#v", have, want)
		}

	}
}

func TestAggregateMetadataSum(t *testing.T) {
	var (
		this = AggregateMetadata{
			"ingress_bytes": 3,
		}
		other = AggregateMetadata{
			"ingress_bytes": 333,
			"egress_bytes":  3,
		}
		want = AggregateMetadata{
			"ingress_bytes": 336,
			"egress_bytes":  3,
		}
	)

	this.Merge(other)
	if have := this; !reflect.DeepEqual(have, want) {
		t.Errorf("have: %#v, want %#v", have, want)
	}
}
