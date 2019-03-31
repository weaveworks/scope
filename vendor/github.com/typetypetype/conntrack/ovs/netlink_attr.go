package ovs

// Netlink attr parsing.

import (
	"encoding/binary"
	"errors"
	"syscall"
)

const attrHdrLength = 4

type NetlinkAttr struct {
	Msg            []byte
	Typ            int
	IsNested       bool
	IsNetByteorder bool
}

func parseAttrs(b []byte, attrs []NetlinkAttr) ([]NetlinkAttr, error) {
	for len(b) >= attrHdrLength {
		var attr NetlinkAttr
		attr, b = parseAttr(b)
		attrs = append(attrs, attr)
	}
	if len(b) != 0 {
		return nil, errors.New("leftover attr bytes")
	}
	return attrs, nil
}

func parseAttr(b []byte) (NetlinkAttr, []byte) {
	l := binary.LittleEndian.Uint16(b[0:2])
	// length is header + payload
	l -= uint16(attrHdrLength)

	typ := binary.LittleEndian.Uint16(b[2:4])
	attr := NetlinkAttr{
		Msg:            b[attrHdrLength : attrHdrLength+int(l)],
		Typ:            int(typ & NLA_TYPE_MASK),
		IsNested:       typ&NLA_F_NESTED > 0,
		IsNetByteorder: typ&NLA_F_NET_BYTEORDER > 0,
	}
	return attr, b[rtaAlignOf(attrHdrLength+int(l)):]
}

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
