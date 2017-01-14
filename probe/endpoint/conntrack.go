package endpoint

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/scope/probe/endpoint/conntrack"
)

const (
	// From https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt
	eventsPath = "sys/net/netfilter/nf_conntrack_events"
)

// flowWalker is something that maintains flows, and provides an accessor
// method to walk them.
type flowWalker interface {
	walkFlows(f func(conntrack.Flow, bool))
	stop()
}

type nilFlowWalker struct{}

func (n nilFlowWalker) stop()                                  {}
func (n nilFlowWalker) walkFlows(f func(conntrack.Flow, bool)) {}

// conntrackWalker uses the conntrack command to track network connections and
// implement flowWalker.
type conntrackWalker struct {
	sync.Mutex
	activeFlows   map[uint32]conntrack.Flow // active flows in state != TIME_WAIT
	bufferedFlows []conntrack.Flow          // flows coming out of activeFlows spend 1 walk cycle here
	bufferSize    int
	args          []string
	quit          chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newConntrackFlowWalker(useConntrack bool, procRoot string, bufferSize int, args ...string) flowWalker {
	if !useConntrack {
		return nilFlowWalker{}
	} else if err := IsConntrackSupported(procRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
		return nilFlowWalker{}
	}
	result := &conntrackWalker{
		activeFlows: map[uint32]conntrack.Flow{},
		bufferSize:  bufferSize,
		args:        args,
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

	c.activeFlows = map[uint32]conntrack.Flow{}
}

func logPipe(prefix string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Error(prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error(prefix, err)
	}
}

func (c *conntrackWalker) run() {
	existingFlows, err := c.existingConnections()
	if err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
		return
	}
	for _, flow := range existingFlows {
		c.handleFlow(flow, true)
	}

	events, stop, err := conntrack.Follow(c.bufferSize)
	if err != nil {
		log.Errorf("conntract Follow error: %v", err)
		return
	}

	c.Lock()
	// We may have stopped in the mean time,
	// so check to see if the channel is open
	// under the lock.
	select {
	default:
	case <-c.quit:
		return
	}
	c.Unlock()

	defer log.Infof("conntrack exiting")

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
			c.handleFlow(f, false)
		}
	}
}

func (c *conntrackWalker) existingConnections() ([]conntrack.Flow, error) {
	flows, err := conntrack.Established(c.bufferSize)
	if err != nil {
		return []conntrack.Flow{}, err
	}
	return flows, nil
}

func (c *conntrackWalker) stop() {
	c.Lock()
	defer c.Unlock()
	close(c.quit)
}

func (c *conntrackWalker) handleFlow(f conntrack.Flow, forceAdd bool) {
	c.Lock()
	defer c.Unlock()

	// Ignore flows for which we never saw an update; they are likely
	// incomplete or wrong.  See #1462.
	switch {
	case forceAdd || f.MsgType == conntrack.NfctMsgUpdate:
		if f.State != conntrack.TCPStateTimeWait {
			c.activeFlows[f.ID] = f
		} else if _, ok := c.activeFlows[f.ID]; ok {
			delete(c.activeFlows, f.ID)
			c.bufferedFlows = append(c.bufferedFlows, f)
		}
	case f.MsgType == conntrack.NfctMsgDestroy:
		if active, ok := c.activeFlows[f.ID]; ok {
			delete(c.activeFlows, f.ID)
			c.bufferedFlows = append(c.bufferedFlows, active)
		}
	}
}

// walkFlows calls f with all active flows and flows that have come and gone
// since the last call to walkFlows
func (c *conntrackWalker) walkFlows(f func(conntrack.Flow, bool)) {
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
