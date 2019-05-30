// +build linux

package endpoint

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/typetypetype/conntrack"
)

const (
	// From https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt
	eventsPath = "sys/net/netfilter/nf_conntrack_events"
	timeWait   = "TIME_WAIT"
	tcpClose   = "CLOSE"
	tcpProto   = 6
)

// flowWalker is something that maintains flows, and provides an accessor
// method to walk them.
type flowWalker interface {
	walkFlows(f func(conntrack.Conn, bool))
	stop()
}

type nilFlowWalker struct{}

func (n nilFlowWalker) stop()                                  {}
func (n nilFlowWalker) walkFlows(f func(conntrack.Conn, bool)) {}

// conntrackWalker uses conntrack (via netlink) to track network connections and
// implement flowWalker.
type conntrackWalker struct {
	sync.Mutex
	activeFlows map[uint32]conntrack.Conn // active flows in state != TIME_WAIT
	//bufferedFlows []conntrack.Conn          // flows coming out of activeFlows spend 1 walk cycle here
	bufferSize int
	quit       chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newConntrackFlowWalker(useConntrack bool, procRoot string, bufferSize int) flowWalker {
	if !useConntrack {
		return nilFlowWalker{}
	} else if err := IsConntrackSupported(procRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
		return nilFlowWalker{}
	}
	result := &conntrackWalker{
		activeFlows: map[uint32]conntrack.Conn{},
		bufferSize:  bufferSize,
		quit:        make(chan struct{}),
	}
	go result.loop()
	return result
}

// IsConntrackSupported returns true if conntrack is suppported by the kernel
var IsConntrackSupported = func(procRoot string) error {
	// Make sure events are enabled, the conntrack CLI doesn't verify it
	f := filepath.Join(procRoot, eventsPath)
	contents, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if string(contents) == "0" {
		return fmt.Errorf("conntrack events (%s) are disabled", f)
	}
	return nil
}

func (c *conntrackWalker) loop() {
	// conntrack can sometimes fail with ENOBUFS, when there is a particularly
	// high connection rate.  In these cases just retry in a loop, so we can
	// survive the spike.  For sustained loads this degrades nicely, as we
	// read the table before starting to handle events - basically degrading to
	// polling.
	for {
		c.run()
		c.clearFlows()

		select {
		case <-time.After(time.Second):
		case <-c.quit:
			return
		}
	}
}

func (c *conntrackWalker) clearFlows() {
	c.Lock()
	defer c.Unlock()

	c.activeFlows = map[uint32]conntrack.Conn{}
}

func (c *conntrackWalker) needsDstNat(f conntrack.Conn) bool {
	return f.Status&conntrack.IPS_DST_NAT == conntrack.IPS_DST_NAT
}

func (c *conntrackWalker) isConnectionInitiator(f conntrack.Conn) bool {
	return f.TCPState == "SYN_SENT"
}

func (c *conntrackWalker) isConnectionEstablished(f conntrack.Conn) bool {
	return f.TCPState == "ESTABLISHED"
}

func (c *conntrackWalker) run() {
	existingFlows, err := conntrack.ConnectionsSize(c.bufferSize)
	if err != nil {
		log.Errorf("conntrack Connections error: %v", err)
		return
	}
	c.Lock()
	for _, flow := range existingFlows {
		if c.needsDstNat(flow) || c.isConnectionInitiator(flow) {
			c.activeFlows[flow.CtId] = flow
		}
	}
	c.Unlock()

	events, stop, err := conntrack.FollowSize(c.bufferSize, conntrack.NF_NETLINK_CONNTRACK_UPDATE|conntrack.NF_NETLINK_CONNTRACK_DESTROY)
	if err != nil {
		log.Errorf("conntrack Follow error: %v", err)
		return
	}

	defer log.Infof("conntrack exiting")

	// Handle conntrack events from netlink socket
	for {
		select {
		case <-c.quit:
			stop()
			return
		case f, ok := <-events:
			if !ok {
				log.Info("error loop")
				return
			}
			c.handleFlow(f)
		}
	}
}

func (c *conntrackWalker) stop() {
	c.Lock()
	defer c.Unlock()
	close(c.quit)
}

func (c *conntrackWalker) handleFlow(f conntrack.Conn) {
	c.Lock()
	defer c.Unlock()

	// Ignore flows for which we never saw an update; they are likely
	// incomplete or wrong.  See #1462.
	switch {
	case f.MsgType == conntrack.NfctMsgUpdate:
		if c.isConnectionInitiator(f) || c.isConnectionEstablished(f) {
			c.activeFlows[f.CtId] = f
		}
	case f.MsgType == conntrack.NfctMsgDestroy:
		if _, ok := c.activeFlows[f.CtId]; ok {
			delete(c.activeFlows, f.CtId)
		}
	}
}

// walkFlows calls f with all active flows and flows that have come and gone
// since the last call to walkFlows
func (c *conntrackWalker) walkFlows(f func(conntrack.Conn, bool)) {
	c.Lock()
	defer c.Unlock()
	for _, flow := range c.activeFlows {
		f(flow, true)
	}
}
