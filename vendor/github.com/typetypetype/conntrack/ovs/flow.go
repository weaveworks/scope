// +build linux

package ovs

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"encoding/binary"

	"golang.org/x/sys/unix"
)

var nativeEndian binary.ByteOrder

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

const (
	sizeofGenmsg = uint32(unsafe.Sizeof(unix.Nfgenmsg{}))
)

func AllBytes(data []byte, x byte) bool {
	for _, y := range data {
		if x != y {
			return false
		}
	}

	return true
}

type FlowKey interface {
	typeId() uint16
	putKeyNlAttr(*NlMsgBuilder)
	putMaskNlAttr(*NlMsgBuilder) error
	Ignored() bool
	Equals(FlowKey) bool
}

type FlowKeys map[uint16]FlowKey

func (a FlowKeys) Equals(b FlowKeys) bool {
	for id, ak := range a {
		bk, ok := b[id]
		if ok {
			if !ak.Equals(bk) {
				return false
			}
		} else {
			if !ak.Ignored() {
				return false
			}
		}
	}

	for id, bk := range b {
		_, ok := a[id]
		if !ok && !bk.Ignored() {
			return false
		}
	}

	return true
}

func (fks FlowKeys) toNlAttrs(msg *NlMsgBuilder) error {
	// The ethernet flow key is mandatory, even if it is
	// completely wildcarded.
	var defaultEthernetFlowKey FlowKey
	if fks[OVS_KEY_ATTR_ETHERNET] == nil {
		defaultEthernetFlowKey = NewEthernetFlowKey()
	}

	msg.PutNestedAttrs(OVS_FLOW_ATTR_KEY, func() {
		for _, k := range fks {
			if !k.Ignored() {
				k.putKeyNlAttr(msg)
			}
		}

		if defaultEthernetFlowKey != nil {
			defaultEthernetFlowKey.putKeyNlAttr(msg)
		}
	})

	var err error
	msg.PutNestedAttrs(OVS_FLOW_ATTR_MASK, func() {
		for _, k := range fks {
			if !k.Ignored() {
				if e := k.putMaskNlAttr(msg); e != nil {
					err = e
				}
			}
		}

		if defaultEthernetFlowKey != nil {
			defaultEthernetFlowKey.putMaskNlAttr(msg)
		}
	})

	return err
}

// A FlowKeyParser describes how to parse a flow key of a particular
// type from a netlnk message
type FlowKeyParser struct {
	// Flow key parsing function
	//
	// key may be nil if the relevant attribute wasn't provided.
	// This generally means that the mask will indicate that the
	// flow key is Ignored.
	parse func(typ uint16, key []byte, mask []byte, exact bool) (FlowKey, error)

	// Special mask values indicating that the flow key is an
	// exact match or Ignored.  The parse function also receives
	// an "exact" flag, to handle cases where the representation
	// of the mask of awkward.
	exactMask  []byte
	ignoreMask []byte
}

// Maps an NL attribute type to the corresponding FlowKeyParser
type FlowKeyParsers map[uint16]FlowKeyParser

func ParseFlowKeys(keys Attrs, masks Attrs) (res FlowKeys, err error) {
	res = make(FlowKeys)

	for typ, key := range keys {
		parser, ok := flowKeyParsers[typ]
		if !ok {
			parser = FlowKeyParser{parse: parseUnknownFlowKey}
		}

		var mask []byte
		exact := false
		if masks == nil {
			// "OVS_FLOW_ATTR_MASK: ... If not present,
			// all flow key bits are exact match bits."
			mask = parser.exactMask
			exact = true
		} else {
			// "Omitting attribute is treated as
			// wildcarding all corresponding fields"
			mask, ok = masks[typ]
			if !ok {
				mask = parser.ignoreMask
			}
		}

		res[typ], err = parser.parse(typ, key, mask, exact)
		if err != nil {
			return nil, err
		}
	}

	if masks != nil {
		for typ, mask := range masks {
			_, ok := keys[typ]
			if ok {
				continue
			}

			// flow key mask without a corresponding flow
			// key value
			parser, ok := flowKeyParsers[typ]
			if !ok {
				parser = FlowKeyParser{parse: parseUnknownFlowKey}
			}

			res[typ], err = parser.parse(typ, nil, mask, false)
			if err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

// A flow key of a type we don't know about
type UnknownFlowKey struct {
	typ   uint16
	key   []byte
	mask  []byte // nil means ignored
	exact bool
}

func parseUnknownFlowKey(typ uint16, key []byte, mask []byte, exact bool) (FlowKey, error) {
	return UnknownFlowKey{typ: typ, key: key, mask: mask, exact: exact}, nil
}

func (key UnknownFlowKey) String() string {
	var mask string
	switch {
	case key.exact:
		mask = "exact"
	case key.mask == nil:
		mask = "ignored"
	default:
		mask = hex.EncodeToString(key.mask)
	}

	return fmt.Sprintf("UnknownFlowKey{type: %d, key: %s, mask: %s}",
		key.typ, hex.EncodeToString(key.key), mask)
}

func (key UnknownFlowKey) typeId() uint16 {
	return key.typ
}

func (key UnknownFlowKey) putKeyNlAttr(msg *NlMsgBuilder) {
	msg.PutSliceAttr(key.typ, key.key)
}

func (key UnknownFlowKey) putMaskNlAttr(msg *NlMsgBuilder) error {
	if key.exact {
		return fmt.Errorf("cannot serialize exact mask for unknown flow key of type %d", key.typ)
	}

	if key.mask != nil {
		msg.PutSliceAttr(key.typ, key.mask)
	}

	return nil
}

func (key UnknownFlowKey) Ignored() bool {
	return key.mask == nil && !key.exact
}

func (a UnknownFlowKey) Equals(gb FlowKey) bool {
	b, ok := gb.(UnknownFlowKey)
	if !ok {
		return false
	}

	if a.typ != b.typ || !bytes.Equal(a.key, b.key) {
		return false
	}

	switch {
	case a.exact:
		return b.exact
	case a.mask == nil:
		return b.mask == nil
	default:
		return bytes.Equal(a.mask, b.mask)
	}
}

// Most flow keys can be handled as opaque bytes.
type BlobFlowKey struct {
	typ uint16

	// This holds the key and the mask concatenated, so it is
	// twice their length
	keyMask []byte
}

func NewBlobFlowKey(typ uint16, size int) BlobFlowKey {
	km := MakeAlignedByteSlice(size * 2)
	mask := km[size:]
	for i := range mask {
		mask[i] = 0xff
	}
	return BlobFlowKey{typ: typ, keyMask: km}
}

func (key BlobFlowKey) String() string {
	return fmt.Sprintf("BlobFlowKey{type: %d, key: %s, mask: %s}", key.typ,
		hex.EncodeToString(key.key()), hex.EncodeToString(key.mask()))
}

func (key BlobFlowKey) typeId() uint16 {
	return key.typ
}

func (key BlobFlowKey) key() []byte {
	return key.keyMask[:len(key.keyMask)/2]
}

func (key BlobFlowKey) mask() []byte {
	return key.keyMask[len(key.keyMask)/2:]
}

func (key BlobFlowKey) putKeyNlAttr(msg *NlMsgBuilder) {
	msg.PutSliceAttr(key.typ, key.key())
}

func (key BlobFlowKey) putMaskNlAttr(msg *NlMsgBuilder) error {
	msg.PutSliceAttr(key.typ, key.mask())
	return nil
}

func (key BlobFlowKey) Ignored() bool {
	return AllBytes(key.mask(), 0)
}

// Go's anonymous struct fields are not quite a replacement for
// inheritance.  We want to have an Equals method for BlobFlowKeys,
// that works even when BlobFlowKeys are embedded as anonymous struct
// fields.  But we can't use a straightforward type assertion to tell
// if another OvsFlowKey is also a BlobFlowKey, because in the embedded
// case, it will say that the OvsFlowKey is not an BlobFlowKey (the "has
// an anonymoys field of X" is not an "is a X" relation).  To work
// around this, we use an interface, implemented by BlobFlowKey, that
// automatically gets promoted to all structs that embed BlobFlowKey.

type BlobFlowKeyish interface {
	toBlobFlowKey() BlobFlowKey
}

func (key BlobFlowKey) toBlobFlowKey() BlobFlowKey { return key }

func (a BlobFlowKey) Equals(gb FlowKey) bool {
	bx, ok := gb.(BlobFlowKeyish)
	if !ok {
		return false
	}
	b := bx.toBlobFlowKey()

	if a.typ != b.typ {
		return false
	}

	size := len(a.keyMask)
	if len(b.keyMask) != size {
		return false
	}
	size /= 2

	amask := a.keyMask[size:]
	bmask := b.keyMask[size:]
	for i := range amask {
		if amask[i] != bmask[i] || ((a.keyMask[i]^b.keyMask[i])&amask[i]) != 0 {
			return false
		}
	}

	return true
}

func parseBlobFlowKey(typ uint16, key []byte, mask []byte, size int) (BlobFlowKey, error) {
	res := BlobFlowKey{typ: typ}

	if len(mask) != size {
		return res, fmt.Errorf("flow key mask type %d has wrong length (expected %d bytes, got %d)", typ, size, len(mask))
	}

	res.keyMask = MakeAlignedByteSlice(size * 2)
	copy(res.keyMask[size:], mask)

	if key != nil {
		if len(key) != size {
			return res, fmt.Errorf("flow key type %d has wrong length (expected %d bytes, got %d)", typ, size, len(key))
		}

		copy(res.keyMask, key)
	} else {
		// The kernel produces masks without a corresponding
		// key, but in such cases the mask should indicate
		// that the key value is ignored.
		if !AllBytes(mask, 0) {
			return res, fmt.Errorf("flow key type %d has non-zero mask without a value (mask %v)", typ, mask)
		}
	}

	return res, nil
}

func blobFlowKeyParser(size int, wrap func(BlobFlowKey) FlowKey) FlowKeyParser {
	exact := make([]byte, size)
	for i := range exact {
		exact[i] = 0xff
	}

	return FlowKeyParser{
		parse: func(typ uint16, key []byte, mask []byte, exact bool) (FlowKey, error) {
			bfk, err := parseBlobFlowKey(typ, key, mask, size)
			if err != nil {
				return nil, err
			}
			if wrap == nil {
				return bfk, nil
			} else {
				return wrap(bfk), nil
			}
		},
		ignoreMask: make([]byte, size),
		exactMask:  exact,
	}
}

// OVS_KEY_ATTR_IN_PORT: Incoming port number
//
// This flow key is problematic.  First, the kernel always does an
// exact match for IN_PORT, i.e. it takes the mask to be 0xffffffff if
// the key is set at all.  Second, when reporting the mask, the kernel
// always sets the upper 16 bits, probably because port numbers are 16
// bits in the kernel, but 32 bits in the ABI to userspace.  It does
// this even if the IN_PORT flow key was not set.  As a result, we
// take any mask other than 0xffffffff to mean ignored.

type InPortFlowKey struct {
	BlobFlowKey
}

func parseInPortFlowKey(typ uint16, key []byte, mask []byte, exact bool) (FlowKey, error) {
	if !AllBytes(mask, 0xff) {
		for i := range mask {
			mask[i] = 0
		}
	}
	fk, err := parseBlobFlowKey(typ, key, mask, 4)
	if err != nil {
		return nil, err
	}
	return InPortFlowKey{fk}, nil
}

func NewInPortFlowKey(vport VportID) FlowKey {
	fk := InPortFlowKey{NewBlobFlowKey(OVS_KEY_ATTR_IN_PORT, 4)}
	*uint32At(fk.key(), 0) = uint32(vport)
	return fk
}

func (key InPortFlowKey) String() string {
	return fmt.Sprintf("InPortFlowKey{vport: %d}", key.VportID())
}

func (k InPortFlowKey) VportID() VportID {
	return VportID(*uint32At(k.key(), 0))
}

// OVS_KEY_ATTR_ETHERNET: Ethernet header flow key

type EthernetFlowKey struct {
	BlobFlowKey
}

func (key EthernetFlowKey) Ignored() bool {
	// An ethernet flow key is mandatory, so don't omit it just
	// because the mask is all zeros
	return false
}

func NewEthernetFlowKey() EthernetFlowKey {
	return EthernetFlowKey{NewBlobFlowKey(OVS_KEY_ATTR_ETHERNET,
		SizeofOvsKeyEthernet)}
}

func (fk *EthernetFlowKey) key() *OvsKeyEthernet {
	return ovsKeyEthernetAt(fk.BlobFlowKey.key(), 0)
}

func (fk *EthernetFlowKey) mask() *OvsKeyEthernet {
	return ovsKeyEthernetAt(fk.BlobFlowKey.mask(), 0)
}

func (fk EthernetFlowKey) Key() OvsKeyEthernet {
	return *fk.key()
}

func (fk EthernetFlowKey) Mask() OvsKeyEthernet {
	return *fk.mask()
}

func (fk *EthernetFlowKey) SetMaskedEthSrc(addr [ETH_ALEN]byte,
	mask [ETH_ALEN]byte) {
	fk.key().EthSrc = addr
	fk.mask().EthSrc = mask
}

func (fk *EthernetFlowKey) SetEthSrc(addr [ETH_ALEN]byte) {
	fk.SetMaskedEthSrc(addr, [...]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
}

func (fk *EthernetFlowKey) SetMaskedEthDst(addr [ETH_ALEN]byte,
	mask [ETH_ALEN]byte) {
	fk.key().EthDst = addr
	fk.mask().EthDst = mask
}

func (fk *EthernetFlowKey) SetEthDst(addr [ETH_ALEN]byte) {
	fk.SetMaskedEthDst(addr, [...]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
}

func (fk EthernetFlowKey) String() string {
	var buf bytes.Buffer
	var sep string
	fmt.Fprint(&buf, "EthernetFlowKey{")

	if fk.typ != OVS_KEY_ATTR_ETHERNET {
		fmt.Fprintf(&buf, "type: %d", fk.typ)
		sep = ", "
	}

	k := fk.Key()
	m := fk.Mask()
	ha := func(s []byte) string { return net.HardwareAddr(s).String() }
	printMaskedBytes(&buf, &sep, "src", k.EthSrc[:], m.EthSrc[:], ha)
	printMaskedBytes(&buf, &sep, "dst", k.EthDst[:], m.EthDst[:], ha)
	fmt.Fprint(&buf, "}")
	return buf.String()
}

func printMaskedBytes(buf *bytes.Buffer, sep *string, n string, k, m []byte,
	s func([]byte) string) {
	if !AllBytes(m, 0) {
		fmt.Fprintf(buf, "%s%s: %s", *sep, n, s(k))
		if !AllBytes(m, 0xff) {
			fmt.Fprintf(buf, "&%s", s(m))
		}

		*sep = ", "
	}
}

var ethernetFlowKeyParser = blobFlowKeyParser(SizeofOvsKeyEthernet,
	func(fk BlobFlowKey) FlowKey { return EthernetFlowKey{fk} })

// OVS_KEY_ATTR_TUNNEL: Tunnel flow key.  This is more elaborate than
// other flow keys because it consists of a set of attributes.

type TunnelAttrs struct {
	TunnelId [8]byte
	Ipv4Src  [4]byte
	Ipv4Dst  [4]byte
	Tos      uint8
	Ttl      uint8
	Df       bool
	Csum     bool
	TpSrc    uint16
	TpDst    uint16
}

type TunnelAttrsPresence struct {
	TunnelId bool
	Ipv4Src  bool
	Ipv4Dst  bool
	Tos      bool
	Ttl      bool
	Df       bool
	Csum     bool
	TpSrc    bool
	TpDst    bool
}

// Extract presence information from a TunnelAttrs mask
func (ta TunnelAttrs) present() TunnelAttrsPresence {
	// The kernel requires Ipv4Dst and Ttl to be present, so we
	// always mark those as present, even if we end up wildcarding
	// them.
	return TunnelAttrsPresence{
		TunnelId: !AllBytes(ta.TunnelId[:], 0),
		Ipv4Src:  !AllBytes(ta.Ipv4Src[:], 0),
		Ipv4Dst:  true,
		Tos:      ta.Tos != 0,
		Ttl:      true,
		Df:       ta.Df,
		Csum:     ta.Csum,
		TpSrc:    ta.TpSrc != 0,
		TpDst:    ta.TpDst != 0,
	}
}

// Convert a TunnelAttrsPresence to a mask
func (tap TunnelAttrsPresence) mask() (res TunnelAttrs) {
	if tap.TunnelId {
		res.TunnelId = [8]byte{
			0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff,
		}
	}

	if tap.Ipv4Src {
		res.Ipv4Src = [4]byte{0xff, 0xff, 0xff, 0xff}
	}

	if tap.Ipv4Dst {
		res.Ipv4Dst = [4]byte{0xff, 0xff, 0xff, 0xff}
	}

	if tap.Tos {
		res.Tos = 0xff
	}

	if tap.Ttl {
		res.Ttl = 0xff
	}

	res.Df = tap.Df
	res.Csum = tap.Csum

	if tap.TpSrc {
		res.TpSrc = 0xffff
	}

	if tap.TpDst {
		res.TpDst = 0xffff
	}

	return
}

func (ta TunnelAttrs) toNlAttrs(msg *NlMsgBuilder, present TunnelAttrsPresence) {
	if present.TunnelId {
		msg.PutSliceAttr(OVS_TUNNEL_KEY_ATTR_ID, ta.TunnelId[:])
	}

	if present.Ipv4Src {
		msg.PutSliceAttr(OVS_TUNNEL_KEY_ATTR_IPV4_SRC, ta.Ipv4Src[:])
	}

	if present.Ipv4Dst {
		msg.PutSliceAttr(OVS_TUNNEL_KEY_ATTR_IPV4_DST, ta.Ipv4Dst[:])
	}

	if present.Tos {
		msg.PutUint8Attr(OVS_TUNNEL_KEY_ATTR_TOS, ta.Tos)
	}

	if present.Ttl {
		msg.PutUint8Attr(OVS_TUNNEL_KEY_ATTR_TTL, ta.Ttl)
	}

	if present.Df && ta.Df {
		msg.PutEmptyAttr(OVS_TUNNEL_KEY_ATTR_DONT_FRAGMENT)
	}

	if present.Csum && ta.Csum {
		msg.PutEmptyAttr(OVS_TUNNEL_KEY_ATTR_CSUM)
	}

	if present.TpSrc {
		msg.PutUint16Attr(OVS_TUNNEL_KEY_ATTR_TP_SRC,
			uint16ToBE(ta.TpSrc))
	}

	if present.TpDst {
		msg.PutUint16Attr(OVS_TUNNEL_KEY_ATTR_TP_DST,
			uint16ToBE(ta.TpDst))
	}
}

func parseTunnelAttrsData(data []byte) (ta TunnelAttrs, present TunnelAttrsPresence, err error) {
	attrs, err := ParseNestedAttrs(data)
	if err != nil {
		return
	}

	return parseTunnelAttrs(attrs)
}

func parseTunnelAttrs(attrs Attrs) (ta TunnelAttrs, present TunnelAttrsPresence, err error) {
	present.TunnelId, err = attrs.GetOptionalBytes(OVS_TUNNEL_KEY_ATTR_ID, ta.TunnelId[:])
	if err != nil {
		return
	}

	present.Ipv4Src, err = attrs.GetOptionalBytes(OVS_TUNNEL_KEY_ATTR_IPV4_SRC, ta.Ipv4Src[:])
	if err != nil {
		return
	}

	present.Ipv4Dst, err = attrs.GetOptionalBytes(OVS_TUNNEL_KEY_ATTR_IPV4_DST, ta.Ipv4Dst[:])

	ta.Tos, present.Tos, err = attrs.GetOptionalUint8(OVS_TUNNEL_KEY_ATTR_TOS)
	if err != nil {
		return
	}

	ta.Ttl, present.Ttl, err = attrs.GetOptionalUint8(OVS_TUNNEL_KEY_ATTR_TTL)
	if err != nil {
		return
	}

	ta.Df, err = attrs.GetEmpty(OVS_TUNNEL_KEY_ATTR_DONT_FRAGMENT)
	present.Df = ta.Df
	if err != nil {
		return
	}

	ta.Csum, err = attrs.GetEmpty(OVS_TUNNEL_KEY_ATTR_CSUM)
	present.Csum = ta.Csum
	if err != nil {
		return
	}

	ta.TpSrc, present.TpSrc, err = attrs.GetOptionalUint16(OVS_TUNNEL_KEY_ATTR_TP_SRC)
	if err != nil {
		return
	}
	ta.TpSrc = uint16FromBE(ta.TpSrc)

	ta.TpDst, present.TpDst, err = attrs.GetOptionalUint16(OVS_TUNNEL_KEY_ATTR_TP_DST)
	if err != nil {
		return
	}
	ta.TpDst = uint16FromBE(ta.TpDst)

	return
}

type TunnelFlowKey struct {
	key  TunnelAttrs
	mask TunnelAttrs
}

func (fk TunnelFlowKey) String() string {
	var buf bytes.Buffer
	var sep string
	fmt.Fprint(&buf, "TunnelFlowKey{")

	printMaskedBytes(&buf, &sep, "id", fk.key.TunnelId[:],
		fk.mask.TunnelId[:], hex.EncodeToString)
	printMaskedBytes(&buf, &sep, "ipv4src", fk.key.Ipv4Src[:],
		fk.mask.Ipv4Src[:], ipv4ToString)
	printMaskedBytes(&buf, &sep, "ipv4dst", fk.key.Ipv4Dst[:],
		fk.mask.Ipv4Dst[:], ipv4ToString)

	printByte := func(n string, k, m byte) {
		if m != 0 {
			fmt.Fprintf(&buf, "%s%s: %d", sep, n, k)
			if m != 0xff {
				fmt.Fprintf(&buf, "&%x", m)
			}
			sep = ", "
		}
	}

	printByte("tos", fk.key.Tos, fk.mask.Tos)
	printByte("ttl", fk.key.Ttl, fk.mask.Ttl)

	if fk.mask.Df {
		fmt.Fprintf(&buf, "%sdf: %t", sep, fk.key.Df)
		sep = ", "
	}

	if fk.mask.Csum {
		fmt.Fprintf(&buf, "%ssum: %t", sep, fk.key.Csum)
		sep = ", "
	}

	printUint16 := func(n string, k, m uint16) {
		if m != 0 {
			fmt.Fprintf(&buf, "%s%s: %d", sep, n, k)
			if m != 0xffff {
				fmt.Fprintf(&buf, "&%x", m)
			}
			sep = ", "
		}
	}

	printUint16("tpsrc", fk.key.TpSrc, fk.mask.TpSrc)
	printUint16("tpdst", fk.key.TpDst, fk.mask.TpDst)

	fmt.Fprint(&buf, "}")
	return buf.String()
}

func ipv4ToString(ip []byte) string {
	return net.IP(ip).To4().String()
}

func (fk TunnelFlowKey) Key() TunnelAttrs {
	return fk.key
}

func (fk TunnelFlowKey) Mask() TunnelAttrs {
	return fk.mask
}

func (TunnelFlowKey) typeId() uint16 {
	return OVS_KEY_ATTR_TUNNEL
}

func (fk *TunnelFlowKey) SetTunnelId(id [8]byte) {
	fk.key.TunnelId = id
	fk.mask.TunnelId = [...]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}
}

func (fk *TunnelFlowKey) SetIpv4Src(addr [4]byte) {
	fk.key.Ipv4Src = addr
	fk.mask.Ipv4Src = [...]byte{0xff, 0xff, 0xff, 0xff}
}

func (fk *TunnelFlowKey) SetIpv4Dst(addr [4]byte) {
	fk.key.Ipv4Dst = addr
	fk.mask.Ipv4Dst = [...]byte{0xff, 0xff, 0xff, 0xff}
}

func (fk *TunnelFlowKey) SetTos(tos uint8) {
	fk.key.Tos = tos
	fk.mask.Tos = 0xff
}

func (fk *TunnelFlowKey) SetTtl(ttl uint8) {
	fk.key.Ttl = ttl
	fk.mask.Ttl = 0xff
}

func (fk *TunnelFlowKey) SetDf(df bool) {
	fk.key.Df = df
	fk.mask.Df = true
}

func (fk *TunnelFlowKey) SetCsum(csum bool) {
	fk.key.Csum = csum
	fk.mask.Csum = true
}

func (fk *TunnelFlowKey) SetTpSrc(port uint16) {
	fk.key.TpSrc = port
	fk.mask.TpSrc = 0xffff
}

func (fk *TunnelFlowKey) SetTpDst(port uint16) {
	fk.key.TpDst = port
	fk.mask.TpDst = 0xffff
}

func (key TunnelFlowKey) putKeyNlAttr(msg *NlMsgBuilder) {
	msg.PutNestedAttrs(OVS_KEY_ATTR_TUNNEL, func() {
		key.key.toNlAttrs(msg, key.mask.present())
	})
}

func (key TunnelFlowKey) putMaskNlAttr(msg *NlMsgBuilder) error {
	msg.PutNestedAttrs(OVS_KEY_ATTR_TUNNEL, func() {
		key.mask.toNlAttrs(msg, key.mask.present())
	})
	return nil
}

func (a TunnelFlowKey) Equals(gb FlowKey) bool {
	b, ok := gb.(TunnelFlowKey)
	if !ok {
		return false
	}
	return a.key == b.key && a.mask == b.mask
}

func (key TunnelFlowKey) Ignored() bool {
	m := key.mask
	return AllBytes(m.TunnelId[:], 0) &&
		AllBytes(m.Ipv4Src[:], 0) &&
		AllBytes(m.Ipv4Dst[:], 0) &&
		m.Tos == 0 &&
		m.Ttl == 0 &&
		!m.Csum &&
		m.TpSrc == 0 && m.TpDst == 0
}

func parseTunnelFlowKey(typ uint16, key []byte, mask []byte, exact bool) (FlowKey, error) {
	var k, m TunnelAttrs
	var kp TunnelAttrsPresence
	var err error

	if key != nil {
		k, kp, err = parseTunnelAttrsData(key)
		if err != nil {
			return nil, err
		}
	}

	if mask != nil {
		// We don't care about mask presence information,
		// because a missing mask attribute means the field is
		// wildcarded
		m, _, err = parseTunnelAttrsData(mask)
		if err != nil {
			return nil, err
		}
	} else {
		// mask being nil means that no mask attributes were
		// provided, which means the mask is implicit in the
		// key attributes provided
		m = kp.mask()
	}

	return TunnelFlowKey{key: k, mask: m}, err
}

var flowKeyParsers = FlowKeyParsers{
	// Packet QoS priority flow key
	OVS_KEY_ATTR_PRIORITY: blobFlowKeyParser(4, nil),

	OVS_KEY_ATTR_IN_PORT: FlowKeyParser{
		parse:      parseInPortFlowKey,
		exactMask:  []byte{0xff, 0xff, 0xff, 0xff},
		ignoreMask: []byte{0, 0, 0, 0},
	},

	OVS_KEY_ATTR_ETHERNET:  ethernetFlowKeyParser,
	OVS_KEY_ATTR_ETHERTYPE: blobFlowKeyParser(2, nil),
	OVS_KEY_ATTR_IPV4:      blobFlowKeyParser(12, nil),
	OVS_KEY_ATTR_IPV6:      blobFlowKeyParser(40, nil),
	OVS_KEY_ATTR_TCP:       blobFlowKeyParser(4, nil),
	OVS_KEY_ATTR_UDP:       blobFlowKeyParser(4, nil),
	OVS_KEY_ATTR_ICMP:      blobFlowKeyParser(2, nil),
	OVS_KEY_ATTR_ICMPV6:    blobFlowKeyParser(2, nil),
	OVS_KEY_ATTR_ARP:       blobFlowKeyParser(24, nil),
	OVS_KEY_ATTR_ND:        blobFlowKeyParser(28, nil),
	OVS_KEY_ATTR_SKB_MARK:  blobFlowKeyParser(4, nil),
	OVS_KEY_ATTR_DP_HASH:   blobFlowKeyParser(4, nil),
	OVS_KEY_ATTR_TCP_FLAGS: blobFlowKeyParser(2, nil),
	OVS_KEY_ATTR_RECIRC_ID: blobFlowKeyParser(4, nil),

	OVS_KEY_ATTR_TUNNEL: FlowKeyParser{
		parse:      parseTunnelFlowKey,
		exactMask:  nil,
		ignoreMask: []byte{},
	},
}

func MakeFlowKeys() FlowKeys {
	return make(FlowKeys)
}

func (keys FlowKeys) Add(k FlowKey) {
	// TODO check for collisions
	keys[k.typeId()] = k
}

// Actions

type Action interface {
	typeId() uint16
	toNlAttr(*NlMsgBuilder)
	Equals(Action) bool
}

type OutputAction VportID

func NewOutputAction(vport VportID) OutputAction {
	return OutputAction(vport)
}

func (oa OutputAction) String() string {
	return fmt.Sprintf("OutputAction{vport: %d}", oa)
}

func (oa OutputAction) VportID() VportID {
	return VportID(oa)
}

func (OutputAction) typeId() uint16 {
	return OVS_ACTION_ATTR_OUTPUT
}

func (oa OutputAction) toNlAttr(msg *NlMsgBuilder) {
	msg.PutUint32Attr(OVS_ACTION_ATTR_OUTPUT, uint32(oa))
}

func (a OutputAction) Equals(bx Action) bool {
	b, ok := bx.(OutputAction)
	if !ok {
		return false
	}
	return a == b
}

func parseOutputAction(typ uint16, data []byte) (Action, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("flow action type %d has wrong length (expects 4 bytes, got %d)", typ, len(data))
	}

	return OutputAction(*uint32At(data, 0)), nil
}

type SetTunnelAction struct {
	TunnelAttrs
	Present TunnelAttrsPresence
}

func (ta SetTunnelAction) String() string {
	var buf bytes.Buffer
	var sep string
	fmt.Fprint(&buf, "SetTunnelAction{")

	if ta.Present.TunnelId {
		fmt.Fprintf(&buf, "%sid: %s", sep,
			hex.EncodeToString(ta.TunnelId[:]))
		sep = ", "
	}

	if ta.Present.Ipv4Src {
		fmt.Fprintf(&buf, "%sipv4src: %s", sep,
			ipv4ToString(ta.Ipv4Src[:]))
		sep = ", "
	}

	if ta.Present.Ipv4Dst {
		fmt.Fprintf(&buf, "%sipv4dst: %s", sep,
			ipv4ToString(ta.Ipv4Dst[:]))
		sep = ", "
	}

	if ta.Present.Tos {
		fmt.Fprintf(&buf, "%stos: %d", sep, ta.Tos)
		sep = ", "
	}

	if ta.Present.Ttl {
		fmt.Fprintf(&buf, "%sttl: %d", sep, ta.Ttl)
		sep = ", "
	}

	if ta.Present.Df {
		fmt.Fprintf(&buf, "%sdf: %t", sep, ta.Df)
		sep = ", "
	}

	if ta.Present.Csum {
		fmt.Fprintf(&buf, "%scsum: %t", sep, ta.Csum)
		sep = ", "
	}

	if ta.Present.TpSrc {
		fmt.Fprintf(&buf, "%stpsrc: %d", sep, ta.TpSrc)
		sep = ", "
	}

	if ta.Present.TpDst {
		fmt.Fprintf(&buf, "%stpdst: %d", sep, ta.TpDst)
		sep = ", "
	}

	fmt.Fprint(&buf, "}")
	return buf.String()
}

func (SetTunnelAction) typeId() uint16 {
	return OVS_ACTION_ATTR_SET
}

func (ta SetTunnelAction) toNlAttr(msg *NlMsgBuilder) {
	msg.PutNestedAttrs(OVS_ACTION_ATTR_SET, func() {
		msg.PutNestedAttrs(OVS_KEY_ATTR_TUNNEL, func() {
			ta.Present.Df = ta.Df
			ta.Present.Csum = ta.Csum
			ta.TunnelAttrs.toNlAttrs(msg, ta.Present)
		})
	})
}

func (a SetTunnelAction) Equals(bx Action) bool {
	b, ok := bx.(SetTunnelAction)
	if !ok {
		return false
	}
	return a.TunnelAttrs == b.TunnelAttrs
}

func (a *SetTunnelAction) SetTunnelId(id [8]byte) {
	a.TunnelId = id
	a.Present.TunnelId = true
}

func (a *SetTunnelAction) SetIpv4Src(addr [4]byte) {
	a.Ipv4Src = addr
	a.Present.Ipv4Src = true
}

func (a *SetTunnelAction) SetIpv4Dst(addr [4]byte) {
	a.Ipv4Dst = addr
	a.Present.Ipv4Dst = true
}

func (a *SetTunnelAction) SetTos(tos uint8) {
	a.Tos = tos
	a.Present.Tos = true
}

func (a *SetTunnelAction) SetTtl(ttl uint8) {
	a.Ttl = ttl
	a.Present.Ttl = true
}

func (a *SetTunnelAction) SetDf(df bool) {
	a.Df = df
	a.Present.Df = true
}

func (a *SetTunnelAction) SetCsum(csum bool) {
	a.Csum = csum
	a.Present.Csum = true
}

func (a *SetTunnelAction) SetTpSrc(port uint16) {
	a.TpSrc = port
	a.Present.TpSrc = true
}

func (a *SetTunnelAction) SetTpDst(port uint16) {
	a.TpDst = port
	a.Present.TpDst = true
}

type SetUnknownAction struct {
	typ  uint16
	data []byte
}

func (a SetUnknownAction) String() string {
	return fmt.Sprintf("SetUnknownAction{type: %d, data: %s}",
		a.typ, hex.EncodeToString(a.data))
}

func (a SetUnknownAction) Equals(bx Action) bool {
	b, ok := bx.(SetUnknownAction)
	if !ok {
		return false
	}
	return a.typ == b.typ && bytes.Equal(a.data, b.data)
}

func (a SetUnknownAction) typeId() uint16 {
	return a.typ
}

func (a SetUnknownAction) toNlAttr(msg *NlMsgBuilder) {
	msg.PutSliceAttr(a.typ, a.data)
}

func parseSetAction(typ uint16, data []byte) (Action, error) {
	attrs, err := ParseNestedAttrs(data)
	if err != nil {
		return nil, err
	}

	// openvswitch.h says "OVS_ACTION_ATTR_SET: Replaces the
	// contents of an existing header.  The single nested
	// OVS_KEY_ATTR_* attribute specifies a header to modify and
	// its value.".  So we only expect single nested attr.
	//
	// But, a kernel bug in 4.3 (fixed by kernel commit
	// e905eabc90a5b7) means that OVS_KEY_ATTR_TUNNEL gets
	// incorrectly encoded, so that the nested attributes directly
	// contain the OVS_TUNNEL_KEY_ATTR attributes.  But we can
	// detect the consequences of that bug: tunnel attributes must
	// contain either OVS_TUNNEL_KEY_ATTR_IPV4_DST or
	// OVS_TUNNEL_KEY_ATTR_IPV4_SRC, and the sizes of those
	// attributes differ from the corresponding OVS_KEY_ATTR
	// attributes.

	if adata := attrs[OVS_TUNNEL_KEY_ATTR_IPV4_DST]; len(adata) == 4 {
		// Not an OVS_KEY_ATTR_PRIORITY, this is the 4.3 bug.
		return makeSetTunnelAction(parseTunnelAttrs(attrs))
	}

	if adata := attrs[OVS_TUNNEL_KEY_ATTR_IPV6_DST]; len(adata) == 16 {
		// Not an OVS_KEY_ATTR_ARP, this is the 4.3 bug.
		return makeSetTunnelAction(parseTunnelAttrs(attrs))
	}

	if adata := attrs[OVS_KEY_ATTR_TUNNEL]; adata != nil {
		return makeSetTunnelAction(parseTunnelAttrsData(adata))
	}

	for atyp, adata := range attrs {
		return SetUnknownAction{typ: atyp, data: adata}, nil
	}

	// Shouldn't happen, but just in case
	return SetUnknownAction{typ: 0}, nil
}

func makeSetTunnelAction(ta TunnelAttrs, present TunnelAttrsPresence, err error) (Action, error) {
	if err != nil {
		return nil, err
	}

	return SetTunnelAction{TunnelAttrs: ta, Present: present}, nil
}

var actionParsers = map[uint16](func(uint16, []byte) (Action, error)){
	OVS_ACTION_ATTR_OUTPUT: parseOutputAction,
	OVS_ACTION_ATTR_SET:    parseSetAction,
}

// Complete flows

type FlowSpec struct {
	FlowKeys
	Actions []Action
}

func NewFlowSpec() FlowSpec {
	return FlowSpec{FlowKeys: make(FlowKeys), Actions: nil}
}

func (f FlowSpec) String() string {
	var keys []FlowKey

	for _, k := range f.FlowKeys {
		keys = append(keys, k)
	}

	return fmt.Sprintf("FlowSpec{keys: %v, actions: %v}", keys, f.Actions)
}

func (f *FlowSpec) AddKey(k FlowKey) {
	f.FlowKeys.Add(k)
}

func (f *FlowSpec) AddAction(a Action) {
	f.Actions = append(f.Actions, a)
}

func (f *FlowSpec) AddActions(as []Action) {
	f.Actions = append(f.Actions, as...)
}

func (f FlowSpec) toNlAttrs(msg *NlMsgBuilder) error {
	if err := f.FlowKeys.toNlAttrs(msg); err != nil {
		return err
	}

	msg.PutNestedAttrs(OVS_FLOW_ATTR_ACTIONS, func() {
		for _, a := range f.Actions {
			a.toNlAttr(msg)
		}
	})

	return nil
}

func (a FlowSpec) Equals(b FlowSpec) bool {
	if !a.FlowKeys.Equals(b.FlowKeys) {
		return false
	}
	if len(a.Actions) != len(b.Actions) {
		return false
	}

	for i := range a.Actions {
		if !a.Actions[i].Equals(b.Actions[i]) {
			return false
		}
	}

	return true
}

func (dp DatapathHandle) parseFlowMsg(msg *NlMsgParser, cmd int) (Attrs, error) {
	if err := dp.checkNlMsgHeaders(msg, FLOW, cmd); err != nil {
		return nil, err
	}

	return msg.TakeAttrs()
}

func parseFlowSpec(attrs Attrs) (f FlowSpec, err error) {
	keys, err := attrs.GetNestedAttrs(OVS_FLOW_ATTR_KEY, false)
	if err != nil {
		return f, err
	}

	masks, err := attrs.GetNestedAttrs(OVS_FLOW_ATTR_MASK, true)
	if err != nil {
		return f, err
	}

	f.FlowKeys, err = ParseFlowKeys(keys, masks)
	if err != nil {
		return f, err
	}

	actattrs, err := attrs.GetOrderedAttrs(OVS_FLOW_ATTR_ACTIONS)
	if err != nil {
		return f, err
	}

	actions := make([]Action, 0)
	for _, actattr := range actattrs {
		parser, ok := actionParsers[actattr.typ]
		if ok {
			action, err := parser(actattr.typ, actattr.val)
			if err != nil {
				return f, err
			}
			actions = append(actions, action)
		} else {
			fmt.Errorf("unknown action type %d (value %v)", actattr.typ, actattr.val)
		}
	}

	f.Actions = actions
	return f, nil
}

func (dp DatapathHandle) CreateFlow(f FlowSpec) error {
	dpif := dp.Dpif

	req := NewNlMsgBuilder(RequestFlags, dpif.families[FLOW].id)
	req.PutGenlMsghdr(OVS_FLOW_CMD_NEW, OVS_FLOW_VERSION)
	req.putOvsHeader(dp.Ifindex)
	if err := f.toNlAttrs(req); err != nil {
		return err
	}

	_, err := dpif.sock.Request(req)
	return err
}

func (dp DatapathHandle) DeleteFlow(fks FlowKeys) error {
	dpif := dp.Dpif

	req := NewNlMsgBuilder(RequestFlags, dpif.families[FLOW].id)
	req.PutGenlMsghdr(OVS_FLOW_CMD_DEL, OVS_FLOW_VERSION)
	req.putOvsHeader(dp.Ifindex)
	if err := fks.toNlAttrs(req); err != nil {
		return err
	}

	_, err := dpif.sock.Request(req)
	return err
}

func (dp DatapathHandle) ClearFlow(f FlowSpec) error {
	dpif := dp.Dpif

	req := NewNlMsgBuilder(RequestFlags, dpif.families[FLOW].id)
	req.PutGenlMsghdr(OVS_FLOW_CMD_SET, OVS_FLOW_VERSION)
	req.putOvsHeader(dp.Ifindex)
	if err := f.toNlAttrs(req); err != nil {
		return err
	}

	req.PutEmptyAttr(OVS_FLOW_ATTR_CLEAR)

	_, err := dpif.sock.Request(req)
	return err
}

func IsNoSuchFlowError(err error) bool {
	return err == NetlinkError(syscall.ENOENT)
}

type FlowInfo struct {
	FlowSpec
	Packets uint64
	Bytes   uint64
	Used    uint64
}

func parseFlowInfo(attrs Attrs) (fi FlowInfo, err error) {
	fi.FlowSpec, err = parseFlowSpec(attrs)
	if err != nil {
		return
	}

	statsBytes, err := attrs.GetFixedBytes(OVS_FLOW_ATTR_STATS,
		SizeofOvsFlowStats, true)
	if err != nil {
		return
	}

	if statsBytes != nil {
		stats := ovsFlowStatsAt(statsBytes, 0)
		fi.Packets = stats.NPackets
		fi.Bytes = stats.NBytes
	}

	used, usedPresent, err := attrs.GetOptionalUint64(OVS_FLOW_ATTR_USED)
	if err != nil {
		return
	} else if usedPresent {
		fi.Used = used
	}

	return
}

func (dp DatapathHandle) EnumerateFlows() ([]FlowInfo, error) {
	dpif := dp.Dpif
	res := make([]FlowInfo, 0)

	req := NewNlMsgBuilder(DumpFlags, dpif.families[FLOW].id)
	req.PutGenlMsghdr(OVS_FLOW_CMD_GET, OVS_FLOW_VERSION)
	req.putOvsHeader(dp.Ifindex)

	//TODO: check if we can send filter to the message (eg only tunnels)
	consumer := func(resp *NlMsgParser) error {
		attrs, err := dp.parseFlowMsg(resp, OVS_FLOW_CMD_GET)
		if err != nil {
			return err
		}

		fi, err := parseFlowInfo(attrs)
		if err != nil {
			return err
		}

		res = append(res, fi)
		return nil
	}

	err := dpif.sock.RequestMulti(req, consumer)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (dp DatapathHandle) FollowFlows() (<-chan *OvsFlowInfo, func(), error) {
	var closing bool
	//TODO: check if we can send filter to the message (eg only tunnels)

	stop := func() {
		closing = true
	}

	res := make(chan *OvsFlowInfo, 1)

	go func() {
		if err := dp.readMsgs(dp.Dpif.sock, func(fi *OvsFlowInfo) {
			res <- fi
		}); err != nil && !closing {
			panic(err)
		}
	}()

	return res, stop, nil

}

func (dp DatapathHandle) readMsgs(s *NetlinkSocket, cb func(fi *OvsFlowInfo)) error {

	rb := make([]byte, 2*syscall.Getpagesize())
loop:
	for {
		nr, _, err := syscall.Recvfrom(s.fd, rb, 0)
		if err == syscall.ENOBUFS {
			fmt.Println("ENOBUFS")
			// ENOBUF means we miss some events here. No way around it. That's life.
			//cb(FlowInfo{Err: syscall.ENOBUFS})
			continue
		} else if err != nil {
			return err
		}

		msgs, err := syscall.ParseNetlinkMessage(rb[:nr])
		if err != nil {
			return err
		}
		for _, msg := range msgs {

			if msg.Header.Type == unix.NLMSG_ERROR {
				return errors.New("NLMSG_ERROR")
			}
			if msg.Header.Type == unix.NLMSG_DONE {
				break loop
			}

			if nflnMsgType(msg.Header.Type) != uint8(dp.Dpif.families[FLOW].id) {
				fmt.Println(fmt.Sprintf("unexpected message_type: %d\n", nflnMsgType(msg.Header.Type)))
				continue
			}

			fi, err := parsePayload(msg.Data[sizeofGenmsg:])
			if err != nil {
				return err
			}

			cb(fi)
		}
	}
	return nil
}

func parsePayload(payload []byte) (*OvsFlowInfo, error) {
	// Most of this comes from libnetfilter_conntrack/src/conntrack/parse_mnl.c
	flow := &OvsFlowInfo{}
	var attrSpace [16]NetlinkAttr
	attrs, err := parseAttrs(payload, attrSpace[0:0])
	if err != nil {
		return flow, err
	}
	for _, attr := range attrs {
		switch attr.Typ {
		case OVS_FLOW_ATTR_ACTIONS:
			flow.Actions, _ = parseActions(attr.Msg)
		case OVS_FLOW_ATTR_KEY:
			flow.Keys, _ = parseFlowKeys(attr.Msg)
		case OVS_FLOW_ATTR_MASK:
			flow.Masks, _ = parseFlowKeys(attr.Msg)
		}
	}
	return flow, nil
}

func parseFlowKeys(flowKeysPayload []byte) (OvsFlowKeys, error) {

	var attrSpace [16]NetlinkAttr
	keys, err := parseAttrs(flowKeysPayload, attrSpace[0:0])
	if err != nil {
		return nil, err
	}

	var res = make(OvsFlowKeys)

	for _, attrKey := range keys {
		switch OvsKeyAttrType(attrKey.Typ) {
		case OvsAttrUnspec:
		case OvsAttrEncap:
		case OvsAttrPrio:
		case OvsAttrInPrt:
			res[OvsAttrInPrt] = OvsAttrInPrtFlowKey{Port: *(*uint32)(unsafe.Pointer(&attrKey.Msg[0]))}
		case OvsAttrEthernet:
		case OvsAttrVlan:
			res[OvsAttrVlan] = OvsAttrVlanFlowKey{Id: binary.BigEndian.Uint16(attrKey.Msg)}
		case OvsAttrEthertype:
		case OvsAttrIpv4:
			ipv4fk := OvsAttrIpv4FlowKey{}

			ipv4fk.Src = *(*uint32)(unsafe.Pointer(&attrKey.Msg[0]))
			ipv4fk.Dst = *(*uint32)(unsafe.Pointer(&attrKey.Msg[4]))
			ipv4fk.Proto = *(*byte)(unsafe.Pointer(&attrKey.Msg[8]))
			ipv4fk.Tos = *(*byte)(unsafe.Pointer(&attrKey.Msg[9]))
			ipv4fk.Ttl = *(*byte)(unsafe.Pointer(&attrKey.Msg[10]))
			ipv4fk.Frag = *(*byte)(unsafe.Pointer(&attrKey.Msg[11]))
			res[OvsAttrIpv4] = ipv4fk

		case OvsAttrIpv6:
			ipv6fk := OvsAttrIpv6FlowKey{}
			binary.Read(bytes.NewReader(attrKey.Msg), binary.BigEndian, &ipv6fk)
			res[OvsAttrIpv6] = ipv6fk
		case OvsAttrTcp:
			tcpfk := OvsAttrTcpFlowKey{}

			tcpfk.Src = *(*uint16)(unsafe.Pointer(&attrKey.Msg[0]))
			tcpfk.Dst = *(*uint16)(unsafe.Pointer(&attrKey.Msg[2]))

			res[OvsAttrTcp] = tcpfk
		case OvsAttrUdp:
			udpfk := OvsAttrUdpFlowKey{}
			udpfk.Src = *(*uint16)(unsafe.Pointer(&attrKey.Msg[0]))
			udpfk.Dst = *(*uint16)(unsafe.Pointer(&attrKey.Msg[2]))
			res[OvsAttrUdp] = udpfk
		case OvsAttrIcmp:
		case OvsAttrIcmpv6:
		case OvsAttrArp:
		case OvsAttrNd:
		case OvsAttrSkbMark:
		case OvsAttrTunnel:
			tunnel, _ := parseOvsSetTunnelAction(attrKey.Msg)
			res[OvsAttrTunnel] = tunnel
		case OvsAttrSctp:
		case OvsAttrTcpFlags:
		case OvsAttrDpHash:
		case OvsAttrRecircId:
		case OvsAttrMpls:
		case OvsAttrCtState:
			ctfk := OvsAttrCtStateFlowKey{CtState: *(*uint32)(unsafe.Pointer(&attrKey.Msg[0]))}
			res[OvsAttrCtState] = ctfk
		case OvsAttrCtZone:
			ctzonefk := OvsAttrCtZoneFlowKey{Zone: *(*uint16)(unsafe.Pointer(&attrKey.Msg[0]))}
			res[OvsAttrCtZone] = ctzonefk
		case OvsAttrCtMark:
			ctmarkfk := OvsAttrCtMarkFlowKey{Mark: *(*uint32)(unsafe.Pointer(&attrKey.Msg[0]))}
			res[OvsAttrCtMark] = ctmarkfk
		case OvsAttrCtLabels:
			ctmarkfk := OvsAttrCtLabelsFlowKey{}
			copy(ctmarkfk.Labels[:], attrKey.Msg)
			res[OvsAttrCtLabels] = ctmarkfk
		case OvsAttrCtOrigTupleIpv4:
			tupIpv4 := OvsAttrCtOrigTupleIpv4FlowKey{}
			binary.Read(bytes.NewReader(attrKey.Msg[:]), binary.BigEndian, &tupIpv4)
			res[OvsAttrCtOrigTupleIpv4] = tupIpv4
		case OvsAttrCtOrigTupleIpv6:
			tupIpv6 := OvsAttrCtOrigTupleIpv6FlowKey{}
			binary.Read(bytes.NewReader(attrKey.Msg[:]), binary.BigEndian, &tupIpv6)
			res[OvsAttrCtOrigTupleIpv6] = tupIpv6
		}
	}

	return res, nil
}

func parseActions(b []byte) (OvsActions, error) {

	var attrSpace [16]NetlinkAttr
	attrs, err := parseAttrs(b, attrSpace[0:0])
	if err != nil {
		return OvsActions{}, fmt.Errorf("invalid action attr: %s", err)
	}

	var res = make(OvsActions)
	for _, attr := range attrs {
		switch attr.Typ {
		case OVS_ACTION_ATTR_OUTPUT:
		case OVS_ACTION_ATTR_SET:
			ovsSetAttributeAction, err := parseOvsSetAction(attr.Msg)
			if err != nil {
				return OvsActions{}, err
			}

			res[OVS_ACTION_ATTR_SET] = ovsSetAttributeAction

		case OVS_ACTION_ATTR_RECIRC:
		case OVS_ACTION_ATTR_CT:
			ovsCtAction, err := parseOvsCtAction(attr.Msg)
			if err != nil {
				return OvsActions{}, err
			}

			res[OVS_ACTION_ATTR_CT] = ovsCtAction

		default:
		}
	}

	return res, nil
}

func parseOvsCtAction(payload []byte) (OvsCtAction, error) {
	var attrSpace [16]NetlinkAttr
	attrs, err := parseAttrs(payload, attrSpace[0:0])
	if err != nil {
		return OvsCtAction{}, fmt.Errorf("invalid ovs ct action attr: %s", err)
	}

	var res OvsCtAction
	for _, attr := range attrs {
		switch OvsCtAttrType(attr.Typ) {

		case OvsCtAttrTypeUnspec:
		case OvsCtAttrTypeCommit:
			res.Commit = true
		case OvsCtAttrTypeZone:
			res.Zone = *(*uint16)(unsafe.Pointer(&attr.Msg[0]))
		case OvsCtAttrTypeMark:
		case OvsCtAttrTypeLabels:
		case OvsCtAttrTypeHelper:
		case OvsCtAttrTypeNat:
		case OvsCtAttrTypeForceCommit:
		case OvsCtAttrTypeEventMask:
			res.EventMask = *(*uint32)(unsafe.Pointer(&attr.Msg[0]))
		default:
			return OvsCtAction{}, fmt.Errorf("invalid ovs ct action attr: %s", err)
		}
	}

	return res, nil
}

func parseOvsSetAction(payload []byte) ([]OvsAction, error) {

	var attrSpace [16]NetlinkAttr
	setAttrs, err := parseAttrs(payload, attrSpace[0:0])
	if err != nil {
		fmt.Println(err)
		return []OvsAction{}, nil
	}

	// openvswitch.h says "OVS_ACTION_ATTR_SET: Replaces the
	// contents of an existing header.  The single nested
	// OVS_KEY_ATTR_* attribute specifies a header to modify and
	// its value.".  So we only expect single nested attr.
	//
	// But, a kernel bug in 4.3 (fixed by kernel commit
	// e905eabc90a5b7) means that OVS_KEY_ATTR_TUNNEL gets
	// incorrectly encoded, so that the nested attributes directly
	// contain the OVS_TUNNEL_KEY_ATTR attributes.  But we can
	// detect the consequences of that bug: tunnel attributes muzst
	// contain either OVS_TUNNEL_KEY_ATTR_IPV4_DST or
	// OVS_TUNNEL_KEY_ATTR_IPV4_SRC, and the sizes of those
	// attributes differ from the corresponding OVS_KEY_ATTR
	// attributes.

	//TODO: Handle this edge case
	//if adata := attrs[OVS_TUNNEL_KEY_ATTR_IPV4_DST]; len(adata) == 4 {
	//	// Not an OVS_KEY_ATTR_PRIORITY, this is the 4.3 bug.
	//	return makeSetTunnelAction(parseTunnelAttrs(attrs))
	//}
	//
	//if adata := attrs[OVS_TUNNEL_KEY_ATTR_IPV6_DST]; len(adata) == 16 {
	//	// Not an OVS_KEY_ATTR_ARP, this is the 4.3 bug.
	//	return makeSetTunnelAction(parseTunnelAttrs(attrs))
	//}
	//

	var res []OvsAction

	for _, setAttr := range setAttrs {
		switch setAttr.Typ {
		case OVS_KEY_ATTR_TUNNEL:
			setTunnel, err := parseOvsSetTunnelAction(setAttr.Msg)
			if err != nil {
				return []OvsAction{}, err
			}
			res = append(res, setTunnel)
		case OVS_TUNNEL_KEY_ATTR_IPV4_DST:
		case OVS_TUNNEL_KEY_ATTR_IPV6_DST:
		default:
		}

	}

	return res, nil
}

func parseOvsSetTunnelAction(payload []byte) (OvsAction, error) {

	var res OvsSetTunnelAction
	var attrSpace [16]NetlinkAttr
	tunnelAttrs, err := parseAttrs(payload, attrSpace[0:0])
	if err != nil {
		fmt.Println(err)
		return OvsSetTunnelAction{}, err
	}

	for _, tunAttr := range tunnelAttrs {

		switch OvsTunnelKeyAttrType(tunAttr.Typ) {

		case OvsTunnelKeyAttrId:
			res.Present.TunnelId = true
			res.TunnelId = binary.BigEndian.Uint64(tunAttr.Msg)
		case OvsTunnelKeyAttrIpv4Src:
			res.Present.Ipv4Src = true
			res.Ipv4Src = *(*uint32)(unsafe.Pointer(&tunAttr.Msg[0]))
		case OvsTunnelKeyAttrIpv4Dst:
			res.Present.Ipv4Dst = true
			res.Ipv4Dst = *(*uint32)(unsafe.Pointer(&tunAttr.Msg[0]))
		case OvsTunnelKeyAttrTos:
			res.Present.Tos = true
			res.Tos = tunAttr.Msg[0]
		case OvsTunnelKeyAttrTtl:
			res.Present.Ttl = true
			res.Ttl = tunAttr.Msg[0]
		case OvsTunnelKeyAttrDontFragment:
			res.Df = true
		case OvsTunnelKeyAttrCsum:
			res.Csum = true
		case OvsTunnelKeyAttrOam:
			res.Oam = true
		case OvsTunnelKeyAttrTpSrc:
			res.Present.TpSrc = true
			res.TpSrc = binary.BigEndian.Uint16(tunAttr.Msg)
		case OvsTunnelKeyAttrTpDst:
			res.Present.TpDst = true
			res.TpDst = binary.BigEndian.Uint16(tunAttr.Msg)
		case OvsTunnelKeyAttrVxlanOpts:
		case OvsTunnelKeyAttrIpv6Src:
			res.IPv6Src = tunAttr.Msg[0:16]
		case OvsTunnelKeyAttrIpv6Dst:
			res.IPv6Dst = tunAttr.Msg[0:16]
		case OvsTunnelKeyAttrGeneveOpts:
		case OvsTunnelKeyAttrPad:
		case OvsTunnelKeyAttrErspanOpts:
		}

	}

	return res, nil

}
