package conntrack

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	sizeofGenmsg = uint32(unsafe.Sizeof(unix.Nfgenmsg{})) // TODO
)

type ConntrackListReq struct {
	Header syscall.NlMsghdr
	Body   unix.Nfgenmsg
}

func (c *ConntrackListReq) toWireFormat() []byte {
	// adapted from syscall/NetlinkRouteRequest.toWireFormat
	b := make([]byte, c.Header.Len)
	*(*uint32)(unsafe.Pointer(&b[0:4][0])) = c.Header.Len
	*(*uint16)(unsafe.Pointer(&b[4:6][0])) = c.Header.Type
	*(*uint16)(unsafe.Pointer(&b[6:8][0])) = c.Header.Flags
	*(*uint32)(unsafe.Pointer(&b[8:12][0])) = c.Header.Seq
	*(*uint32)(unsafe.Pointer(&b[12:16][0])) = c.Header.Pid
	b[16] = byte(c.Body.Nfgen_family)
	b[17] = byte(c.Body.Version)
	*(*uint16)(unsafe.Pointer(&b[18:20][0])) = c.Body.Res_id
	return b
}

func connectNetfilter(bufferSize int, groups uint32) (int, *syscall.SockaddrNetlink, error) {
	s, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_NETFILTER)
	if err != nil {
		return 0, nil, err
	}
	lsa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: groups,
	}
	if err := syscall.Bind(s, lsa); err != nil {
		return 0, nil, err
	}
	if bufferSize > 0 {
		if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_RCVBUFFORCE, bufferSize); err != nil {
			if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_RCVBUF, bufferSize); err != nil {
				return 0, nil, err
			}
		}
	}
	return s, lsa, nil
}

// Established lists all established TCP connections.
func Established(bufferSize int) ([]Flow, error) {
	s, lsa, err := connectNetfilter(bufferSize, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(s)

	var flows []Flow
	msg := ConntrackListReq{
		Header: syscall.NlMsghdr{
			Len:   syscall.NLMSG_HDRLEN + sizeofGenmsg,
			Type:  (NFNL_SUBSYS_CTNETLINK << 8) | uint16(IpctnlMsgCtGet),
			Flags: syscall.NLM_F_REQUEST | syscall.NLM_F_DUMP,
			Pid:   0,
			Seq:   0,
		},
		Body: unix.Nfgenmsg{
			Nfgen_family: syscall.AF_INET,
			Version:      NFNETLINK_V0,
			Res_id:       0,
		},
	}
	wb := msg.toWireFormat()
	if err := syscall.Sendto(s, wb, 0, lsa); err != nil {
		return nil, err
	}

	readMsgs(s, func(c Flow) {
		if c.MsgType != NfctMsgUpdate {
			return
		}
		flows = append(flows, c)
	})
	return flows, nil
}

// Follow gives a channel with all changes.
func Follow(bufferSize int) (<-chan Flow, func(), error) {
	s, _, err := connectNetfilter(bufferSize, NF_NETLINK_CONNTRACK_NEW|NF_NETLINK_CONNTRACK_UPDATE|NF_NETLINK_CONNTRACK_DESTROY)
	stop := func() {
		syscall.Close(s)
	}
	if err != nil {
		return nil, stop, err
	}

	res := make(chan Flow, 1)
	go func() {
		defer syscall.Close(s)
		if err := readMsgs(s, func(c Flow) {
			res <- c
		}); err != nil {
			close(res)
			return
		}
	}()
	return res, stop, nil
}

func readMsgs(s int, cb func(Flow)) error {
	for {
		// TODO(jpg): Re-use the receive buffer.
		// Will require copying any byte slices we will take pointers to
		rb := make([]byte, syscall.Getpagesize())
		nr, _, err := syscall.Recvfrom(s, rb, 0)
		if err != nil {
			return err
		}

		msgs, err := syscall.ParseNetlinkMessage(rb[:nr])
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			if err := nfnlIsError(msg.Header); err != nil {
			}
			if nflnSubsysID(msg.Header.Type) != NFNL_SUBSYS_CTNETLINK {
				return fmt.Errorf(
					"unexpected subsys_id: %d\n",
					nflnSubsysID(msg.Header.Type),
				)
			}
			flow, err := parsePayload(msg.Data[sizeofGenmsg:])
			if err != nil {
				return err
			}

			// Disregard non-TCP flows for now
			if flow.Original.Proto != syscall.IPPROTO_TCP {
				continue
			}

			// Taken from conntrack/parse.c:__parse_message_type
			switch CntlMsgTypes(nflnMsgType(msg.Header.Type)) {
			case IpctnlMsgCtNew:
				flow.MsgType = NfctMsgUpdate
				if msg.Header.Flags&(syscall.NLM_F_CREATE|syscall.NLM_F_EXCL) > 0 {
					flow.MsgType = NfctMsgNew
				}
			case IpctnlMsgCtDelete:
				flow.MsgType = NfctMsgDestroy
			}

			cb(*flow)
		}
	}
}

type Flow struct {
	MsgType  NfConntrackMsg
	Original Meta
	Reply    Meta
	State    TCPState
	ID       uint32
}

type Layer3 struct {
	SrcIP net.IP
	DstIP net.IP
}

type Layer4 struct {
	SrcPort uint16
	DstPort uint16
	Proto   uint8
}

type Meta struct {
	Layer3
	Layer4
}

func parsePayload(b []byte) (*Flow, error) {
	// Adapted from libnetfilter_conntrack/src/conntrack/parse_mnl.c
	flow := &Flow{}
	attrs, err := parseAttrs(b)
	if err != nil {
		return flow, err
	}
	for _, attr := range attrs {
		switch CtattrType(attr.Typ) {
		case CtaTupleOrig:
			parseTuple(attr.Msg, &flow.Original)
		case CtaTupleReply:
			parseTuple(attr.Msg, &flow.Reply)
		case CtaProtoinfo:
			parseProtoinfo(attr.Msg, flow)
		case CtaId:
			flow.ID = binary.BigEndian.Uint32(attr.Msg)
		}
	}
	return flow, nil
}

func parseTuple(b []byte, meta *Meta) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrTuple(attr.Typ) {
		case CtaTupleUnspec:
		case CtaTupleIp:
			if err := parseIP(attr.Msg, meta); err != nil {
				return err
			}
		case CtaTupleProto:
			parseProto(attr.Msg, meta)
		}
	}
	return nil
}

func parseIP(b []byte, meta *Meta) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		// TODO(jpg) IPv6 support
		switch CtattrIp(attr.Typ) {
		case CtaIpV4Src:
			meta.SrcIP = net.IP(attr.Msg) // TODO: copy so we can reuse the buffer?
		case CtaIpV4Dst:
			meta.DstIP = net.IP(attr.Msg) // TODO: copy so we can reuse the buffer?
		}
	}
	return nil
}

func parseProto(b []byte, meta *Meta) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrL4proto(attr.Typ) {
		case CtaProtoNum:
			meta.Proto = uint8(attr.Msg[0])
		case CtaProtoSrcPort:
			meta.SrcPort = binary.BigEndian.Uint16(attr.Msg)
		case CtaProtoDstPort:
			meta.DstPort = binary.BigEndian.Uint16(attr.Msg)
		}
	}
	return nil
}

func parseProtoinfo(b []byte, flow *Flow) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrProtoinfo(attr.Typ) {
		case CtaProtoinfoTcp:
			if err := parseProtoinfoTCP(attr.Msg, flow); err != nil {
				return err
			}
		default:
			// we're not interested in other protocols
		}
	}
	return nil
}

func parseProtoinfoTCP(b []byte, flow *Flow) error {
	attrs, err := parseAttrs(b)
	if err != nil {
		return fmt.Errorf("invalid tuple attr: %s", err)
	}
	for _, attr := range attrs {
		switch CtattrProtoinfoTcp(attr.Typ) {
		case CtaProtoinfoTcpState:
			flow.State = TCPState(attr.Msg[0])
		default:
			// we're not interested in other protocols
		}
	}
	return nil
}
