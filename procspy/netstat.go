package procspy

// netstat reading.

import (
	"net"
	"strconv"
	"strings"
)

// parseDarwinNetstat parses netstat output. (Linux has ip:port, darwin
// ip.port. The 'Proto' column value also differs.)
func parseDarwinNetstat(out string) []Connection {
	//
	//  Active Internet connections
	//  Proto Recv-Q Send-Q  Local Address          Foreign Address        (state)
	//  tcp4       0      0  10.0.1.6.58287         1.2.3.4.443      		ESTABLISHED
	//
	res := []Connection{}
	for i, line := range strings.Split(out, "\n") {
		if i == 0 || i == 1 {
			// Skip header
			continue
		}

		// Fields are:
		fields := strings.Fields(line)
		if len(fields) != 6 {
			continue
		}

		if fields[5] != "ESTABLISHED" {
			continue
		}

		t := Connection{
			Transport: "tcp",
		}

		// Format is <ip>.<port>
		locals := strings.Split(fields[3], ".")
		if len(locals) < 2 {
			continue
		}

		var (
			localAddress = strings.Join(locals[:len(locals)-1], ".")
			localPort    = locals[len(locals)-1]
		)

		t.LocalAddress = net.ParseIP(localAddress)

		p, err := strconv.Atoi(localPort)
		if err != nil {
			return nil
		}

		t.LocalPort = uint16(p)

		remotes := strings.Split(fields[4], ".")
		if len(remotes) < 2 {
			continue
		}

		var (
			remoteAddress = strings.Join(remotes[:len(remotes)-1], ".")
			remotePort    = remotes[len(remotes)-1]
		)

		t.RemoteAddress = net.ParseIP(remoteAddress)

		p, err = strconv.Atoi(remotePort)
		if err != nil {
			return nil
		}

		t.RemotePort = uint16(p)

		res = append(res, t)
	}

	return res
}
