// +build linux

package endpoint

import (
	"net"
	"strconv"

	"github.com/typetypetype/conntrack"

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

func toMapping(f conntrack.Conn) *endpointMapping {
	var mapping endpointMapping
	if f.Orig.Src.Equal(f.Reply.Dst) {
		mapping = endpointMapping{
			originalIP:    f.Reply.Src,
			originalPort:  f.Reply.SrcPort,
			rewrittenIP:   f.Orig.Dst,
			rewrittenPort: f.Orig.DstPort,
		}
	} else {
		mapping = endpointMapping{
			originalIP:    f.Orig.Src,
			originalPort:  f.Orig.SrcPort,
			rewrittenIP:   f.Reply.Dst,
			rewrittenPort: f.Reply.DstPort,
		}
	}

	return &mapping
}

// applyNAT duplicates Nodes in the endpoint topology of a report, based on
// the NAT table.
func (n natMapper) applyNAT(rpt report.Report, scope string) {
	n.flowWalker.walkFlows(func(f conntrack.Conn, _ bool) {
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
