package conntrack

const (
	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink.h
	NFNL_SUBSYS_CTNETLINK = 1
	NFNETLINK_V0          = 0

	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink_compat.h
	NF_NETLINK_CONNTRACK_NEW     = 0x00000001
	NF_NETLINK_CONNTRACK_UPDATE  = 0x00000002
	NF_NETLINK_CONNTRACK_DESTROY = 0x00000004

	// #defined in libnfnetlink/include/libnfnetlink/libnfnetlink.h
	NLA_F_NESTED        = uint16(1 << 15)
	NLA_F_NET_BYTEORDER = uint16(1 << 14)
	NLA_TYPE_MASK       = ^(NLA_F_NESTED | NLA_F_NET_BYTEORDER)
)

type NfConntrackMsg int

const (
	NfctMsgUnknown NfConntrackMsg = 0
	NfctMsgNew     NfConntrackMsg = 1 << 0
	NfctMsgUpdate  NfConntrackMsg = 1 << 1
	NfctMsgDestroy NfConntrackMsg = 1 << 2
)

// Taken from libnetfilter_conntrack/src/conntrack/snprintf.c
type TCPState uint8

const (
	TCPStateNone = iota
	TCPStateSynSent
	TCPStateSynRecv
	TCPStateEstablished
	TCPStateFinWait
	TCPStateCloseWait
	TCPStateLastAck
	TCPStateTimeWait
	TCPStateClose
	TCPStateListen
	TCPStateMax
	TCPStateIgnore
)

func (s TCPState) String() string {
	return map[TCPState]string{
		TCPStateNone:        "NONE",
		TCPStateSynSent:     "SYN_SENT",
		TCPStateSynRecv:     "SYN_RECV",
		TCPStateEstablished: "ESTABLISHED",
		TCPStateFinWait:     "FIN_WAIT",
		TCPStateCloseWait:   "CLOSE_WAIT",
		TCPStateLastAck:     "LAST_ACK",
		TCPStateTimeWait:    "TIME_WAIT",
		TCPStateClose:       "CLOSE",
		TCPStateListen:      "LISTEN",
		TCPStateMax:         "MAX",
		TCPStateIgnore:      "IGNORE",
	}[s]
}

// Taken from libnetfilter_conntrack: git://git.netfilter.org/libnetfilter_conntrack
// include/libnetfilter_conntrack/linux_nfnetlink_conntrack.h

type CtattrType int

const (
	CtaUnspec         CtattrType = 0
	CtaTupleOrig      CtattrType = 1
	CtaTupleReply     CtattrType = 2
	CtaStatus         CtattrType = 3
	CtaProtoinfo      CtattrType = 4
	CtaHelp           CtattrType = 5
	CtaNatSrc         CtattrType = 6
	CtaTimeout        CtattrType = 7
	CtaMark           CtattrType = 8
	CtaCountersOrig   CtattrType = 9
	CtaCountersReply  CtattrType = 10
	CtaUse            CtattrType = 11
	CtaId             CtattrType = 12
	CtaNatDst         CtattrType = 13
	CtaTupleMaster    CtattrType = 14
	CtaNatSeqAdjOrig  CtattrType = 15
	CtaNatSeqAdjReply CtattrType = 16
	CtaSecmark        CtattrType = 17
	CtaZone           CtattrType = 18
	CtaSecctx         CtattrType = 19
	CtaTimestamp      CtattrType = 20
	CtaMarkMask       CtattrType = 21
	CtaLabels         CtattrType = 22
	CtaLabelsMask     CtattrType = 23
	CtaMax            CtattrType = 24
)

type CtattrTuple int

const (
	CtaTupleUnspec CtattrTuple = 0
	CtaTupleIp     CtattrTuple = 1
	CtaTupleProto  CtattrTuple = 2
	CtaTupleMax    CtattrTuple = 3
)

type CtattrIp int

const (
	CtaIpUnspec CtattrIp = 0
	CtaIpV4Src  CtattrIp = 1
	CtaIpV4Dst  CtattrIp = 2
	CtaIpV6Src  CtattrIp = 3
	CtaIpV6Dst  CtattrIp = 4
	CtaIpMax    CtattrIp = 5
)

type CtattrL4proto int

const (
	CtaProtoUnspec     CtattrL4proto = 0
	CtaProtoNum        CtattrL4proto = 1
	CtaProtoSrcPort    CtattrL4proto = 2
	CtaProtoDstPort    CtattrL4proto = 3
	CtaProtoIcmpId     CtattrL4proto = 4
	CtaProtoIcmpType   CtattrL4proto = 5
	CtaProtoIcmpCode   CtattrL4proto = 6
	CtaProtoIcmpv6Id   CtattrL4proto = 7
	CtaProtoIcmpv6Type CtattrL4proto = 8
	CtaProtoIcmpv6Code CtattrL4proto = 9
	CtaProtoMax        CtattrL4proto = 10
)

type CtattrProtoinfo int

const (
	CtaProtoinfoUnspec CtattrProtoinfo = 0
	CtaProtoinfoTcp    CtattrProtoinfo = 1
	CtaProtoinfoDccp   CtattrProtoinfo = 2
	CtaProtoinfoSctp   CtattrProtoinfo = 3
	CtaProtoinfoMax    CtattrProtoinfo = 4
)

type CtattrProtoinfoTcp int

const (
	CtaProtoinfoTcpUnspec         CtattrProtoinfoTcp = 0
	CtaProtoinfoTcpState          CtattrProtoinfoTcp = 1
	CtaProtoinfoTcpWscaleOriginal CtattrProtoinfoTcp = 2
	CtaProtoinfoTcpWscaleReply    CtattrProtoinfoTcp = 3
	CtaProtoinfoTcpFlagsOriginal  CtattrProtoinfoTcp = 4
	CtaProtoinfoTcpFlagsReply     CtattrProtoinfoTcp = 5
	CtaProtoinfoTcpMax            CtattrProtoinfoTcp = 6
)

type NfConntrackAttrGrp int

const (
	AttrGrpOrigIpv4     NfConntrackAttrGrp = 0
	AttrGrpReplIpv4     NfConntrackAttrGrp = 1
	AttrGrpOrigIpv6     NfConntrackAttrGrp = 2
	AttrGrpReplIpv6     NfConntrackAttrGrp = 3
	AttrGrpOrigPort     NfConntrackAttrGrp = 4
	AttrGrpReplPort     NfConntrackAttrGrp = 5
	AttrGrpIcmp         NfConntrackAttrGrp = 6
	AttrGrpMasterIpv4   NfConntrackAttrGrp = 7
	AttrGrpMasterIpv6   NfConntrackAttrGrp = 8
	AttrGrpMasterPort   NfConntrackAttrGrp = 9
	AttrGrpOrigCounters NfConntrackAttrGrp = 10
	AttrGrpReplCounters NfConntrackAttrGrp = 11
	AttrGrpOrigAddrSrc  NfConntrackAttrGrp = 12
	AttrGrpOrigAddrDst  NfConntrackAttrGrp = 13
	AttrGrpReplAddrSrc  NfConntrackAttrGrp = 14
	AttrGrpReplAddrDst  NfConntrackAttrGrp = 15
	AttrGrpMax          NfConntrackAttrGrp = 16
)

type NfConntrackQuery int

const (
	NfctQCreate          NfConntrackQuery = 0
	NfctQUpdate          NfConntrackQuery = 1
	NfctQDestroy         NfConntrackQuery = 2
	NfctQGet             NfConntrackQuery = 3
	NfctQFlush           NfConntrackQuery = 4
	NfctQDump            NfConntrackQuery = 5
	NfctQDumpReset       NfConntrackQuery = 6
	NfctQCreateUpdate    NfConntrackQuery = 7
	NfctQDumpFilter      NfConntrackQuery = 8
	NfctQDumpFilterReset NfConntrackQuery = 9
)

// Taken from libnetfilter_conntrack: git://git.netfilter.org/libnetfilter_conntrack
// include/libnetfilter_conntrack/linux_nfnetlink_conntrack.h

type CntlMsgTypes int

const (
	IpctnlMsgCtNew            CntlMsgTypes = 0
	IpctnlMsgCtGet            CntlMsgTypes = 1
	IpctnlMsgCtDelete         CntlMsgTypes = 2
	IpctnlMsgCtGetCtrzero     CntlMsgTypes = 3
	IpctnlMsgCtGetStatsCpu    CntlMsgTypes = 4
	IpctnlMsgCtGetStats       CntlMsgTypes = 5
	IpctnlMsgCtGetDying       CntlMsgTypes = 6
	IpctnlMsgCtGetUnconfirmed CntlMsgTypes = 7
	IpctnlMsgMax              CntlMsgTypes = 8
)
