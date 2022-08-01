//go:build linux
// +build linux

package endpoint

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/typetypetype/conntrack"
	"os"
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
	activeFlows   map[uint32]conntrack.Conn // active flows in state != TIME_WAIT
	bufferedFlows []conntrack.Conn          // flows coming out of activeFlows spend 1 walk cycle here
	bufferSize    int
	natOnly       bool
	quit          chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newConntrackFlowWalker(useConntrack bool, procRoot string, bufferSize int, natOnly bool) flowWalker {
	if !useConntrack {
		return nilFlowWalker{}
	} else if err := IsConntrackSupported(procRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
		return nilFlowWalker{}
	}
	result := &conntrackWalker{
		activeFlows: map[uint32]conntrack.Conn{},
		bufferSize:  bufferSize,
		natOnly:     natOnly,
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

	for _, f := range c.activeFlows {
		c.bufferedFlows = append(c.bufferedFlows, f)
	}

	c.activeFlows = map[uint32]conntrack.Conn{}
}

func (c *conntrackWalker) relevant(f conntrack.Conn) bool {
	// For now, we're only interested in tcp connections - there is too much udp
	// traffic going on (every container talking to dns, for example) to
	// render nicely. TODO: revisit this.
	if f.Orig.Proto != tcpProto {
		return false
	}
	return !(c.natOnly && (f.Status&conntrack.IPS_NAT_MASK) == 0)
}

func (c *conntrackWalker) run() {
	var hostname, err = os.Hostname()
	var reverseType = map[int]string{1: "New", 1 << 1: "Update", 1 << 2: "Destroy"}
	if err != nil {
		log.Errorf("error retrieveing hostname, uses default: %v", err)
		hostname = "invalid"
	}
	existingFlows, err := conntrack.ConnectionsSize(c.bufferSize)
	if err != nil {
		log.Errorf("conntrack Connections error: %v", err)
		return
	}
	c.Lock()
	for _, flow := range existingFlows {
		if c.relevant(flow) && flow.TCPState != tcpClose && flow.TCPState != timeWait {
			c.activeFlows[flow.CtId] = flow
		}
	}
	c.Unlock()

	events, stop, err := conntrack.FollowSize(c.bufferSize, conntrack.NF_NETLINK_CONNTRACK_UPDATE|conntrack.NF_NETLINK_CONNTRACK_DESTROY)
	if err != nil {
		log.Errorf("conntrack Follow error: %v", err)
		return
	}

	periodicRestart := time.After(6 * time.Hour)
	// Handle conntrack events from netlink socket
	for {
		select {
		case <-periodicRestart:
			log.Debugf("conntrack periodic restart")
			return
		case <-c.quit:
			log.Infof("conntrack quit signal - exiting")
			stop()
			return
		case f, ok := <-events:
			if !ok {
				log.Errorf("conntrack events read failed - exiting")
				return
			}
			if f.Err != nil {
				metrics.IncrCounter([]string{"conntrack", "errors"}, 1)
				log.Errorf("conntrack event error: %v", f.Err)
				stop()
				return
			}
			if c.relevant(f) {
				// ========= PRINT ==========
				//s, _ := json.Marshal(f)
				//var out bytes.Buffer
				//json.Indent(&out, s, "", "\t")
				//log.Debugf("conntrack get connection: %v", out.String())
				log.Infof("[CONN] [conntrack] {%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v}", hostname, reverseType[int(f.MsgType)], f.TCPState, f.Orig.Src, f.Orig.SrcPort, f.Orig.Dst, f.Orig.DstPort, f.Reply.Src, f.Reply.SrcPort, f.Reply.Dst, f.Reply.DstPort, f.CtId)
				// ========= PRINT ==========
				c.handleFlow(f)
			}
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
		if f.TCPState != timeWait {
			c.activeFlows[f.CtId] = f
		} else if _, ok := c.activeFlows[f.CtId]; ok {
			delete(c.activeFlows, f.CtId)
			c.bufferedFlows = append(c.bufferedFlows, f)
		}
	case f.MsgType == conntrack.NfctMsgDestroy:
		if active, ok := c.activeFlows[f.CtId]; ok {
			delete(c.activeFlows, f.CtId)
			c.bufferedFlows = append(c.bufferedFlows, active)
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
	for _, flow := range c.bufferedFlows {
		f(flow, false)
	}
	c.bufferedFlows = c.bufferedFlows[:0]
}
