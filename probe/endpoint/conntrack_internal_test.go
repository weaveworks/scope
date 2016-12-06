package endpoint

import (
	"bufio"
	"encoding/xml"
	"io"
	"testing"
	"time"

	"github.com/weaveworks/common/exec"
	"github.com/weaveworks/scope/test"
	testexec "github.com/weaveworks/scope/test/exec"
)

const conntrackCloseTag = "</conntrack>\n"
const bufferSize = 1024 * 1024

func makeFlow(ty string) flow {
	return flow{
		XMLName: xml.Name{
			Local: "flow",
		},
		Type: ty,
	}
}

func addMeta(f *flow, dir, srcIP, dstIP string, srcPort, dstPort int) *meta {
	meta := meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: dir,
		Layer3: layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
			SrcIP: srcIP,
			DstIP: dstIP,
		},
		Layer4: layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
			SrcPort: srcPort,
			DstPort: dstPort,
			Proto:   tcpProto,
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

func addIndependant(f *flow, id int64, state string) *meta {
	meta := meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: "independent",
		ID:        id,
		State:     state,
		Layer3: layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
		},
		Layer4: layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

func TestConntracker(t *testing.T) {
	oldExecCmd, oldIsConntrackSupported := exec.Command, IsConntrackSupported
	defer func() { exec.Command, IsConntrackSupported = oldExecCmd, oldIsConntrackSupported }()

	IsConntrackSupported = func(_ string) error {
		return nil
	}

	first := true
	existingConnectionsReader, existingConnectionsWriter := io.Pipe()
	reader, writer := io.Pipe()
	exec.Command = func(name string, args ...string) exec.Cmd {
		if first {
			first = false
			return testexec.NewMockCmd(existingConnectionsReader)
		}
		return testexec.NewMockCmd(reader)
	}

	flowWalker := newConntrackFlowWalker(true, "", bufferSize)
	defer flowWalker.stop()

	// First write out some empty xml for the existing connections
	ecbw := bufio.NewWriter(existingConnectionsWriter)
	if _, err := ecbw.WriteString(xmlHeader); err != nil {
		t.Fatal(err)
	}
	if _, err := ecbw.WriteString(conntrackOpenTag); err != nil {
		t.Fatal(err)
	}
	if _, err := ecbw.WriteString(conntrackCloseTag); err != nil {
		t.Fatal(err)
	}
	if err := ecbw.Flush(); err != nil {
		t.Fatal(err)
	}

	// Then write out eventa
	bw := bufio.NewWriter(writer)
	if _, err := bw.WriteString(xmlHeader); err != nil {
		t.Fatal(err)
	}
	if _, err := bw.WriteString(conntrackOpenTag); err != nil {
		t.Fatal(err)
	}
	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}

	have := func() interface{} {
		result := []flow{}
		flowWalker.walkFlows(func(f flow) {
			f.Original = nil
			f.Reply = nil
			f.Independent = nil
			result = append(result, f)
		})
		return result
	}
	ts := 100 * time.Millisecond

	// First, assert we have no flows
	test.Poll(t, ts, []flow{}, have)

	// Now add some flows
	xmlEncoder := xml.NewEncoder(bw)
	writeFlow := func(f flow) {
		if err := xmlEncoder.Encode(f); err != nil {
			t.Fatal(err)
		}
		if _, err := bw.WriteString("\n"); err != nil {
			t.Fatal(err)
		}
		if err := bw.Flush(); err != nil {
			t.Fatal(err)
		}
	}

	flow1 := makeFlow(updateType)
	addMeta(&flow1, "original", "1.2.3.4", "2.3.4.5", 2, 3)
	addIndependant(&flow1, 1, "")
	writeFlow(flow1)
	test.Poll(t, ts, []flow{flow1}, have)

	// Now check when we remove the flow, we still get it in the next Walk
	flow2 := makeFlow(destroyType)
	addMeta(&flow2, "original", "1.2.3.4", "2.3.4.5", 2, 3)
	addIndependant(&flow2, 1, "")
	writeFlow(flow2)
	test.Poll(t, ts, []flow{flow1}, have)
	test.Poll(t, ts, []flow{}, have)

	// This time we're not going to remove it, but put it in state TIME_WAIT
	flow1.Type = updateType
	writeFlow(flow1)
	test.Poll(t, ts, []flow{flow1}, have)

	flow1.Metas[1].State = timeWait
	writeFlow(flow1)
	test.Poll(t, ts, []flow{flow1}, have)
	test.Poll(t, ts, []flow{}, have)
}
