// +build linux

package endpoint

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/typetypetype/conntrack/ovs"
)

type TunnelAttrs struct {
	TunIpSrc uint32
	TunIpDst uint32

	MaskSrc uint32
	MaskDst uint32

	IpDst   uint32
	PortDst uint16
}

// ovsFlowWalker uses conntrack (via netlink) to track network connections and
// implement flowWalker.
type ovsFlowWalker struct {
	sync.Mutex
	activeFlows   map[uint64]TunnelAttrs
	bufferedFlows []TunnelAttrs // flows coming out of activeFlows spend 1 walk cycle here
	bufferSize    int
	natOnly       bool
	quit          chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newOvsFlowWalker(bufferSize int) *ovsFlowWalker {
	log.Info("creating ovs flow walker")
	result := &ovsFlowWalker{
		activeFlows: map[uint64]TunnelAttrs{},
		bufferSize:  bufferSize,
		quit:        make(chan struct{}),
	}
	go result.loop()
	return result
}

func (c *ovsFlowWalker) loop() {
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

func (c *ovsFlowWalker) clearFlows() {
	c.Lock()
	defer c.Unlock()

	for _, f := range c.activeFlows {
		c.bufferedFlows = append(c.bufferedFlows, f)
	}

	c.activeFlows = map[uint64]TunnelAttrs{}
}

func (c *ovsFlowWalker) relevant(fi *ovs.OvsFlowInfo) bool {

	key, ok := fi.Keys[ovs.OvsAttrIpv4]
	if !ok {
		return false
	}

	if _, ok := key.(ovs.OvsAttrIpv4FlowKey); !ok {
		return false
	}

	if _, ok := fi.Masks[ovs.OvsAttrIpv4]; !ok {
		return false
	}

	if _, ok := fi.Actions[ovs.OVS_ACTION_ATTR_SET]; !ok {
		return false
	}

	return true
}

func (c *ovsFlowWalker) run() {
	//existingFlows, err := conntrack.ConnectionsSize(c.bufferSize)
	//if err != nil {
	//	log.Errorf("conntrack Connections error: %v", err)
	//	return
	//}
	//c.Lock()
	//for _, flow := range existingFlows {
	//	if c.relevant(flow) && flow.TCPState != tcpClose && flow.TCPState != timeWait {
	//		c.activeFlows[flow.CtId] = flow
	//	}
	//}
	//c.Unlock()
	//

	events, stop, err := ovs.FollowOvsFlows(c.bufferSize, 0)
	if err != nil {
		log.Errorf("ovs Follow error: %v", err)
		return
	}

	defer log.Infof("ovs exiting")

	// Handle conntrack events from netlink socket
	for {
		select {
		case <-c.quit:
			stop()
			return
		case f, ok := <-events:
			if !ok {
				return
			}
			if c.relevant(f) {
				c.handleFlow(f)
			}
		}
	}
}

func (c *ovsFlowWalker) stop() {
	c.Lock()
	defer c.Unlock()
	close(c.quit)
}

func (c *ovsFlowWalker) handleFlow(fi *ovs.OvsFlowInfo) {
	c.Lock()
	defer c.Unlock()

	log.Info("handling")
	key, ok := fi.Keys[ovs.OvsAttrIpv4]
	if !ok {
		return
	}

	ipv4fk, ok := key.(ovs.OvsAttrIpv4FlowKey)
	if !ok {
		return
	}

	maskIpv4, ok := fi.Masks[ovs.OvsAttrIpv4]
	if !ok {
		return
	}

	maskIpv4Fk, ok := maskIpv4.(ovs.OvsAttrIpv4FlowKey)
	if !ok {
		return
	}

	setTunnel, ok := fi.Actions[ovs.OVS_ACTION_ATTR_SET]
	if !ok {
		return
	}

	setTunnelFk, ok := setTunnel.(ovs.OvsSetTunnelAction)
	if !ok {
		return
	}

	if _, exists := c.activeFlows[setTunnelFk.TunnelId]; !exists {
		c.activeFlows[setTunnelFk.TunnelId] = TunnelAttrs{TunIpSrc: ipv4fk.Src,
			TunIpDst: ipv4fk.Dst,
			MaskSrc:  maskIpv4Fk.Src,
			MaskDst:  maskIpv4Fk.Dst, IpDst: setTunnelFk.Ipv4Dst, PortDst: setTunnelFk.TpDst}
	}
}

// walkFlows calls f with all active flows and flows that have come and gone
// since the last call to walkFlows
func (c *ovsFlowWalker) walkFlows(f func(TunnelAttrs, bool)) {
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
