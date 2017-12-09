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

// natMapper rewrites a report to deal with NAT'd connections.
type natMapper struct {
	flowWalker
}

func makeNATMapper(fw flowWalker) natMapper {
	return natMapper{fw}
}

func toMapping(f flow) *endpointMapping {
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

// applyNAT duplicates Nodes in the endpoint topology of a report, based on
// the NAT table.
func (n natMapper) applyNAT(rpt report.Report, scope string) {
	n.flowWalker.walkFlows(func(f flow, active bool) {
		mapping := toMapping(f)

		realEndpointPort := strconv.Itoa(mapping.originalPort)
		copyEndpointPort := strconv.Itoa(mapping.rewrittenPort)
		realEndpointID := report.MakeEndpointNodeID(scope, "", mapping.originalIP, realEndpointPort)
		copyEndpointID := report.MakeEndpointNodeID(scope, "", mapping.rewrittenIP, copyEndpointPort)

		node, ok := rpt.Endpoint.Nodes[realEndpointID]
		if !ok {
			return
		}

		rpt.Endpoint.AddNode(node.WithID(copyEndpointID).WithLatests(map[string]string{
			CopyOf: realEndpointID,
		}))
	})
}
