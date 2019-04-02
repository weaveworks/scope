package ovs

import (
	"net"
)

type OvsKeyAttrType int

const ( // ovs_key_attr include/linux/openvswitch.h

	OvsAttrUnspec          OvsKeyAttrType = 0
	OvsAttrEncap           OvsKeyAttrType = 1
	OvsAttrPrio            OvsKeyAttrType = 2
	OvsAttrInPrt           OvsKeyAttrType = 3
	OvsAttrEthernet        OvsKeyAttrType = 4
	OvsAttrVlan            OvsKeyAttrType = 5
	OvsAttrEthertype       OvsKeyAttrType = 6
	OvsAttrIpv4            OvsKeyAttrType = 7
	OvsAttrIpv6            OvsKeyAttrType = 8
	OvsAttrTcp             OvsKeyAttrType = 9
	OvsAttrUdp             OvsKeyAttrType = 10
	OvsAttrIcmp            OvsKeyAttrType = 11
	OvsAttrIcmpv6          OvsKeyAttrType = 12
	OvsAttrArp             OvsKeyAttrType = 13
	OvsAttrNd              OvsKeyAttrType = 14
	OvsAttrSkbMark         OvsKeyAttrType = 15
	OvsAttrTunnel          OvsKeyAttrType = 16
	OvsAttrSctp            OvsKeyAttrType = 17
	OvsAttrTcpFlags        OvsKeyAttrType = 18
	OvsAttrDpHash          OvsKeyAttrType = 19
	OvsAttrRecircId        OvsKeyAttrType = 20
	OvsAttrMpls            OvsKeyAttrType = 21
	OvsAttrCtState         OvsKeyAttrType = 22
	OvsAttrCtZone          OvsKeyAttrType = 23
	OvsAttrCtMark          OvsKeyAttrType = 24
	OvsAttrCtLabels        OvsKeyAttrType = 25
	OvsAttrCtOrigTupleIpv4 OvsKeyAttrType = 26
	OvsAttrCtOrigTupleIpv6 OvsKeyAttrType = 27
)

type OvsTunnelKeyAttrType int

const ( // ovs_tunnel_key_attr include/linux/openvswitch.h

	OvsTunnelKeyAttrId           OvsTunnelKeyAttrType = 0
	OvsTunnelKeyAttrIpv4Src      OvsTunnelKeyAttrType = 1
	OvsTunnelKeyAttrIpv4Dst      OvsTunnelKeyAttrType = 2
	OvsTunnelKeyAttrTos          OvsTunnelKeyAttrType = 3
	OvsTunnelKeyAttrTtl          OvsTunnelKeyAttrType = 4
	OvsTunnelKeyAttrDontFragment OvsTunnelKeyAttrType = 5
	OvsTunnelKeyAttrCsum         OvsTunnelKeyAttrType = 6
	OvsTunnelKeyAttrOam          OvsTunnelKeyAttrType = 7
	OvsTunnelKeyAttrGeneveOpts   OvsTunnelKeyAttrType = 8
	OvsTunnelKeyAttrTpSrc        OvsTunnelKeyAttrType = 9
	OvsTunnelKeyAttrTpDst        OvsTunnelKeyAttrType = 10
	OvsTunnelKeyAttrVxlanOpts    OvsTunnelKeyAttrType = 11
	OvsTunnelKeyAttrIpv6Src      OvsTunnelKeyAttrType = 12
	OvsTunnelKeyAttrIpv6Dst      OvsTunnelKeyAttrType = 13
	OvsTunnelKeyAttrPad          OvsTunnelKeyAttrType = 14
	OvsTunnelKeyAttrErspanOpts   OvsTunnelKeyAttrType = 15
)

type OvsCtAttrType uint32

const (
	OvsCtAttrTypeUnspec      OvsCtAttrType = iota
	OvsCtAttrTypeCommit      OvsCtAttrType = iota
	OvsCtAttrTypeZone        OvsCtAttrType = iota
	OvsCtAttrTypeMark        OvsCtAttrType = iota
	OvsCtAttrTypeLabels      OvsCtAttrType = iota
	OvsCtAttrTypeHelper      OvsCtAttrType = iota
	OvsCtAttrTypeNat         OvsCtAttrType = iota
	OvsCtAttrTypeForceCommit OvsCtAttrType = iota
	OvsCtAttrTypeEventMask   OvsCtAttrType = iota
)

/* Connection tracking event types */
type IpConntrackEvents uint32

const (
	New       IpConntrackEvents = iota /* new conntrack */
	Related   IpConntrackEvents = iota /* related conntrack */
	Destroy   IpConntrackEvents = iota /* destroyed conntrack */
	Reply     IpConntrackEvents = iota /* connection has seen two-way traffic */
	Assured   IpConntrackEvents = iota /* connection status has changed to assured */
	Protoinfo IpConntrackEvents = iota /* protocol information has changed */
	Helper    IpConntrackEvents = iota /* new helper has been set */
	Mark      IpConntrackEvents = iota /* new mark has been set */
	Seqadj    IpConntrackEvents = iota /* sequence adjustment has changed */
	Secmark   IpConntrackEvents = iota /* new security mark has been set */
	Label     IpConntrackEvents = iota /* new connlabel has been set */
	Synproxy  IpConntrackEvents = iota /* synproxy has been set */
	Max       IpConntrackEvents = iota
)

type OvsCtState uint32

const (
	/* OVS_KEY_ATTR_CT_STATE flags */
	OVS_CS_F_NEW         OvsCtState = 0x01 /* Beginning of a new connection. */
	OVS_CS_F_ESTABLISHED OvsCtState = 0x02 /* Part of an existing connection. */
	OVS_CS_F_RELATED     OvsCtState = 0x04 /* Related to an established connection. */
	OVS_CS_F_REPLY_DIR   OvsCtState = 0x08 /* Flow is in the reply direction. */
	OVS_CS_F_INVALID     OvsCtState = 0x10 /* Could not track connection. */
	OVS_CS_F_TRACKED     OvsCtState = 0x20 /* Conntrack has occurred. */
	OVS_CS_F_SRC_NAT     OvsCtState = 0x40 /* Packet's source address/port was mangled by NAT. */
	OVS_CS_F_DST_NAT     OvsCtState = 0x80 /* Packet's destination address/port was mangled by NAT. */
)

type Tuple struct {
	Proto   int
	Src     net.IP
	SrcPort uint16
	Dst     net.IP
	DstPort uint16

	// ICMP stuff.
	IcmpId   uint16
	IcmpType uint8
	IcmpCode uint8
}

type Conn struct {
	MsgType  NfConntrackMsg
	TCPState string
	Status   CtStatus
	Orig     Tuple
	Reply    Tuple

	// ct.mark, used to set permission type of the flow.
	CtMark uint32

	// ct.id, used to identify connections.
	CtId uint32

	// For multitenancy.
	Zone uint16

	// Flow stats.
	ReplyPktLen   uint64
	ReplyPktCount uint64
	OrigPktLen    uint64
	OrigPktCount  uint64

	// Error, if any.
	Err error
}

type OvsFlowKeys []OvsFlowKey

type OvsFlowKey interface {
}

type OvsAttrUnspecFlowKey struct {
	OvsFlowKey
}

type OvsAttrEncapFlowKey struct {
	OvsFlowKey
}

type OvsAttrPrioFlowKey struct {
	OvsFlowKey
}

type OvsAttrInPrtFlowKey struct {
	Port uint32
	OvsFlowKey
}

type OvsAttrEthernetFlowKey struct {
	OvsFlowKey
}

type OvsAttrVlanFlowKey struct {
	Id uint16
	OvsFlowKey
}

type OvsAttrEthertypeFlowKey struct {
	OvsFlowKey
}

type OvsAttrIpv4FlowKey struct {
	Src   uint32
	Dst   uint32
	Proto byte
	Tos   byte
	Ttl   byte
	Frag  byte
	OvsFlowKey
}

type OvsAttrIpv6FlowKey struct {
	Src    [4]uint32
	Dst    [4]uint32
	Label  uint32
	Proto  byte
	TClass byte
	Hlimit byte
	Frag   byte
	OvsFlowKey
}

type OvsAttrTcpFlowKey struct {
	Src uint16
	Dst uint16
	OvsFlowKey
}

type OvsAttrUdpFlowKey struct {
	Src uint16
	Dst uint16
	OvsFlowKey
}

type OvsAttrIcmpFlowKey struct {
	OvsFlowKey
}

type OvsAttrIcmpv6FlowKey struct {
	OvsFlowKey
}

type OvsAttrArpFlowKey struct {
	OvsFlowKey
}

type OvsAttrNdFlowKey struct {
	OvsFlowKey
}

type OvsAttrSkbMarkFlowKey struct {
	OvsFlowKey
}

type OvsAttrTunnelFlowKey struct {
	OvsFlowKey
}

type OvsAttrSctpFlowKey struct {
	OvsFlowKey
}

type OvsAttrTcpFlagsFlowKey struct {
	OvsFlowKey
}

type OvsAttrDpHashFlowKey struct {
	OvsFlowKey
}

type OvsAttrRecircIdFlowKey struct {
	OvsFlowKey
}

type OvsAttrMplsFlowKey struct {
	OvsFlowKey
}

type OvsAttrCtStateFlowKey struct {
	CtState uint32
	OvsFlowKey
}

type OvsAttrCtZoneFlowKey struct {
	Zone uint16
	OvsFlowKey
}

type OvsAttrCtMarkFlowKey struct {
	Mark uint32
	OvsFlowKey
}

type OvsAttrCtLabelsFlowKey struct {
	Labels [16]byte
	OvsFlowKey
}

type OvsAttrCtOrigTupleIpv4FlowKey struct {
	Src     uint32
	Dst     uint32
	SrcPort uint16
	DstPort uint16
	Proto   byte
	OvsFlowKey
}

type OvsAttrCtOrigTupleIpv6FlowKey struct {
	Src     [4]uint32
	Dst     [4]uint32
	SrcPort uint16
	DstPort uint16
	Proto   byte
	OvsFlowKey
}

type OvsFlowInfo struct {
	OvsFlowSpec
	Packets uint64
	Bytes   uint64
	Used    uint64
}

type OvsFlowSpec struct {
	OvsFlowKeys
	Actions []OvsAction
}

type OvsAction interface {
}

// ovs_ct_attr include/linux/openvswitch.h
type OvsCtAction struct {
	OvsAction
	Commit    bool
	Zone      uint16
	EventMask uint32
}

type OvsSetTunnelAction struct {
	OvsAction
	OvsTunnelAttrs
	Present OvsTunnelAttrsPresence
}

type OvsTunnelAttrs struct {
	TunnelId uint64
	Ipv4Src  uint32
	Ipv4Dst  uint32
	Tos      uint8
	Ttl      uint8
	Df       bool
	Csum     bool
	Oam      bool
	TpSrc    uint16
	TpDst    uint16
	IPv6Src  net.IP
	IPv6Dst  net.IP
}

type OvsTunnelAttrsPresence struct {
	TunnelId bool
	Ipv4Src  bool
	Ipv4Dst  bool
	Tos      bool
	Ttl      bool
	TpSrc    bool
	TpDst    bool
}
