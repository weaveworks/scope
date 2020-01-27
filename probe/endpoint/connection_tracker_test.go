package endpoint

import (
	"testing"
)

// mapPortToPids collects info about connections between specific
// address pairs and destination port.

// a set of connections from a single pid to a destination off-box (pid is zero)
func fakeConnectionsOffBox(count int, fromPid uint32, startPort uint16) mapPortToPids {
	ret := make(mapPortToPids, count)
	for i := 0; i < count; i++ {
		ret[uint16(i)+startPort] = pidPair{fromPid: fromPid, toPid: 0}
	}
	return ret
}

// a set of connections to a single pid from a destination off-box (pid is zero)
func fakeConnectionsFromOffBox(count int, toPid uint32, startPort uint16) mapPortToPids {
	ret := make(mapPortToPids, count)
	for i := 0; i < count; i++ {
		ret[uint16(i)+startPort] = pidPair{fromPid: 0, toPid: toPid}
	}
	return ret
}

// a set of connections from a range of pids and a range of ports between two pids
func fakeConnections(count int, fromPid, toPid uint32, startPort uint16) mapPortToPids {
	ret := make(mapPortToPids, count)
	for i := 0; i < count; i++ {
		ret[uint16(i)+startPort] = pidPair{fromPid: fromPid, toPid: toPid}
	}
	return ret
}

// union N existing sets
func concatMapPortToPids(maps ...mapPortToPids) mapPortToPids {
	ret := make(mapPortToPids)
	for _, m := range maps {
		for port, pair := range m {
			ret[port] = pair
		}
	}
	return ret
}

func TestConnectionThinning(t *testing.T) {
	for _, d := range []struct {
		name string
		m    mapPortToPids
		expC int
	}{
		{
			name: "0 ports",
			m:    mapPortToPids{},
			expC: 0,
		},
		{
			name: "5 connections off-box",
			m:    fakeConnectionsOffBox(5, 1000, 30000),
			expC: 5, // 5 is too few to thin down
		},
		{
			name: "50 connections off-box",
			m:    fakeConnectionsOffBox(50, 1000, 30000),
			expC: 5, // 50 connections should be thinned down
		},
		{
			name: "50 connections from off-box",
			m:    fakeConnectionsFromOffBox(50, 1000, 30000),
			expC: 5, // 50 connections should be thinned down
		},
		{
			name: "5 connections from pid 1000 to 2000",
			m:    fakeConnections(5, 1000, 2000, 30000),
			expC: 5, // 5 is too few to thin down
		},
		{
			name: "50 connections from pid 1000 to 2000",
			m:    fakeConnections(50, 1000, 2000, 30000),
			expC: 5, // 50 connections should be thinned down
		},
		{
			name: "connections both ways",
			m: concatMapPortToPids(
				fakeConnections(50, 1000, 2000, 30000),
				fakeConnections(50, 2000, 1000, 40000)),
			expC: 100, // no thinning because in both directions
		},
		{
			name: "connections to two different pids",
			m: concatMapPortToPids(
				fakeConnections(50, 1000, 2000, 30000),
				fakeConnections(50, 1000, 3000, 40000)),
			expC: 100, // no thinning because two different pids
		},
		{
			name: "connections from off-box and on-box",
			m:    concatMapPortToPids(fakeConnectionsFromOffBox(50, 1000, 30000), fakeConnections(50, 2000, 1000, 40000)),
			expC: 100, // no thinning because pid and no-pid
		},
	} {
		t.Run(d.name, func(t *testing.T) {
			filter, count := makeFilter(d.m)
			_ = filter
			if d.expC != count {
				t.Errorf("expected count %d, got %d", d.expC, count)
			}
		})
	}
}
