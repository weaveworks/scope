package ovs

const (
	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink.h
	NFNL_SUBSYS_CTNETLINK = 1
	NFNETLINK_V0          = 0

	// #defined in libnfnetlink/include/libnfnetlink/linux_nfnetlink_compat.h
	NF_NETLINK_CONNTRACK_NEW         = 0x00000001
	NF_NETLINK_CONNTRACK_UPDATE      = 0x00000002
	NF_NETLINK_CONNTRACK_DESTROY     = 0x00000004
	NF_NETLINK_CONNTRACK_EXP_NEW     = 0x00000008
	NF_NETLINK_CONNTRACK_EXP_UPDATE  = 0x00000010
	NF_NETLINK_CONNTRACK_EXP_DESTROY = 0x00000020

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

// taken from libnetfilter_conntrack/src/conntrack/snprintf.c
var tcpState = []string{
	"NONE",
	"SYN_SENT",
	"SYN_RECV",
	"ESTABLISHED",
	"FIN_WAIT",
	"CLOSE_WAIT",
	"LAST_ACK",
	"TIME_WAIT",
	"CLOSE",
	"LISTEN",
	"MAX",
	"IGNORE",
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
