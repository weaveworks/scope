package endpoint

import (
	"strconv"

	"github.com/weaveworks/scope/report"
)

// This is our 'abstraction' of the endpoint that have been rewritten by NAT.
// Original is the private IP that has been rewritten.
type endpointMapping struct {
	originalIP   string
	originalPort int

	rewrittenIP   string
	rewrittenPort int
}

type natmapper struct {
	*Conntracker
}

func newNATMapper() (*natmapper, error) {
	ct, err := NewConntracker("--any-nat")
	if err != nil {
		return nil, err
	}
	return &natmapper{ct}, nil
}

func toMapping(f Flow) *endpointMapping {
	var mapping endpointMapping
	if f.Original.Layer3.SrcIP == f.Reply.Layer3.DstIP {
		mapping = endpointMapping{
			originalIP:    f.Reply.Layer3.SrcIP,
			originalPort:  f.Reply.Layer4.SrcPort,
			rewrittenIP:   f.Original.Layer3.DstIP,
			rewrittenPort: f.Original.Layer4.DstPort,
		}
	} else {
		mapping = endpointMapping{
			originalIP:    f.Original.Layer3.SrcIP,
			originalPort:  f.Original.Layer4.SrcPort,
			rewrittenIP:   f.Reply.Layer3.DstIP,
			rewrittenPort: f.Reply.Layer4.DstPort,
		}
	}

	return &mapping
}

// applyNAT duplicates NodeMetadatas in the endpoint topology of a
// report, based on the NAT table as returns by natTable.
func (n *natmapper) applyNAT(rpt report.Report, scope string) {
	n.WalkFlows(func(f Flow) {
		mapping := toMapping(f)
		realEndpointID := report.MakeEndpointNodeID(scope, mapping.originalIP, strconv.Itoa(mapping.originalPort))
		copyEndpointID := report.MakeEndpointNodeID(scope, mapping.rewrittenIP, strconv.Itoa(mapping.rewrittenPort))
		nmd, ok := rpt.Endpoint.NodeMetadatas[realEndpointID]
		if !ok {
			return
		}

		rpt.Endpoint.NodeMetadatas[copyEndpointID] = nmd.Copy()
	})
}
