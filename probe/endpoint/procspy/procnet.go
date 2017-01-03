package procspy

import (
	"bytes"
	"net"
)

// ProcNet is an iterator to parse /proc/net/tcp{,6} files.
type ProcNet struct {
	b                       []byte
	c                       Connection
	bytesLocal, bytesRemote [16]byte
	seen                    map[uint64]struct{}
}

// NewProcNet gives a new ProcNet parser.
func NewProcNet(b []byte) *ProcNet {
	return &ProcNet{
		b:           b,
		c:           Connection{},
		seen:        map[uint64]struct{}{},
	}
}

// Next returns the next connection. All buffers are re-used, so if you want
// to keep the IPs you have to copy them.
func (p *ProcNet) Next() *Connection {
again:
	if len(p.b) == 0 {
		return nil
	}
	b := p.b

	if p.b[2] == 's' {
		// Skip header
		p.b = nextLine(b)
		goto again
	}

	var (
		local, remote, inode []byte
	)
	_, b = nextField(b) // 'sl' column
	local, b = nextField(b)
	remote, b = nextField(b)
	_, b = nextField(b) // 'st' column
	_, b = nextField(b) // 'tx_queue' column
	_, b = nextField(b) // 'rx_queue' column
	_, b = nextField(b) // 'tr' column
	_, b = nextField(b) // 'uid' column
	_, b = nextField(b) // 'timeout' column
	inode, b = nextField(b)

	p.c.LocalAddress, p.c.LocalPort = scanAddressNA(local, &p.bytesLocal)
	p.c.RemoteAddress, p.c.RemotePort = scanAddressNA(remote, &p.bytesRemote)
	p.c.inode = parseDec(inode)
	p.b = nextLine(b)
	if _, alreadySeen := p.seen[p.c.inode]; alreadySeen {
		goto again
	}
	p.seen[p.c.inode] = struct{}{}
	return &p.c
}

// scanAddressNA parses 'A12CF62E:00AA' to the address/port. Handles IPv4 and
// IPv6 addresses. The address is a big endian 32 bit ints, hex encoded. We
// just decode the hex and flip the bytes in every group of 4.
func scanAddressNA(in []byte, buf *[16]byte) (net.IP, uint16) {
	col := bytes.IndexByte(in, ':')
	if col == -1 {
		return nil, 0
	}

	// Network address is big endian. Can be either ipv4 or ipv6.
	address := hexDecode32bigNA(in[:col], buf)
	return net.IP(address), uint16(parseHex(in[col+1:]))
}

// hexDecode32big decodes sequences of 32bit big endian bytes.
func hexDecode32bigNA(src []byte, buf *[16]byte) []byte {
	blocks := len(src) / 8
	for block := 0; block < blocks; block++ {
		for i := 0; i < 4; i++ {
			a := fromHexChar(src[block*8+i*2])
			b := fromHexChar(src[block*8+i*2+1])
			buf[block*4+3-i] = (a << 4) | b
		}
	}
	return buf[:blocks*4]
}

func nextField(s []byte) ([]byte, []byte) {
	// Skip whitespace.
	for i, b := range s {
		if b != ' ' {
			s = s[i:]
			break
		}
	}

	// Up until the next whitespace field.
	for i, b := range s {
		if b == ' ' {
			return s[:i], s[i:]
		}
	}

	return nil, nil
}

func nextLine(s []byte) []byte {
	i := bytes.IndexByte(s, '\n')
	if i == -1 {
		return nil
	}
	return s[i+1:]
}

// Simplified copy of strconv.ParseUint(16).
func parseHex(s []byte) uint {
	n := uint(0)
	for i := 0; i < len(s); i++ {
		n *= 16
		n += uint(fromHexChar(s[i]))
	}
	return n
}

// Simplified copy of strconv.ParseUint(10).
func parseDec(s []byte) uint64 {
	n := uint64(0)
	for _, c := range s {
		n *= 10
		n += uint64(c - '0')
	}
	return n
}

// hexDecode32big decodes sequences of 32bit big endian bytes.
func hexDecode32big(src []byte) []byte {
	dst := make([]byte, len(src)/2)
	blocks := len(src) / 8
	for block := 0; block < blocks; block++ {
		for i := 0; i < 4; i++ {
			a := fromHexChar(src[block*8+i*2])
			b := fromHexChar(src[block*8+i*2+1])
			dst[block*4+3-i] = (a << 4) | b
		}
	}
	return dst
}

// fromHexChar converts a hex character into its value.
func fromHexChar(c byte) uint8 {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
