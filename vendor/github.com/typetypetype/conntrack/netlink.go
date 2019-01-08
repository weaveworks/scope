package conntrack

import (
	"syscall"
)

// NFNL_MSG_TYPE
func nflnMsgType(x uint16) uint8 {
	return uint8(x & 0x00ff)
}

// NFNL_SUBSYS_ID
func nflnSubsysID(x uint16) uint8 {
	return uint8((x & 0xff00) >> 8)
}

// Round the length of a netlink route attribute up to align it
// properly.
func rtaAlignOf(attrlen int) int {
	return (attrlen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}
