package endpoint_test

import (
	"bufio"
	"encoding/xml"
	"io"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/exec"
	. "github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/test"
	testExec "github.com/weaveworks/scope/test/exec"
)

func makeFlow(ty string) Flow {
	return Flow{
		XMLName: xml.Name{
			Local: "flow",
		},
		Type: ty,
	}
}

func addMeta(f *Flow, dir, srcIP, dstIP string, srcPort, dstPort int) *Meta {
	meta := Meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: dir,
		Layer3: Layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
			SrcIP: srcIP,
			DstIP: dstIP,
		},
		Layer4: Layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
			SrcPort: srcPort,
			DstPort: dstPort,
			Proto:   TCP,
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

func addIndependant(f *Flow, id int64, state string) *Meta {
	meta := Meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: "independent",
		ID:        id,
		State:     state,
		Layer3: Layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
		},
		Layer4: Layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

func TestConntracker(t *testing.T) {
	oldExecCmd, oldConntrackPresent := exec.Command, ConntrackModulePresent
	defer func() { exec.Command, ConntrackModulePresent = oldExecCmd, oldConntrackPresent }()

	ConntrackModulePresent = func() bool {
		return true
	}

	reader, writer := io.Pipe()
	exec.Command = func(name string, args ...string) exec.Cmd {
		return testExec.NewMockCmd(reader)
	}

	conntracker, err := NewConntracker(false)
	if err != nil {
		t.Fatal(err)
	}

	bw := bufio.NewWriter(writer)
	if _, err := bw.WriteString(XMLHeader); err != nil {
		t.Fatal(err)
	}
	if _, err := bw.WriteString(ConntrackOpenTag); err != nil {
		t.Fatal(err)
	}
	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}

	have := func() interface{} {
		result := []Flow{}
		conntracker.WalkFlows(func(f Flow) {
			f.Original = nil
			f.Reply = nil
			f.Independent = nil
			result = append(result, f)
		})
		return result
	}
	ts := 100 * time.Millisecond

	// First, assert we have no flows
	test.Poll(t, ts, []Flow{}, have)

	// Now add some flows
	xmlEncoder := xml.NewEncoder(bw)
	writeFlow := func(f Flow) {
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

	flow1 := makeFlow(New)
	addMeta(&flow1, "original", "1.2.3.4", "2.3.4.5", 2, 3)
	addIndependant(&flow1, 1, "")
	writeFlow(flow1)
	test.Poll(t, ts, []Flow{flow1}, have)

	// Now check when we remove the flow, we still get it in the next Walk
	flow1.Type = Destroy
	writeFlow(flow1)
	test.Poll(t, ts, []Flow{flow1}, have)
	test.Poll(t, ts, []Flow{}, have)

	// This time we're not going to remove it, but put it in state TIME_WAIT
	flow1.Type = New
	writeFlow(flow1)
	test.Poll(t, ts, []Flow{flow1}, have)

	flow1.Metas[1].State = TimeWait
	writeFlow(flow1)
	test.Poll(t, ts, []Flow{flow1}, have)
	test.Poll(t, ts, []Flow{}, have)
}
