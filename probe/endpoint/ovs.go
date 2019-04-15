// +build linux

package endpoint

import (
	"sync"
	"time"

	"fmt"
	"net"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"github.com/typetypetype/conntrack/ovs"
)

type TunnelAttrs struct {
	TunnelID uint64
	TunIpSrc uint32
	TunIpDst uint32

	MaskSrc uint32
	MaskDst uint32

	IpDst   uint32
	PortDst uint16
}

func (ta TunnelAttrs) SrcIP() string {
	return fmt.Sprintf("%s/%s", ipv4ToString(ta.TunIpSrc), ipv4ToString(ta.MaskSrc))
}

func (ta TunnelAttrs) DstIP() string {
	return fmt.Sprintf("%s/%s", ipv4ToString(ta.TunIpDst), ipv4ToString(ta.MaskDst))
}

func (ta TunnelAttrs) DstFlow() string {
	return ipv4ToString(ta.IpDst)
}

func ipv4ToString(ip uint32) string {
	ipAsByte := (*[4]byte)(unsafe.Pointer(&ip))[:]
	return net.IP(ipAsByte).To4().String()
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

	events, stop, err := ovs.FollowOvsFlows()
	if err != nil {
		log.Errorf("ovs follow error: %v", err)
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

	actions, ok := setTunnel.([]ovs.OvsAction)
	if !ok {
		return
	}

	var setTunnelFk *ovs.OvsSetTunnelAction

	for _, action := range actions {
		if setTunFk, ok := action.(ovs.OvsSetTunnelAction); ok {
			setTunnelFk = &setTunFk
			break
		}

	}

	if setTunnelFk == nil {
		return
	}

	if ipv4fk.Src == 0 || ipv4fk.Dst == 0 {
		return
	}

	c.activeFlows[setTunnelFk.TunnelId] = TunnelAttrs{
		TunnelID: setTunnelFk.TunnelId,
		TunIpSrc: ipv4fk.Src,
		TunIpDst: ipv4fk.Dst,
		MaskSrc:  maskIpv4Fk.Src,
		MaskDst:  maskIpv4Fk.Dst,
		IpDst:    setTunnelFk.Ipv4Dst,
		PortDst:  setTunnelFk.TpDst}

}

// walkFlows calls f with all active flows and flows that have come and gone
// since the last call to walkFlows
func (c *ovsFlowWalker) walkFlows(f func(TunnelAttrs)) {
	c.Lock()
	defer c.Unlock()

	for _, flow := range c.activeFlows {
		f(flow)
	}
}
