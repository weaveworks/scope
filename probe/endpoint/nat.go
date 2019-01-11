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

func endpointNodeID(scope string, ip net.IP, port uint16) string {
	return report.MakeEndpointNodeID(scope, "", ip.String(), strconv.Itoa(int(port)))
}

/*

Some examples of connections with NAT:

Here 10.32.0.X are pod addresses; 172.31.X.X are node addresses; 10.10X.X.X are service virtual addresses.

Pod to pod via Kubernetes service
  picked up by ebpf as 10.32.0.16:47600->10.105.173.176:5432 and 10.32.0.6:5432 (??)
  NAT IPS_DST_NAT orig: 10.32.0.16:47600->10.105.173.176:5432, reply: 10.32.0.6:5432->10.32.0.16:47600
  We want: 10.32.0.16:47600->10.32.0.6:5432
   - replace the destination (== NAT orig dst) with the NAT reply source (A)

Incoming from outside the cluster to a NodePort:
  picked up by ebpf as 10.32.0.1:13488->10.32.0.7:80
  NAT: IPS_SRC_NAT IPS_DST_NAT orig: 37.157.33.76:13488->172.31.2.17:30081, reply: 10.32.0.7:80->10.32.0.1:13488
  We want: 37.157.33.76:13488->10.32.0.7:80
   - replace the source (== NAT reply dst) with the NAT original source (B)
  To match another probe with the other side of this connection, also want 37.157.33.76:13488->172.31.2.17:30081
   - add NAT original dst as a copy of nat reply source (C)

Outgoing from a pod:
  picked up by ebpf as 10.32.0.7:36078->18.221.99.178:443
  NAT:  IPS_SRC_NAT orig: 10.32.0.7:36078->18.221.99.178:443, reply: 18.221.99.178:443->172.31.2.17:36078
  We want: 10.32.0.7:36078->18.221.99.178:443
   - leave it alone. (D)

Docker container exposing port to similar on different host
host1:
  picked up by ebpf as ip-172-31-5-80;172.17.0.2:43042->172.31.2.17:8080
  NAT: IPS_SRC_NAT orig: 172.17.0.2:43042->172.31.2.17:8080, reply: 172.31.2.17:8080-> 172.31.5.80:43042
  We want: 172.31.5.80:43042->172.31.2.17:8080
   - can't have a blanket rule to replace NAT original source with NAT reply destination, because that breaks case D.
  we could add 172.31.5.80:43042 (nat reply destination) as a copy of ip-172-31-5-80;172.17.0.2:43042 (nat orig source) (E)
host2:
  picked up by ebpf as 172.31.5.80:43042->ip-172-31-2-17;172.17.0.2:80
  NAT: IPS_DST_NAT orig: 172.31.5.80:43042->172.31.2.17:8080, reply: 172.17.0.2:80->172.31.5.80:43042
  Rule A doesn't match and rule B is a no-op because the addresses are the same.
  To match another probe with the other side of this connection, also want 172.31.5.80:43042->172.31.2.17:8080
   - add NAT original dst as a copy of nat reply source (C)

All of the above can be satisfied by these rules:
  For SRC_NAT
    replace the source (== NAT reply dst) with the NAT original source (B)
    or add NAT reply destination as a copy of NAT original source (E)
  For DST_NAT
    replace NAT original destination in adjacencies with the NAT reply source (A)
    or add NAT original destination as a copy of NAT reply source (C)
*/

// applyNAT modifies Nodes in the endpoint topology of a report, based on
// the NAT table.
func (n natMapper) applyNAT(rpt report.Report, scope string) {
	n.flowWalker.walkFlows(func(f conntrack.Conn, _ bool) {

		if (f.Status & conntrack.IPS_SRC_NAT) != 0 {
			origSrcID := endpointNodeID(scope, f.Orig.Src, f.Orig.SrcPort)
			replyDstID := endpointNodeID(scope, f.Reply.Dst, f.Reply.DstPort)
			if replyDstID != origSrcID {
				if fromNode, ok := rpt.Endpoint.Nodes[replyDstID]; ok {
					// replace the source (== NAT reply dst) with the NAT original source (B)
					delete(rpt.Endpoint.Nodes, replyDstID)
					rpt.Endpoint.AddNode(fromNode.WithID(origSrcID))
				} else if origSrcNode, ok := rpt.Endpoint.Nodes[origSrcID]; ok {
					// add NAT reply destination as a copy of NAT original source (E)
					newNode := origSrcNode.WithID(replyDstID).WithLatests(map[string]string{
						CopyOf: origSrcID,
					})
					rpt.Endpoint.AddNode(newNode)
				}
			}
		}

		if (f.Status & conntrack.IPS_DST_NAT) != 0 {
			replySrcID := endpointNodeID(scope, f.Reply.Src, f.Reply.SrcPort)
			origDstID := endpointNodeID(scope, f.Orig.Dst, f.Orig.DstPort)
			if replySrcID != origDstID {
				fromID := endpointNodeID(scope, f.Reply.Dst, f.Reply.DstPort)
				fromNode, ok := rpt.Endpoint.Nodes[fromID]
				if ok && fromNode.Adjacency.Contains(origDstID) {
					// replace NAT original destination in adjacencies with the NAT reply source (A)
					fromNode.Adjacency = fromNode.Adjacency.Minus(origDstID)
					fromNode = fromNode.WithAdjacent(replySrcID)
					rpt.Endpoint.Nodes[fromID] = fromNode
				} else if replySrcNode, ok := rpt.Endpoint.Nodes[replySrcID]; ok {
					// add NAT original destination as a copy of NAT reply source (C)
					newNode := replySrcNode.WithID(origDstID).WithLatests(map[string]string{
						CopyOf: replySrcID,
					})
					rpt.Endpoint.AddNode(newNode)
				}
			}

		}
	})
}
