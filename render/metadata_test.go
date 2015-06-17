package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func TestAggregateMetadata(t *testing.T) {
	for from, want := range map[report.EdgeMetadata]render.AggregateMetadata{

		// Simple connection count
		report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  400,
		}: {
			render.KeyMaxConnCountTCP: 400,
		},

		// Connection count rounding
		report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  4,
		}: {
			render.KeyMaxConnCountTCP: 4,
		},

		// 0 connections.
		report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  0,
		}: {
			render.KeyMaxConnCountTCP: 0,
		},

		// Egress
		report.EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  24,
			BytesIngress: 0,
		}: {
			render.KeyBytesEgress:  24,
			render.KeyBytesIngress: 0,
		},

		// Ingress
		report.EdgeMetadata{
			WithBytes:    true,
			BytesEgress:  0,
			BytesIngress: 1200,
		}: {
			render.KeyBytesEgress:  0,
			render.KeyBytesIngress: 1200,
		},

		// Nothing there.
		report.EdgeMetadata{}: {},
	} {
		if have := render.AggregateMetadataOf(from); !reflect.DeepEqual(have, want) {
			t.Errorf("have: %#v, want %#v", have, want)
		}

	}
}

func TestAggregateMetadataSum(t *testing.T) {
	var (
		this = render.AggregateMetadata{
			"ingress_bytes": 3,
		}
		other = render.AggregateMetadata{
			"ingress_bytes": 333,
			"egress_bytes":  3,
		}
		want = render.AggregateMetadata{
			"ingress_bytes": 336,
			"egress_bytes":  3,
		}
	)

	this.Merge(other)
	if have := this; !reflect.DeepEqual(have, want) {
		t.Errorf("have: %#v, want %#v", have, want)
	}
}
