package endpoint

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

// fourTuple is an (IP, port, IP, port) tuple, representing a connection
// active tells whether the connection belongs to an activeFlow (see
// conntrack.go)
type fourTuple struct {
	fromAddr, toAddr [net.IPv4len]byte
	fromPort, toPort uint16
}

func makeFourTuple(fromAddr, toAddr net.IP, fromPort, toPort uint16) fourTuple {
	tuple := fourTuple{fromPort: fromPort, toPort: toPort}
	copy(tuple.fromAddr[:], fromAddr.To4())
	copy(tuple.toAddr[:], toAddr.To4())
	return tuple
}

func (t fourTuple) String() string {
	return fmt.Sprintf("%s:%d-%s:%d", t.fromAddr, t.fromPort, t.toAddr, t.toPort)
}

// key is a sortable direction-independent key for tuples, used to look up a
// fourTuple when you are unsure of its direction.
func (t fourTuple) key() string {
	key := []string{
		fmt.Sprintf("%s:%d", t.fromAddr, t.fromPort),
		fmt.Sprintf("%s:%d", t.toAddr, t.toPort),
	}
	sort.Strings(key)
	return strings.Join(key, " ")
}

// reverse flips the direction of the tuple
func (t *fourTuple) reverse() {
	t.fromAddr, t.fromPort, t.toAddr, t.toPort = t.toAddr, t.toPort, t.fromAddr, t.fromPort
}

// reverse flips the direction of a tuple, without side effects
func reverse(tuple fourTuple) fourTuple {
	return fourTuple{
		fromAddr: tuple.toAddr,
		toAddr:   tuple.fromAddr,
		fromPort: tuple.toPort,
		toPort:   tuple.fromPort,
	}
}
