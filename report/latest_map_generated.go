// Generated file, do not edit.
// To regenerate, run ./tools/generate_latest_map ./report/latest_map_generated.go string

package report

import (
	"time"

	"github.com/ugorji/go/codec"
)

type wireStringLatestEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type stringLatestEntryDecoder struct{}

func (d *stringLatestEntryDecoder) Decode(decoder *codec.Decoder, entry *LatestEntry) {
	wire := wireStringLatestEntry{}
	decoder.Decode(&wire)
	entry.Timestamp = wire.Timestamp
	entry.Value = wire.Value
}

// StringLatestEntryDecoder is an implementation of LatestEntryDecoder
// that decodes the LatestEntry instances having a string value.
var StringLatestEntryDecoder LatestEntryDecoder = &stringLatestEntryDecoder{}

// StringLatestMap holds latest string instances.
type StringLatestMap LatestMap

// EmptyStringLatestMap is an empty StringLatestMap. Start with this.
var EmptyStringLatestMap = (StringLatestMap)(MakeLatestMapWithDecoder(StringLatestEntryDecoder))

// MakeStringLatestMap makes an empty StringLatestMap.
func MakeStringLatestMap() StringLatestMap {
	return EmptyStringLatestMap
}

// Copy is a noop, as StringLatestMaps are immutable.
func (m StringLatestMap) Copy() StringLatestMap {
	return (StringLatestMap)((LatestMap)(m).Copy())
}

// Size returns the number of elements.
func (m StringLatestMap) Size() int {
	return (LatestMap)(m).Size()
}

// Merge produces a fresh StringLatestMap containing the keys from both inputs.
// When both inputs contain the same key, the newer value is used.
func (m StringLatestMap) Merge(other StringLatestMap) StringLatestMap {
	return (StringLatestMap)((LatestMap)(m).Merge((LatestMap)(other)))
}

// Lookup the value for the given key.
func (m StringLatestMap) Lookup(key string) (string, bool) {
	v, ok := (LatestMap)(m).Lookup(key)
	if !ok {
		var zero string
		return zero, false
	}
	return v.(string), true
}

// LookupEntry returns the raw entry for the given key.
func (m StringLatestMap) LookupEntry(key string) (string, time.Time, bool) {
	v, timestamp, ok := (LatestMap)(m).LookupEntry(key)
	if !ok {
		var zero string
		return zero, timestamp, false
	}
	return v.(string), timestamp, true
}

// Set the value for the given key.
func (m StringLatestMap) Set(key string, timestamp time.Time, value string) StringLatestMap {
	return (StringLatestMap)((LatestMap)(m).Set(key, timestamp, value))
}

// Delete the value for the given key.
func (m StringLatestMap) Delete(key string) StringLatestMap {
	return (StringLatestMap)((LatestMap)(m).Delete(key))
}

// ForEach executes fn on each key value pair in the map.
func (m StringLatestMap) ForEach(fn func(k string, timestamp time.Time, v string)) {
	(LatestMap)(m).ForEach(func(key string, ts time.Time, value interface{}) {
		fn(key, ts, value.(string))
	})
}

// String returns the StringLatestMap's string representation.
func (m StringLatestMap) String() string {
	return (LatestMap)(m).String()
}

// DeepEqual tests equality with other StringLatestMap.
func (m StringLatestMap) DeepEqual(n StringLatestMap) bool {
	return (LatestMap)(m).DeepEqual((LatestMap)(n))
}

// CodecEncodeSelf implements codec.Selfer.
func (m *StringLatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	(*LatestMap)(m).CodecEncodeSelf(encoder)
}

// CodecDecodeSelf implements codec.Selfer.
func (m *StringLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	bm := (*LatestMap)(m)
	bm.decoder = StringLatestEntryDecoder
	bm.CodecDecodeSelf(decoder)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead.
func (StringLatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead.
func (*StringLatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
