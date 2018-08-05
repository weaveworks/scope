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

// Taken from include/uapi/linux/netfilter/nf_conntrack_common.h
type CtStatus uint32

const (
	IPS_EXPECTED CtStatus = 1 << iota
	IPS_SEEN_REPLY
	IPS_ASSURED
	IPS_CONFIRMED
	IPS_SRC_NAT
	IPS_DST_NAT
	IPS_SEQ_ADJUST
	IPS_SRC_NAT_DONE
	IPS_DST_NAT_DONE
	IPS_DYING
	IPS_FIXED_TIMEOUT
	IPS_TEMPLATE
	IPS_UNTRACKED
	IPS_HELPER
	IPS_OFFLOAD

	IPS_NAT_MASK      = (IPS_DST_NAT | IPS_SRC_NAT)
	IPS_NAT_DONE_MASK = (IPS_DST_NAT_DONE | IPS_SRC_NAT_DONE)
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
