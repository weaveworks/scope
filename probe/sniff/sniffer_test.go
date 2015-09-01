package sniff_test

import (
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/gopacket"

	"github.com/weaveworks/scope/probe/sniff"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestSnifferShutdown(t *testing.T) {
	var (
		hostID = "abcd"
		src    = newMockSource([]byte{}, nil)
		on     = time.Millisecond
		off    = time.Millisecond
		s      = sniff.New(hostID, report.Networks{}, src, on, off)
	)

	// Stopping the source should terminate the sniffer.
	src.Close()
	time.Sleep(10 * time.Millisecond)

	// Try to get a report from the sniffer. It should block forever, as the
	// loop goroutine should have exited.
	report := make(chan struct{})
	go func() { _, _ = s.Report(); close(report) }()
	select {
	case <-time.After(time.Millisecond):
	case <-report:
		t.Errorf("shouldn't get report after Close")
	}
}

func TestMerge(t *testing.T) {
	var (
		hostID = "xyz"
		src    = newMockSource([]byte{}, nil)
		on     = time.Millisecond
		off    = time.Millisecond
		rpt    = report.MakeReport()
		p      = sniff.Packet{
			SrcIP:     "1.0.0.0",
			SrcPort:   "1000",
			DstIP:     "2.0.0.0",
			DstPort:   "2000",
			Network:   512,
			Transport: 256,
		}

		_, ipnet, _ = net.ParseCIDR(p.SrcIP + "/24") // ;)
		localNets   = report.Networks([]*net.IPNet{ipnet})
	)
	sniff.New(hostID, localNets, src, on, off).Merge(p, &rpt)

	var (
		srcEndpointNodeID = report.MakeEndpointNodeID(hostID, p.SrcIP, p.SrcPort)
		dstEndpointNodeID = report.MakeEndpointNodeID(hostID, p.DstIP, p.DstPort)
	)
	if want, have := (report.Topology{
		NodeMetadatas: report.NodeMetadatas{
			srcEndpointNodeID: report.MakeNodeMetadata().WithEdgeMetadata(dstEndpointNodeID, report.EdgeMetadata{
				EgressPacketCount: newu64(1),
				EgressByteCount:   newu64(256),
			}),
			dstEndpointNodeID: report.MakeNodeMetadata(),
		},
	}), rpt.Endpoint; !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}

	var (
		srcAddressNodeID = report.MakeAddressNodeID(hostID, p.SrcIP)
		dstAddressNodeID = report.MakeAddressNodeID(hostID, p.DstIP)
	)
	if want, have := (report.Topology{
		NodeMetadatas: report.NodeMetadatas{
			srcAddressNodeID: report.MakeNodeMetadata().WithEdgeMetadata(dstAddressNodeID, report.EdgeMetadata{
				EgressPacketCount: newu64(1),
				EgressByteCount:   newu64(512),
			}),
			dstAddressNodeID: report.MakeNodeMetadata(),
		},
	}), rpt.Address; !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

type mockSource struct {
	mtx  sync.RWMutex
	data []byte
	err  error
}

func newMockSource(data []byte, err error) *mockSource {
	return &mockSource{
		data: data,
		err:  err,
	}
}

func (s *mockSource) ZeroCopyReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.data, gopacket.CaptureInfo{
		Timestamp:     time.Now(),
		CaptureLength: len(s.data),
		Length:        len(s.data),
	}, s.err
}

func (s *mockSource) Close() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.err = io.EOF
}

func newu64(value uint64) *uint64 { return &value }
