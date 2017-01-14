package conntrack

import (
	"errors"
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

// from src/libnfnetlink.c
func nfnlIsError(hdr syscall.NlMsghdr) error {
	if hdr.Type == syscall.NLMSG_ERROR {
		return errors.New("NLMSG_ERROR")
	}
	if hdr.Type == syscall.NLMSG_DONE && hdr.Flags&syscall.NLM_F_MULTI > 0 {
		return errors.New("Done!")
	}
	return nil
}

// Round the length of a netlink route attribute up to align it
// properly.
func rtaAlignOf(attrlen int) int {
	return (attrlen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}
