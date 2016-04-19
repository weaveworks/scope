package sniff

import (
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestInterpolateCounts(t *testing.T) {
	var (
		hostID        = "macbook-air"
		srcNodeID     = report.MakeEndpointNodeID(hostID, "1.2.3.4", "5678")
		dstNodeID     = report.MakeEndpointNodeID(hostID, "5.6.7.8", "9012")
		samplingCount = uint64(200)
		samplingTotal = uint64(2345)
		packetCount   = uint64(123)
		byteCount     = uint64(4096)
	)

	r := report.MakeReport()
	r.Sampling.Count = samplingCount
	r.Sampling.Total = samplingTotal
	r.Endpoint.AddNode(report.MakeNode(srcNodeID).WithEdge(dstNodeID, report.EdgeMetadata{
		EgressPacketCount:  newu64(packetCount),
		IngressPacketCount: newu64(packetCount),
		EgressByteCount:    newu64(byteCount),
		IngressByteCount:   newu64(byteCount),
	}))

	interpolateCounts(r)

	var (
		rate   = float64(samplingCount) / float64(samplingTotal)
		factor = 1.0 / rate
		apply  = func(v uint64) uint64 { return uint64(factor * float64(v)) }
		emd    = r.Endpoint.Nodes[srcNodeID].Edges[dstNodeID]
	)
	if want, have := apply(packetCount), (*emd.EgressPacketCount); want != have {
		t.Errorf("want %d packets, have %d", want, have)
	}
	if want, have := apply(packetCount), (*emd.IngressPacketCount); want != have {
		t.Errorf("want %d packets, have %d", want, have)
	}
	if want, have := apply(byteCount), (*emd.EgressByteCount); want != have {
		t.Errorf("want %d bytes, have %d", want, have)
	}
	if want, have := apply(byteCount), (*emd.IngressByteCount); want != have {
		t.Errorf("want %d bytes, have %d", want, have)
	}
}

func newu64(value uint64) *uint64 { return &value }
