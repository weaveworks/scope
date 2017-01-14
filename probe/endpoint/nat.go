package endpoint

import (
	"net"
	"strconv"

	"github.com/weaveworks/scope/probe/endpoint/conntrack"
	"github.com/weaveworks/scope/report"
)

// This is our 'abstraction' of the endpoint that have been rewritten by NAT.
// Original is the private IP that has been rewritten.
type endpointMapping struct {
	originalIP   net.IP
	originalPort uint16

	rewrittenIP   net.IP
	rewrittenPort uint16
}

// natMapper rewrites a report to deal with NAT'd connections.
type natMapper struct {
	flowWalker
}

func makeNATMapper(fw flowWalker) natMapper {
	return natMapper{fw}
}

func toMapping(f conntrack.Flow) *endpointMapping {
	var mapping endpointMapping
	if f.Original.Layer3.SrcIP.Equal(f.Reply.Layer3.DstIP) {
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

// applyNAT duplicates Nodes in the endpoint topology of a report, based on
// the NAT table.
func (n natMapper) applyNAT(rpt report.Report, scope string) {
	n.flowWalker.walkFlows(func(f conntrack.Flow, _ bool) {
		mapping := toMapping(f)

		realEndpointPort := strconv.Itoa(int(mapping.originalPort))
		copyEndpointPort := strconv.Itoa(int(mapping.rewrittenPort))
		realEndpointID := report.MakeEndpointNodeID(scope, "", mapping.originalIP.String(), realEndpointPort)
		copyEndpointID := report.MakeEndpointNodeID(scope, "", mapping.rewrittenIP.String(), copyEndpointPort)

		node, ok := rpt.Endpoint.Nodes[realEndpointID]
		if !ok {
			return
		}

		rpt.Endpoint.AddNode(node.WithID(copyEndpointID).WithLatests(map[string]string{
			CopyOf: realEndpointID,
		}))
	})
}
