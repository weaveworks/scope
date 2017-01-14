package conntrack

import (
	"encoding/binary"
	"errors"
)

const attrHdrLength = 4

type Attr struct {
	Msg            []byte
	Typ            int
	IsNested       bool
	IsNetByteorder bool
}

func parseAttrs(b []byte) ([]Attr, error) {
	var attrs []Attr
	for len(b) >= attrHdrLength {
		var attr Attr
		attr, b = parseAttr(b)
		attrs = append(attrs, attr)
	}
	if len(b) != 0 {
		return nil, errors.New("leftover attr bytes")
	}
	return attrs, nil
}

func parseAttr(b []byte) (Attr, []byte) {
	l := binary.LittleEndian.Uint16(b[0:2])
	// length is header + payload
	l -= uint16(attrHdrLength)

	typ := binary.LittleEndian.Uint16(b[2:4])
	attr := Attr{
		Msg:            b[attrHdrLength : attrHdrLength+int(l)],
		Typ:            int(typ & NLA_TYPE_MASK),
		IsNested:       typ&NLA_F_NESTED > 0,
		IsNetByteorder: typ&NLA_F_NET_BYTEORDER > 0,
	}
	return attr, b[rtaAlignOf(attrHdrLength+int(l)):]
}
