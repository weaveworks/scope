// +build !arm

package marshal

// FromUtsname reads the C-strings from syscall.Utsname and transforms them to Go
// strings
func FromUtsname(ca [65]int8) string {
	s := make([]byte, len(ca))
	var lens int
	for ; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}
