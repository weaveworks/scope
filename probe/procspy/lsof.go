package procspy

// lsof-executing implementation.

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	lsofFields = "cn" // parseLSOF() depends on the order
)

// parseLsof parses lsof out with `-F cn` argument.
//
// Format description: the first letter is the type of record, records are
// newline seperated, the record starting with 'p' (pid) is a new processid.
// There can be multiple connections for the same 'p' record in which case the
// 'p' is not repeated.
//
// For example, this is one process with two listens and one connection:
//
//   p13100
//   cmpd
//   n[::1]:6600
//   n127.0.0.1:6600
//   n[::1]:6600->[::1]:50992
//
func parseLSOF(out string) (map[string]Proc, error) {
	var (
		res = map[string]Proc{} // Local addr -> Proc
		cp  = Proc{}
	)
	for _, line := range strings.Split(out, "\n") {
		if len(line) <= 1 {
			continue
		}

		var (
			field = line[0]
			value = line[1:]
		)
		switch field {
		case 'p':
			pid, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid 'p' field in lsof output: %#v", value)
			}
			cp.PID = uint(pid)

		case 'n':
			// 'n' is the last field, with '-F cn'
			// format examples:
			// "192.168.2.111:44013->54.229.241.196:80"
			// "[2003:45:2b57:8900:1869:2947:f942:aba7]:55711->[2a00:1450:4008:c01::11]:443"
			// "*:111" <- a listen
			addresses := strings.SplitN(value, "->", 2)
			if len(addresses) != 2 {
				// That's a listen entry.
				continue
			}
			res[addresses[0]] = Proc{
				Name: cp.Name,
				PID:  cp.PID,
			}

		case 'c':
			cp.Name = value

		default:
			return nil, fmt.Errorf("unexpected lsof field: %c in %#v", field, value)
		}
	}

	return res, nil
}
