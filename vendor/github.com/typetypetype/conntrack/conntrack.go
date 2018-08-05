package conntrack

import (
	"fmt"
	"log"
	"time"
)

// ConnTCP is a connection
type ConnTCP struct {
	Local      string // net.IP
	LocalPort  string // int
	Remote     string // net.IP
	RemotePort string // int
}

func (c ConnTCP) String() string {
	return fmt.Sprintf("%s:%s->%s:%s", c.Local, c.LocalPort, c.Remote, c.RemotePort)
}

// ConnTrack monitors the connections. It is build with Established() and
// Follow().
type ConnTrack struct {
	connReq chan chan []ConnTCP
	quit    chan struct{}
}

// New returns a ConnTrack.
func New() (*ConnTrack, error) {
	c := ConnTrack{
		connReq: make(chan chan []ConnTCP),
		quit:    make(chan struct{}),
	}
	go func() {
		for {
			err := c.track()
			select {
			case <-c.quit:
				return
			default:
			}
			if err != nil {
				log.Printf("conntrack: %s\n", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return &c, nil
}

// Close stops all monitoring and executables.
func (c ConnTrack) Close() {
	close(c.quit)
}

// Connections returns the list of all connections seen since last time you
// called it.
func (c *ConnTrack) Connections() []ConnTCP {
	r := make(chan []ConnTCP)
	c.connReq <- r
	return <-r
}

// track is the main loop
func (c *ConnTrack) track() error {
	// We use Follow() to keep track of conn state changes, but it doesn't give
	// us the initial state. If we first look at the established connections
	// and then start the follow process we might miss events.
	var flags uint32 = NF_NETLINK_CONNTRACK_NEW | NF_NETLINK_CONNTRACK_UPDATE |
		NF_NETLINK_CONNTRACK_DESTROY
	events, stop, err := Follow(flags)
	if err != nil {
		return err
	}

	established := map[ConnTCP]struct{}{}
	cs, err := Established()
	if err != nil {
		return err
	}
	for _, c := range cs {
		established[c] = struct{}{}
	}
	// we keep track of deleted so we can report them
	deleted := map[ConnTCP]struct{}{}

	local := localIPs()
	updateLocalIPs := time.Tick(time.Minute)

	for {
		select {

		case <-c.quit:
			stop()
			return nil

		case <-updateLocalIPs:
			local = localIPs()

		case e, ok := <-events:
			if !ok {
				return nil
			}
			switch {

			default:
				// not interested

			case e.TCPState == "ESTABLISHED":
				cn := e.ConnTCP(local)
				if cn == nil {
					// log.Printf("not a local connection: %+v\n", e)
					continue
				}
				established[*cn] = struct{}{}

			case e.MsgType == NfctMsgDestroy, e.TCPState == "TIME_WAIT", e.TCPState == "CLOSE":
				cn := e.ConnTCP(local)
				if cn == nil {
					// log.Printf("not a local connection: %+v\n", e)
					continue
				}
				if _, ok := established[*cn]; !ok {
					continue
				}
				delete(established, *cn)
				deleted[*cn] = struct{}{}

			}

		case r := <-c.connReq:
			cs := make([]ConnTCP, 0, len(established)+len(deleted))
			for c := range established {
				cs = append(cs, c)
			}
			for c := range deleted {
				cs = append(cs, c)
			}
			r <- cs
			deleted = map[ConnTCP]struct{}{}

		}
	}
}
