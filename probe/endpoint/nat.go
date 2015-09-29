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

// NATMapper rewrites a report to deal with NAT's connections
type NATMapper struct {
	Conntracker
}

// MakeNATMapper is exposed for testing
func MakeNATMapper(ct Conntracker) NATMapper {
	return NATMapper{ct}
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

// ApplyNAT duplicates Nodes in the endpoint topology of a
// report, based on the NAT table as returns by natTable.
func (n NATMapper) ApplyNAT(rpt report.Report, scope string) {
	if n.Conntracker == nil {
		return
	}
	n.Conntracker.WalkFlows(func(f Flow) {
		var (
			mapping          = toMapping(f)
			realEndpointID   = report.MakeEndpointNodeID(scope, mapping.originalIP, strconv.Itoa(mapping.originalPort))
			copyEndpointPort = strconv.Itoa(mapping.rewrittenPort)
			copyEndpointID   = report.MakeEndpointNodeID(scope, mapping.rewrittenIP, copyEndpointPort)
			node, ok         = rpt.Endpoint.Nodes[realEndpointID]
		)
		if !ok {
			return
		}

		node = node.Copy()
		node.Metadata[Addr] = mapping.rewrittenIP
		node.Metadata[Port] = copyEndpointPort
		node.Metadata["copy_of"] = realEndpointID
		rpt.Endpoint.AddNode(copyEndpointID, node)
	})
}
