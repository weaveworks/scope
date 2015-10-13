Go module to list all TCP connections, with an option to try to find the owning PID and processname.

Works by reading /proc directly on Linux, and by executing `netstat` and `lsof -i` on Darwin.

Works for IPv4 and IPv6 TCP connections. Only established connections are listed; ports where something is only listening or TIME_WAITs are skipped.

If you want to find all processes you'll need to run this as root.

Status:
-------

Tested on Linux and Darwin (10.9).

Install:
--------

`go install`

Usage:
------

Only list the connections:

`conns, err := procspy.Connections(false)`

List the connections and try to find the owning process:

`conns, err := procspy.Connections(true)`

(See ./example\_test.go)

``` go

package main

import (
	"fmt"

	"github.com/alicebob/procspy"
)

func main() {
	cs, err := procspy.Connections(false)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connections:\n")
	for _, c := range cs {
		fmt.Printf(" - %v\n", c)
	}
}
```
